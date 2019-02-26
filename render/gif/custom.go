// This is our custom encoder for streaming gif frames to the browser.
// These are the only changes we've made to the standard image/gif.

package gif

import (
	"bufio"
	"image"
	"io"
)

type GIFFrame struct {
	Image    *image.Paletted
	FrameNum int
	Delay    int
}

func EncodeAllConcurrent(w io.Writer, c chan GIFFrame) error {
	g := &GIF{}

	// This is a hack to trick the encoder into letting us loop the animation
	// by making it think there's multiple images. Has no impact on frames rendered.
	g.Image = []*image.Paletted{nil, nil}

	e := encoder{g: *g, w: bufio.NewWriter(w)}
	for f := range c {
		if f.FrameNum == 0 {
			// Write header on first frame
			p := f.Image.Bounds().Max
			e.g.Config.Width = p.X
			e.g.Config.Height = p.Y
			e.writeHeader()
		}

		disposal := uint8(0)
		e.writeImageBlock(f.Image, f.Delay, disposal)
	}
	e.writeByte(sTrailer)
	e.flush()
	return e.err
}
