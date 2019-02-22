package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	giforiginal "image/gif"
	"image/png"
	"io"
	"strings"

	engine "github.com/battlesnakeio/exporter/engine"
	"github.com/battlesnakeio/exporter/gif"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
)

const (
	squareColor = "#f1f1f1"
)

// pP = previous point, p = current point, nP next point.
func corner(pP *engine.Point, p *engine.Point, nP *engine.Point) string {
	if (pP == nil) || (nP == nil) {
		return "none"
	}
	coords := fmt.Sprintf("%d,%d:%d,%d", pP.X-p.X, pP.Y-p.Y, nP.X-p.X, nP.Y-p.Y)
	switch coords {
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

// pP = previous point, p = current point, nP next point.
func direction(p *engine.Point, nP *engine.Point) string {
	if nP == nil {
		return "up"
	}
	coords := fmt.Sprintf("%d,%d", p.X-nP.X, p.Y-nP.Y)
	switch coords {
	case "-1,0":
		return "left"
	case "1,0":
		return "right"
	case "0,-1":
		return "up"
	case "0,1":
		return "down"
	case "0,0":
		return "up"
	default:

		panic("snake went weird direction" + coords)
	}

}

// ConvertFrameToPNG takes a frame and makes a png
func ConvertFrameToPNG(w io.Writer, gameFrame *engine.GameFrame, gameStatus *engine.StatusResponse) {
	width, height := getDimensions(gameStatus)
	square := int32(20)
	border := int(square / 8)
	textSize := 15
	dc := gg.NewContext(width*int(square)+border, height*int(square)+border+textSize)
	dc.DrawRectangle(0, 0, float64(width*int(square)+border), float64(height*int(square)+border+textSize))
	dc.SetHexColor("#ffffff")
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

	drawWatermark(dc)

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

	dc.SetColor(color.Black)
	dc.DrawString("play.battlesnake.io", float64(border), float64(height*int(square)+border+textSize-5))
	dc.Fill()

	err := dc.EncodePNG(w)
	if err != nil {
		log.WithError(err).Error("error while encoding PNG")
	}
}

func drawWatermark(dc *gg.Context) {
	width, height := dc.Width(), dc.Height()
	//_, _, watermark := GetWatermarkImage(width, height)
	centerX := width / 2
	centerY := height / 2
	byteImage, err := SnakeImages.Find("watermark3.png")
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(byteImage)
	p, _ := png.Decode(r)
	dc.DrawImageAnchored(p, centerX, centerY, 0.5, 0.5)
}

// ConvertGameToGif reads all frames from the engine and outputs an animated gif.
func ConvertGameToGif(w io.Writer, gameStatus *engine.StatusResponse, gameID string, batchSize, startFrame, frameRange int) error {
	outGif := &gif.GIF{}
	c := make(chan gif.PalettAndDelay, 50)
	outGif.Image = c
	gameFrames, err := GetGameFramesWithLength(gameID, startFrame, batchSize)
	if err != nil {
		return err
	}
	outGif.SampleImage = createGif(&gameFrames.Frames[0], gameStatus).(*image.Paletted)
	go getGifFrames(c, gameFrames, outGif, gameStatus, gameID, startFrame, batchSize, frameRange)
	outGif.LoopCount = 0
	err = gif.EncodeAll(w, outGif)
	if err != nil {
		return err
	}
	return nil
}

func getGifFrames(c chan gif.PalettAndDelay, firstSet *engine.ListGameFramesResponse, outGif *gif.GIF, gameStatus *engine.StatusResponse, gameID string, startingOffset, batchSize, frameRange int) {
	currentOffset := startingOffset
	totalFrameCount := 0
mainLoop:
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
			totalFrameCount++
			imageGif := createGif(&frame, gameStatus)
			delay := 8
			if frameRange <= 10 {
				delay = 16
			}

			if totalFrameCount == frameRange {
				delay = 100
			}

			if gameStatus.LastFrame.Turn == frame.Turn {
				delay = 500
			}
			outGif.Image <- gif.PalettAndDelay{
				Palett: imageGif.(*image.Paletted),
				Delay:  delay,
				I:      frameCount + currentOffset - startingOffset,
			}
			if totalFrameCount >= frameRange {
				break mainLoop
			}
		}
		if frameCount < batchSize {
			break mainLoop
		}
		currentOffset += batchSize
		fmt.Printf("giffing game: %s: frames: %d of %d\n", gameID, currentOffset, gameStatus.LastFrame.Turn)
	}

	close(c)
}

