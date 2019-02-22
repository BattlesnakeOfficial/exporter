package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"strings"

	engine "github.com/battlesnakeio/exporter/engine"
	"github.com/fogleman/gg"
	"github.com/gobuffalo/packr/v2"
	"gopkg.in/go-playground/colors.v1"
)

// SnakeImages box of the custom snake parts.
var SnakeImages = packr.New("imagebox", "../snake-images")

var snakeImagesCache = make(map[string]image.Image)
var watermarkeCache = make(map[string]*Watermark)

// Watermark struct for cache map
type Watermark struct {
	logo   image.Image
	height int
	width  int
}

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

// GetWatermarkImage Returns a transparent watermark image
func GetWatermarkImage(width, height int) (int, int, image.Image) {
	key := fmt.Sprintf("%d,%d", width, height)
	cachedResult, ok := watermarkeCache[key]
	if ok {
		return cachedResult.width, cachedResult.height, cachedResult.logo
	}
	byteImage, err := SnakeImages.Find("watermark.png")
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(byteImage)
	img, _, err := image.Decode(r)
	if err != nil {
		panic(err)
	}

	ic := gg.NewContext(img.Bounds().Max.X, img.Bounds().Max.Y)
	ac := gg.NewContext(ic.Width(), ic.Height())
	ac.DrawRectangle(0, 0, float64(ac.Width()), float64(ac.Height()))
	ac.SetHexColor("#000000FF")
	ac.Fill()
	expectedWidth := float64(width) / float64(1.4)
	scale := expectedWidth / float64(img.Bounds().Max.X)
	expectedHeight := scale * float64(img.Bounds().Max.Y)
	result := ic.Image()
	ic.Scale(scale, scale)
	ic.DrawImage(img, 0, 0)
	ic.Clip()
	result = setAlpha(result)
	watermark := &Watermark{
		width:  int(expectedWidth),
		height: int(expectedHeight),
		logo:   result,
	}
	watermarkeCache[key] = watermark
	return int(expectedWidth), int(expectedHeight), result
}

func setAlpha(logo image.Image) image.Image {

	bounds := logo.Bounds()
	if cimg, ok := logo.(changeable); ok {
		for x := 0; x < bounds.Max.X; x++ {
			for y := 0; y < bounds.Max.Y; y++ {
				currentPoint := logo.At(x, y)
				r, g, b, a := currentPoint.RGBA()
				// ratio := (float64(r) + float64(g) + float64(b)) / float64(3)
				if a > 200 {
					cimg.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(140)})
				}
			}
		}
	}
	return logo
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
func GetOrCreateRotatedSnakeImage(segmentType SegmentType, snake *engine.Snake, backgroundColor string, rotation float64, square float64) (image.Image, error) {
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
				isLight := ratio < 0.5
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
		x, y := float64(h.Bounds().Max.X), float64(h.Bounds().Max.Y)

		ic := gg.NewContext(int(x), int(y))
		if square > x {
			ic = gg.NewContext(int(square), int(square))
		}
		hr := ic.Image()
		ic.Scale((square-(square/8))/x, (square-(square/8))/x)
		ic.RotateAbout(gg.Radians(rotation), x/float64(2), y/float64(2))
		ic.DrawImage(h, 0, 0)
		ic.Clip()
		// TODO remove the possible 6TB image cache and make it more sensible per game not global
		snakeImagesCache[key] = hr
		return hr, nil
	}
	return nil, fmt.Errorf("not an image that can be modified. Type: %s", imageType)

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
