package handlers

import (
	"testing"

	openapi "github.com/battlesnakeio/exporter/model"
)

func TestFrameToGif(t *testing.T) {
	frame := createFrame()
	frame.Food = []openapi.EnginePoint{
		openapi.EnginePoint{X: 0, Y: 2},
	}
	status := createStatus()
	gif := FrameToGif(frame, status)

}
