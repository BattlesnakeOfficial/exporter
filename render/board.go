package render

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/parse"
	log "github.com/sirupsen/logrus"
)

type snakeCorner string

const (
	cornerBottomLeft  snakeCorner = "bottom-left"  // ╚
	cornerBottomRight snakeCorner = "bottom-right" // ╝
	cornerTopLeft     snakeCorner = "top-left"     // ╔
	cornerTopRight    snakeCorner = "top-right"    // ╗
	cornerNone        snakeCorner = "none"
)

func (c snakeCorner) isBottom() bool {
	return strings.HasPrefix(string(c), "bottom")
}

func (c snakeCorner) isLeft() bool {
	return strings.HasSuffix(string(c), "left")
}

type snakeDirection int

const (
	movingUp snakeDirection = iota
	movingDown
	movingLeft
	movingRight
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
	Color     color.Color
	SnakeType string
	Direction snakeDirection
	Corner    snakeCorner
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

// removeIfExists checks whether the board square at coordinate (x,y)
// contains the given content type and removes it if it does
func (b Board) removeIfExists(x, y int, t BoardSquareContentType) {
	s := b.getSquare(x, y)
	if s == nil {
		return
	}

	foundIdx := -1
	for i, c := range s.Contents {
		if c.Type == t {
			foundIdx = i
			break
		}
	}

	if foundIdx < 0 {
		// not found, we're done
		return
	}

	// if only 1 item, reset contents to nil
	if len(s.Contents) == 1 {
		s.Contents = nil
		return
	}

	// more than one item, we need to update the slice
	// this is using the "append based removal" strategy
	// which re-uses the underlying slice
	s.Contents = append(s.Contents[:foundIdx], s.Contents[foundIdx+1:]...)
}

func (b *Board) addFood(p *engine.Point) {
	b.addContent(p, BoardSquareContent{
		Type: BoardSquareFood,
	})
}

func (b *Board) addSnakeTail(p *engine.Point, c color.Color, snakeType string, direction snakeDirection) {
	// when a snake eats and grows, the tail is placed on the same square as a body
	// this makes sure we remove the body segment if that condition is hit
	b.removeIfExists(p.X, p.Y, BoardSquareSnakeBody)

	b.addContent(p, BoardSquareContent{
		Type:      BoardSquareSnakeTail,
		Color:     c,
		SnakeType: snakeType,
		Direction: direction,
	})
}

func (b *Board) addSnakeHead(p *engine.Point, c color.Color, snakeType string, dir snakeDirection) {
	b.addContent(p, BoardSquareContent{
		Type:      BoardSquareSnakeHead,
		Color:     c,
		SnakeType: snakeType,
		Direction: dir,
	})
}

func (b *Board) addSnakeBody(p *engine.Point, c color.Color, dir snakeDirection, corner snakeCorner) {
	b.addContent(p, BoardSquareContent{
		Type:      BoardSquareSnakeBody,
		Color:     c,
		Direction: dir,
		Corner:    corner,
	})
}

func (b *Board) addHazard(p *engine.Point) {
	b.addContent(p, BoardSquareContent{
		Type: BoardSquareHazard,
	})
}

func getDirection(p engine.Point, nP engine.Point) snakeDirection {
	// handle when the points haven't changed (a common case) by defaulting to "right"
	if p == nP {
		return movingRight
	}

	// handle cases where we aren't wrapping around the board
	if p.X+1 == nP.X {
		return movingRight
	}

	if p.X-1 == nP.X {
		return movingLeft
	}

	if p.Y+1 == nP.Y {
		return movingUp
	}

	if p.Y-1 == nP.Y {
		return movingDown
	}

	// handle cases where we are wrapping around the board
	if p.X > nP.X && nP.X == 0 {
		return movingRight
	}

	if p.X < nP.X && p.X == 0 {
		return movingLeft
	}

	if p.Y > nP.Y && nP.Y == 0 {
		return movingUp
	}

	if p.Y < nP.Y && p.Y == 0 {
		return movingDown
	}

	// default to "right" when invalid moves are passed
	log.Errorf("Unable to determine snake direction: %v to %v", p, nP)
	return movingRight
}

// getCorner gets the corner type for the given 3 segments.
// pP = previous point, p = current point, nP next point.
// note: p is also the "corner point" ;)
func getCorner(pP engine.Point, p engine.Point, nP engine.Point) snakeCorner {
	// for a corner, there needs to be an X AND a Y change
	if (pP.X == p.X && pP.X == nP.X) || (pP.Y == p.Y && pP.Y == nP.Y) {
		// either X or Y hasn't changed, so no corner
		return cornerNone
	}

	// okay, we have a corner - time to figure out what kind!
	yType := "top"
	xType := "right"

	yDiff := p.Y - pP.Y
	if yDiff == 0 {
		yDiff = p.Y - nP.Y
	}

	// it's a bottom corner if one point is above the corner
	// wrapped mode makes "above" a bit trickier ;)
	// NOTE: "above" means a larger Y value on the Battlesnake board
	if abs(yDiff) == 1 {
		// non-wrapped
		if yDiff < 0 {
			// corner is below a point
			yType = "bottom"
		}
	} else {
		// wrapped
		if yDiff > 0 {
			yType = "bottom"
		}
	}

	xDiff := p.X - pP.X
	if xDiff == 0 {
		xDiff = p.X - nP.X
	}

	// it's a left corner if either point is "right" of corner point
	// wrapped mode also makes this trickier ;)
	// NOTE: "right" means a larger X value on the Battlesnake board
	if abs(xDiff) == 1 {
		// non-wrapped
		if xDiff < 0 {
			xType = "left"
		}
	} else {
		// wrapped
		if xDiff > 0 {
			xType = "left"
		}
	}

	return snakeCorner(fmt.Sprintf("%s-%s", yType, xType))
}

func (b *Board) placeSnake(snake engine.Snake) {
	// Default head type
	head := "default"
	if len(snake.Head) > 0 {
		head = snake.Head
	}

	// Default tail type
	tail := "default"
	if len(snake.Tail) > 0 {
		tail = snake.Tail
	}

	// Death color
	color := parse.HexColor(snake.Color)
	if snake.Death != nil {
		color = parse.HexColor(ColorDeadSnake)
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
