package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/battlesnakeio/exporter/engine"
)

const (
	ASCIIEmpty     = " "
	ASCIIFood      = "*"
	ASCIISnakeHead = "H"
	ASCIISnakeBody = "O"
	ASCIISnakeTail = "T"
)

func GameFrameToASCII(w io.Writer, g *engine.Game, gf *engine.GameFrame) {
	board := GameFrameToBoard(g, gf)

	fmt.Fprint(w, strings.Repeat("-", board.Width+2)+"\n")
	for y := 0; y < board.Height; y++ {
		fmt.Fprint(w, "|")
		for x := 0; x < board.Width; x++ {
			switch board.Squares[x][y].Content {
			case BoardSquareSnakeHead:
				fmt.Fprint(w, ASCIISnakeHead)
			case BoardSquareSnakeBody:
				fmt.Fprint(w, ASCIISnakeBody)
			case BoardSquareSnakeTail:
				fmt.Fprint(w, ASCIISnakeTail)
			case BoardSquareFood:
				fmt.Fprint(w, ASCIIFood)
			case BoardSquareEmpty:
				fmt.Fprint(w, ASCIIEmpty)
			}
		}
		fmt.Fprint(w, "|\n")
	}
	fmt.Fprint(w, strings.Repeat("-", board.Width+2)+"\n")
}
