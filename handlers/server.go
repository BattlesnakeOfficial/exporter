package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	raw           string = "raw"
	board         string = "board"
	boardAnimated string = "board-animated"
	move          string = "move"
)

//  main function
func main() {
	router := mux.NewRouter()
	SetupRoutes(router)
	log.Fatal(http.ListenAndServe(":8000", router))
}

// SetupRoutes defines all the routs that this server will handle.
func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/games/{id}/frames/{frame}", getFrame).
		Methods("GET").
		Queries("output", "{output:move}").
		Queries("youId", "{youId}")
	router.HandleFunc("/games/{id}/frames/{frame}", getFrame).
		Methods("GET").
		Queries("output", "{output:board|board-animated|raw}")
	router.NotFoundHandler = http.HandlerFunc(readMe)
}

// gets a frame from the engine, writes it out in a supported format.
func getFrame(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if paramsNotOk(w, params) {
		return
	}

	gameFrames, err := GetGameFrames(params["id"], params["frame"])
	if err != nil {
		response(w, 500, "Problem getting game frames: "+err.Error())
		return
	}
	if len(gameFrames.Frames) == 0 {
		response(w, 404, "No frames")
		return
	}
	obj, _ := json.Marshal(gameFrames.Frames[0])
	if params["output"] == raw {
		w.Header().Set("Content-Type", "application/json")
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
	if params["output"] == raw ||
		params["output"] == board ||
		params["output"] == boardAnimated ||
		params["output"] == move {
		return false
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
