package render

import (
	"fmt"
	"image"
	"image/color"
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
	BoardBorder        float64 = 2
	SquareBorderPixels float64 = 1
	ColorEmptySquare           = "#f0f0f0"
	ColorFood                  = "#ff5c75"
	ColorHazard                = "#00000066"
)

type boardContext struct {
	*gg.Context
	squareSizePixels int
	squareSizeHalfPx float64
	squareFoodRadius float64
}

// func (dc *boardContext)

// cache for storing image.Image objects to speed up rendering
var imageCache = cache.New(6*time.Hour, 10*time.Minute)

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

func drawWatermark(dc *boardContext) {
	watermarkImage, err := media.GetWatermarkPNG(dc.Width()*2/3, dc.Height()*2/3)
	if err != nil {
		log.WithError(err).Error("Unable to load watermark image")
		return
	}
	dc.DrawImageAnchored(watermarkImage, dc.Width()/2, dc.Height()/2, 0.5, 0.5)
}

func drawEmptySquare(dc *boardContext, bx int, by int) {
	dc.SetHexColor(ColorEmptySquare)
	dc.DrawRectangle(
		boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
		boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
		float64(dc.squareSizePixels)-SquareBorderPixels*2,
		float64(dc.squareSizePixels)-SquareBorderPixels*2,
	)
	dc.Fill()
}

func drawFood(dc *boardContext, bx int, by int) {
	dc.SetHexColor(ColorFood)
	dc.DrawCircle(
		boardXToDrawX(dc, bx)+dc.squareSizeHalfPx+BoardBorder,
		boardYToDrawY(dc, by)+dc.squareSizeHalfPx+BoardBorder,
		dc.squareFoodRadius,
	)
	dc.Fill()
}

func drawHazard(dc *boardContext, bx int, by int) {
	dc.SetHexColor(ColorHazard)
	dc.DrawRectangle(
		boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
		boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
		float64(dc.squareSizePixels)-SquareBorderPixels*2,
		float64(dc.squareSizePixels)-SquareBorderPixels*2,
	)
	dc.Fill()
}

func drawSnakeImage(name string, st snakeImageType, dc *boardContext, bx int, by int, c color.Color, dir snakeDirection) {

	width := dc.squareSizePixels - int(SquareBorderPixels*2)
	height := dc.squareSizePixels - int(SquareBorderPixels*2)

	var snakeImg image.Image
	var err error
	switch st {
	case snakeHead:
		snakeImg, err = media.GetHeadPNG(name, width, height, c)
	case snakeTail:
		snakeImg, err = media.GetTailPNG(name, width, height, c)
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

	dx := int(boardXToDrawX(dc, bx) + SquareBorderPixels + BoardBorder)
	dy := int(boardYToDrawY(dc, by) + SquareBorderPixels + BoardBorder)
	dc.DrawImage(snakeImg, dx, dy)
}

func drawSnakeBody(dc *boardContext, bx int, by int, c color.Color, corner snakeCorner) {
	dc.SetColor(c)
	if corner == "none" {
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
		)
	} else {
		dc.DrawRoundedRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
			dc.squareSizeHalfPx,
		)
		if corner.isBottom() {
			dc.DrawRectangle(
				boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
				boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
				float64(dc.squareSizePixels)-SquareBorderPixels*2,
				dc.squareSizeHalfPx,
			)
			if corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+dc.squareSizeHalfPx+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+dc.squareSizeHalfPx+BoardBorder,
					dc.squareSizeHalfPx-SquareBorderPixels,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
				)
			}
			if !corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+dc.squareSizeHalfPx+BoardBorder,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
				)
			}
		}
		if !corner.isBottom() {
			dc.DrawRectangle(
				boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
				boardYToDrawY(dc, by)+SquareBorderPixels+dc.squareSizeHalfPx+BoardBorder,
				float64(dc.squareSizePixels)-SquareBorderPixels*2,
				dc.squareSizeHalfPx,
			)
			if corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+dc.squareSizeHalfPx+SquareBorderPixels+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
				)
			}
			if !corner.isLeft() {
				dc.DrawRectangle(
					boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
					boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
					dc.squareSizeHalfPx-SquareBorderPixels*2,
				)
			}
		}
	}
	dc.Fill()
}

