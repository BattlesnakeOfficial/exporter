package handlers

import openapi "github.com/battlesnakeio/exporter/model"

func getDimensions(gameStatus *openapi.EngineStatusResponse) (int, int) {
	width := int(gameStatus.Game.Width)
	height := int(gameStatus.Game.Height)
	return width, height
}
