package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"runtime"
	"time"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/render/gif"
	"github.com/ericpauley/go-quantize/quantize"
	log "github.com/sirupsen/logrus"
)

const (
	GIFFrameDelay        = 8
	GIFLoopDelay         = 200
	GIFMaxColorsPerFrame = 256
)

func gameFrameToPalettedImage(g *engine.Game, gf *engine.GameFrame, w, h int) *image.Paletted {
	board := GameFrameToBoard(g, gf)

	// This is where the bulk of GIF creation CPU is spent.
	// First, Board is rendered to RGBA Image
	// Second, RGBA Image converted to Paletted Image (lossy)
	rgbaImage := DrawBoard(board, w, h)
	q := quantize.MedianCutQuantizer{}
	p := q.Quantize(make([]color.Color, 0, 256), rgbaImage)
	palettedImage := image.NewPaletted(rgbaImage.Bounds(), p)

	// No Dithering
	draw.Draw(palettedImage, rgbaImage.Bounds(), rgbaImage, image.Point{}, draw.Src)
	// FloydSteinberg Dithering (for future reference)
	// draw.FloydSteinberg.Draw(palettedImage, rgbaImage.Bounds(), rgbaImage, image.ZP)

	return palettedImage
}

func GameFrameToGIF(w io.Writer, g *engine.Game, gf *engine.GameFrame, width, height int) error {
	i := gameFrameToPalettedImage(g, gf, width, height)
	err := gif.Encode(w, i, nil)
	if err != nil {
		return err
	}
	return nil
}

func GameFramesToAnimatedGIF(w io.Writer, g *engine.Game, gameFrames []*engine.GameFrame, frameDelay, loopDelay, width, height int) error {
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
		start := time.Now()
		for i, gf := range gameFrames {
			delay := frameDelay
			if i == len(gameFrames)-1 {
				delay = loopDelay
			}
			c <- gif.GIFFrame{
				Image:    gameFrameToPalettedImage(g, gf, width, height),
				FrameNum: i,
				Delay:    delay,
			}
		}

		elapsed := time.Since(start)
		fps := 0.0
		// guard against divide by 0 in the unlikely event the elapsed time was 0.
		if elapsed.Seconds() > 0 {
			fps = float64(len(gameFrames)) / elapsed.Seconds()
		}
		log.WithFields(log.Fields{
			"game":     g.ID,
			"duration": elapsed,
			"fps":      fps,
		}).Infof("GIF render complete")

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
