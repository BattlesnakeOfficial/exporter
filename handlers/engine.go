package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	engine "github.com/battlesnakeio/exporter/engine"
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
func GetGameFrames(gameID string, offset int) (*engine.ListGameFramesResponse, error) {
	return GetGameFramesWithLength(gameID, offset, 1)
}

// GetGameFramesWithLength returns a game frame object with length frames
func GetGameFramesWithLength(gameID string, offset, length int) (*engine.ListGameFramesResponse, error) {
	url := fmt.Sprintf("https://engine.battlesnake.io/games/%s/frames?offset=%d&limit=%d", gameID, offset, length)
	body, err := MakeEngineCall(url)
	if err != nil {
		return nil, err
	}

	var gameFrames *engine.ListGameFramesResponse
	if err := json.Unmarshal(body, &gameFrames); err != nil {
		return nil, err
	}
	return gameFrames, nil
}

// GetGameStatus returns a game status object from the engine.
func GetGameStatus(gameID string) (*engine.StatusResponse, error) {
	body, err := MakeEngineCall(fmt.Sprintf("https://engine.battlesnake.io/games/%s", gameID))
	if err != nil {
		return nil, err
	}

	var gameStatus *engine.StatusResponse
	if err := json.Unmarshal(body, &gameStatus); err != nil {
		return nil, err
	}
	return gameStatus, nil
}
