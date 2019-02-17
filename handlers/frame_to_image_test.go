package handlers

import (
	"bufio"
	"os"
	"testing"

	engine "github.com/battlesnakeio/exporter/engine"
)

func TestFrameToPNG(t *testing.T) {
	frame := createFrame()
	frame.Food = []engine.Point{
		engine.Point{X: 0, Y: 2},
	}
	status := createGameStatus(2, 2)
	f, _ := os.Create("/tmp/temp.png")
	defer f.Close()
	w := bufio.NewWriter(f)
	ConvertFrameToPNG(w, frame, status)
	w.Flush()
}
