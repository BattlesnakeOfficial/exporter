package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	engine "github.com/battlesnakeio/exporter/engine"
	gock "gopkg.in/h2non/gock.v1"
)

func TestFrameToPNG(t *testing.T) {
	frame := createFrame()
	frame.Food = []engine.Point{
		engine.Point{X: 0, Y: 2},
	}
	status := createGameStatus(2, 2)
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	ConvertFrameToPNG(w, frame, status)
	w.Flush()
	require.True(t, b.Len() > 0)
}

func TestGetGif(t *testing.T) {
	defer gock.Off()
	frameList, _ := json.Marshal(createFrameList())
	frameList5, _ := json.Marshal(createFrameList5())
	Gock15Frames(string(frameList5), string(frameList))
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	ConvertGameToGif(w, createGameStatus(3, 3), GameID, 5)
	w.Flush()
	require.True(t, b.Len() > 0)
}
