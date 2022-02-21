package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	"github.com/BattlesnakeOfficial/exporter/media"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

type snakeImageType int

const (
	snakeHead snakeImageType = iota
	snakeTail
)

type rotations int

const (
	rotate0 rotations = iota
	rotate90
	rotate180
	rotate270
)

const (
	BoardBorder        = 2
	SquareSizePixels   = 20
	SquareBorderPixels = 1
	SquareFoodRadius   = SquareSizePixels / 3
	ColorEmptySquare   = "#f0f0f0"
	ColorFood          = "#ff5c75"
	ColorHazard        = "#00000066"
)

// cache for storing image.Image objects to speed up rendering
var imageCache = cache.New(6*time.Hour, 10*time.Minute)

// From github.com/fogleman/gg
func parseHexColor(x string) color.Color {
	var r, g, b, a uint8

	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return color.RGBA{r, g, b, a}
}

func rotateImage(src image.Image, rot rotations) image.Image {
	switch rot {
	case rotate90:
		return imaging.Rotate90(src)
	case rotate180:
		return imaging.FlipH(src)
	case rotate270:
		return imaging.Rotate270(src)
	}
	return src
}

func drawWatermark(dc *gg.Context) {
	watermarkImage, err := media.GetWatermarkPNG(dc.Width()*2/3, dc.Height()*2/3)
	if err != nil {
		log.WithError(err).Error("Unable to load watermark image")
		return
	}
	dc.DrawImageAnchored(watermarkImage, dc.Width()/2, dc.Height()/2, 0.5, 0.5)
}

func drawEmptySquare(dc *gg.Context, bx int, by int) {
	dc.SetHexColor(ColorEmptySquare)
	dc.DrawRectangle(
		boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
		boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
		SquareSizePixels-SquareBorderPixels*2,
		SquareSizePixels-SquareBorderPixels*2,
	)
	dc.Fill()
}

func drawFood(dc *gg.Context, bx int, by int) {
	dc.SetHexColor(ColorFood)
	dc.DrawCircle(
		boardXToDrawX(dc, bx)+SquareSizePixels/2+BoardBorder,
		boardYToDrawY(dc, by)+SquareSizePixels/2+BoardBorder,
		SquareFoodRadius,
	)
	dc.Fill()
}

func drawHazard(dc *gg.Context, bx int, by int) {
	dc.SetHexColor(ColorHazard)
	dc.DrawRectangle(
		boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
		boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
		SquareSizePixels-SquareBorderPixels*2,
		SquareSizePixels-SquareBorderPixels*2,
	)
	dc.Fill()
}

func drawSnakeImage(name string, st snakeImageType, dc *gg.Context, bx int, by int, hexColor string, dir snakeDirection) {

	width := SquareSizePixels - SquareBorderPixels*2
	height := SquareSizePixels - SquareBorderPixels*2

	var snakeImg image.Image
	var err error
	switch st {
	case snakeHead:
		snakeImg, err = media.GetHeadPNG(name, width, height, hexColor)
	case snakeTail:
		snakeImg, err = media.GetTailPNG(name, width, height, hexColor)
	default:
		log.WithField("snakeImageType", st).Error("unable to draw an unrecognized snake image type")
	}

	if err != nil {
		log.WithError(err).Error("Unable to get snake image - aborting draw")
		return
	}

	var rot rotations
	switch dir {
	case movingRight:
		rot = rotate0
	case movingDown:
		rot = rotate270
	case movingLeft:
		rot = rotate180
	case movingUp:
		rot = rotate90
	}
	snakeImg = rotateImage(snakeImg, rot)

	// dst := dc.Image().(draw.Image)
	// dstRect := image.Rect(
	// 	int(boardXToDrawX(dc, bx))+SquareBorderPixels+BoardBorder,
	// 	int(boardYToDrawY(dc, by))+SquareBorderPixels+BoardBorder,
	// 	int(boardXToDrawX(dc, bx+1))-SquareBorderPixels+BoardBorder,
	// 	int(boardYToDrawY(dc, by-1))-SquareBorderPixels+BoardBorder,
	// )

	// srcImage := &image.Uniform{parseHexColor(hexColor)}

	// draw.DrawMask(dst, dstRect, srcImage, image.Point{}, snakeImg, image.Point{}, draw.Over)
	dx := int(boardXToDrawX(dc, bx)) + SquareBorderPixels + BoardBorder
	dy := int(boardYToDrawY(dc, by)) + SquareBorderPixels + BoardBorder
	dc.DrawImage(snakeImg, dx, dy)
}

