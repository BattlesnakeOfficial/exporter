package render

import (
	"image"
	"os"

	"github.com/fogleman/gg"
)

const (
	SquareSizePixels   = 60
	SquareBorderPixels = 3
	SquareFoodRadius   = SquareSizePixels / 4
)

func drawWatermark(dc *gg.Context) {
	f, err := os.Open("render/assets/watermark.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	watermarkImage, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	dc.DrawImageAnchored(watermarkImage, dc.Width()/2, dc.Height()/2, 0.5, 0.5)
}

func drawEmptySquare(dc *gg.Context, x int, y int) {
	dc.SetRGB255(240, 240, 240)
	dc.DrawRectangle(
		float64(x*SquareSizePixels+SquareBorderPixels),
		float64(y*SquareSizePixels+SquareBorderPixels),
		float64(SquareSizePixels-SquareBorderPixels*2),
		float64(SquareSizePixels-SquareBorderPixels*2),
	)
	dc.Fill()
}

func drawFood(dc *gg.Context, x int, y int) {
	dc.SetRGB255(255, 92, 117)
	dc.DrawCircle(
		float64(x*SquareSizePixels+SquareSizePixels/2),
		float64(y*SquareSizePixels+SquareSizePixels/2),
		SquareFoodRadius,
	)
	dc.Fill()
}

func drawSnakeBody(dc *gg.Context, x int, y int, hexColor string) {
	dc.SetHexColor(hexColor)
	dc.DrawRectangle(
		float64(x*SquareSizePixels+SquareBorderPixels),
		float64(y*SquareSizePixels+SquareBorderPixels),
		float64(SquareSizePixels-SquareBorderPixels*2),
		float64(SquareSizePixels-SquareBorderPixels*2),
	)
	dc.Fill()
}

func drawBoard(b *Board) image.Image {
	dc := gg.NewContext(SquareSizePixels*b.Width, SquareSizePixels*b.Height)

	dc.SetRGB255(255, 255, 255)
	dc.Clear()

	// Draw empty squares under watermark
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			switch b.Squares[x][y].Content {
			case BoardSquareFood:
				drawEmptySquare(dc, x, y)
			case BoardSquareEmpty:
				drawEmptySquare(dc, x, y)
			}
		}
	}

	drawWatermark(dc)

	// Draw food and snakes over watermark
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			switch b.Squares[x][y].Content {
			case BoardSquareSnakeHead:
				drawSnakeBody(dc, x, y, b.Squares[x][y].HexColor)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, x, y, b.Squares[x][y].HexColor)
			case BoardSquareSnakeTail:
				drawSnakeBody(dc, x, y, b.Squares[x][y].HexColor)
			case BoardSquareFood:
				drawFood(dc, x, y)
			}
		}
	}

	return dc.Image()
}
