package handlers

import (
	"encoding/json"
	"testing"

	openapi "github.com/battlesnakeio/exporter/model"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestBadUrls(t *testing.T) {
	rr := serveURL("output=aoeu")
	assert.Equal(t, 404, rr.Code)
	rr = serveURL("output=move")
	assert.Equal(t, 404, rr.Code)
	rr = serveURL("youId=id1")
	assert.Equal(t, 404, rr.Code)
	rr = serveURL("")
	assert.Equal(t, 404, rr.Code)
}

func TestGetMove(t *testing.T) {
	defer gock.Off()
	frameList, _ := json.Marshal(createFrameList())
	GockFrame(string(frameList))

	gameStatus, _ := json.Marshal(openapi.EngineStatusResponse{
		Game: openapi.EngineGame{
			Height: 2,
			Width:  2,
		},
	})
	GockStatus(string(gameStatus))
	rr := serveURL("output=move&youId=1")
	assert.Equal(t, "{\"game\":{},\"board\":{\"height\":2,\"width\":2,\"snakes\":[{\"id\":\"1\",\"body\":[{\"y\":1}]}]},\"you\":{\"id\":\"1\",\"body\":[{\"y\":1}]}}", rr.Body.String())
}
func TestGetBoard(t *testing.T) {
	defer gock.Off()
	frameList, _ := json.Marshal(createFrameList())
	GockFrame(string(frameList))

	gameStatus, _ := json.Marshal(openapi.EngineStatusResponse{
		Game: openapi.EngineGame{
			Height: 2,
			Width:  2,
		},
	})
	GockStatus(string(gameStatus))

	rr := serveURL("output=board")
	assert.Equal(t, "------\n|    |\n|H1  |\n------\n", rr.Body.String())
}

func TestGetBoardAnimated(t *testing.T) {
	defer gock.Off()
	frameList, _ := json.Marshal(createFrameList())
	GockFrame(string(frameList))

	gameStatus, _ := json.Marshal(openapi.EngineStatusResponse{
		Game: openapi.EngineGame{
			Height: 2,
			Width:  2,
		},
	})
	GockStatus(string(gameStatus))

	rr := serveURL("output=board-animated")
	assert.Equal(t, "<html><head></head><body><pre>"+
		"------\n"+
		"|    |\n"+
		"|H1  |\n"+
		"------\n"+
		"</pre></body><script>window.location.assign(window.location.href.replace(/0\\?/g, '1?'));</script></html>", rr.Body.String())
}

func TestGetFrameWithTurn(t *testing.T) {
	defer gock.Off()
	GockFrame("{ \"Frames\": [ { \"Turn\": 5 }] }")
	rr := serveURL("output=raw")
	assert.Equal(t, "{\"Turn\":5}", rr.Body.String())
}

func TestNoFrames(t *testing.T) {
	defer gock.Off()
	GockFrame("{ \"Frames\": [ ] }")
	rr := serveURL("output=raw")
	assert.Equal(t, 404, rr.Code)
	assert.Equal(t, "No frames", rr.Body.String())
}

func createFrameList() *openapi.EngineListGameFramesResponse {
	return &openapi.EngineListGameFramesResponse{
		Frames: []openapi.EngineGameFrame{
			openapi.EngineGameFrame{
				Snakes: []openapi.EngineSnake{
					openapi.EngineSnake{
						ID: "1",
						Body: []openapi.EnginePoint{
							openapi.EnginePoint{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
		},
	}
}
