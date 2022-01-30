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
	BoardBorderBottom        = 15
	SquareSizePixels         = 20
	SquareBorderPixels       = 1
	SquareFoodRadius         = SquareSizePixels / 3
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

func drawEmptySquare(dc *gg.Context, x int, y int) {
	dc.SetRGB255(240, 240, 240)
	dc.DrawRectangle(
		float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
		float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
		float64(SquareSizePixels-SquareBorderPixels*2),
		float64(SquareSizePixels-SquareBorderPixels*2),
	)
	dc.Fill()
}

func drawFood(dc *gg.Context, x int, y int) {
	dc.SetRGBA255(255, 92, 117, 255)
	dc.DrawCircle(
		float64(x*SquareSizePixels+SquareSizePixels/2+BoardBorder),
		float64(y*SquareSizePixels+SquareSizePixels/2+BoardBorder),
		SquareFoodRadius,
	)
	dc.Fill()
}

func drawHazard(dc *gg.Context, x int, y int) {
	dc.SetRGBA255(0, 0, 0, 102)
	dc.DrawRectangle(
		float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
		float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
		float64(SquareSizePixels-SquareBorderPixels*2),
		float64(SquareSizePixels-SquareBorderPixels*2),
	)
	dc.Fill()
}

func drawSnakeImage(filename string, fallbackFilename string, dc *gg.Context, x int, y int, hexColor string, direction string) {
	var rotation int
	switch direction {
	case "right":
		rotation = 0
	case "down":
		rotation = 270
	case "left":
		rotation = 180
	case "up":
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
		x*SquareSizePixels+SquareBorderPixels+BoardBorder,
		y*SquareSizePixels+SquareBorderPixels+BoardBorder,
		(x+1)*SquareSizePixels-SquareBorderPixels+BoardBorder,
		(y+1)*SquareSizePixels-SquareBorderPixels+BoardBorder,
	)

	srcImage := &image.Uniform{parseHexColor(hexColor)}

	draw.DrawMask(dst, dstRect, srcImage, image.Point{}, maskImage, image.Point{}, draw.Over)
}

func drawSnakeBody(dc *gg.Context, x int, y int, hexColor, corner string) {
	dc.SetHexColor(hexColor)
	if corner == "none" {
		dc.DrawRectangle(
			float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(SquareSizePixels-SquareBorderPixels*2),
			float64(SquareSizePixels-SquareBorderPixels*2),
		)
	} else {
		dc.DrawRoundedRectangle(
			float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(SquareSizePixels-SquareBorderPixels*2),
			float64(SquareSizePixels-SquareBorderPixels*2),
			float64(SquareSizePixels/2),
		)
		if strings.HasPrefix(corner, "bottom") {
			dc.DrawRectangle(
				float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
				float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
				float64(SquareSizePixels-SquareBorderPixels*2),
				float64(SquareSizePixels/2),
			)
			if strings.HasSuffix(corner, "left") {
				dc.DrawRectangle(
					float64(x*SquareSizePixels+SquareSizePixels/2+BoardBorder),
					float64(y*SquareSizePixels+SquareBorderPixels+SquareSizePixels/2+BoardBorder),
					float64(SquareSizePixels/2-SquareBorderPixels),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
				)
			}
			if strings.HasSuffix(corner, "right") {
				dc.DrawRectangle(
					float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
					float64(y*SquareSizePixels+SquareBorderPixels+SquareSizePixels/2+BoardBorder),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
				)
			}
		}
		if strings.HasPrefix(corner, "top") {
			dc.DrawRectangle(
				float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
				float64(y*SquareSizePixels+SquareBorderPixels+SquareSizePixels/2+BoardBorder),
				float64(SquareSizePixels-SquareBorderPixels*2),
				float64(SquareSizePixels/2),
			)
			if strings.HasSuffix(corner, "left") {
				dc.DrawRectangle(
					float64(x*SquareSizePixels+SquareSizePixels/2+SquareBorderPixels+BoardBorder),
					float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
				)
			}
			if strings.HasSuffix(corner, "right") {
				dc.DrawRectangle(
					float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
					float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
					float64(SquareSizePixels/2-SquareBorderPixels*2),
				)
			}
		}
	}
	dc.Fill()
}

func drawGaps(dc *gg.Context, x, y int, direction, hexColor string) {
	dc.SetHexColor(hexColor)
	switch direction {
	case "up":
		dc.DrawRectangle(
			float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64((y+1)*SquareSizePixels-SquareBorderPixels+BoardBorder),
			float64(SquareSizePixels-SquareBorderPixels*2),
			float64(SquareBorderPixels*2),
		)
	case "down":
		dc.DrawRectangle(
			float64(x*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(y*SquareSizePixels-SquareBorderPixels+BoardBorder),
			float64(SquareSizePixels-SquareBorderPixels*2),
			float64(SquareBorderPixels*2),
		)
	case "right":
		dc.DrawRectangle(
			float64(x*SquareSizePixels-SquareBorderPixels+BoardBorder),
			float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(SquareBorderPixels*2),
			float64(SquareSizePixels-SquareBorderPixels*2),
		)
	case "left":
		dc.DrawRectangle(
			float64((x+1)*SquareSizePixels-SquareBorderPixels+BoardBorder),
			float64(y*SquareSizePixels+SquareBorderPixels+BoardBorder),
			float64(SquareBorderPixels*2),
			float64(SquareSizePixels-SquareBorderPixels*2),
		)
	}
	dc.Fill()
}

func createBoardContext(b *Board) *gg.Context {
	dc := gg.NewContext(
		SquareSizePixels*b.Width+BoardBorder*2,
		SquareSizePixels*b.Height+BoardBorder*2+BoardBorderBottom,
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
			if len(b.getContents(x, y)) == 0 {
				drawEmptySquare(dc, x, y)
			}
		}
	}

	// Draw watermark
	drawWatermark(dc)

	// Draw subtitle
	dc.SetColor(color.Black)
	dc.DrawStringAnchored(
		"play.battlesnake.com",
		float64(dc.Width()/2),
		float64(dc.Height()-10),
		0.5, 0.5,
	)
	dc.Fill()

	// Cache for next time
	cacheDC := gg.NewContext(dc.Width(), dc.Height())
	cacheDC.DrawImage(dc.Image(), 0, 0)
	boardImageCache[cacheKey] = cacheDC.Image()

	return dc
}

func drawBoard(b *Board) image.Image {
	dc := createBoardContext(b)

	// Draw food and snakes over watermark
	var snakeAsset string
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			for _, c := range b.getContents(x, y) {
				switch c.Type {
				case BoardSquareSnakeHead:
					snakeAsset = fmt.Sprintf("heads/%s.png", c.SnakeType)
					drawSnakeImage(snakeAsset, AssetFallbackHead, dc, x, y, c.HexColor, c.Direction)
					drawGaps(dc, x, y, c.Direction, c.HexColor)
				case BoardSquareSnakeBody:
					drawSnakeBody(dc, x, y, c.HexColor, c.Corner)
					drawGaps(dc, x, y, c.Direction, c.HexColor)
				case BoardSquareSnakeTail:
					snakeAsset = fmt.Sprintf("tails/%s.png", c.SnakeType)
					drawSnakeImage(snakeAsset, AssetFallbackTail, dc, x, y, c.HexColor, c.Direction)
				case BoardSquareFood:
					drawFood(dc, x, y)
				case BoardSquareHazard:
					drawHazard(dc, x, y)
				}
			}
		}
	}

	return dc.Image()
}
