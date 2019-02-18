package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	gif_original "image/gif"
	"io"
	"strings"

	"github.com/fogleman/gg"
	"gopkg.in/go-playground/colors.v1"

	engine "github.com/battlesnakeio/exporter/engine"
	gif "github.com/battlesnakeio/exporter/gif"
)

// pP = previous point, p = current point, nP next point.
func corner(pP engine.Point, p engine.Point, nP engine.Point) string {
	if (pP == engine.Point{}) || (nP == engine.Point{}) {
		return "none"
	}

	switch fmt.Sprintf("%d,%d:%d,%d", pP.X-p.X, pP.Y-p.Y, nP.X-p.X, nP.Y-p.Y) {
	case "0,-1:1,0", "1,0:0,-1":
		return "bottom-left"
	case "-1,0:0,-1", "0,-1:-1,0":
		return "bottom-right"
	case "-1,0:0,1", "0,1:-1,0":
		return "top-right"
	case "0,1:1,0", "1,0:0,1":
		return "top-left"
	default:
		return "none"
	}
}

// ConvertFrameToPNG takes a frame and makes a png
func ConvertFrameToPNG(w io.Writer, gameFrame *engine.GameFrame, gameStatus *engine.StatusResponse) {
	width, height := getDimensions(gameStatus)
	square := int32(40)
	border := int(square / 8)
	dc := gg.NewContext(width*int(square)+border, height*int(square)+border)
	dc.DrawRectangle(0, 0, float64(width*int(square)+border), float64(height*int(square)+border))
	dc.SetHexColor("#000000")
	dc.Fill()

	// create board
	for x := int32(0); x < int32(width); x++ {
		for y := int32(0); y < int32(height); y++ {
			drawSquare(dc, x, y, square)
		}
	}

	// draw dead snakes
	for _, snake := range gameFrame.Snakes {
		if snake.Death.Cause != "" {
			drawSnake(dc, &snake, square)
		}
	}
	// draw alive snakes
	for _, snake := range gameFrame.Snakes {
		if snake.Death.Cause == "" {
			drawSnake(dc, &snake, square)
		}
	}
	// draw food
	for _, point := range gameFrame.Food {
		drawFood(dc, &point, square)
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
		fmt.Printf("giffing game: %s: frames: %d of %d\n", gameID, currentOffset, gameStatus.LastFrame.Turn)
	}
}

func createGif(frame *engine.GameFrame, gameStatus *engine.StatusResponse) image.Image {
	var framePng bytes.Buffer
	ConvertFrameToPNG(&framePng, frame, gameStatus)
	imagePng, _, _ := image.Decode(&framePng)
	var frameGif bytes.Buffer
	gif_original.Encode(&frameGif, imagePng, nil)
	imageGif, _, _ := image.Decode(&frameGif)
	return imageGif
}

func drawSquare(dc *gg.Context, x int32, y int32, square int32) {
	border := float64(square / 8)
	dc.DrawRectangle(float64(x*square)+border, float64(y*square)+border, float64(square)-border, float64(square)-border)
	dc.SetHexColor("#111111")
	dc.Fill()
}

func drawFood(dc *gg.Context, point *engine.Point, square int32) {
	borderHalf := float64(square) / float64(20)
	dc.DrawCircle(float64(point.X*square)+float64(square)/2.0+borderHalf, float64(point.Y*square+square/2.0)+borderHalf, float64(square)/float64(4))
	dc.SetHexColor("#FFA500")
	dc.Fill()
}

type changeable interface {
	Set(x, y int, c color.Color)
}

func placeHead(dc *gg.Context, point *engine.Point, square int32, snakeColor string) {
	border := float64(square) / float64(8)
	h, err := GetSnakeHeadImage("beluga")
	if err != nil {
		panic(err)
	}
	bounds := h.Bounds()
	if cimg, ok := h.(changeable); ok {
		for x := 0; x < bounds.Max.X; x++ {
			for y := 0; y < bounds.Max.Y; y++ {
				// colors.ParseRGBA(h.At(x, y))
				r, g, b, _ := h.At(x, y).RGBA()
				ratio := (float64(r) + float64(g) + float64(b)) / float64(3*65535)
				if !strings.HasPrefix(snakeColor, "#") {
					snakeColor = "#" + snakeColor
				}
				snakeColorHex, err := colors.ParseHEX(snakeColor)
				if err != nil {
					panic(err)
				}
				newR := float64(snakeColorHex.ToRGBA().R) * (1 - ratio)
				newG := float64(snakeColorHex.ToRGBA().G) * (1 - ratio)
				newB := float64(snakeColorHex.ToRGBA().B) * (1 - ratio)
				fmt.Println(snakeColorHex)
				fmt.Printf("%d, %d, %d\n", uint8(newR), uint8(newG), uint8(newB))
				// fmt.Println(avg)
				// color.RG
				cimg.Set(x, y, color.RGBA{uint8(newR), uint8(newG), uint8(newB), 255})
			}
		}
		dc.DrawImage(h, int(point.X*square+int32(border)), int(point.Y*square+int32(border)))
	}
}

func drawSnake(dc *gg.Context, snake *engine.Snake, square int32) {
	halfSquare := float64(square) / float64(2)

	borderHalf := float64(square) / float64(20)
	border := float64(square) / float64(8)
	previousPoint := engine.Point{}
	nextPoint := engine.Point{}
	for i, point := range snake.Body {
		if i < len(snake.Body)-1 {
			nextPoint = snake.Body[i+1]
		} else {
			nextPoint = engine.Point{}
		}
		corner := corner(previousPoint, point, nextPoint)
		color := snake.Color
		if snake.Death.Cause != "" {
			color = "#333333"
		}
		if i == 0 {
			placeHead(dc, &point, square, color)
		} else if corner == "none" {
			dc.DrawRectangle(float64(point.X*square)+border, float64(point.Y*square)+border, float64(square)-border, float64(square)-border)
		} else {
			dc.DrawRoundedRectangle(float64(point.X*square)+border, float64(point.Y*square)+border, float64(square)-border, float64(square)-border, halfSquare)
			if strings.HasPrefix(corner, "bottom") {
				dc.DrawRectangle(float64(point.X*square)+border, float64(point.Y*square)+border, float64(square)-border, float64(square)-border-halfSquare+borderHalf)
				if strings.HasSuffix(corner, "left") {
					dc.DrawRectangle(float64(point.X*square)+border+halfSquare-borderHalf, float64(point.Y*square)+border, float64(square)-border-halfSquare+borderHalf, float64(square)-border)
				}
				if strings.HasSuffix(corner, "right") {
					dc.DrawRectangle(float64(point.X*square)+border, float64(point.Y*square)+border, float64(square)-border-halfSquare+borderHalf, float64(square)-border)
				}
			}
			if strings.HasPrefix(corner, "top") {
				dc.DrawRectangle(float64(point.X*square)+border, float64(point.Y*square)+border+halfSquare-borderHalf, float64(square)-border, float64(square)-border-halfSquare+borderHalf)
				if strings.HasSuffix(corner, "left") {
					dc.DrawRectangle(float64(point.X*square)+border+halfSquare-borderHalf, float64(point.Y*square)+border, float64(square)-border-halfSquare+borderHalf, float64(square)-border)
				}
				if strings.HasSuffix(corner, "right") {
					dc.DrawRectangle(float64(point.X*square)+border, float64(point.Y*square)+border, float64(square)-border-halfSquare+borderHalf, float64(square)-border)
				}
			}
		}
		dc.SetHexColor(color)
		dc.Fill()
		previousPoint = point
	}
}
