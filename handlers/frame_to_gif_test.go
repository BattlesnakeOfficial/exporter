package handlers

import (
	"bufio"
	"os"
	"testing"

	openapi "github.com/battlesnakeio/exporter/model"
)

func TestFrameTopng(t *testing.T) {
	frame := createFrame()
	frame.Food = []openapi.EnginePoint{
		openapi.EnginePoint{X: 0, Y: 2},
	}
	status := createGameStatus(10, 10)
	f, _ := os.Create("/tmp/temp.png")
	defer f.Close()
	w := bufio.NewWriter(f)
	ConvertFrameToPng(w, frame, status)
	w.Flush()
}
