package handlers

import (
	"encoding/json"
	"testing"

	engine "github.com/battlesnakeio/exporter/engine"
	"github.com/stretchr/testify/require"
)

func TestConvertGridToString(t *testing.T) {
	frame := createFrame()
	frame.Food = []engine.Point{
		engine.Point{X: 0, Y: 2},
	}
	grid := ConvertFrameToGrid(frame, createGameStatus(3, 3))
	board := ConvertGridToString(grid)
	require.Equal(t,
		""+
			"--------\n"+
			"|H1B1  |\n"+
			"|  T1  |\n"+
			"|FF  H3|\n"+
			"--------\n", board)
}
func TestConvertFrameToMove(t *testing.T) {
	frame := createFrame()
	frame.Food = []engine.Point{
		engine.Point{X: 0, Y: 2},
	}
	gameStatus := createGameStatus(2, 2)

	move, _ := ConvertFrameToMove(frame, gameStatus, "id2")
	json, _ := json.Marshal(move)
	require.Equal(t, "{\"game\":{\"id\":\"GameID\"},\"turn\":0,\"board\":{\"height\":2,\"width\":2,\"food\":[{\"x\":0,\"y\":2}],\"snakes\":[{\"id\":\"id1\",\"name\":\"\",\"health\":0,\"body\":[{\"x\":0,\"y\":0},{\"x\":1,\"y\":0},{\"x\":1,\"y\":1}]},{\"id\":\"id2\",\"name\":\"\",\"health\":0,\"body\":[{\"x\":2,\"y\":2}]}]},\"you\":{\"id\":\"id2\",\"name\":\"\",\"health\":0,\"body\":[{\"x\":2,\"y\":2}]}}", string(json))
}

func createGameStatus(width int, height int) *engine.StatusResponse {
	return &engine.StatusResponse{
		Game: engine.Game{
			Height: int32(height),
			Width:  int32(width),
			ID:     "GameID",
		},
	}
}
func TestFrameToGridOneSnake(t *testing.T) {
	frame := &engine.GameFrame{
		Snakes: []engine.Snake{
			engine.Snake{
				ID:    "id1",
				Color: "6611FF",
				Body: []engine.Point{
					engine.Point{
						X: 0, Y: 0,
					},
				},
			},
		},
	}
	grid := ConvertFrameToGrid(frame, createGameStatus(2, 2))
	require.Equal(t, [][]Pixel{
		{Pixel{ID: "id1", Colour: "6611FF", PixelType: "H"}, Pixel{PixelType: Space}},
		{Pixel{PixelType: Space}, Pixel{PixelType: Space}}}, grid)
}

func TestFrameToGridOneSnakeAndFood(t *testing.T) {
	frame := &engine.GameFrame{
		Food: []engine.Point{
			engine.Point{
				X: 1, Y: 1,
			},
		},
		Snakes: []engine.Snake{
			engine.Snake{
				ID:    "id1",
				Color: "6611FF",
				Body: []engine.Point{
					engine.Point{
						X: 0, Y: 0,
					},
				},
			},
		},
	}
	grid := ConvertFrameToGrid(frame, createGameStatus(2, 2))
	require.Equal(t, [][]Pixel{
		{Pixel{ID: "id1", Colour: "6611FF", PixelType: Head}, Pixel{PixelType: Space}},
		{Pixel{PixelType: Space}, Pixel{PixelType: Food}}}, grid)
}

func TestFrameToGridEmpty(t *testing.T) {
	frame := &engine.GameFrame{}
	grid := ConvertFrameToGrid(frame, createGameStatus(1, 1))
	require.Equal(t, [][]Pixel{
		{Pixel{PixelType: Space}}}, grid)
}

func createFrame() *engine.GameFrame {
	return &engine.GameFrame{
		Snakes: []engine.Snake{
			engine.Snake{
				ID:       "id1",
				Color:    "6611FF",
				HeadType: "fang",
				Body: []engine.Point{
					engine.Point{
						X: 0, Y: 0,
					},
					engine.Point{
						X: 1, Y: 0,
					},
					engine.Point{
						X: 1, Y: 1,
					},
				},
			},
			engine.Snake{
				ID:    "id2",
				Color: "FFFFFF",
				Body: []engine.Point{
					engine.Point{
						X: 2, Y: 2,
					},
				},
			},
			engine.Snake{
				ID: "id3",
				Death: engine.DeathCause{
					Cause: "DOA",
				},
				Color: "FF0000",
				Body: []engine.Point{
					engine.Point{
						X: 2, Y: 0,
					},
				},
			},
		},
	}
}
