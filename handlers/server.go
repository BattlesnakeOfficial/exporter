package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

const (
	raw           string = "raw"
	board         string = "board"
	boardAnimated string = "board-animated"
	move          string = "move"
	pngOutputType string = "png"
	gifImage      string = "gif"
)

// OutputTypes All the output types
var OutputTypes = []string{
	raw, board, boardAnimated, move, pngOutputType, gifImage,
}

// SetupRoutes defines all the routs that this server will handle.
func SetupRoutes(router *mux.Router) {

	router.HandleFunc("/games/{id}/frames/{frame:[0-9]+}", getFrame).
		Methods("GET").
		Queries("output", "{output:move}").
		Queries("youId", "{youId}")
	router.HandleFunc("/games/{id}/frames/{frame:[0-9]*}", getPNG).
		Methods("GET").
		Queries("output", "{output:png}")
	router.HandleFunc("/games/{id}/frames/{frame:[0-9]*}", getFrame).
		Methods("GET").
		Queries("output", "{output:board|board-animated|raw}")
	router.HandleFunc("/games/{id}", getGIF).
		Methods("GET").
		Queries("output", "{output:gif}")
	router.NotFoundHandler = http.HandlerFunc(readMe)
}

// create a png of the game
func getGIF(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if paramsNotOk(w, params) {
		return
	}
	frames := strings.Split(r.FormValue("frames"), "-")
	offset := 0
	frameRange := -1
	if len(frames) == 2 {
		var err error
		var endFrame int
		offset, err = strconv.Atoi(frames[0])
		if err != nil {
			logrus.WithError(err).Errorf("unable to convert offset: %s", frames[0])
			offset = 0
		}
		endFrame, err = strconv.Atoi(frames[1])
		if err != nil {
			logrus.WithError(err).Errorf("unable to convert ending frame: %s", frames[1])
		} else {
			frameRange = endFrame - offset + 1
		}
	}
	batchSize, err := strconv.Atoi(r.FormValue("batchSize"))
	if err != nil {
		batchSize = 100
	}
	gameFrames, err := GetGameFrames(params["id"], 0)
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
	}
	if len(gameFrames.Frames) == 0 {
		response(w, 404, "No frames")
		return
	}
	gameStatus, err := GetGameStatus(params["id"])
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	delay, err := strconv.Atoi(r.FormValue("frameDelay"))
	if err != nil || delay == 0 {
		delay = 8
	}
	loopDelay, err := strconv.Atoi(r.FormValue("loopDelay"))
	if err != nil || loopDelay == 0 {
		loopDelay = 500
	}
	if frameRange <= 10 && frameRange > 0 {
		delay = 16
		loopDelay = 100
	}
	err = ConvertGameToGif(w, gameStatus, params["id"], batchSize, offset, frameRange, delay, loopDelay)

	if err != nil {
		response(w, 500, "Could not export to gif: "+err.Error())
	}
}

// create a png of the game
func getPNG(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if paramsNotOk(w, params) {
		return
	}
	offset, _ := strconv.Atoi(params["frame"])
	gameFrames, err := GetGameFrames(params["id"], offset)
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
		return
	}
	if len(gameFrames.Frames) == 0 {
		response(w, 404, "No frames")
		return
	}
	gameStatus, err := GetGameStatus(params["id"])
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	ConvertFrameToPNG(w, &gameFrames.Frames[0], gameStatus)
}

// gets a frame from the engine, writes it out in a supported format.
func getFrame(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if paramsNotOk(w, params) {
		return
	}

	offset, _ := strconv.Atoi(params["frame"])
	gameFrames, err := GetGameFrames(params["id"], offset)
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
		return
	}
	if len(gameFrames.Frames) == 0 {
		response(w, 404, "No frames")
		return
	}
	if params["output"] == raw {
		w.Header().Set("Content-Type", "application/json")
		obj, _ := json.Marshal(gameFrames.Frames[0])
		w.Write(obj)
		return
	}

	gameStatus, err := GetGameStatus(params["id"])
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
		return
	}
	if params["output"] == move {
		move, err := ConvertFrameToMove(&gameFrames.Frames[0], gameStatus, params["youId"])
		if err != nil {
			response(w, 500, "Problem creating move: "+err.Error())
			return
		}
		json, _ := json.Marshal(move)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
	if strings.HasPrefix(params["output"], board) {
		grid := ConvertFrameToGrid(&gameFrames.Frames[0], gameStatus)
		turn := int(gameFrames.Frames[0].Turn)
		board := ConvertGridToString(grid)
		if params["output"] == boardAnimated {
			board = fmt.Sprintf("<html><head></head><body><pre>%s</pre></body>"+
				"<script>window.location.assign(window.location.href.replace(/%d\\?/g, '%d?'));</script></html>", board, turn, turn+1)
		}
		w.Write([]byte(board))
	}
}

func paramsNotOk(w http.ResponseWriter, params map[string]string) bool {
	for _, output := range OutputTypes {
		if params["output"] == output {
			return false
		}
	}
	response(w, 404, fmt.Sprintf("Unsupported output type: %s support types are %s|%s|%s|%s", params["output"], raw, board, boardAnimated, move))
	return true
}

func response(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}

func readMe(w http.ResponseWriter, r *http.Request) {
	response(w, 404, `<html>
	<head>
    <script src="https://code.jquery.com/jquery-3.3.1.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/showdown/1.9.0/showdown.min.js"></script>
	</head>
	<body>
		<div></div>
  </body>
  <script> 
    $.get("https://raw.githubusercontent.com/battlesnakeio/exporter/master/README.md", function( text ) {
      $('div').html(new showdown.Converter().makeHtml(text))
    }, 'text')
  </script>
</html>`)
}
