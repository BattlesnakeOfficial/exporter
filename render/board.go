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

func (b *Board) setSquare(p *engine.Point, s BoardSquare) {
	b.Squares[p.X][b.Height - 1 - p.Y] = s
}

func (b *Board) getDirection(p engine.Point, nP engine.Point) string {
	d := fmt.Sprintf("%d,%d", nP.X-p.X, p.Y-nP.Y)
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
func (b *Board) getCorner(pP engine.Point, p engine.Point, nP engine.Point) string {
	coords := fmt.Sprintf("%d,%d:%d,%d", pP.X-p.X, p.Y-pP.Y, nP.X-p.X, p.Y-nP.Y)
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

func (b *Board) placeSnake(snake engine.Snake) {
	// Default head type
	head := "regular"
	if len(snake.Head) > 0 {
		head = snake.Head
	}

	// Default tail type
	tail := "regular"
	if len(snake.Tail) > 0 {
		tail = snake.Tail
	}

	// Death color
	color := snake.Color
	if snake.Death != nil {
		color = "#cdcdcd"
	}

	for i, point := range snake.Body {
		if i == 0 {
			// Snake heads can exist off-board after colliding with a wall
			if point.X < 0 || point.X >= b.Width {
				continue
			}
			if point.Y < 0 || point.X >= b.Height {
				continue
			}

			square := BoardSquare{
				Content:   BoardSquareSnakeHead,
				HexColor:  color,
				SnakeType: head,
				Direction: b.getDirection(snake.Body[i+1], point),
			}
			b.setSquare(&point, square)
		} else if i == (len(snake.Body) - 1) {
			prev := snake.Body[i-1]
			direction := b.getDirection(prev, point)
			if prev.X == point.X && prev.Y == point.Y {
				direction = b.getDirection(snake.Body[i-2], point)
			}
			square := BoardSquare{
				Content:   BoardSquareSnakeTail,
				HexColor:  color,
				SnakeType: tail,
				Direction: direction,
			}
			b.setSquare(&point, square)
		} else {
			square := BoardSquare{
				Content:   BoardSquareSnakeBody,
				HexColor:  color,
				Direction: b.getDirection(snake.Body[i+1], point),
				Corner:    b.getCorner(snake.Body[i-1], point, snake.Body[i+1]),
			}
			b.setSquare(&point, square)
		}
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

	// First place dead snakes (up to 10 turns after death)
	for _, snake := range gf.Snakes {
		if snake.Death != nil && (gf.Turn-snake.Death.Turn) <= 10 {
			board.placeSnake(snake)
		}
	}

	// Second, place food
	for _, point := range gf.Food {
		board.setSquare(&point, BoardSquare{Content: BoardSquareFood})
	}

	// Third, place alive snakes
	for _, snake := range gf.Snakes {
		if snake.Death == nil {
			board.placeSnake(snake)
		}
	}

	return board
}
