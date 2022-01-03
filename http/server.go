package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	router     *httprouter.Router
	httpServer *http.Server
}

func NewServer() *Server {
	router := httprouter.New()

	// System routes
	router.GET("/", handleVersion)
	router.GET("/version", handleVersion)
	router.GET("/healthz/alive", handleAlive)
	router.GET("/healthz/ready", handleReady)

	// Export routes
	router.GET("/avatars/*params", withCaching(handleAvatar))
	router.GET("/games/:game/gif", withCaching(handleGIFGame))
	router.GET("/games/:game/frames/:frame/ascii", withCaching(handleASCIIFrame))
	router.GET("/games/:game/frames/:frame/gif", withCaching(handleGIFFrame))

	router.PanicHandler = panicHandler

	return &Server{
		router: router,
		httpServer: &http.Server{
			Handler: router,
		},
	}
}

func withCaching(wrappedHandler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		maxAgeSeconds := 60 * 60 * 24 // 24 Hours
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age:%d", maxAgeSeconds))
		wrappedHandler(w, r, p)
	}
}

func (s *Server) Run() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = ":8000"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	s.httpServer.Addr = port
	logger := log.WithField("listen", port)

	sigHandler := make(chan os.Signal, 1)
	signal.Notify(sigHandler, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	connectionsClosed := make(chan struct{})
	go func() {
		receivedSignal := <-sigHandler
		logger.WithField("signal", receivedSignal.String()).
			Warn("Exporter shutdown signal received")
		if err := s.Shutdown(time.Second * 20); err != nil {
			logger.WithError(err).
				Fatal("Failed to shut down exporter gracefully")
		}
		close(connectionsClosed)
	}()

	logger.Info("Exporter serving")
	err := s.WaitForExit()
	if err != nil && err != http.ErrServerClosed {
		logger.WithError(err).
			Fatal("Exporter failed to start")
	}
	<-connectionsClosed
	logger.Info("Exporter shutdown successfully")
}

// WaitForExit starts up the server and blocks until the server shuts down.
func (s *Server) WaitForExit() error { return s.httpServer.ListenAndServe() }

// Shutdown gracefully shuts down the server, waiting until a timeout for active requests to complete.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
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
