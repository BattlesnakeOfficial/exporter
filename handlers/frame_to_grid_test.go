package handlers

import (
	"encoding/json"
	"testing"

	openapi "github.com/battlesnakeio/exporter/model"
	"github.com/stretchr/testify/assert"
)

func TestConvertGridToString(t *testing.T) {
	frame := createFrame()
	frame.Food = []openapi.EnginePoint{
		openapi.EnginePoint{X: 0, Y: 2},
	}
	grid := ConvertFrameToGrid(&frame, createGameStatus(3, 3))
	board := ConvertGridToString(grid)
	assert.Equal(t,
		""+
			"--------\n"+
			"|H1B1  |\n"+
			"|  T1  |\n"+
			"|FF  H3|\n"+
			"--------\n", board)
}
func TestConvertFrameToMove(t *testing.T) {
	frame := createFrame()
	frame.Food = []openapi.EnginePoint{
		openapi.EnginePoint{X: 0, Y: 2},
	}
	gameStatus := createGameStatus(2, 2)

	move, _ := ConvertFrameToMove(&frame, gameStatus, "id2")
	json, _ := json.Marshal(move)
	assert.Equal(t, "{\"game\":{\"id\":\"GameID\"},\"board\":{\"height\":2,\"width\":2,\"food\":[{\"y\":2}],\"snakes\":[{\"id\":\"id1\",\"body\":[{},{\"x\":1},{\"x\":1,\"y\":1}]},{\"id\":\"id2\",\"body\":[{\"x\":2,\"y\":2}]}]},\"you\":{\"id\":\"id2\",\"body\":[{\"x\":2,\"y\":2}]}}", string(json))
}

func createGameStatus(width int, height int) *openapi.EngineStatusResponse {
	return &openapi.EngineStatusResponse{
		Game: openapi.EngineGame{
			Height: int32(height),
			Width:  int32(width),
			ID:     "GameID",
		},
	}
}
func TestFrameToGridOneSnake(t *testing.T) {
	frame := &openapi.EngineGameFrame{
		Snakes: []openapi.EngineSnake{
			openapi.EngineSnake{
				ID:    "id1",
				Color: "6611FF",
				Body: []openapi.EnginePoint{
					openapi.EnginePoint{
						X: 0, Y: 0,
					},
				},
			},
		},
	}
	grid := ConvertFrameToGrid(frame, createGameStatus(2, 2))
	assert.Equal(t, [][]Pixel{
		{Pixel{ID: "id1", Colour: "6611FF", PixelType: "H"}, Pixel{PixelType: Space}},
		{Pixel{PixelType: Space}, Pixel{PixelType: Space}}}, grid)
}

func TestFrameToGridOneSnakeAndFood(t *testing.T) {
	frame := &openapi.EngineGameFrame{
		Food: []openapi.EnginePoint{
			openapi.EnginePoint{
				X: 1, Y: 1,
			},
		},
		Snakes: []openapi.EngineSnake{
			openapi.EngineSnake{
				ID:    "id1",
				Color: "6611FF",
				Body: []openapi.EnginePoint{
					openapi.EnginePoint{
						X: 0, Y: 0,
					},
				},
			},
		},
	}
	grid := ConvertFrameToGrid(frame, createGameStatus(2, 2))
	assert.Equal(t, [][]Pixel{
		{Pixel{ID: "id1", Colour: "6611FF", PixelType: Head}, Pixel{PixelType: Space}},
		{Pixel{PixelType: Space}, Pixel{PixelType: Food}}}, grid)
}

func TestFrameToGridEmpty(t *testing.T) {
	frame := &openapi.EngineGameFrame{}
	grid := ConvertFrameToGrid(frame, createGameStatus(1, 1))
	assert.Equal(t, [][]Pixel{
		{Pixel{PixelType: Space}}}, grid)
}

func createFrame() openapi.EngineGameFrame {
	return openapi.EngineGameFrame{
		Snakes: []openapi.EngineSnake{
			openapi.EngineSnake{
				ID:    "id1",
				Color: "6611FF",
				Body: []openapi.EnginePoint{
					openapi.EnginePoint{
						X: 0, Y: 0,
					},
					openapi.EnginePoint{
						X: 1, Y: 0,
					},
					openapi.EnginePoint{
						X: 1, Y: 1,
					},
				},
			},
			openapi.EngineSnake{
				ID:    "id2",
				Color: "FFFFFF",
				Body: []openapi.EnginePoint{
					openapi.EnginePoint{
						X: 2, Y: 2,
					},
				},
			},
			openapi.EngineSnake{
				ID: "id3",
				Death: openapi.EngineDeathCause{
					Cause: "DOA",
				},
				Color: "FF0000",
				Body: []openapi.EnginePoint{
					openapi.EnginePoint{
						X: 2, Y: 0,
					},
				},
			},
		},
	}
}
