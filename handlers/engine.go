package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	openapi "github.com/battlesnakeio/exporter/model"
)

// EngineURL External URL of engine
const EngineURL = "https://engine.battlesnake.io"

// MakeEngineCall returns a by array from an engine call.
func MakeEngineCall(url string) ([]byte, error) {
	netClient := &http.Client{}
	getResponse, err := netClient.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(getResponse.Body)
	if err != nil {
		return nil, err
	}

	if getResponse.StatusCode != 200 {
		return nil, fmt.Errorf("Got non 200 response code: %d, message: %s", getResponse.StatusCode, string(body))
	}
	return body, nil
}

// GetGameFrames returns a game frame object
func GetGameFrames(gameID string, offset string) (*openapi.EngineListGameFramesResponse, error) {
	body, err := MakeEngineCall(fmt.Sprintf("https://engine.battlesnake.io/games/%s/frames?offset=%s&limit=1", gameID, offset))
	if err != nil {
		return nil, err
	}

	var gameFrames *openapi.EngineListGameFramesResponse
	if err := json.Unmarshal(body, &gameFrames); err != nil {
		return nil, err
	}
	return gameFrames, nil
}

// GetGameStatus returns a game status object from the engine.
func GetGameStatus(gameID string) (*openapi.EngineStatusResponse, error) {
	body, err := MakeEngineCall(fmt.Sprintf("https://engine.battlesnake.io/games/%s", gameID))
	if err != nil {
		return nil, err
	}

	var gameStatus *openapi.EngineStatusResponse
	if err := json.Unmarshal(body, &gameStatus); err != nil {
		return nil, err
	}
	return gameStatus, nil
}
