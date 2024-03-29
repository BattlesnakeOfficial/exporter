package render

import (
	"testing"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCorner(t *testing.T) {

	// shifts the corner points around to test different edge cases
	shift := func(p engine.Point, x, y int) engine.Point {
		nP := engine.Point{X: p.X + x, Y: p.Y + y}

		// wrap around (assuming 3x3 board here)
		if nP.X < 0 {
			nP.X += 3
		}
		if nP.Y < 0 {
			nP.Y += 2
		}
		if nP.X > 2 {
			nP.X -= 3
		}
		if nP.Y > 2 {
			nP.Y -= 3
		}
		return nP
	}

	// This tries all the permutations of the corner and straight pieces being placed on a 3x3 board
	// It should make sure that all the wrapped cases work.
	// It checks 200 different permutations! 😎
	for _, x := range []int{-2, -1, 0, 1, 2} {
		for y := range []int{-2, -1, 0, 1, 2} {
			t.Logf("shifting x by %d, y by %d", x, y)
			// none
			assert.Equal(t, cornerNone, getCorner(shift(engine.Point{X: 0, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 1}, x, y), shift(engine.Point{X: 0, Y: 2}, x, y)))
			assert.Equal(t, cornerNone, getCorner(shift(engine.Point{X: 0, Y: 2}, x, y), shift(engine.Point{X: 0, Y: 1}, x, y), shift(engine.Point{X: 0, Y: 0}, x, y)))
			assert.Equal(t, cornerNone, getCorner(shift(engine.Point{X: 0, Y: 0}, x, y), shift(engine.Point{X: 1, Y: 0}, x, y), shift(engine.Point{X: 2, Y: 0}, x, y)))
			assert.Equal(t, cornerNone, getCorner(shift(engine.Point{X: 2, Y: 0}, x, y), shift(engine.Point{X: 1, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 0}, x, y)))

			// ╔
			assert.Equal(t, cornerTopLeft, getCorner(shift(engine.Point{X: 0, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 1}, x, y)))
			// ╗
			assert.Equal(t, cornerTopRight, getCorner(shift(engine.Point{X: 0, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 0}, x, y)))
			// ╝
			assert.Equal(t, cornerBottomRight, getCorner(shift(engine.Point{X: 1, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 0}, x, y)))
			// ╚
			assert.Equal(t, cornerBottomLeft, getCorner(shift(engine.Point{X: 1, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 1}, x, y)))
		}
	}
}

func TestGetDirection(t *testing.T) {

	cases := []struct {
		p1   engine.Point   // initial point
		p2   engine.Point   // point moved to
		want snakeDirection // direction of movement
		desc string         // description of test case
	}{
		// easy cases
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 3, Y: 4}, want: movingUp, desc: "non-wrapped up"},
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 3, Y: 2}, want: movingDown, desc: "non-wrapped down"},
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 4, Y: 3}, want: movingRight, desc: "non-wrapped right"},
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 2, Y: 3}, want: movingLeft, desc: "non-wrapped left"},

		// wrapped cases
		{p1: engine.Point{X: 1, Y: 10}, p2: engine.Point{X: 1, Y: 0}, want: movingUp, desc: "wrapped up"},
		{p1: engine.Point{X: 1, Y: 0}, p2: engine.Point{X: 1, Y: 10}, want: movingDown, desc: "wrapped down"},
		{p1: engine.Point{X: 0, Y: 4}, p2: engine.Point{X: 10, Y: 4}, want: movingLeft, desc: "wrapped left"},
		{p1: engine.Point{X: 10, Y: 0}, p2: engine.Point{X: 0, Y: 0}, want: movingRight, desc: "wrapped right"},
	}

	for _, tc := range cases {
		got := getDirection(tc.p1, tc.p2)
		assert.Equal(t, tc.want, got, tc.desc)
	}
}

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
	b.addSnakeTail(&engine.Point{X: 0, Y: 0}, parse.HexColor("#0acc33"), "regular", movingRight)
	assert.Equal(t, BoardSquareSnakeTail, b.getContents(0, 0)[0].Type, "(0,0) should have tail content")

	b.addSnakeBody(&engine.Point{X: 1, Y: 0}, parse.HexColor("#0acc33"), movingRight, "none")
	assert.Equal(t, BoardSquareSnakeBody, b.getContents(1, 0)[0].Type, "(1,0) should have body content")

	b.addSnakeHead(&engine.Point{X: 2, Y: 0}, parse.HexColor("#0acc33"), "regular", movingRight)
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
	assert.Equal(t, movingDown, c[0].Direction, "the head should be pointing down")
	assert.Equal(t, parse.HexColor("#3B194D"), c[0].Color, "the head should have the snake colour")
	assert.Equal(t, "beluga", c[0].SnakeType, "the head should be customised")

	// BODY
	c = b.getContents(0, 1)
	require.Len(t, c, 1, "there should only be a body here")
	assert.Equal(t, BoardSquareSnakeBody, c[0].Type, "this should be a body")
	assert.Equal(t, parse.HexColor("#3B194D"), c[0].Color, "the body should have the snake colour")
	assert.Equal(t, "", c[0].SnakeType, "the body should not have a customization")

	// TAIL
	c = b.getContents(1, 1)
	require.Len(t, c, 1, "there should only be a tail here")
	assert.Equal(t, BoardSquareSnakeTail, c[0].Type, "this should be a tail")
	assert.Equal(t, movingRight, c[0].Direction, "the tail should be pointing right")
	assert.Equal(t, parse.HexColor("#3B194D"), c[0].Color, "the tail should have the snake colour")
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
	assert.Equal(t, movingUp, c[0].Direction, "the head should be pointing up")
	assert.Equal(t, parse.HexColor(ColorDeadSnake), c[0].Color, "the head should have the dead snake colour")
	assert.Equal(t, "default", c[0].SnakeType, "the head should be default")

	// BODY
	c = b.getContents(5, 8)
	require.Len(t, c, 1, "there should only be a body here")
	assert.Equal(t, BoardSquareSnakeBody, c[0].Type, "this should be a body")
	assert.Equal(t, parse.HexColor(ColorDeadSnake), c[0].Color, "the body should have the dead snake colour")
	assert.Equal(t, "", c[0].SnakeType, "the body should not have a customization")

	// TAIL
	c = b.getContents(4, 8)
	require.Len(t, c, 1, "there should only be a tail here")
	assert.Equal(t, BoardSquareSnakeTail, c[0].Type, "this should be a tail")
	assert.Equal(t, movingLeft, c[0].Direction, "the tail should be pointing left")
	assert.Equal(t, parse.HexColor(ColorDeadSnake), c[0].Color, "the tail should have the dead snake colour")
	assert.Equal(t, "default", c[0].SnakeType, "the tail should be default")
}

