package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"gopkg.in/h2non/gock.v1"
)

const (
	// GameID test game id
	GameID = "15799e31-cd98-4e87-9d49-40ceb4eb439e"
)

// GockStatus create a mock http call to the engine for status.
func GockStatus(response string) *gock.Request {
	mockRequest := *gock.New(EngineURL).Get(fmt.Sprintf("/games/%s", GameID))
	mockRequest.Reply(200).BodyString(response)
	return &mockRequest
}

// GockFrame create a mock http call to the engine for a frame.
func GockFrame(response string) {
	GockStatus(response).MatchParam("offset", "29")
}

// Gock15Frames create a mock http call to the engine for 15 consequitive frames.
func Gock15Frames(response string, lastResponse string) {
	GockStatus(response).MatchParam("offset", "0").MatchParam("limit", "1")
	GockStatus(response).MatchParam("offset", "0").MatchParam("limit", "5")
	GockStatus(response).MatchParam("offset", "5").MatchParam("limit", "5")
	GockStatus(lastResponse).MatchParam("offset", "10").MatchParam("limit", "5")
}

func initialize() (*mux.Router, *httptest.ResponseRecorder) {
	router := mux.NewRouter()
	rr := httptest.NewRecorder()
	SetupRoutes(router)
	return router, rr
}
func serveURL(params string) *httptest.ResponseRecorder {
	router, rr := initialize()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/games/%s/frames/29?%s", GameID, params), nil)
	router.ServeHTTP(rr, req)
	return rr
}
