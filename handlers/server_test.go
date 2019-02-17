package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	engine "github.com/battlesnakeio/exporter/engine"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestGetGif(t *testing.T) {
	defer gock.Off()
	frameList, _ := json.Marshal(createFrameList())
	frameList5, _ := json.Marshal(createFrameList5())
	gameStatus, _ := json.Marshal(createGameStatus(3, 3))
	Gock15Frames(string(frameList5), string(frameList))
	GockStatus(string(gameStatus))
	router, rr := initialize()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/games/%s?output=gif&batchSize=5", GameID), nil)
	router.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	if rr.Code != 200 {
		fmt.Println(rr.Body.String())
	}
	assert.True(t, rr.Body.Len() > 0)
}
func TestGetPNG(t *testing.T) {
	defer gock.Off()
	frameList, _ := json.Marshal(createFrameList())
	gameStatus, _ := json.Marshal(createGameStatus(3, 3))
	GockFrame(string(frameList))
	GockStatus(string(gameStatus))
	rr := serveURL("output=png")
	assert.Equal(t, 200, rr.Code)
	assert.True(t, rr.Body.Len() > 0)
}
func TestBadURLs(t *testing.T) {
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

	gameStatus, _ := json.Marshal(engine.StatusResponse{
		Game: engine.Game{
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

	gameStatus, _ := json.Marshal(engine.StatusResponse{
		Game: engine.Game{
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

	gameStatus, _ := json.Marshal(engine.StatusResponse{
		Game: engine.Game{
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

func createFrameList() *engine.ListGameFramesResponse {
	return &engine.ListGameFramesResponse{
		Frames: []engine.GameFrame{
			engine.GameFrame{
				Snakes: []engine.Snake{
					engine.Snake{
						ID: "1",
						Body: []engine.Point{
							engine.Point{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
		},
	}
}

func createFrameList5() *engine.ListGameFramesResponse {
	return &engine.ListGameFramesResponse{
		Frames: []engine.GameFrame{
			engine.GameFrame{
				Snakes: []engine.Snake{
					engine.Snake{
						ID: "1",
						Body: []engine.Point{
							engine.Point{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
			engine.GameFrame{
				Snakes: []engine.Snake{
					engine.Snake{
						ID: "1",
						Body: []engine.Point{
							engine.Point{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
			engine.GameFrame{
				Snakes: []engine.Snake{
					engine.Snake{
						ID: "1",
						Body: []engine.Point{
							engine.Point{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
			engine.GameFrame{
				Snakes: []engine.Snake{
					engine.Snake{
						ID: "1",
						Body: []engine.Point{
							engine.Point{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
			engine.GameFrame{
				Snakes: []engine.Snake{
					engine.Snake{
						ID: "1",
						Body: []engine.Point{
							engine.Point{
								X: 0, Y: 1,
							},
						},
					},
				},
			},
		},
	}
}
