package http

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/media"
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

func handleVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	version := os.Getenv("APP_VERSION")
	if len(version) == 0 {
		version = "unknown"
	}
	fmt.Fprint(w, version)
}

var reAvatarParams = regexp.MustCompile(`^/(?:[a-z-]{1,32}:[a-z-0-9#]{1,32}/)*(?P<width>[0-9]{2,4})x(?P<height>[0-9]{2,4}).(?P<ext>[a-z]{3,4})$`)
var reAvatarCustomizations = regexp.MustCompile(`(?P<key>[a-z-]{1,32}):(?P<value>[a-z-0-9#]{1,32})`)

func handleAvatar(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	errBadRequest := fmt.Errorf("bad request")
	avatarSettings := render.AvatarSettings{}

	// Extract width, height, and filetype
	reParamsResult := reAvatarParams.FindStringSubmatch(p.ByName("params"))
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
	if pExt != "svg" {
		handleBadRequest(w, r, errBadRequest)
		return
	}

	// Extract customization params
	reCustomizationResults := reAvatarCustomizations.FindAllStringSubmatch(p.ByName("params"), -1)
	for _, match := range reCustomizationResults {
		cKey, cValue := match[1], match[2]
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

	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, avatarSVG)
}

func handleASCIIFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	engineURL := r.URL.Query().Get("engine_url")
	frameID, err := strconv.Atoi(p.ByName("frame"))
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

func handleGIFFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	width, height, err := getGameDimensions(p)
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}
	frameID, err := strconv.Atoi(p.ByName("frame"))
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

func getGameDimensions(p httprouter.Params) (int, int, error) {
	sizeParam := p.ByName("size")
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

func handleGIFGame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	width, height, err := getGameDimensions(p)
	if err != nil {
		handleBadRequest(w, r, err)
		return
	}

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

func handleAlive(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "alive")
}

func handleReady(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

// parseSizeParam parses a path parameter that is expected to be in the form "<WIDTH>x<HEIGHT>"
// if size is empty, 0,0 is returned
func parseSizeParam(param string) (int, int, error) {
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
