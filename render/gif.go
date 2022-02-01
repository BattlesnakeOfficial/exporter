package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"runtime"
	"sort"
	"time"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/render/gif"
	"github.com/sirupsen/logrus"
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

	// pallete := []color.Color{
	// 	parseHexColor(ColorFood),
	// 	parseHexColor(ColorEmptySquare),
	// 	parseHexColor(ColorHazard),
	// 	parseHexColor(ColorDeadSnake),
	// 	color.RGBA{136, 136, 136, 255},
	// 	color.RGBA{136, 68, 68, 255},
	// 	color.RGBA{238, 238, 238, 255},
	// 	color.RGBA{61, 47, 144, 255},
	// }
	// printPallete(rgbaImage)
	// for _, s := range gf.Snakes {
	// 	// GIFS can't support more than 256 colours per-frame
	// 	// if len(pallete) >= 256 {
	// 	// 	break
	// 	// }
	// 	pallete = append(pallete, parseHexColor(s.Color))
	// }
	palettedImage := image.NewPaletted(rgbaImage.Bounds(), getPallete(rgbaImage))

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
		start := time.Now()
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
		elapsed := time.Since(start)
		fps := float64(len(gameFrames)) / elapsed.Seconds()
		logrus.WithFields(logrus.Fields{
			"game":     g.ID,
			"duration": elapsed,
			"fps":      fps,
		}).Infof("render complete")
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

func getPallete(img image.Image) color.Palette {
	colors := map[color.Color]int{}
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			colors[img.At(x, y)]++
		}
	}

	p := make(PairList, len(colors))

	i := 0
	for k, v := range colors {
		p[i] = Pair{k, v}
		i++
	}

	sort.Sort(p)

	pal := make(color.Palette, min(256, len(p)))
	for i := 0; i < len(pal); i++ {
		pal[i] = p[i].Key
	}

	return pal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Pair struct {
	Key   color.Color
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value > p[j].Value }
