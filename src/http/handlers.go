package http

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"github.com/battlesnakeio/exporter/engine"
	"github.com/battlesnakeio/exporter/render"
)

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "https://battlesnake.io", 302)
}

func handleASCIIFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	frameID, err := strconv.Atoi(p.ByName("frame"))
	if err != nil {
		panic(err)
	}

	game, err := engine.GetGame(gameID)
	if err != nil {
		panic(err)
	}

	gameFrame, err := engine.GetGameFrame(game.ID, frameID)
	if err != nil {
		panic(err)
	}

	render.GameFrameToASCII(w, game, gameFrame)
}

func handleGIFFrame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")
	frameID, err := strconv.Atoi(p.ByName("frame"))
	if err != nil {
		panic(err)
	}

	game, err := engine.GetGame(gameID)
	if err != nil {
		panic(err)
	}

	gameFrame, err := engine.GetGameFrame(game.ID, frameID)
	if err != nil {
		panic(err)
	}

	render.GameFrameToGIF(w, game, gameFrame)
	w.Header().Set("Content-Type", "image/gif")
}

func handleGIFGame(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gameID := p.ByName("game")

	game, err := engine.GetGame(gameID)
	if err != nil {
		panic(err)
	}

	gameFrames, err := engine.GetGameFrames(game.ID)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "image/gif")
	render.GameFramesToAnimatedGIF(w, game, gameFrames)
}
