package render

import (
	"testing"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameFrameToBoard(t *testing.T) {
	g := engine.Game{
		Width:  7,
		Height: 7,
	}
	gf := engine.GameFrame{
		Turn: 50,
		Food: []engine.Point{{X: 3, Y: 3}},
		Snakes: []engine.Snake{
			// alive corner snake
			{
				Body: []engine.Point{{X: 0, Y: 0}, {X: 0, Y: 1}},
			},
			// expired dead snake (> 10 turns since death)
			{
				Death: &engine.Death{Cause: "", Turn: 12},
				Body:  []engine.Point{{X: 2, Y: 0}, {X: 2, Y: 1}},
			},
			// recently dead snake
			{
				Death: &engine.Death{Cause: "", Turn: 48},
				Body:  []engine.Point{{X: 4, Y: 2}, {X: 3, Y: 2}},
			},
		},
		Hazards: []engine.Point{{X: 0, Y: 6}, {X: 1, Y: 6}, {X: 2, Y: 6}, {X: 3, Y: 6}, {X: 4, Y: 6}, {X: 5, Y: 6}, {X: 6, Y: 6}},
	}
	b := GameFrameToBoard(&g, &gf)
	require.NotNil(t, b)
	assert.Equal(t, 7, b.Width, "board width should match game")
	assert.Equal(t, 7, b.Height, "board height should match game")
	assert.Len(t, b.squares, 1+2+2+7) // 1 food, snake of 2, snake of 2, hazard of 7

	// food
	assert.Len(t, b.getContents(3, 3), 1, "food should be placed")
	assert.Equal(t, BoardSquareFood, b.getContents(3, 3)[0].Type)

	// expired dead snake
	assert.Len(t, b.getContents(2, 0), 0, "expired snake shouldn't get placed")
	assert.Len(t, b.getContents(2, 1), 0, "expired snake shouldn't get placed")

	// non expired, dead snake
	require.Len(t, b.getContents(4, 2), 1, "recently dead snake should be placed")
	assert.Equal(t, BoardSquareSnakeHead, b.getContents(4, 2)[0].Type)
	require.Len(t, b.getContents(3, 2), 1, "recently dead snake should be placed")
	assert.Equal(t, BoardSquareSnakeTail, b.getContents(3, 2)[0].Type)

	// alive snake
	require.Len(t, b.getContents(0, 0), 1, "alive snake should be placed")
	assert.Equal(t, BoardSquareSnakeHead, b.getContents(0, 0)[0].Type)
	require.Len(t, b.getContents(0, 1), 1, "alive snake should be placed")
	assert.Equal(t, BoardSquareSnakeTail, b.getContents(0, 1)[0].Type)

	// hazard line
	for i := 0; i < 7; i++ {
		require.Len(t, b.getContents(i, 6), 1, "hazard (%d,6) should exist", i)
		assert.Equal(t, BoardSquareHazard, b.getContents(i, 6)[0].Type)
	}
}
