package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func apiCall(path string) ([]byte, error) {
	url := fmt.Sprintf("https://engine.battlesnake.io/%s", path)
	client := http.Client{}

	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Got non 200 from engine: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getFrames(gameID string, offset int, limit int) ([]*GameFrame, error) {
	path := fmt.Sprintf("/games/%s/frames?offset=%d&limit=%d", gameID, offset, limit)
	body, err := apiCall(path)
	if err != nil {
		return nil, err
	}

	response := gameFramesResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Frames, nil
}

// GetGame is commented
func GetGame(gameID string) (*Game, error) {
	path := fmt.Sprintf("/games/%s", gameID)
	body, err := apiCall(path)
	if err != nil {
		return nil, err
	}

	response := gameStatusResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Game, nil
}

// GetGameFrame is commented
func GetGameFrame(gameID string, frameNum int) (*GameFrame, error) {
	gameFrames, err := getFrames(gameID, frameNum, 1)
	if err != nil {
		return nil, err
	}

	return gameFrames[0], nil
}

// GetGameFrames is commented
func GetGameFrames(gameID string) ([]*GameFrame, error) {
	var gameFrames []*GameFrame

	batchSize := 100
	for {
		newFrames, err := getFrames(gameID, len(gameFrames), batchSize)
		if err != nil {
			return nil, err
		}

		gameFrames = append(gameFrames, newFrames...)
		if len(newFrames) < batchSize {
			break
		}
	}

	return gameFrames, nil
}

// // EngineURL External URL of engine
// const EngineURL = "https://engine.battlesnake.io"

// // MakeEngineCall returns a by array from an engine call.
// func MakeEngineCall(url string) ([]byte, error) {
// 	netClient := &http.Client{}
// 	getResponse, err := netClient.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	body, err := ioutil.ReadAll(getResponse.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if getResponse.StatusCode != 200 {
// 		return nil, fmt.Errorf("Got non 200 response code: %d, message: %s", getResponse.StatusCode, string(body))
// 	}
// 	return body, nil
// }

// // GetGameFrames returns a game frame object
// func GetGameFrames(gameID string, offset int) (*engine.ListGameFramesResponse, error) {
// 	return GetGameFramesWithLength(gameID, offset, 1)
// }

// // GetGameFramesWithLength returns a game frame object with length frames
// func GetGameFramesWithLength(gameID string, offset, length int) (*engine.ListGameFramesResponse, error) {
// 	url := fmt.Sprintf("https://engine.battlesnake.io/games/%s/frames?offset=%d&limit=%d", gameID, offset, length)
// 	body, err := MakeEngineCall(url)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var gameFrames *engine.ListGameFramesResponse
// 	if err := json.Unmarshal(body, &gameFrames); err != nil {
// 		return nil, err
// 	}
// 	return gameFrames, nil
// }

// // GetGameStatus returns a game status object from the engine.
// func GetGameStatus(gameID string) (*engine.StatusResponse, error) {
// 	body, err := MakeEngineCall(fmt.Sprintf("https://engine.battlesnake.io/games/%s", gameID))
// 	if err != nil {
// 		return nil, err
// 	}

// 	var gameStatus *engine.StatusResponse
// 	if err := json.Unmarshal(body, &gameStatus); err != nil {
// 		return nil, err
// 	}
// 	return gameStatus, nil
// }
