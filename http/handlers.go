package http

import (
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"goji.io/v3/pat"

	log "github.com/sirupsen/logrus"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/media"
	"github.com/BattlesnakeOfficial/exporter/parse"
	"github.com/BattlesnakeOfficial/exporter/render"
)

// maxGIFResolution is the maximum resolution of GIF that we want to support.
// This is an important limit to set, because the GIF rendering takes a lot more
// IO, CPU and memory resources for larger resolutions.
// This resolution was chosen as a safe upper-limit after which the rendering starts
// to get really slow and the GIF sizes start to get too big.
const maxGIFResolution = 504 * 504

// allowedPixelsPerSquare is a list of resolutions that the API will allow.
var allowedPixelsPerSquare = []int{10, 20, 30, 40}

var errBadRequest = fmt.Errorf("bad request")
var errBadColor = fmt.Errorf("color parameter should have the format #FFFFFF")

func handleVersion(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("APP_VERSION")
	if len(version) == 0 {
		version = "unknown"
	}
	fmt.Fprint(w, version)
}

var reAvatarParams = regexp.MustCompile(`^/(?:[a-z-]{1,32}:[A-Za-z-0-9#]{0,32}/)*(?P<width>[0-9]{2,4})x(?P<height>[0-9]{2,4}).(?P<ext>[a-z]{3,4})$`)
var reAvatarCustomizations = regexp.MustCompile(`(?P<key>[a-z-]{1,32}):(?P<value>[A-Za-z-0-9#]{0,32})`)

