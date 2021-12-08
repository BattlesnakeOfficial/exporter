package http

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Router *httprouter.Router
}

func NewServer() *Server {
	router := httprouter.New()

	router.GET("/", versionHandler)
	router.GET("/version", versionHandler)

	router.GET("/games/:game/gif", handleGIFGame)

	router.GET("/games/:game/frames/:frame/ascii", handleASCIIFrame)
	router.GET("/games/:game/frames/:frame/gif", handleGIFFrame)

	router.GET("/healthz/alive", handleAlive)
	router.GET("/healthz/ready", handleReady)

	router.PanicHandler = panicHandler

	return &Server{router}
}

func (s *Server) Run() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = ":8000"
	}
	log.WithField("port", port).Info("http server listening")
	if err := http.ListenAndServe(port, s.Router); err != nil {
		log.WithError(err).WithField("port", port).Error("error while trying to listen on port")
	}
}

func panicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	source := "unknown"
	if _, filename, line, ok := runtime.Caller(3); ok {
		source = fmt.Sprintf("%s:%d", filename, line)
	}
	log.WithField("err", err).
		WithField("url", r.URL.String()).
		WithField("source", source).
		Error("unhandled panic")

	w.WriteHeader(http.StatusInternalServerError)
}
