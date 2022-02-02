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
	GIFFrameDelay        = 8
	GIFLoopDelay         = 200
	GIFMaxColorsPerFrame = 256
)

func gameFrameToPalettedImage(g *engine.Game, gf *engine.GameFrame) *image.Paletted {
	board := GameFrameToBoard(g, gf)

	// This is where the bulk of GIF creation CPU is spent.
	// First, Board is rendered to RGBA Image
	// Second, RGBA Image converted to Paletted Image (lossy)
	rgbaImage := drawBoard(board)
	palettedImage := image.NewPaletted(rgbaImage.Bounds(), buildGIFPallete(rgbaImage))

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
		fps := 0.0
		// guard against divide by 0 in the unlikely event the elapsed time was 0.
		if elapsed.Seconds() > 0 {
			fps = float64(len(gameFrames)) / elapsed.Seconds()
		}
		logrus.WithFields(logrus.Fields{
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

// getColorCounts finds all unique colours in an image and returns a count
// of how often those colours are used, sorted in descending order.
func getColorCounts(img image.Image) usageList {
	counts := map[color.Color]int{}
	m := img.Bounds().Max
	for x := 0; x < m.X; x++ {
		for y := 0; y < m.Y; y++ {
			counts[img.At(x, y)]++
		}
	}

	l := make(usageList, len(counts))
	i := 0
	for k, v := range counts {
		l[i] = colorUsage{k, v}
		i++
	}

	sort.Sort(l)

	return l
}

// buildGIFPallete builds a colour pallete that can be used to convert the given image to a GIF frame.
// Any image with any number of colours can be used. If the image has more colours than a GIF frame can
// support, the pallete will be a subset of the source image colours.
func buildGIFPallete(src image.Image) color.Palette {
	counts := getColorCounts(src)

	pal := make(color.Palette, min(GIFMaxColorsPerFrame, len(counts)))
	for i := 0; i < len(pal); i++ {
		pal[i] = counts[i].Key
	}

	return pal
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// colorUsage is simple pair of color and the number of times it's used
// It exists to make it easy to sort a slice of colors ordered by how much they are used.
type colorUsage struct {
	Key   color.Color
	Value int
}

// usageList is a type that is used to satisfy sort.Interface so we can sort colors by usage
type usageList []colorUsage

func (p usageList) Len() int           { return len(p) }
func (p usageList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p usageList) Less(i, j int) bool { return p[i].Value > p[j].Value } // should be descending order
