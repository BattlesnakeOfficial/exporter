package handlers

import (
	"testing"
	"time"

	engine "github.com/battlesnakeio/exporter/engine"
	"github.com/stretchr/testify/require"
)

func TestBox(t *testing.T) {
	_, imageType, err := GetSnakeHeadImage("beluga")
	if err != nil {
		panic(err)
	}
	require.Equal(t, "png", imageType)
}

func TestHeadTailCache(t *testing.T) {
	start1 := time.Now()
	GetOrCreateRotatedSnakeImage(HeadSegment, &engine.Snake{}, "FFFFFF", 90, 10)
	finish1 := time.Now()
	duration1 := finish1.Nanosecond() - start1.Nanosecond()

	start2 := time.Now()
	GetOrCreateRotatedSnakeImage(HeadSegment, &engine.Snake{}, "FFFFFF", 90, 10)
	finish2 := time.Now()
	duration2 := finish2.Nanosecond() - start2.Nanosecond()

	require.True(t, duration2 < duration1)

}

func TestGetSafeColorEdges(t *testing.T) {
	require.Equal(t, "#ffffff", getSafeHexColour("", "FFFFFF").ToHEX().String())
	require.Equal(t, "#ffffff", getSafeHexColour("aoeu", "FFFFFF").ToHEX().String())
	defer func() {
		if err := recover(); err != nil {
			require.Equal(t, "default colour couldn't be parsed: #aoeu", err)
		} else {
			require.Fail(t, "should have thrown an error for bad default")
		}
	}()
	require.Equal(t, "#ffffff", getSafeHexColour("", "aoeu").ToHEX().String())

}
