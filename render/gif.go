package render

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"io"
	"runtime"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/render/gif"
)

const (
	GIFFrameDelay = 8
	GIFLoopDelay  = 200
)

func gameFrameToPalettedImage(g *engine.Game, gf *engine.GameFrame) *image.Paletted {
	board := GameFrameToBoard(g, gf)

	// This is where the bulk of GIF creation CPU is spent.
	// First, Board is rendered to RGBA Image
	// Second, RGBA Image converted to Paletted Image (lossy)
	rgbaImage := drawBoard(board)
	palettedImage := image.NewPaletted(rgbaImage.Bounds(), palette.Plan9)

	// No Dithering
	draw.Draw(palettedImage, rgbaImage.Bounds(), rgbaImage, image.Point{}, draw.Src)
	// FloydSteinberg Dithering (for future reference)
	// draw.FloydSteinberg.Draw(palettedImage, rgbaImage.Bounds(), rgbaImage, image.ZP)

	return palettedImage
}

func GameFrameToGIF(w io.Writer, g *engine.Game, gf *engine.GameFrame) error {
	i := gameFrameToPalettedImage(g, gf)
	err := gif.Encode(w, i, nil)
	if err != nil {
		return err
	}
	return nil
}

func GameFramesToAnimatedGIF(w io.Writer, g *engine.Game, gameFrames []*engine.GameFrame, frameDelay, loopDelay int) error {
	c := make(chan gif.GIFFrame)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := recoverToError(r)
				c <- gif.GIFFrame{
					Error: err,
				}
			}
		}()
		for i, gf := range gameFrames {
			delay := frameDelay
			if i == len(gameFrames)-1 {
				delay = loopDelay
			}
			c <- gif.GIFFrame{
				Image:    gameFrameToPalettedImage(g, gf),
				FrameNum: i,
				Delay:    delay,
			}
		}
		close(c)
	}()
	return gif.EncodeAllConcurrent(w, c)
}

func recoverToError(panicArg interface{}) error {
	var err error
	if panicErr, ok := panicArg.(error); ok {
		err = panicErr
	} else {
		err = fmt.Errorf("%v", panicArg)
	}
	source := "unknown"
	if _, filename, line, ok := runtime.Caller(4); ok {
		source = fmt.Sprintf("%s:%d", filename, line)
	}
	err = fmt.Errorf("panic at %s: %w", source, err)
	return err
}
