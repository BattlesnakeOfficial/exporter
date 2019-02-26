package main

import (
	"github.com/battlesnakeio/exporter/http"
)

func main() {
	httpServer := http.NewServer()
	httpServer.Run()
}