func createGif(frame *engine.GameFrame, gameStatus *engine.StatusResponse) image.Image {
	var framePng bytes.Buffer
	ConvertFrameToPNG(&framePng, frame, gameStatus)
	imagePng, _, _ := image.Decode(&framePng)
	var frameGif bytes.Buffer
	err := giforiginal.Encode(&frameGif, imagePng, nil)
	if err != nil {
		log.WithError(err).Error("unable to encode gif")
	}
	imageGif, _, _ := image.Decode(&frameGif)
	return imageGif
}

func drawSquare(dc *gg.Context, x int32, y int32, square int32) {
	border := float64(square / 8)
	dc.DrawRectangle(float64(x*square)+border, float64(y*square)+border, float64(square)-border, float64(square)-border)
	dc.SetHexColor(squareColor)
	dc.Fill()
}

func drawFood(dc *gg.Context, point *engine.Point, square int32) {
	borderHalf := float64(square) / float64(20)
	dc.DrawCircle(float64(point.X*square)+float64(square)/2.0+borderHalf, float64(point.Y*square+square/2.0)+borderHalf, float64(square)/3.0)
	dc.SetHexColor("#ff5c75")
	dc.Fill()
}

func getRotation(point *engine.Point, nextPoint *engine.Point) float64 {
	direction := direction(point, nextPoint)
	switch direction {
	case "up":
		return -90
	case "down":
		return 90
	case "left":
		return -180
	case "right":
		return 0
		// if somehow a snake moves diagonally or some other direction due to future move mechanics.
	default:
		return 0
	}
}
func placeHead(snake *engine.Snake, dc *gg.Context, point *engine.Point, nextPoint *engine.Point, square int32, backgroundColor string) {
	segmentImage, err := GetOrCreateRotatedSnakeImage(HeadSegment, snake, backgroundColor, getRotation(point, nextPoint), float64(square))
	if err != nil {
		fmt.Printf("Couldn't load head segment: %s\n", err)
		return
	}
	drawSegment(dc, segmentImage, square, point)
}

func placeTail(snake *engine.Snake, dc *gg.Context, point *engine.Point, nextPoint *engine.Point, square int32, backgroundColor string) {
	segmentImage, err := GetOrCreateRotatedSnakeImage(TailSegment, snake, backgroundColor, getRotation(point, nextPoint), float64(square))
	if err != nil {
		fmt.Printf("Couldn't load tail segment: %s\n", err)
		return
	}
	drawSegment(dc, segmentImage, square, point)
}
func drawSegment(dc *gg.Context, segmentImage image.Image, square int32, point *engine.Point) {
	border := float64(square) / float64(8)
	dc.DrawImage(segmentImage, int(point.X*square+int32(border)), int(point.Y*square+int32(border)))
}

func drawSnake(dc *gg.Context, snake *engine.Snake, square int32) {
	halfSquare := float64(square) / float64(2)
	borderHalf := float64(square) / float64(20)
	border := float64(square) / float64(8)
	var previousPoint engine.Point
	var nextPoint *engine.Point
	if snake.Death.Cause != "" {
		snake.Color = "#cdcdcd"
	}
	for i, point := range snake.Body {

		if i < len(snake.Body)-1 {
			nextPoint = &snake.Body[i+1]
		} else {
			nextPoint = nil
		}
		if nextPoint != nil && (point.X == nextPoint.X && point.Y == nextPoint.Y) && i > 0 {
			continue
		}
		corner := corner(&previousPoint, &point, nextPoint)
		if i == 0 {
			placeHead(snake, dc, &point, nextPoint, square, squareColor)
		} else if i == len(snake.Body)-1 {
			placeTail(snake, dc, &point, &previousPoint, square, squareColor)
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

		dir := direction(&point, nextPoint)
		if nextPoint != nil {
			switch dir {
			case "up":
				dc.DrawRectangle(float64(point.X*square)+border, float64((point.Y+1)*square), float64(square)-border, border)
			case "down":
				dc.DrawRectangle(float64(point.X*square)+border, float64(point.Y*square), float64(square)-border, border)
			case "left":
				dc.DrawRectangle(float64((point.X+1)*square), float64(point.Y*square)+border, border, float64(square)-border)
			case "right":
				dc.DrawRectangle(float64(point.X*square), float64(point.Y*square)+border, border, float64(square)-border)
			}
		}
		dc.SetHexColor(snake.Color)
		dc.Fill()
		previousPoint = point
	}
}
