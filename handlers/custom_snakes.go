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

	byteImage, err := SnakeImages.Find(fmt.Sprintf("head/%s.svg.png", name))
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(byteImage)
	image, imageType, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	fmt.Println(imageType)
	return image, nil
}
