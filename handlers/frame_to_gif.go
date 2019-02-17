package handlers

import (
	"fmt"
	"io"

	"github.com/fogleman/gg"

	openapi "github.com/battlesnakeio/exporter/model"
)

// ConvertFrameToPng takes a frame and makes a png
func ConvertFrameToPng(w io.Writer, gameFrame *openapi.EngineGameFrame, gameStatus *openapi.EngineStatusResponse) {
	width, height := getDimensions(gameStatus)
	square := int32(20)
	filled := make(map[string]bool)
	dc := gg.NewContext(width*int(square)+2, height*int(square)+2)
	dc.DrawRectangle(0, 0, float64(width*int(square)+2), float64(height*int(square)+2))
	dc.SetHexColor("#000000")
	dc.Fill()
	for _, snake := range gameFrame.Snakes {
		for i, point := range snake.Body {
			filled[fmt.Sprintf("%d,%d", point.X, point.Y)] = true
			transparancy := "AA"
			if i == 0 {
				transparancy = "FF"
			}
			color := snake.Color + transparancy
			if snake.Death.Cause != "" {
				color = "#555555" + transparancy
			}
			dc.DrawRectangle(float64(point.X*square)+2, float64(point.Y*square)+2, float64(square)-2, float64(square)-2)
			dc.SetHexColor(color)
			dc.Fill()
		}
	}

	for _, point := range gameFrame.Food {
		filled[fmt.Sprintf("%d,%d", point.X, point.Y)] = true
		dc.DrawRectangle(float64(point.X*square)+2, float64(point.Y*square)+2, float64(square)-2, float64(square)-2)
		dc.SetHexColor("#111111")
		dc.Fill()
		dc.DrawCircle(float64(point.X*square)+float64(square)/2.0+1, float64(point.Y*square+square/2.0)+1, float64(square)/float64(4))
		dc.SetHexColor("#FFA500")
		dc.Fill()
	}
	for x := int32(0); x < int32(width); x++ {
		for y := int32(0); y < int32(height); y++ {
			if !filled[fmt.Sprintf("%d,%d", x, y)] {
				dc.DrawRectangle(float64(x*square)+2, float64(y*square)+2, float64(square)-2, float64(square)-2)
				dc.SetHexColor("#111111")
				dc.Fill()
			}
		}
	}

	dc.EncodePNG(w)
}
