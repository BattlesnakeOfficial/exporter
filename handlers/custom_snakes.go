package handlers

import (
	"bytes"
	"fmt"
	"image"

	"github.com/gobuffalo/packr"
)

// SnakeImages box of the custom snake parts.
var SnakeImages = packr.NewBox("../snake-images")

// GetSnakeHeadImage returns a snake head or the default if not found.
func GetSnakeHeadImage(name string) (image.Image, error) {
	return getSnakeImage(name, "head")
}

// GetSnakeTailImage returns a snake head or the default if not found.
func GetSnakeTailImage(name string) (image.Image, error) {
	return getSnakeImage(name, "tail")
}

func getSnakeImage(name string, dir string) (image.Image, error) {

	byteImage, err := SnakeImages.Find(fmt.Sprintf("%s/%s.svg.png", dir, name))
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(byteImage)
	image, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return image, nil
}
