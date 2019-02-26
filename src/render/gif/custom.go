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
