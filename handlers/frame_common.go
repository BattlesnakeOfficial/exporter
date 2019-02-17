package handlers

import engine "github.com/battlesnakeio/exporter/engine"

func getDimensions(gameStatus *engine.StatusResponse) (int, int) {
	width := int(gameStatus.Game.Width)
	height := int(gameStatus.Game.Height)
	return width, height
}
