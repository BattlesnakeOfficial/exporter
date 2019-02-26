package render

import (
	"fmt"

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
	Content   BoardSquareContent
	HexColor  string
	SnakeType string
	Direction string
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
		if snake.Death != nil {
			continue
		}

		// Default snake types
		if len(snake.Head) == 0 {
			snake.Head = "regular"
		}
		if len(snake.Tail) == 0 {
			snake.Tail = "regular"
		}

		for i, point := range snake.Body {
			if i == 0 {
				square := BoardSquare{
					Content:   BoardSquareSnakeHead,
					HexColor:  snake.Color,
					SnakeType: snake.Head,
					Direction: getDirection(snake.Body[i+1], point),
				}
				board.SetSquare(&point, square)
			} else if i == (len(snake.Body) - 1) {
				square := BoardSquare{
					Content:   BoardSquareSnakeTail,
					HexColor:  snake.Color,
					SnakeType: snake.Tail,
					Direction: getDirection(snake.Body[i-1], point),
				}
				board.SetSquare(&point, square)
			} else {
				square := BoardSquare{
					Content:  BoardSquareSnakeBody,
					HexColor: snake.Color,
				}
				board.SetSquare(&point, square)
			}
		}
	}

	for _, point := range gf.Food {
		board.SetSquare(&point, BoardSquare{Content: BoardSquareFood})
	}

	return board
}

func getDirection(p engine.Point, nP engine.Point) string {
	d := fmt.Sprintf("%d,%d", nP.X-p.X, nP.Y-p.Y)
	switch d {
	case "1,0":
		return "right"
	case "0,1":
		return "down"
	case "-1,0":
		return "left"
	case "0,-1":
		return "up"
	case "0,0":
		return "right"
	default:
		panic(fmt.Errorf("Unable to deterine snake direction: %s", d))
	}
}
