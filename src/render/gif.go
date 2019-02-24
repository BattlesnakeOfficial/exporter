package render

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io"

	"github.com/battlesnakeio/exporter/engine"
)

const (
	GIFFrameDelay = 8
)

func gameFrameToPalettedImage(g *engine.Game, gf *engine.GameFrame) *image.Paletted {
	board := GameFrameToBoard(g, gf)

	// This is where the bulk of GIF creation CPU is spent.
	// First, Board is rendered to RGBA Image
	// Second, RGBA Image converted to Paletted Image (lossy)
	rgbaImage := drawBoard(board)
	palettedImage := image.NewPaletted(rgbaImage.Bounds(), palette.Plan9)

	// No Dithering
	draw.Draw(palettedImage, rgbaImage.Bounds(), rgbaImage, image.ZP, draw.Src)
	// FloydSteinberg Dithering (for future reference)
	// draw.FloydSteinberg.Draw(palettedImage, rgbaImage.Bounds(), rgbaImage, image.ZP)

	return palettedImage
}

func GameFrameToGIF(w io.Writer, g *engine.Game, gf *engine.GameFrame) {
	i := gameFrameToPalettedImage(g, gf)
	gif.Encode(w, i, nil)
}

func GameFramesToAnimatedGIF(w io.Writer, g *engine.Game, gameFrames []*engine.GameFrame) {
	animatedGIF := &gif.GIF{}
	for _, gf := range gameFrames {
		i := gameFrameToPalettedImage(g, gf)
		animatedGIF.Image = append(animatedGIF.Image, i)
		animatedGIF.Delay = append(animatedGIF.Delay, GIFFrameDelay)
	}
	gif.EncodeAll(w, animatedGIF)
}
