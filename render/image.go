package render

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BattlesnakeOfficial/exporter/inkscape"
	"github.com/BattlesnakeOfficial/exporter/media"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// AssetType is intended to be used like an enum for the type of asset we are loading
type AssetType string

const (
	// AssetHead is a snake head
	AssetHead AssetType = "head"
	// AssetTail is a snake tail
	AssetTail AssetType = "tail"
)

type rotations int

const (
	rotate0 rotations = iota
	rotate90
	rotate180
	rotate270
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
var inkscapeClient = &inkscape.Client{}
var svgMgr = &svgManager{baseDir: "render/assets/"}

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

func fetchSVGAsset(name string, aType AssetType) (string, error) {
	switch aType {
	case AssetHead:
		return media.GetHeadSVG(name) // already potentially cached
	case AssetTail:
		return media.GetTailSVG(name) // already potentially cached
	default:
		return "", fmt.Errorf("unrecognised SVG asset type '%s'", aType)
	}
}

func scaleImage(src image.Image, w, h int) image.Image {
	return imaging.Resize(src, w, h, imaging.Lanczos)
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

func imageCacheKey(path string, w, h int, rot rotations) string {
	return fmt.Sprintf("%s:%d:%d:%d", path, w, h, rot)
}

func loadLocalImageAsset(path string, w, h int, rot rotations) (image.Image, error) {
	key := imageCacheKey(path, w, h, rot)
	cachedImage, ok := assetImageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	img, err := loadImageFile(path)
	if err != nil {
		log.WithField("path", path).WithError(err).Errorf("Error loading asset from file")
		return nil, err
	}
	img = scaleImage(img, w, h)
	assetImageCache.Set(key, img, 0)

	return img, nil
}

func drawWatermark(dc *gg.Context) {
	watermarkImage, err := loadLocalImageAsset("render/assets/watermark.png", dc.Width()*2/3, dc.Height()*2/3, 0)
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

func drawSnakeImage(name string, aType AssetType, dc *gg.Context, bx int, by int, hexColor string, dir snakeDirection) {
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
	width := SquareSizePixels - SquareBorderPixels*2
	height := SquareSizePixels - SquareBorderPixels*2

	// first we try to load from the media server SVG's
	maskImage, err := svgMgr.loadSVGImageAsset(name, aType, width, height, rot)
	if err != nil {
		// log at info, because this could error just for people specifying snake types that don't exist
		log.WithFields(log.Fields{
			"name": name,
			"type": aType,
		}).WithError(err).Info("unable to load SVG image asset")

		// attempt a graceful fall back to local, default PNG
		var filename string
		switch aType {
		case AssetHead:
			filename = fmt.Sprintf("render/assets/heads/%s.png", AssetFallbackHeadName)
		case AssetTail:
			filename = fmt.Sprintf("render/assets/tails/%s.png", AssetFallbackTailName)
		default:
			// something went wrong if we are asked for types we don't know about (log at error)
			log.WithField("type", aType).Error("unrecognized snake image type - aborting draw")
			return
		}
		maskImage, err = loadLocalImageAsset(filename, width, height, rot)
		if err != nil {
			// at this point we are unable to draw correctly, so we should log at error level
			log.WithField("type", aType).WithError(err).Error("Unable to load local fallback image from file - aborting draw")
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
	for p, s := range b.squares { // cool, we can iterate ONLY the non-empty squares!
		for _, c := range s.Contents {
			switch c.Type {
			case BoardSquareSnakeHead:
				drawSnakeImage(c.SnakeType, AssetHead, dc, p.X, p.Y, c.HexColor, c.Direction)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeBody:
				drawSnakeBody(dc, p.X, p.Y, c.HexColor, c.Corner)
				drawGaps(dc, p.X, p.Y, c.Direction, c.HexColor)
			case BoardSquareSnakeTail:
				drawSnakeImage(c.SnakeType, AssetTail, dc, p.X, p.Y, c.HexColor, c.Direction)
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

type svgManager struct {
	baseDir string
}

func (sm svgManager) ensureSubdirExists(subDir string) error {
	path := filepath.Join(sm.baseDir, subDir)
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(path, os.ModePerm)
	}

	return err
}

func (sm svgManager) loadSVGImageAsset(name string, aType AssetType, w, h int, rot rotations) (image.Image, error) {
	// determine path to SVG given whatever asset type it is
	var path string
	var err error
	switch aType {
	case AssetHead:
		err = sm.ensureSubdirExists("heads")
		path = filepath.Join(sm.baseDir, "heads", fmt.Sprintf("%s.svg", name))
	case AssetTail:
		err = sm.ensureSubdirExists("tails")
		path = filepath.Join(sm.baseDir, "tails", fmt.Sprintf("%s.svg", name))
	default:
		return nil, fmt.Errorf("unable to load SVG - unrecognized asset type '%s'", aType)
	}

	if err != nil {
		return nil, err
	}

	key := imageCacheKey(path, w, h, rot)
	cachedImage, ok := assetImageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	// make sure inkscape is available, otherwise we can't create an image from an SVG
	if !inkscapeClient.IsAvailable() {
		return nil, errors.New("inkscape is not available - unable to load SVG")
	}

	// check if we need to download the SVG from the media server
	_, err = os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		svg, err := fetchSVGAsset(name, aType)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(path, []byte(svg), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	// rasterize the SVG
	img, err := inkscapeClient.SVGToPNG(path, w, h)
	if err != nil {
		log.WithField("path", path).WithError(err).Info("unable to rasterize SVG")
		return nil, err
	}

	img = rotateImage(img, rot)
	assetImageCache.Set(key, img, 0)
	return img, nil
}
