package http

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/alitto/pond"
	log "github.com/sirupsen/logrus"
	"goji.io/v3"
	"goji.io/v3/pat"
)

// How many requests can be queued up for any GIF render before we start rejecting with HTTP 429
// Half of the current max in-flight requests
const DEFAULT_RENDER_BACKLOG = 40

type Server struct {
	router     *goji.Mux
	httpServer *http.Server
}

func NewServer() *Server {
	log.WithField("size", runtime.NumCPU()).Info("Starting GIF render pool")
	renderPool := pond.New(runtime.NumCPU(), DEFAULT_RENDER_BACKLOG)

	mux := goji.NewMux()
	mux.Use(Recovery) // captures panics

	// // System routes
	mux.HandleFunc(pat.Get("/"), handleVersion)
	mux.HandleFunc(pat.Get("/version"), handleVersion)
	mux.HandleFunc(pat.Get("/healthz/alive"), handleAlive)
	mux.HandleFunc(pat.Get("/healthz/ready"), handleReady)
	mux.HandleFunc(pat.Get("/robots.txt"), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	// Export routes
	mux.HandleFunc(pat.Get("/avatars/*"), withCaching(handleAvatar))

	mux.HandleFunc(pat.Get("/customizations/:type/:name.:ext"), withCaching(handleCustomization))

	mux.HandleFunc(pat.Get("/games/:game/:size.gif"), withConcurrencyLimit(renderPool, withCaching(handleGIFGameDimensions)))
	mux.HandleFunc(pat.Get("/games/:game/gif"), withConcurrencyLimit(renderPool, withCaching(handleGIFGame)))
	mux.HandleFunc(pat.Get("/games/:game/frames/:frame/:size.gif"), withConcurrencyLimit(renderPool, withCaching(handleGIFFrameDimensions)))
	mux.HandleFunc(pat.Get("/games/:game/frames/:frame/gif"), withConcurrencyLimit(renderPool, withCaching(handleGIFFrame)))

	mux.HandleFunc(pat.Get("/games/:game/frames/:frame.txt"), withCaching(handleASCIIFrame))
	mux.HandleFunc(pat.Get("/games/:game/frames/:frame/ascii"), withCaching(handleASCIIFrame))

	return &Server{
		router: mux,
		httpServer: &http.Server{
			Handler: mux,
		},
	}
}

func withConcurrencyLimit(pool *pond.WorkerPool, wrappedHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})

		// Try to submit a job to the pool asynchronously, and if it fails, reject the request
		submitted := pool.TrySubmit(func() {
			defer close(done)
			wrappedHandler(w, r)
		})

		if !submitted {
			close(done)
			log.WithField("url", r.URL.String()).Print("No worker available from pool, rejecting request")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		} else {
			// Block until the job that was submitted to the pool is done
			<-done
		}
	}
}

func withCaching(wrappedHandler http.HandlerFunc) http.HandlerFunc {
	appVersion, ok := os.LookupEnv("APP_VERSION")
	if !ok {
		appVersion = "0.0.0"
	}

	cacheControlMaxAgeSeconds, ok := os.LookupEnv("CACHE_CONTROL_MAX_AGE_SECONDS")
	if !ok {
		cacheControlMaxAgeSeconds = "86400" // 24 Hours
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%s", cacheControlMaxAgeSeconds))

		// Set etag based on URL path and App Version
		etagString := fmt.Sprintf("%s/%s", appVersion, r.URL.Path)
		w.Header().Set("Etag", fmt.Sprintf(`"%x"`, md5.Sum([]byte(etagString))))

		wrappedHandler(w, r)
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

func Recovery(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				source := "unknown"
				if _, filename, line, ok := runtime.Caller(2); ok {
					source = fmt.Sprintf("%s:%d", filename, line)
				}
				log.WithField("err", err).
					WithField("url", r.URL.String()).
					WithField("source", source).
					Error("unhandled panic")

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
