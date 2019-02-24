package http

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Server is commented
type Server struct{}

// NewServer is commented
func NewServer() *Server {
	return &Server{}
}

// Run is commented
func (s *Server) Run() {
	router := httprouter.New()

	router.GET("/", indexHandler)

	router.GET("/games/:game/frames/:frame/ascii", handleASCIIFrame)
	router.GET("/games/:game/frames/:frame/gif", handleGIFFrame)

	log.Fatal(http.ListenAndServe(":8000", router))
}
