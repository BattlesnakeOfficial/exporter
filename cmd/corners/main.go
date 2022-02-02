package main

import (
	"log"

	"github.com/BattlesnakeOfficial/exporter/render"
	"github.com/fogleman/gg"
)

func main() {
	dc := gg.NewContext(100, 100)
	b := render.NewBoard(3, 3)
	// b.AddSnakeHead(&engine.Point{X: 0, Y: 2}, "", "", up)
	// b.AddSnakeBody(&engine.Point{X: 0, Y: 1}, "", "", up)
	// b.AddSnakeTail(&engine.Point{X: 2, Y: 1}, "", "", left)
	dc.DrawImage(render.DrawBoard(b), 0, 0)
	err := dc.SavePNG("corners.png")
	if err != nil {
		log.Fatalf("Unable to save PNG: %v\n", err)
	}
}
