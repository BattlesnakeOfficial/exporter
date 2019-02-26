package render

import (
	"fmt"

	"github.com/battlesnakeio/exporter/engine"
	log "github.com/sirupsen/logrus"
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
	Corner    string
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
				prev := snake.Body[i-1]
				direction := getDirection(prev, point)
				if prev.X == point.X && prev.Y == point.Y {
					direction = getDirection(snake.Body[i-2], point)
				}
				square := BoardSquare{
					Content:   BoardSquareSnakeTail,
					HexColor:  snake.Color,
					SnakeType: snake.Tail,
					Direction: direction,
				}
				board.SetSquare(&point, square)
			} else {
				square := BoardSquare{
					Content:   BoardSquareSnakeBody,
					HexColor:  snake.Color,
					Direction: getDirection(snake.Body[i+1], point),
					Corner:    getCorner(snake.Body[i-1], point, snake.Body[i+1]),
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
		log.Errorf("Unable to deterine snake direction: %s", d)
		return "up"
	}
}

// pP = previous point, p = current point, nP next point.
func getCorner(pP engine.Point, p engine.Point, nP engine.Point) string {
	coords := fmt.Sprintf("%d,%d:%d,%d", pP.X-p.X, pP.Y-p.Y, nP.X-p.X, nP.Y-p.Y)
	switch coords {
	case "0,-1:1,0", "1,0:0,-1":
		return "bottom-left"
	case "-1,0:0,-1", "0,-1:-1,0":
		return "bottom-right"
	case "-1,0:0,1", "0,1:-1,0":
		return "top-right"
	case "0,1:1,0", "1,0:0,1":
		return "top-left"
	default:
		return "none"
	}
}
