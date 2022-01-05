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

func handleGIFFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
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
	if err = render.GameFrameToGIF(w, game, gameFrame); err != nil {
		handleError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func handleGIFGame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
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
	err = render.GameFramesToAnimatedGIF(w, game, gameFrames, frameDelay, loopDelay)
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
