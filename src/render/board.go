package render

import (
	"github.com/battlesnakeio/exporter/engine"
)

const (
	BoardSquareEmpty     = 0 // Zero State (Default)
	BoardSquareFood      = 1
	BoardSquareSnakeBody = 2
	BoardSquareSnakeHead = 3
	BoardSquareSnakeTail = 4
	BoardSquareDeadSnake = 5
)

type BoardSquareContent int

type BoardSquare struct {
	Content  BoardSquareContent
	HexColor string
}

type Board struct {
	Width   int
	Height  int
	Squares [][]BoardSquare
}

func (b *Board) SetSquare(p *engine.Point, s BoardSquare) {
	b.Squares[p.X][p.Y] = s
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
		for i, point := range snake.Body {
			if i == 0 {
				board.SetSquare(&point, BoardSquare{BoardSquareSnakeHead, snake.Color})
			} else if i == (len(snake.Body) - 1) {
				board.SetSquare(&point, BoardSquare{BoardSquareSnakeTail, snake.Color})
			} else {
				board.SetSquare(&point, BoardSquare{BoardSquareSnakeBody, snake.Color})
			}
		}
	}

	for _, point := range gf.Food {
		board.SetSquare(&point, BoardSquare{BoardSquareFood, ""})
	}

	return board
}
