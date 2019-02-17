package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"io"

	"github.com/fogleman/gg"

	engine "github.com/battlesnakeio/exporter/engine"
)

// ConvertFrameToPNG takes a frame and makes a png
func ConvertFrameToPNG(w io.Writer, gameFrame *engine.GameFrame, gameStatus *engine.StatusResponse) {
	width, height := getDimensions(gameStatus)
	square := int32(20)
	filled := make(map[string]bool)
	dc := gg.NewContext(width*int(square)+2, height*int(square)+2)
	dc.DrawRectangle(0, 0, float64(width*int(square)+2), float64(height*int(square)+2))
	dc.SetHexColor("#000000")
	dc.Fill()
	for _, snake := range gameFrame.Snakes {
		for i, point := range snake.Body {
			filled[fmt.Sprintf("%d,%d", point.X, point.Y)] = true
			transparancy := "AA"
			if i == 0 {
				transparancy = "FF"
			}
			color := snake.Color + transparancy
			if snake.Death.Cause != "" {
				color = "#555555" + transparancy
			}
			dc.DrawRectangle(float64(point.X*square)+2, float64(point.Y*square)+2, float64(square)-2, float64(square)-2)
			dc.SetHexColor(color)
			dc.Fill()
		}
	}

	for _, point := range gameFrame.Food {
		filled[fmt.Sprintf("%d,%d", point.X, point.Y)] = true
		dc.DrawRectangle(float64(point.X*square)+2, float64(point.Y*square)+2, float64(square)-2, float64(square)-2)
		dc.SetHexColor("#111111")
		dc.Fill()
		dc.DrawCircle(float64(point.X*square)+float64(square)/2.0+1, float64(point.Y*square+square/2.0)+1, float64(square)/float64(4))
		dc.SetHexColor("#FFA500")
		dc.Fill()
	}
	for x := int32(0); x < int32(width); x++ {
		for y := int32(0); y < int32(height); y++ {
			if !filled[fmt.Sprintf("%d,%d", x, y)] {
				dc.DrawRectangle(float64(x*square)+2, float64(y*square)+2, float64(square)-2, float64(square)-2)
				dc.SetHexColor("#111111")
				dc.Fill()
			}
		}
	}

	dc.EncodePNG(w)
}

// ConvertGameToGif reads all frames from the engine and outputs an animated gif.
func ConvertGameToGif(w io.Writer, gameStatus *engine.StatusResponse, gameID string, batchSize int) error {
	currentOffset := 0
	outGif := &gif.GIF{}
	for {
		gameFrames, err := GetGameFramesWithLength(gameID, currentOffset, batchSize)
		if err != nil {
			return err
		}
		frameCount := 0

		for _, frame := range gameFrames.Frames {
			var framePng bytes.Buffer
			frameCount++
			ConvertFrameToPNG(&framePng, &frame, gameStatus)
			imagePng, _, _ := image.Decode(&framePng)
			var frameGif bytes.Buffer
			gif.Encode(&frameGif, imagePng, nil)
			imageGif, _, _ := image.Decode(&frameGif)
			outGif.Image = append(outGif.Image, imageGif.(*image.Paletted))
			outGif.Delay = append(outGif.Delay, 10)
		}
		if frameCount < batchSize {
			break
		}
		currentOffset += batchSize
		fmt.Printf("game: %s: frames: %d of %d\n", gameID, currentOffset, gameStatus.LastFrame.Turn)
	}
	outGif.LoopCount = -1
	gif.EncodeAll(w, outGif)
	return nil
}
