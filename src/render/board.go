package render

import (
	"github.com/battlesnakeio/exporter/engine"
)

const (
	BoardSquareEmpty     = 0 // Zero State (Default)
	BoardSquareFood      = 1
	BoardSquareSnake     = 2
	BoardSquareDeadSnake = 3
)

type BoardSquareContent int

type BoardSquare struct {
	Content BoardSquareContent
}

type Board struct {
	Width   int
	Height  int
	Squares [][]BoardSquare
}

func (b *Board) SetSquare(p *engine.Point, c BoardSquareContent) {
	b.Squares[p.X][p.Y] = BoardSquare{
		Content: c,
	}
}

func NewBoard(w int, h int) *Board {
	b := Board{Width: w, Height: h}

	b.Squares = make([][]BoardSquare, w)
	for i := range b.Squares {
		b.Squares[i] = make([]BoardSquare, h)
	}

	return &b
}

func GameFrameToBoard(g *engine.Game, gf *engine.GameFrame) *Board {
	board := NewBoard(g.Width, g.Height)

	for _, snake := range gf.Snakes {
		for _, point := range snake.Body {
			board.SetSquare(&point, BoardSquareSnake)
		}
	}

	for _, point := range gf.Food {
		board.SetSquare(&point, BoardSquareFood)
	}

	return board
}
