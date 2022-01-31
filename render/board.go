package render

import (
	"fmt"

	"github.com/BattlesnakeOfficial/exporter/engine"
	log "github.com/sirupsen/logrus"
)

const (
	BoardSquareFood BoardSquareContentType = iota
	BoardSquareSnakeBody
	BoardSquareSnakeHead
	BoardSquareSnakeTail
	BoardSquareHazard
)

// ColorDeadSnake is the default hex colour used for displaying snakes that have died
const ColorDeadSnake = "#cdcdcd"

// BoardSquareContentType works like an enum.
// It provides a restricted set of types of content that can be placed in a board square.
type BoardSquareContentType int

// BoardSquareContent represents a single piece of content in a single square of the game board.
// Examples of content are food, snake body parts and hazard squares
type BoardSquareContent struct {
	Type      BoardSquareContentType
	HexColor  string
	SnakeType string
	Direction string
	Corner    string
}

// BoardSquare represents a unique location on the game board.
type BoardSquare struct {
	Contents []BoardSquareContent
}

// Board is the root datastructure that represents a game board
type Board struct {
	Width   int
	Height  int
	squares map[engine.Point]*BoardSquare
}

// getSquare gets the BoardSquare at the given coordinates.
// It returns nil if the square is empty (or if the coordinate is out of bounds).
func (b *Board) getSquare(x, y int) *BoardSquare {
	return b.squares[engine.Point{X: x, Y: y}]
}

func (b *Board) addContent(p *engine.Point, c BoardSquareContent) {
	s := b.getSquare(p.X, p.Y)

	// initialise squares for empty locations
	if s == nil {
		s = &BoardSquare{}
		b.squares[*p] = s
	}

	s.Contents = append(s.Contents, c)
}

// getContents gets the contents of the board at the specified position.
// It is safe to call for any position.
// Empty squares will have an empty list.
func (b Board) getContents(x, y int) []BoardSquareContent {
	s := b.getSquare(x, y)
	if s == nil {
		return nil
	}

	return s.Contents
}

func (b *Board) addFood(p *engine.Point) {
	b.addContent(p, BoardSquareContent{
		Type: BoardSquareFood,
	})
}

func (b *Board) addSnakeTail(p *engine.Point, color, snakeType, direction string) {
	b.addContent(p, BoardSquareContent{
		Type:      BoardSquareSnakeTail,
		HexColor:  color,
		SnakeType: snakeType,
		Direction: direction,
	})
}

func (b *Board) addSnakeHead(p *engine.Point, color, snakeType, direction string) {
	b.addContent(p, BoardSquareContent{
		Type:      BoardSquareSnakeHead,
		HexColor:  color,
		SnakeType: snakeType,
		Direction: direction,
	})
}

func (b *Board) addSnakeBody(p *engine.Point, color, direction, corner string) {
	b.addContent(p, BoardSquareContent{
		Type:      BoardSquareSnakeBody,
		HexColor:  color,
		Direction: direction,
		Corner:    corner,
	})
}

func (b *Board) addHazard(p *engine.Point) {
	b.addContent(p, BoardSquareContent{
		Type: BoardSquareHazard,
	})
}

func getDirection(p engine.Point, nP engine.Point) string {
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
func getCorner(pP engine.Point, p engine.Point, nP engine.Point) string {
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
		color = ColorDeadSnake
	}

	for i, point := range snake.Body {
		if i == 0 {
			// Snake heads can exist off-board after colliding with a wall
			if point.X < 0 || point.X >= b.Width {
				continue
			}
			if point.Y < 0 || point.Y >= b.Height {
				continue
			}

			b.addSnakeHead(&point, color, head, getDirection(snake.Body[i+1], point))
			continue
		}

		// Skip any body parts which overlap the head
		if point == snake.Body[0] {
			continue
		}

		if i == (len(snake.Body) - 1) {
			prev := snake.Body[i-1]
			direction := getDirection(prev, point)
			if prev.X == point.X && prev.Y == point.Y {
				direction = getDirection(snake.Body[i-2], point)
			}
			b.addSnakeTail(&point, color, tail, direction)
		} else {
			direction := getDirection(snake.Body[i+1], point)
			corner := getCorner(snake.Body[i-1], point, snake.Body[i+1])
			b.addSnakeBody(&point, color, direction, corner)
		}
	}
}

func NewBoard(w int, h int) *Board {
	b := Board{Width: w, Height: h, squares: make(map[engine.Point]*BoardSquare)}
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
		board.addFood(&point)
	}

	// Third, place alive snakes
	for _, snake := range gf.Snakes {
		if snake.Death == nil {
			board.placeSnake(snake)
		}
	}

	// Fourth, place hazards
	for _, point := range gf.Hazards {
		board.addHazard(&point)
	}

	return board
}
