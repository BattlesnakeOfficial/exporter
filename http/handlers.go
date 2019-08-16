package http

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/battlesnakeio/exporter/engine"
	"github.com/battlesnakeio/exporter/render"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "https://battlesnake.com", 302)
}

func handleASCIIFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	engineURL := r.URL.Query().Get("engine_url")
	frameID, err := strconv.Atoi(p.ByName("frame"))
	if err != nil {
		handleError(w, r, err)
		return
	}

	game, err := engine.GetGame(gameID, engineURL)
	if err != nil {
		handleError(w, r, err)
		return
	}

	gameFrame, err := engine.GetGameFrame(game.ID, engineURL, frameID)
	if err != nil {
		handleError(w, r, err)
		return
	}

	if err = render.GameFrameToASCII(w, game, gameFrame); err != nil {
		handleError(w, r, err)
		return
	}
}

func handleGIFFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	frameID, err := strconv.Atoi(p.ByName("frame"))
	if err != nil {
		handleError(w, r, err)
		return
	}

	engineURL := r.URL.Query().Get("engine_url")
	game, err := engine.GetGame(gameID, engineURL)
	if err != nil {
		handleError(w, r, err)
		return
	}

	gameFrame, err := engine.GetGameFrame(game.ID, engineURL, frameID)
	if err != nil {
		handleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	if err = render.GameFrameToGIF(w, game, gameFrame); err != nil {
		handleError(w, r, err)
		return
	}
}

func handleGIFGame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	engineURL := r.URL.Query().Get("engine_url")
	game, err := engine.GetGame(gameID, engineURL)
	if err != nil {
		handleError(w, r, err)
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
		handleError(w, r, err)
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
		handleError(w, r, err)
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

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.WithError(err).
		WithFields(log.Fields{
			"httpRequest": map[string]interface{}{
				"method":    r.Method,
				"url":       r.URL.String(),
				"userAgent": r.Header.Get("User-Agent"),
				"referrer":  r.Header.Get("Referer"),
			},
		}).Error("unable to process request")
	w.WriteHeader(http.StatusInternalServerError)
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
