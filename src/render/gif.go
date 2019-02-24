package render

import (
	"image/gif"
	"io"

	"github.com/battlesnakeio/exporter/engine"
)

func GameFrameToGIF(w io.Writer, g *engine.Game, gf *engine.GameFrame) {
	board := GameFrameToBoard(g, gf)
	image := BoardToImage(board)
	gif.Encode(w, image, nil)
}
