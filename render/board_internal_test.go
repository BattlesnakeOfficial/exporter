package render

import (
	"testing"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoard(t *testing.T) {
	b := NewBoard(11, 11)

	// ensure initial state is clean and correct
	require.Len(t, b.squares, 0)
	for x := 0; x < 11; x++ {
		for y := 0; y < 11; y++ {
			s := b.getSquare(x, y)
			require.Nil(t, s, 0, "board square (%d,%d) should be empty", x, y)
		}
	}

	// ensure adding content works
	b.addSnakeTail(&engine.Point{X: 0, Y: 0}, "#0acc33", "regular", "right")
	assert.Equal(t, BoardSquareSnakeTail, b.getContents(0, 0)[0].Type, "(0,0) should have tail content")

	b.addSnakeBody(&engine.Point{X: 1, Y: 0}, "#0acc33", "right", "none")
	assert.Equal(t, BoardSquareSnakeBody, b.getContents(1, 0)[0].Type, "(1,0) should have body content")

	b.addSnakeHead(&engine.Point{X: 2, Y: 0}, "#0acc33", "regular", "right")
	assert.Equal(t, BoardSquareSnakeHead, b.getContents(2, 0)[0].Type, "(2,0) should have head content")

	b.addFood(&engine.Point{X: 3, Y: 0})
	assert.Equal(t, BoardSquareFood, b.getContents(3, 0)[0].Type, "(3,0) should have food content")

	b.addHazard(&engine.Point{X: 3, Y: 0})
	assert.Equal(t, BoardSquareHazard, b.getContents(3, 0)[1].Type, "(3,0) should ALSO have hazard content")
}

func TestPlaceSnake(t *testing.T) {
	b := NewBoard(11, 11)

	t.Log("Placing an alive snake")
	s := engine.Snake{
		// define properties that matter to the board
		Color: "#3B194D",
		Body: []engine.Point{
			{X: 0, Y: 0}, // head
			{X: 0, Y: 1},
			{X: 1, Y: 1}, // tail
		},
		Head: "beluga",
		Tail: "rattle",
	}
	b.placeSnake(s)

	// HEAD
	c := b.getContents(0, 0)
	require.Len(t, c, 1, "there should only be a head here")
	assert.Equal(t, BoardSquareSnakeHead, c[0].Type, "this should be a head")
	assert.Equal(t, "down", c[0].Direction, "the head should be pointing down")
	assert.Equal(t, "#3B194D", c[0].HexColor, "the head should have the snake colour")
	assert.Equal(t, "beluga", c[0].SnakeType, "the head should be customised")

	// BODY
	c = b.getContents(0, 1)
	require.Len(t, c, 1, "there should only be a body here")
	assert.Equal(t, BoardSquareSnakeBody, c[0].Type, "this should be a body")
	assert.Equal(t, "#3B194D", c[0].HexColor, "the body should have the snake colour")
	assert.Equal(t, "", c[0].SnakeType, "the body should not have a customization")

	// TAIL
	c = b.getContents(1, 1)
	require.Len(t, c, 1, "there should only be a tail here")
	assert.Equal(t, BoardSquareSnakeTail, c[0].Type, "this should be a tail")
	assert.Equal(t, "right", c[0].Direction, "the tail should be pointing right")
	assert.Equal(t, "#3B194D", c[0].HexColor, "the tail should have the snake colour")
	assert.Equal(t, "rattle", c[0].SnakeType, "the tail should be customised")

	t.Log("Placing a dead snake")
	s = engine.Snake{
		Death: &engine.Death{Cause: "", Turn: 10},
		// define properties that matter to the board
		Color: "#FFFFFF",
		Body: []engine.Point{
			{X: 5, Y: 9}, // head
			{X: 5, Y: 8},
			{X: 4, Y: 8}, // tail
		},
	}
	b.placeSnake(s)

	c = b.getContents(5, 9)
	require.Len(t, c, 1, "there should only be a head here")
	assert.Equal(t, BoardSquareSnakeHead, c[0].Type, "this should be a head")
	assert.Equal(t, "up", c[0].Direction, "the head should be pointing up")
	assert.Equal(t, "#cdcdcd", c[0].HexColor, "the head should have the dead snake colour")
	assert.Equal(t, "regular", c[0].SnakeType, "the head should be default")

	// BODY
	c = b.getContents(5, 8)
	require.Len(t, c, 1, "there should only be a body here")
	assert.Equal(t, BoardSquareSnakeBody, c[0].Type, "this should be a body")
	assert.Equal(t, "#cdcdcd", c[0].HexColor, "the body should have the dead snake colour")
	assert.Equal(t, "", c[0].SnakeType, "the body should not have a customization")

	// TAIL
	c = b.getContents(4, 8)
	require.Len(t, c, 1, "there should only be a tail here")
	assert.Equal(t, BoardSquareSnakeTail, c[0].Type, "this should be a tail")
	assert.Equal(t, "left", c[0].Direction, "the tail should be pointing left")
	assert.Equal(t, "#cdcdcd", c[0].HexColor, "the tail should have the dead snake colour")
	assert.Equal(t, "regular", c[0].SnakeType, "the tail should be default")
}