func TestGrowingSnakePlacement(t *testing.T) {
	b := NewBoard(11, 11)

	// This test reproduces the issue found in DEV-1041.
	// The issue was caused by the snake board structure allowing multiple contents per-square.
	// When a snake eats and grows, it results in a body segment being placed on the same square as the tail.
	// Previously, the squares in the game board only supported one content item, so the tail would overwrite the body.
	// But now that squares support multiple contents, they both get rendered.
	// This results in the tail "disappearing" because the body fills up the whole square.

	t.Log("Placing a snake that just ate")
	s := engine.Snake{ID: "test_123",
		Name:   "A hungry snake",
		Body:   []engine.Point{{X: 10, Y: 8}, {X: 9, Y: 8}, {X: 9, Y: 9}, {X: 9, Y: 9}},
		Health: 100,
		Color:  "#4b0082",
		Head:   "bendr",
		Tail:   "freckled",
	}
	b.placeSnake(s)

	require.Len(t, b.getContents(9, 9), 1, "the snake tail should replace the body")
	require.Equal(t, BoardSquareSnakeTail, b.getContents(9, 9)[0].Type, "the snake tail should replace the body")
}

func TestRemoveIfExists(t *testing.T) {
	b := NewBoard(11, 11)

	// empty case
	// ensure removing something that doesn't exist doesn't cause a panic
	b.removeIfExists(0, 0, BoardSquareSnakeBody)

	// ensure a non-matching type doesn't get removed
	require.Len(t, b.getContents(0, 0), 0)
	b.addSnakeBody(&engine.Point{X: 0, Y: 0}, nil, movingUp, cornerBottomLeft)
	require.Len(t, b.getContents(0, 0), 1)
	b.removeIfExists(0, 0, BoardSquareFood)
	require.Len(t, b.getContents(0, 0), 1)

	// ensure that a matching type gets removed
	b.removeIfExists(0, 0, BoardSquareSnakeBody)
	require.Len(t, b.getContents(0, 0), 0)

	// ensure that removal works okay when there is more than one content
	b.addSnakeBody(&engine.Point{X: 0, Y: 0}, nil, movingUp, cornerBottomLeft)
	b.addHazard(&engine.Point{X: 0, Y: 0})
	require.Len(t, b.getContents(0, 0), 2)
	b.removeIfExists(0, 0, BoardSquareSnakeHead)
	require.Len(t, b.getContents(0, 0), 2, "shouldn't change when removing something that doesnt exist")
	b.removeIfExists(0, 0, BoardSquareSnakeBody)
	require.Len(t, b.getContents(0, 0), 1, "body should be gone now")
	require.Equal(t, BoardSquareHazard, b.getContents(0, 0)[0].Type, "just hazard should be left")
}