func handleAvatar(w http.ResponseWriter, r *http.Request) {
	subPath := strings.TrimPrefix(r.URL.Path, "/avatars")
	avatarSettings := render.AvatarSettings{}

	// Extract width, height, and filetype
	reParamsResult := reAvatarParams.FindStringSubmatch(subPath)
	if len(reParamsResult) != 4 {
		handleBadRequest(w, r, errBadRequest)
		return
	}

	pWidth, err := strconv.Atoi(reParamsResult[1])
	if err != nil {
		handleBadRequest(w, r, errBadRequest)
		return
	}
	avatarSettings.Width = pWidth

	pHeight, err := strconv.Atoi(reParamsResult[2])
	if err != nil {
		handleBadRequest(w, r, errBadRequest)
		return
	}
	avatarSettings.Height = pHeight

	pExt := reParamsResult[3]
	if pExt != "svg" && pExt != "png" {
		handleBadRequest(w, r, errBadRequest)
		return
	}

	// Extract customization params
	reCustomizationResults := reAvatarCustomizations.FindAllStringSubmatch(subPath, -1)
	for _, match := range reCustomizationResults {
		cKey, cValue := match[1], match[2]
		if cValue == "" {
			// ignore empty values
			continue
		}
		switch cKey {
		case "head":
			avatarSettings.HeadSVG, err = media.GetHeadSVG(cValue)
			if err != nil {
				if errors.Is(err, media.ErrNotFound) {
					handleBadRequest(w, r, errBadRequest)
				} else {
					handleError(w, r, err, http.StatusInternalServerError)
				}
				return
			}
		case "tail":
			avatarSettings.TailSVG, err = media.GetTailSVG(cValue)
			if err != nil {
				if errors.Is(err, media.ErrNotFound) {
					handleBadRequest(w, r, errBadRequest)
				} else {
					handleError(w, r, err, http.StatusInternalServerError)
				}
				return
			}
		case "color":
			if len(cValue) != 7 || string(cValue[0]) != "#" {
				handleBadRequest(w, r, errBadRequest)
				return
			}
			avatarSettings.Color = cValue
		default:
			handleBadRequest(w, r, errBadRequest)
			return
		}
	}

	// Render SVG
	avatarSVG, err := render.AvatarSVG(avatarSettings)
	if err != nil {
		if errors.Is(err, render.ErrInvalidAvatarSettings) {
			handleBadRequest(w, r, errBadRequest)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	if pExt == "png" {
		image, err := media.ConvertSVGStringToPNG(avatarSVG, avatarSettings.Width, avatarSettings.Height)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, image); err != nil {
			log.WithError(err).Error("unable to write PNG to response stream")
		}
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, avatarSVG)
}

var reCustomizationParam = regexp.MustCompile(`^[A-Za-z-0-9#]{1,32}$`)

func handleCustomization(w http.ResponseWriter, r *http.Request) {
	customizationType := pat.Param(r, "type")
	customizationName := pat.Param(r, "name")
	ext := pat.Param(r, "ext")

	if ext != "svg" {
		handleBadRequest(w, r, errBadRequest)
		return
	}

	if customizationType != "head" && customizationType != "tail" {
		handleBadRequest(w, r, errBadRequest)
		return
	}

	if !reCustomizationParam.MatchString(customizationName) {
		handleBadRequest(w, r, errBadRequest)
		return
	}

	var customizationColor color.Color = color.Black
	colorParam := r.URL.Query().Get("color")
	if colorParam != "" {
		if len(colorParam) != 7 || string(colorParam[0]) != "#" {
			handleBadRequest(w, r, errBadColor)
			return
		}

		customizationColor = parse.HexColor(colorParam)
	}

	flippedParam := r.URL.Query().Get("flipped") != ""

	var svg string
	var err error
	var shouldFlip bool
	switch customizationType {
	case "head":
		svg, err = media.GetHeadSVG(customizationName)
		shouldFlip = flippedParam
	case "tail":
		svg, err = media.GetTailSVG(customizationName)
		shouldFlip = !flippedParam
	}

	if err != nil {
		if err == media.ErrNotFound {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	svg = media.CustomizeSnakeSVG(svg, customizationColor, shouldFlip)

	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, svg)
}

func handleASCIIFrame(w http.ResponseWriter, r *http.Request) {
	gameID := pat.Param(r, "game")
	engineURL := r.URL.Query().Get("engine_url")
	frameID, err := strconv.Atoi(pat.Param(r, "frame"))
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}

	game, err := engine.GetGame(gameID, engineURL)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	gameFrame, err := engine.GetGameFrame(game.ID, engineURL, frameID)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	if err = render.GameFrameToASCII(w, game, gameFrame); err != nil {
		handleError(w, r, err, http.StatusInternalServerError)
		return
	}
}

// validateDimensionsForBoard checks whether the width/height is valid for the given board width/height.
func validateDimensionsForBoard(game *engine.Game, w, h int) error {

	// handle the legacy case where w/h are 0
	if w == 0 || h == 0 {
		return nil
	}

	b := int(render.BoardBorder * 2)
	options := make([]string, 0, len(allowedPixelsPerSquare)) // used to build a helpful error message
	for _, r := range allowedPixelsPerSquare {
		// should match one of the allowed resolutions
		aw := (game.Width*r + b)
		ah := (game.Height*r + b)
		options = append(options, fmt.Sprintf("%dx%d", aw, ah))
		if aw == w && ah == h {
			return nil
		}
	}

	return fmt.Errorf("Dimensions %dx%d invalid - valid options are: %s", w, h, strings.Join(options, ", "))
}

func handleGIFFrameDimensions(w http.ResponseWriter, r *http.Request) {
	width, height, err := getGameDimensions(r)
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}
	handleGIFFrameCommon(w, r, width, height)
}

func handleGIFFrame(w http.ResponseWriter, r *http.Request) {
	handleGIFFrameCommon(w, r, 0, 0)
}

