package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/BattlesnakeOfficial/exporter/engine"
)

const (
	ASCIIEmpty     = " "
	ASCIIFood      = "*"
	ASCIISnakeHead = "H"
	ASCIISnakeBody = "O"
	ASCIISnakeTail = "T"
)

func GameFrameToASCII(w io.Writer, g *engine.Game, gf *engine.GameFrame) error {
	board := GameFrameToBoard(g, gf)

	_, err := fmt.Fprint(w, strings.Repeat("-", board.Width+2)+"\n")
	if err != nil {
		return err
	}
	for y := board.Height - 1; y >= 0; y-- {
		_, err := fmt.Fprint(w, "|")
		if err != nil {
			return err
		}
		for x := 0; x < board.Width; x++ {
			// empty squares
			if len(board.getContents(x, y)) == 0 {
				_, err = fmt.Fprint(w, ASCIIEmpty)
				if err != nil {
					return err
				}
				continue
			}

			contents := board.getContents(x, y)
			last := contents[len(contents)-1] // since ascii can't have overlapping items, we take the last thing on the square
			switch last.Type {
			case BoardSquareSnakeHead:
				_, err := fmt.Fprint(w, ASCIISnakeHead)
				if err != nil {
					return err
				}
			case BoardSquareSnakeBody:
				_, err = fmt.Fprint(w, ASCIISnakeBody)
				if err != nil {
					return err
				}
			case BoardSquareSnakeTail:
				_, err = fmt.Fprint(w, ASCIISnakeTail)
				if err != nil {
					return err
				}
			case BoardSquareFood:
				_, err = fmt.Fprint(w, ASCIIFood)
				if err != nil {
					return err
				}
			}
		}
		_, err = fmt.Fprint(w, "|\n")
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, strings.Repeat("-", board.Width+2)+"\n")
	if err != nil {
		return err
	}
	return nil
}
