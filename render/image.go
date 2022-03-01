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
	// boardOffsetX is the x offset for the bottom-left corner of the board in the image.
	// For boards that are not perfectly fit in an image, it will not be > 0 to center the board.
	boardOffsetX int
	// boardOffsetY is the y offset for the bottom-left corner of the board in the image.
	// For boards that are not perfectly fit in an image, it will not be > 0 to center the board.
	boardOffsetY int
	// boardWidthPx is the width of the actual game board, in pixels
	// This is different than dc.Width, which is the width of the entire image
	boardWidthPx int
	// boardHeightPx is the height of the actual game board, in pixels
	// This is different than dc.Height, which is the height of the entire image
	boardHeightPx int
	// squareSizePx is the size of a single game board square, in pixels
	squareSizePx int
	// squareSizeHalfPx is the half the size of a single game board square
	// We pre-calculate because it's a common value and it's needed in float precision.
	squareSizeHalfPx float64
}

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

	// The watermark is a square image.
	// We want to scale it close to the maximum size of square that can fit within the board.
	wmSize := min(dc.boardWidthPx, dc.boardHeightPx)

	watermarkImage, err := media.GetWatermarkPNG(wmSize*2/3, wmSize*2/3)
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
		float64(dc.squareSizePx)-SquareBorderPixels*2,
		float64(dc.squareSizePx)-SquareBorderPixels*2,
	)
	dc.Fill()
}

func drawFood(dc *boardContext, bx int, by int) {
	dc.SetHexColor(ColorFood)
	dc.DrawCircle(
		boardXToDrawX(dc, bx)+dc.squareSizeHalfPx+BoardBorder,
		boardYToDrawY(dc, by)+dc.squareSizeHalfPx+BoardBorder,
		float64(dc.squareSizePx)/3,
	)
	dc.Fill()
}

func drawHazard(dc *boardContext, bx int, by int) {
	dc.SetHexColor(ColorHazard)
	dc.DrawRectangle(
		boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
		boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
		float64(dc.squareSizePx)-SquareBorderPixels*2,
		float64(dc.squareSizePx)-SquareBorderPixels*2,
	)
	dc.Fill()
}

func drawSnakeImage(name string, st snakeImageType, dc *boardContext, bx int, by int, c color.Color, dir snakeDirection) {

	width := dc.squareSizePx - int(SquareBorderPixels*2)
	height := dc.squareSizePx - int(SquareBorderPixels*2)

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
			float64(dc.squareSizePx)-SquareBorderPixels*2,
			float64(dc.squareSizePx)-SquareBorderPixels*2,
		)
	} else {
		dc.DrawRoundedRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			float64(dc.squareSizePx)-SquareBorderPixels*2,
			float64(dc.squareSizePx)-SquareBorderPixels*2,
			dc.squareSizeHalfPx,
		)
		if corner.isBottom() {
			dc.DrawRectangle(
				boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
				boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
				float64(dc.squareSizePx)-SquareBorderPixels*2,
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
				float64(dc.squareSizePx)-SquareBorderPixels*2,
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
			float64(dc.squareSizePx)-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case movingDown:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)-SquareBorderPixels+BoardBorder,
			float64(dc.squareSizePx)-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case movingRight:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			float64(dc.squareSizePx)-SquareBorderPixels*2,
		)
	case movingLeft:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx+1)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			float64(dc.squareSizePx)-SquareBorderPixels*2,
		)
	}
	dc.Fill()
}

func createBoardContext(b *Board, w, h int) *boardContext {
	ss := calcSquarePx(w, h, b.Width, b.Height)

	boardWidthPx := ss*b.Width + int(BoardBorder)*2
	boardHeightPx := ss*b.Height + int(BoardBorder)*2
	offsetX := (w - boardWidthPx) / 2
	offsetY := (h - boardHeightPx) / 2

	dc := &boardContext{
		Context:          gg.NewContext(w, h),
		squareSizePx:     ss,
		boardWidthPx:     boardWidthPx,
		boardHeightPx:    boardHeightPx,
		boardOffsetX:     offsetX,
		boardOffsetY:     offsetY,
		squareSizeHalfPx: float64(ss) / 2, // float to avoid rounding errors
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
	// dc.DrawRectangle(float64(bx), float64(by), float64(ss*b.Width), float64(ss*b.Height))

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
		// and we should default to the old size which was 20 * num squares + 2 * border size

		imageWidth = b.Width*20 + int(BoardBorder)*2
		imageHeight = b.Height*20 + int(BoardBorder)*2
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
	return float64(dc.boardOffsetX + x*dc.squareSizePx)
}

// boardYToDrawY converts a y coordinate in "board space" to the y coordinate used by graphics.
// More specifically, it assumes the board coordinates are the indexes of squares and it returns the upper left
// corner for that square.
func boardYToDrawY(dc *boardContext, y int) float64 {
	// Note: the Battlesnake board coordinates have (0,0) at the bottom left
	// so we need to flip the y-axis to convert to the graphics, which follows the convention
	// of (0,0) being the top left.
	drawY := (dc.Height() - int(BoardBorder)*2 - dc.squareSizePx) - (y * dc.squareSizePx)
	drawY = drawY - dc.boardOffsetY // add offset to center board in GIF frame
	return float64(drawY)
}
