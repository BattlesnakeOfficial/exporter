package main

import (
	"fmt"
	"image"
	"log"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/render"
	"github.com/fogleman/gg"
)

func main() {
	shift := func(p engine.Point, x, y int) engine.Point {
		nP := engine.Point{X: p.X + x, Y: p.Y + y}

		// wrap around (assuming 3x3 board here)
		if nP.X < 0 {
			nP.X += 3
		}
		if nP.Y < 0 {
			nP.Y += 2
		}
		if nP.X > 2 {
			nP.X -= 3
		}
		if nP.Y > 2 {
			nP.Y -= 3
		}
		return nP
	}

	dc := gg.NewContext(400, 2500)
	drawY := 20
	i := 0
	for _, x := range []int{-2, -1, 0, 1, 2} {
		for _, y := range []int{-2, -1, 0, 1, 2} {
			i += 4
			fmt.Printf("%d,%d,%d\n", x, y, i)
			drawX := 20

			// ╔
			img := drawSnake(shift(engine.Point{X: 0, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 1}, x, y))
			dc.DrawImage(img, drawX, drawY)
			drawX += 100
			// ╗
			img = drawSnake(shift(engine.Point{X: 0, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 0}, x, y))
			dc.DrawImage(img, drawX, drawY)
			drawX += 100
			// ╝
			img = drawSnake(shift(engine.Point{X: 1, Y: 1}, x, y), shift(engine.Point{X: 1, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 0}, x, y))
			dc.DrawImage(img, drawX, drawY)
			drawX += 100
			// ╚
			img = drawSnake(shift(engine.Point{X: 1, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 0}, x, y), shift(engine.Point{X: 0, Y: 1}, x, y))
			dc.DrawImage(img, drawX, drawY)
			drawY += 100
		}
	}
	err := dc.SavePNG("corners.png")
	if err != nil {
		log.Fatalf("Unable to save PNG: %v\n", err)
	}
}

func drawSnake(p1, p2, p3 engine.Point) image.Image {
	b := render.NewBoard(3, 3)
	b.PlaceSnake(engine.Snake{
		Color: "#000000",
		Body:  []engine.Point{p1, p2, p3},
	})
	return render.DrawBoard(b)
}
