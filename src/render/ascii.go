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
	ASCIISnakeBody = "O"
)

func GameFrameToASCII(w io.Writer, g *engine.Game, gf *engine.GameFrame) {
	board := GameFrameToBoard(g, gf)

	var result string

	result += strings.Repeat("-", board.Width+2)
	result += "\n"

	for y := 0; y < board.Height; y++ {
		result += "|"
		for x := 0; x < board.Width; x++ {
			switch board.Squares[x][y].Content {
			case BoardSquareSnake:
				result += ASCIISnakeBody
			case BoardSquareFood:
				result += ASCIIFood
			case BoardSquareEmpty:
				result += ASCIIEmpty
			}
		}
		result += "|\n"
	}

	result += strings.Repeat("-", board.Width+2)

	fmt.Fprint(w, result)
}
