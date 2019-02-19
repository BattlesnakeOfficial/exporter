package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"

	engine "github.com/battlesnakeio/exporter/engine"
	"github.com/fogleman/gg"
	"github.com/gobuffalo/packr/v2"
	"gopkg.in/go-playground/colors.v1"
)

// SnakeImages box of the custom snake parts.
var SnakeImages = packr.NewBox("../snake-images")

var snakeImagesCache = make(map[string]image.Image)

// SegmentType representing head or tail
type SegmentType string

const (
	// HeadSegment a head segment
	HeadSegment SegmentType = "head"
	// TailSegment a tail segment.
	TailSegment SegmentType = "tail"
)

type changeable interface {
	Set(x, y int, c color.Color)
}

func getHeadOrDefault(snake *engine.Snake) string {
	if snake.HeadType == "" {
		return "tongue"
	}
	return snake.HeadType
}

func getTailOrDefault(snake *engine.Snake) string {
	if snake.TailType == "" {
		return "bolt"
	}
	return snake.TailType
}

func getSafeHexColour(color string, defaultColor string) *colors.HEXColor {
	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}
	colorHex, err := colors.ParseHEX(color)
	if err != nil {
		if !strings.HasPrefix(defaultColor, "#") {
			defaultColor = "#" + defaultColor
		}
		colorHex, err = colors.ParseHEX(defaultColor)
		if err != nil {
			panic("default colour couldn't be parsed: " + defaultColor)
		}
	}
	return colorHex
}

// GetOrCreateRotatedSnakeImage holds all the asked for coloured / rotated segments in a cache in memory
func GetOrCreateRotatedSnakeImage(segmentType SegmentType, snake *engine.Snake, backgroundColor string, rotation float64) (image.Image, error) {
	name := getHeadOrDefault(snake)
	if segmentType == TailSegment {
		name = getTailOrDefault(snake)
	}
	h, imageType, err := getSnakeImage(name, segmentType)
	if err != nil {
		return nil, err
	}
	backgroundColorHex := getSafeHexColour(backgroundColor, "#111111")
	snakeColorHex := getSafeHexColour(snake.Color, "#FFFFFF")
	key := fmt.Sprintf("%s:%s:%s:%f", segmentType, snakeColorHex.String(), backgroundColorHex.String(), rotation)
	cachedResult, ok := snakeImagesCache[key]
	if ok {
		return cachedResult, nil
	}
	bounds := h.Bounds()
	if cimg, ok := h.(changeable); ok {
		for x := 0; x < bounds.Max.X; x++ {
			for y := 0; y < bounds.Max.Y; y++ {
				currentPoint := h.At(x, y)
				r, g, b, _ := currentPoint.RGBA()
				ratio := (float64(r) + float64(g) + float64(b)) / float64(3*65535)

				newR := float64(snakeColorHex.ToRGBA().R) * (1 - ratio)
				newG := float64(snakeColorHex.ToRGBA().G) * (1 - ratio)
				newB := float64(snakeColorHex.ToRGBA().B) * (1 - ratio)
				alpha := 255
				isLight := (ratio < 0.5)
				if !isLight {
					newR = float64(backgroundColorHex.ToRGBA().R) * ratio
					newG = float64(backgroundColorHex.ToRGBA().G) * ratio
					newB = float64(backgroundColorHex.ToRGBA().B) * ratio
					alpha = 255
					if snake.Death.Cause == "" {
						alpha = 0
					}
				}
				cimg.Set(x, y, color.RGBA{uint8(newR), uint8(newG), uint8(newB), uint8(alpha)})
			}
		}
		ic := gg.NewContext(h.Bounds().Max.X, h.Bounds().Max.Y)
		ic.RotateAbout(gg.Radians(rotation), 17, 17)
		ic.DrawImage(h, 0, 0)
		hr := ic.Image()
		// TODO remove the possible 6GB image cache and make it more sensible per game not global
		snakeImagesCache[key] = hr
		return hr, nil
	}
	return nil, fmt.Errorf("Not an image that can be modified. Type: %s", imageType)

}

// GetSnakeHeadImage returns a snake head or the default if not found.
func GetSnakeHeadImage(name string) (image.Image, string, error) {
	return getSnakeImage(name, "head")
}

// GetSnakeTailImage returns a snake head or the default if not found.
func GetSnakeTailImage(name string) (image.Image, string, error) {
	return getSnakeImage(name, "tail")
}

func getSnakeImage(name string, dir SegmentType) (image.Image, string, error) {

	byteImage, err := SnakeImages.Find(fmt.Sprintf("%s/%s.svg.png", dir, name))
	if err != nil {
		return nil, "", err
	}
	r := bytes.NewReader(byteImage)
	return image.Decode(r)
}