func drawSnakeBody(dc *gg.Context, bx int, by int, hexColor string, corner snakeCorner) {
	dc.SetHexColor(hexColor)
	if corner == "none" {
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareSizePixels-SquareBorderPixels*2,
			SquareSizePixels-SquareBorderPixels*2,
		)
	} else {
		dc.DrawRoundedRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareSizePixels-SquareBorderPixels*2,
			SquareSizePixels-SquareBorderPixels*2,
			SquareSizePixels/2,
		)
		if corner.isBottom() {
			dc.DrawRectangle(
				boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
				boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
				SquareSizePixels-SquareBorderPixels*2,
				SquareSizePixels/2,
			)
			if corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+SquareSizePixels/2+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+SquareSizePixels/2+BoardBorder,
					SquareSizePixels/2-SquareBorderPixels,
					SquareSizePixels/2-SquareBorderPixels*2,
				)
			}
			if !corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+SquareSizePixels/2+BoardBorder,
					SquareSizePixels/2-SquareBorderPixels*2,
					SquareSizePixels/2-SquareBorderPixels*2,
				)
			}
		}
		if !corner.isBottom() {
			dc.DrawRectangle(
				boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
				boardYToDrawY(dc, by)+SquareBorderPixels+SquareSizePixels/2+BoardBorder,
				SquareSizePixels-SquareBorderPixels*2,
				SquareSizePixels/2,
			)
			if corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+SquareSizePixels/2+SquareBorderPixels+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
					SquareSizePixels/2-SquareBorderPixels*2,
					SquareSizePixels/2-SquareBorderPixels*2,
				)
			}
			if !corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
					SquareSizePixels/2-SquareBorderPixels*2,
					SquareSizePixels/2-SquareBorderPixels*2,
				)
			}
		}
	}
	dc.Fill()
}

func drawGaps(dc *gg.Context, bx, by int, dir snakeDirection, hexColor string) {
	dc.SetHexColor(hexColor)
	switch dir {
	case movingUp:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by-1)-SquareBorderPixels+BoardBorder,
			SquareSizePixels-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case movingDown:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)-SquareBorderPixels+BoardBorder,
			SquareSizePixels-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case movingRight:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			SquareSizePixels-SquareBorderPixels*2,
		)
	case movingLeft:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx+1)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			SquareSizePixels-SquareBorderPixels*2,
		)
	}
	dc.Fill()
}

func createBoardContext(b *Board) *gg.Context {
	dc := gg.NewContext(
		SquareSizePixels*b.Width+BoardBorder*2,
		SquareSizePixels*b.Height+BoardBorder*2,
	)

	cacheKey := fmt.Sprintf("board:%d:%d", b.Width, b.Height)
	cachedBoardImage, ok := imageCache.Get(cacheKey)
	if ok {
		dc.DrawImage(cachedBoardImage.(image.Image), 0, 0)
		return dc
	}

	// Clear to white
	dc.SetColor(color.White)
	dc.Clear()

	// Draw empty squares
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			drawEmptySquare(dc, x, y)
		}
	}

	// Draw watermark
	drawWatermark(dc)

	// Cache for next time
	cacheDC := gg.NewContext(dc.Width(), dc.Height())
	cacheDC.DrawImage(dc.Image(), 0, 0)
	imageCache.Set(cacheKey, cacheDC.Image(), cache.DefaultExpiration)

	return dc
}

func DrawBoard(b *Board) image.Image {
	dc := createBoardContext(b)

	// Draw food and snakes over watermark
	for p, s := range b.squares { // cool, we can iterate ONLY the non-empty squares!
		for _, c := range s.Contents {
			switch c.Type {
			case BoardSquareSnakeHead:
				drawSnakeImage(c.SnakeType, snakeHead, dc, p.X, p.Y, c.HexColor, c.Direction)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, p.X, p.Y, c.HexColor, c.Corner)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeTail:
				drawSnakeImage(c.SnakeType, snakeTail, dc, p.X, p.Y, c.HexColor, c.Direction)
			case BoardSquareFood:
				drawFood(dc, p.X, p.Y)
			case BoardSquareHazard:
				drawHazard(dc, p.X, p.Y)
			}
		}

	}

	return dc.Image()
}

// boardXToDrawX converts an x coordinate in "board space" to the x coordinate used by graphics.
// More specifically, it assumes the board coordinates are the indexes of squares and it returns the upper left
// corner for that square.
func boardXToDrawX(dc *gg.Context, x int) float64 {
	return float64(x * SquareSizePixels)
}

// boardYToDrawY converts a y coordinate in "board space" to the y coordinate used by graphics.
// More specifically, it assumes the board coordinates are the indexes of squares and it returns the upper left
// corner for that square.
func boardYToDrawY(dc *gg.Context, y int) float64 {
	// Note: the Battlesnake board coordinates have (0,0) at the bottom left
	// so we need to flip the y-axis to convert to the graphics, which follows the convention
	// of (0,0) being the top left.
	return float64((dc.Height() - BoardBorder*2 - SquareSizePixels) - (y * SquareSizePixels)) // flip!
}
