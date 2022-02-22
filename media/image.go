package media

import (
	"image"
	"image/color"
	"image/draw"
)

func changeImageColor(src image.Image, c color.Color) image.Image {
	dstRect := image.Rect(src.Bounds().Min.X, src.Bounds().Min.Y, src.Bounds().Max.X, src.Bounds().Max.Y)
	dst := image.NewNRGBA(dstRect)

	srcImage := &image.Uniform{c}

	draw.DrawMask(dst, dstRect, srcImage, image.Point{}, src, image.Point{}, draw.Over)
	return dst
}
