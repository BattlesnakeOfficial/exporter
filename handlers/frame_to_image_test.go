package handlers

import (
	"bufio"
	"os"
	"testing"

	openapi "github.com/battlesnakeio/exporter/model"
)

func TestFrameToPNG(t *testing.T) {
	frame := createFrame()
	frame.Food = []openapi.EnginePoint{
		openapi.EnginePoint{X: 0, Y: 2},
	}
	status := createGameStatus(2, 2)
	f, _ := os.Create("/tmp/temp.png")
	defer f.Close()
	w := bufio.NewWriter(f)
	ConvertFrameToPNG(w, frame, status)
	w.Flush()
}
