package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// AssetType is intended to be used like an enum for the type of asset we are loading
type AssetType string

const (
	// AssetHead is a snake head
	AssetHead AssetType = "head"
	// AssetTail is a snake tail
	AssetTail = "tail"
	// AssetGeneric is a general type of asset
	AssetGeneric = "generic"
)

const (
	AssetFallbackHeadName = "default"
	AssetFallbackTailName = "default"
	BoardBorder           = 2
	SquareSizePixels      = 20
	SquareBorderPixels    = 1
	SquareFoodRadius      = SquareSizePixels / 3
	ColorEmptySquare      = "#f0f0f0"
	ColorFood             = "#ff5c75"
	ColorHazard           = "#00000066"
)

var boardImageCache = cache.New(6*time.Hour, time.Minute)
var assetImageCache = cache.New(time.Hour, time.Minute)

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

func loadImageFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func transformImage(src image.Image, w, h, rot int) image.Image {
	var dstImage image.Image
	dstImage = imaging.Resize(src, w, h, imaging.Lanczos)

	if rot == 180 {
		dstImage = imaging.FlipH(dstImage)
	} else if rot == 90 {
		dstImage = imaging.Rotate90(dstImage)
	} else if rot == 270 {
		dstImage = imaging.Rotate270(dstImage)
	}

	return dstImage
}

func imageCacheKey(path string, w, h, rot int) string {
	return fmt.Sprintf("local:%s:%d:%d:%d", path, w, h, rot)
}

func loadLocalImageAsset(path string, w, h, rot int) (image.Image, error) {
	key := imageCacheKey(path, w, h, rot)
	cachedImage, ok := assetImageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	img, err := loadImageFile(path)
	if err != nil {
		logrus.WithField("path", path).WithError(err).Errorf("Error loading asset from file")
		return nil, err
	}
	img = transformImage(img, w, h, rot)
	assetImageCache.Set(key, img, 0)

	return img, nil
}

func drawWatermark(dc *gg.Context) {
	watermarkImage, err := loadLocalImageAsset("render/assets/watermark.png", dc.Width()*2/3, dc.Height()*2/3, 0)
	if err != nil {
		logrus.WithError(err).Error("Unable to load watermark image")
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

func drawSnakeImage(filename string, fallbackFilename string, dc *gg.Context, bx int, by int, hexColor string, dir snakeDirection) {
	var rotation int
	switch dir {
	case movingRight:
		rotation = 0
	case movingDown:
		rotation = 270
	case movingLeft:
		rotation = 180
	case movingUp:
		rotation = 90
	}

	width := SquareSizePixels - SquareBorderPixels*2
	height := SquareSizePixels - SquareBorderPixels*2
	maskImage, err := loadLocalImageAsset(fmt.Sprintf("render/assets/%s", filename), width, height, rotation)
	if err != nil {
		logrus.WithError(err).Error("Unable to load asset, trying fallback")
		maskImage, err = loadLocalImageAsset(fmt.Sprintf("render/assets/%s", fallbackFilename), width, height, rotation)
		if err != nil {
			logrus.WithError(err).Error("Unable to load fallback image - aborting draw")
			return
		}
	}

	dst := dc.Image().(draw.Image)
	dstRect := image.Rect(
		int(boardXToDrawX(dc, bx))+SquareBorderPixels+BoardBorder,
		int(boardYToDrawY(dc, by))+SquareBorderPixels+BoardBorder,
		int(boardXToDrawX(dc, bx+1))-SquareBorderPixels+BoardBorder,
		int(boardYToDrawY(dc, by-1))-SquareBorderPixels+BoardBorder,
	)

	srcImage := &image.Uniform{parseHexColor(hexColor)}

	draw.DrawMask(dst, dstRect, srcImage, image.Point{}, maskImage, image.Point{}, draw.Over)
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

	cacheKey := fmt.Sprintf("%d:%d", b.Width, b.Height)
	cachedBoardImage, ok := boardImageCache.Get(cacheKey)
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
	boardImageCache.Set(cacheKey, cacheDC.Image(), 0)

	return dc
}

func DrawBoard(b *Board) image.Image {
	dc := createBoardContext(b)

	// Draw food and snakes over watermark
	var snakeAsset string
	for p, s := range b.squares { // cool, we can iterate ONLY the non-empty squares!
		for _, c := range s.Contents {
			switch c.Type {
			case BoardSquareSnakeHead:
				snakeAsset = fmt.Sprintf("heads/%s.png", c.SnakeType)
				drawSnakeImage(snakeAsset, AssetFallbackHeadName, dc, p.X, p.Y, c.HexColor, c.Direction)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, p.X, p.Y, c.HexColor, c.Corner)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeTail:
				snakeAsset = fmt.Sprintf("tails/%s.png", c.SnakeType)
				drawSnakeImage(snakeAsset, AssetFallbackTailName, dc, p.X, p.Y, c.HexColor, c.Direction)
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
