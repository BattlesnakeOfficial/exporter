package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
)

const (
	AssetFallbackHead        = "heads/regular.png"
	AssetFallbackTail        = "tails/regular.png"
	AssetFallbackUnspecified = ""
	BoardBorder              = 2
	SquareSizePixels         = 20
	SquareBorderPixels       = 1
	SquareFoodRadius         = SquareSizePixels / 3
	ColorEmptySquare         = "#f0f0f0"
	ColorFood                = "#ff5c75"
	ColorHazard              = "#00000066"
)

var boardImageCache = make(map[string]image.Image)
var boardImageCacheLock sync.Mutex

var assetImageCache = make(map[string]image.Image)
var assetImageCacheLock sync.Mutex

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

func loadRawImageAsset(filename string) image.Image {
	f, err := os.Open(fmt.Sprintf("render/assets/%s", filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	assetImage, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return assetImage
}

func loadRawImageAssetWithFallback(filename string, fallbackFilename string) image.Image {
	if fallbackFilename == AssetFallbackUnspecified {
		return loadRawImageAsset(filename)
	}

	f, err := os.Open(fmt.Sprintf("render/assets/%s", filename))
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"fallback": fallbackFilename}).Warn("unable to open image asset file")
		return loadRawImageAsset(fallbackFilename)
	}
	defer f.Close()

	assetImage, _, err := image.Decode(f)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"fallback": fallbackFilename}).Warn("unable to decode image asset")
		return loadRawImageAsset(fallbackFilename)
	}

	return assetImage
}

func loadImageAsset(filename string, fallbackFilename string, w int, h int, rot int) image.Image {
	cacheKey := fmt.Sprintf("%s:%d:%d:%d", filename, w, h, rot)
	cachedImage, ok := assetImageCache[cacheKey]
	if ok {
		return cachedImage
	}

	assetImageCacheLock.Lock()
	defer assetImageCacheLock.Unlock()

	srcImage := loadRawImageAssetWithFallback(filename, fallbackFilename)

	var dstImage image.Image
	dstImage = imaging.Resize(srcImage, w, h, imaging.Lanczos)

	if rot == 180 {
		dstImage = imaging.FlipH(dstImage)
	} else if rot == 90 {
		dstImage = imaging.Rotate90(dstImage)
	} else if rot == 270 {
		dstImage = imaging.Rotate270(dstImage)
	}

	assetImageCache[cacheKey] = dstImage

	return dstImage
}

func drawWatermark(dc *gg.Context) {
	watermarkImage := loadImageAsset("watermark.png", AssetFallbackUnspecified, dc.Width()*2/3, dc.Height()*2/3, 0)
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
	case moveRight:
		rotation = 0
	case moveDown:
		rotation = 270
	case moveLeft:
		rotation = 180
	case moveUp:
		rotation = 90
	}

	maskImage := loadImageAsset(
		filename,
		fallbackFilename,
		SquareSizePixels-SquareBorderPixels*2,
		SquareSizePixels-SquareBorderPixels*2,
		rotation,
	)

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
	case moveUp:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by-1)-SquareBorderPixels+BoardBorder,
			SquareSizePixels-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case moveDown:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)+SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)-SquareBorderPixels+BoardBorder,
			SquareSizePixels-SquareBorderPixels*2,
			SquareBorderPixels*2,
		)
	case moveRight:
		dc.DrawRectangle(
			boardXToDrawX(dc, bx)-SquareBorderPixels+BoardBorder,
			boardYToDrawY(dc, by)+SquareBorderPixels+BoardBorder,
			SquareBorderPixels*2,
			SquareSizePixels-SquareBorderPixels*2,
		)
	case moveLeft:
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
	cachedBoardImage, ok := boardImageCache[cacheKey]
	if ok {
		dc.DrawImage(cachedBoardImage, 0, 0)
		return dc
	}

	boardImageCacheLock.Lock()
	defer boardImageCacheLock.Unlock()

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
	boardImageCache[cacheKey] = cacheDC.Image()

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
				drawSnakeImage(snakeAsset, AssetFallbackHead, dc, p.X, p.Y, c.HexColor, c.Direction)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, p.X, p.Y, c.HexColor, c.Corner)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeTail:
				snakeAsset = fmt.Sprintf("tails/%s.png", c.SnakeType)
				drawSnakeImage(snakeAsset, AssetFallbackTail, dc, p.X, p.Y, c.HexColor, c.Direction)
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