func drawGaps(dc *boardContext, bx, by int, dir snakeDirection, c color.Color) {
	dc.SetColor(c)
	switch dir {
	case movingUp:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by-1)-SquareBorderPixels+BoardBorder,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case movingDown:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)-SquareBorderPixels+BoardBorder,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case movingRight:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
		)
	case movingLeft:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx+1)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			float64(dc.squareSizePixels)-SquareBorderPixels*2,
		)
	}
	dc.Fill()
}

func createBoardContext(b *Board, w, h int) *boardContext {
	ss := calcSquarePx(w, h, b.Width, b.Height)
	dc := &boardContext{
		Context:          gg.NewContext(w, h),
		squareSizePixels: ss,
		squareSizeHalfPx: float64(ss) / 2,
		squareFoodRadius: float64(ss) / 3,
	}

	cacheKey := fmt.Sprintf("board:%d:%d:%d:%d", b.Width, b.Height, w, h)
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

// calcSquarePx calculates the size of a game board square (in pixels).
// It figures out the maximum size in pixels for a square given the total pixel width/height
// and the number of squares wide/high.
func calcSquarePx(wPx, hPx, wS, hS int) int {
	// in order to accommodate boards that are rectangles, we need to figure out
	// what the square size should be to fit in the width and within the height
	// We'll use whichever is smaller to make sure the squares fit within the height/width constraints
	maxW := squareFit(wPx, wS)
	maxH := squareFit(hPx, hS)

	return min(maxW, maxH)
}

func squareFit(bounds, numSquares int) int {
	return (bounds - int(BoardBorder)*2) / numSquares
}

// DrawBoard draws the given board data into an image.
// Width and height values are in pixels.
// If the image width/height is invalid (<= 0) a valid width/height
// is calculated using the number of squares in the board.
func DrawBoard(b *Board, imageWidth, imageHeight int) image.Image {
	// check if we need to calculate the image width
	if imageWidth <= 0 || imageHeight <= 0 {
		// the legacy endpoints don't accept width/height parameters
		// in those cases, the height/width is the Go zero value (0)
		// and we should default to the old size which was 20 x num squares
		imageWidth = b.Width * 20
		imageHeight = b.Height * 20
	}
	dc := createBoardContext(b, imageWidth, imageHeight)

	// Draw food and snakes over watermark
	for p, s := range b.squares { // cool, we can iterate ONLY the non-empty squares!
		for _, c := range s.Contents {
			switch c.Type {
			case BoardSquareSnakeHead:
				drawSnakeImage(c.SnakeType, snakeHead, dc, p.X, p.Y, c.Color, c.Direction)
				drawGaps(dc, p.X, p.Y, c.Direction, c.Color)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, p.X, p.Y, c.Color, c.Corner)
				drawGaps(dc, p.X, p.Y, c.Direction, c.Color)
			case BoardSquareSnakeTail:
				drawSnakeImage(c.SnakeType, snakeTail, dc, p.X, p.Y, c.Color, c.Direction)
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
func boardXToDrawX(dc *boardContext, x int) float64 {
	return float64(x * dc.squareSizePixels)
}

// boardYToDrawY converts a y coordinate in "board space" to the y coordinate used by graphics.
// More specifically, it assumes the board coordinates are the indexes of squares and it returns the upper left
// corner for that square.
func boardYToDrawY(dc *boardContext, y int) float64 {
	// Note: the Battlesnake board coordinates have (0,0) at the bottom left
	// so we need to flip the y-axis to convert to the graphics, which follows the convention
	// of (0,0) being the top left.
	return float64((dc.Height() - int(BoardBorder)*2 - dc.squareSizePixels) - (y * dc.squareSizePixels)) // flip!
}
