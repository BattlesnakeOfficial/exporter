package main

import (
	"github.com/BattlesnakeOfficial/exporter/http"
)

func main() {
	httpServer := http.NewServer()
	httpServer.Run()
}
