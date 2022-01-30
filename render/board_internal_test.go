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
	require.Equal(t, 11, b.Width)
	require.Equal(t, 11, b.Height)
	require.Len(t, b.squares, 11)
	for y := 0; y < 11; y++ {
		require.Len(t, b.squares[y], 11, "row %d should have length 11", y)
		for x := 0; x < 11; x++ {
			s := b.getSquare(x, y)
			require.Len(t, s.Contents, 0, "board square (%d,%d) should be empty", x, y)
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
