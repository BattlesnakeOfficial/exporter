package main

import (
	"github.com/TV4/logrus-stackdriver-formatter"
	"github.com/battlesnakeio/exporter/http"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(stackdriver.NewFormatter())

	httpServer := http.NewServer()
	httpServer.Run()
}