func handleGIFFrameCommon(w http.ResponseWriter, r *http.Request, width, height int) {
	gameID := pat.Param(r, "game")
	frameID, err := strconv.Atoi(pat.Param(r, "frame"))
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}

	log.Infof("exporting frame %s:%d", gameID, frameID)

	engineURL := r.URL.Query().Get("engine_url")
	game, err := engine.GetGame(gameID, engineURL)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}
	err = validateDimensionsForBoard(game, width, height)
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}

	gameFrame, err := engine.GetGameFrame(game.ID, engineURL, frameID)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	if err = render.GameFrameToGIF(w, game, gameFrame, width, height); err != nil {
		handleError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func getGameDimensions(r *http.Request) (int, int, error) {
	sizeParam := pat.Param(r, "size")
	width, height, err := parseSizeParam(sizeParam)
	if err != nil {
		return 0, 0, err
	}

	// ensure width/height are within allowable limits
	err = validateGIFSize(width, height)
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

func handleGIFGameDimensions(w http.ResponseWriter, r *http.Request) {
	width, height, err := getGameDimensions(r)
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}
	handleCommonGIFGame(w, r, width, height)
}

func handleGIFGame(w http.ResponseWriter, r *http.Request) {
	handleCommonGIFGame(w, r, 0, 0)
}

func handleCommonGIFGame(w http.ResponseWriter, r *http.Request, width, height int) {

	gameID := pat.Param(r, "game")
	engineURL := r.URL.Query().Get("engine_url")

	log.WithField("game", gameID).WithField("engine_url", engineURL).Info("exporting game")

	game, err := engine.GetGame(gameID, engineURL)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}
	err = validateDimensionsForBoard(game, width, height)
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}

	offset := 0
	limit := math.MaxInt32
	frames := strings.Split(r.URL.Query().Get("frames"), "-")
	if len(frames) == 2 {
		valOne, errOne := strconv.Atoi(frames[0])
		valTwo, errTwo := strconv.Atoi(frames[1])
		if errOne != nil || errTwo != nil {
			handleBadRequest(w, r, fmt.Errorf("invalid frames parameter: %s", r.URL.Query().Get("frames")))
		}

		offset = valOne
		limit = valTwo - valOne + 1
	}

	gameFrames, err := engine.GetGameFrames(game.ID, engineURL, offset, limit)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			handleError(w, r, err, http.StatusNotFound)
		} else {
			handleError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	frameDelay, err := strconv.Atoi(r.URL.Query().Get("frameDelay"))
	if err != nil {
		frameDelay = render.GIFFrameDelay
	}

	loopDelay, err := strconv.Atoi(r.URL.Query().Get("loopDelay"))
	if err != nil {
		loopDelay = render.GIFLoopDelay
	}

	w.Header().Set("Content-Type", "image/gif")
	err = render.GameFramesToAnimatedGIF(w, game, gameFrames, frameDelay, loopDelay, width, height)
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func handleBadRequest(w http.ResponseWriter, r *http.Request, e error) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte(e.Error()))
	if err != nil {
		log.WithError(err).Error("unable to write to response stream")
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.WithError(err).
		WithFields(log.Fields{
			"httpRequest": map[string]interface{}{
				"method":    r.Method,
				"url":       r.URL.String(),
				"userAgent": r.Header.Get("User-Agent"),
				"referrer":  r.Header.Get("Referer"),
			},
		}).Error("unable to process request")

	w.WriteHeader(statusCode)

	if _, err := w.Write([]byte(err.Error())); err != nil {
		log.WithError(err).Error("unable to write to response stream")
	}
}

func handleAlive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "alive")
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ready")
}

var sizeRegex = regexp.MustCompile(`^(\d+)x(\d+)$`)

// validateGIFSize checks that the dimension of the GIF is within a safe range that we can allow.
func validateGIFSize(w, h int) error {

	// ensure the max resolution is not exceeded
	res := w * h
	if res > maxGIFResolution {
		return fmt.Errorf(`Too many pixels! Dimensions %dx%d having resolution %d exceeds maximum allowable resolution of %d.`, w, h, res, maxGIFResolution)
	}

	// ensure the minimum dimensions are met
	if w < 0 {
		return fmt.Errorf(`Invalid width %d: cannot be < 0.`, w)
	}

	// ensure the minimum dimensions are met
	if h < 0 {
		return fmt.Errorf(`Invalid height %d: cannot be < 0`, h)
	}

	return nil
}

// parseSizeParam parses a path parameter that is expected to be in the form "<WIDTH>x<HEIGHT>".
// If the size param is empty, 0,0 is returned.
func parseSizeParam(param string) (int, int, error) {
	// check for legacy case where size params are not included
	if param == "" {
		return 0, 0, nil
	}

	m := sizeRegex.FindStringSubmatch(param)
	if len(m) != 3 {
		return 0, 0, fmt.Errorf(`Invalid dimensions: "%s" not of the format <WIDTH>x<HEIGHT>.`, param)
	}

	w, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, err
	}
	h, err := strconv.Atoi(m[2])
	if err != nil {
		return 0, 0, err
	}

	return w, h, nil
}
