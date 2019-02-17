package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gobuffalo/packr/v2"

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

func TestBox(t *testing.T) {
	box := packr.New("My Box", "../snake-images")
	beluga, err := box.FindString("head/beluga.svg")
	if err != nil {
		panic(err)
	}
	require.Equal(t, "<svg id=\"root\" xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 100 100\">\n  <path d=\"M0 100h56L32 88l-5-14 73 2-10-48L50 0H0zm23-61a9 9 0 1 1-10 10 9 9 0 0 1 10-10z\"/>\n</svg>\n", beluga)
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
