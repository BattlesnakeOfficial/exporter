package handlers

import (
	"bytes"
	"fmt"
	"image"
	gifOriginal "image/gif"
	"io"

	"github.com/fogleman/gg"

	engine "github.com/battlesnakeio/exporter/engine"
	gif "github.com/battlesnakeio/exporter/gif"
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
	outGif := &gif.GIF{}
	c := make(chan gif.PalettAndDelay, 50)
	outGif.Image = c
	gameFrames, _ := GetGameFramesWithLength(gameID, 0, batchSize)
	outGif.SampleImage = createGif(&gameFrames.Frames[0], gameStatus).(*image.Paletted)
	go getGifFrames(c, gameFrames, outGif, gameStatus, gameID, batchSize)
	outGif.LoopCount = -1
	gif.EncodeAll(w, outGif)
	return nil
}

func getGifFrames(c chan gif.PalettAndDelay, firstSet *engine.ListGameFramesResponse, outGif *gif.GIF, gameStatus *engine.StatusResponse, gameID string, batchSize int) {
	currentOffset := 0
	for {
		var gameFrames *engine.ListGameFramesResponse
		if currentOffset == 0 {
			gameFrames = firstSet
		} else {
			gameFrames, _ = GetGameFramesWithLength(gameID, currentOffset, batchSize)
		}
		frameCount := 0

		for _, frame := range gameFrames.Frames {
			frameCount++
			imageGif := createGif(&frame, gameStatus)
			outGif.Image <- gif.PalettAndDelay{
				Palett: imageGif.(*image.Paletted),
				Delay:  10,
				I:      frameCount + currentOffset,
			}
		}
		if frameCount < batchSize {
			close(c)
			break
		}
		currentOffset += batchSize
		fmt.Printf("game: %s: frames: %d of %d\n", gameID, currentOffset, gameStatus.LastFrame.Turn)
	}
}

func createGif(frame *engine.GameFrame, gameStatus *engine.StatusResponse) image.Image {
	var framePng bytes.Buffer
	ConvertFrameToPNG(&framePng, frame, gameStatus)
	imagePng, _, _ := image.Decode(&framePng)
	var frameGif bytes.Buffer
	gifOriginal.Encode(&frameGif, imagePng, nil)
	imageGif, _, _ := image.Decode(&frameGif)
	return imageGif
}
