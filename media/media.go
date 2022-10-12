package media

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BattlesnakeOfficial/exporter/inkscape"
	"github.com/disintegration/imaging"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

const (
	fallbackHead = "heads/default.png" // relative to base path
	fallbackTail = "tails/default.png" // relative to base path
)

// imageCache is a cache that contains image.Image values
var imageCache = cache.New(time.Hour, 10*time.Minute)

var inkscapeClient = &inkscape.Client{}

var baseDir = "media/assets"
var svgMgr = &svgManager{
	baseDir:  filepath.Join(baseDir, "downloads"),
	inkscape: inkscapeClient,
}

// GetWatermarkPNG gets the watermark asset, scaled to the requested width/height
func GetWatermarkPNG(w, h int) (image.Image, error) {
	return loadLocalImageAsset("watermark.png", w, h)
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

// imageCacheKey creates a cache key that is unique to the given parameters.
// color can be nil when there is no color.
func imageCacheKey(path string, w, h int, c color.Color) string {
	return fmt.Sprintf("%s:%d:%d:%s", path, w, h, colorToHex6(c))
}

// loadLocalImageAsset loads the specified media asset from the local filesystem.
// It assumes the "mediaPath" is relative to the base path.
// The base path is the directory where all media assets should be located within.
func loadLocalImageAsset(mediaPath string, w, h int) (image.Image, error) {
	key := imageCacheKey(mediaPath, w, h, nil)
	cachedImage, ok := imageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	fullPath := filepath.Join(baseDir, mediaPath) // file is within the baseDir on disk
	img, err := loadImageFile(fullPath)
	if err != nil {
		log.WithField("path", fullPath).WithError(err).Errorf("Error loading asset from file")
		return nil, err
	}
	img = scaleImage(img, w, h)
	imageCache.Set(key, img, cache.DefaultExpiration)

	return img, nil
}

func getSnakeSVGImage(path, fallbackPath string, w, h int, c color.Color) (image.Image, error) {
	// first we try to load from the media server SVG's
	img, err := svgMgr.loadSnakeSVGImage(path, w, h, c)
	if err != nil {
		// log at info, because this could error just for people specifying snake types that don't exist
		log.WithFields(log.Fields{
			"path":     path,
			"fallback": fallbackPath,
		}).WithError(err).Info("unable to load SVG image asset - loading fallback")

		img, err = loadLocalImageAsset(fallbackPath, w, h)
		if err != nil {
			// at this point we are unable to draw correctly, so we should log at error level
			log.WithFields(log.Fields{
				"path":     path,
				"fallback": fallbackPath,
			}).WithError(err).Error("Unable to load local fallback image from file")
			return nil, err
		}
		img = changeImageColor(img, c)
	}

	return img, err
}

func ConvertSVGStringToPNG(svg string, w, h int) (image.Image, error) {
	// make sure inkscape is available, otherwise we can't create an image from an SVG
	if !inkscapeClient.IsAvailable() {
		return nil, errors.New("inkscape is not available - unable to convert SVG")
	}

	img, err := inkscapeClient.SVGStringToPNG(svg, w, h)
	if err != nil {
		log.WithError(err).Info("unable to rasterize SVG")
		return nil, err
	}
	return img, nil
}

type svgManager struct {
	baseDir  string
	inkscape *inkscape.Client
}

func (sm svgManager) loadSnakeSVGImage(mediaPath string, w, h int, c color.Color) (image.Image, error) {
	key := imageCacheKey(mediaPath, w, h, c)
	cachedImage, ok := imageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	// make sure inkscape is available, otherwise we can't create an image from an SVG
	if !sm.inkscape.IsAvailable() {
		return nil, errors.New("inkscape is not available - unable to load SVG")
	}

	mediaPath, err := sm.ensureDownloaded(mediaPath, c)
	if err != nil {
		return nil, err
	}

	path := sm.getFullPath(mediaPath)

	// rasterize the SVG
	img, err := sm.inkscape.SVGToPNG(path, w, h)
	if err != nil {
		log.WithField("path", path).WithError(err).Info("unable to rasterize SVG")
		return nil, err
	}

	imageCache.Set(key, img, cache.DefaultExpiration)
	return img, nil
}

func (sm svgManager) ensureSubdirExists(subDir string) error {
	path := sm.getFullPath(subDir)
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(path, os.ModePerm)
	}

	return err
}

func (sm svgManager) writeFile(mediaPath string, b []byte) error {
	mediaPath = filepath.Clean(mediaPath)
	err := sm.ensureSubdirExists(filepath.Dir(mediaPath))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(sm.getFullPath(mediaPath), b, os.ModePerm)
}

func (sm svgManager) getFullPath(mediaPath string) string {
	return filepath.Join(sm.baseDir, mediaPath)
}

func (sm svgManager) ensureDownloaded(mediaPath string, c color.Color) (string, error) {
	// use the colour as a directory to separate different colours of SVG's
	customizedMediaPath := path.Join(colorToHex6(c), mediaPath)

	// check if we need to download the SVG from the media server
	_, err := os.Stat(sm.getFullPath(customizedMediaPath))
	if errors.Is(err, fs.ErrNotExist) {
		svg, err := getCachedMediaResource(mediaPath)
		if err != nil {
			return "", err
		}

		svg = customiseSnakeSVG(svg, c)

		err = sm.writeFile(customizedMediaPath, []byte(svg))
		if err != nil {
			return "", err
		}
	}

	// return the new media path which includes the colour directory
	return customizedMediaPath, nil
}

// customiseSnakeSVG sets the fill colour for the outer SVG tag
func customiseSnakeSVG(svg string, c color.Color) string {
	var buf bytes.Buffer
	decoder := xml.NewDecoder(strings.NewReader(svg))
	encoder := xml.NewEncoder(&buf)

	rootSVGFound := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithError(err).Info("error while decoding SVG token")
			break // skip this token
		}
		token = xml.CopyToken(token)
		switch v := (token).(type) {
		case xml.StartElement:

			// check if this is the root SVG tag that we can change the colour on
			if !rootSVGFound && v.Name.Local == "svg" {
				rootSVGFound = true
				attrs := append(v.Attr, xml.Attr{Name: xml.Name{Local: "fill"}, Value: colorToHex6(c)})
				(&v).Attr = attrs
			}

			// this is necessary to prevent a weird behavior in Go's XML serialization where every tag gets the
			// "xmlns" set, even if it already has that as an attribute
			// see also: https://github.com/golang/go/issues/7535
			(&v).Name.Space = ""
			token = v
		case xml.EndElement:
			// this is necessary to prevent a weird behavior in Go's XML serialization where every tag gets the
			// "xmlns" set, even if it already has that as an attribute
			// see also: https://github.com/golang/go/issues/7535
			(&v).Name.Space = ""
			token = v
		}

		if err := encoder.EncodeToken(token); err != nil {
			log.Fatal(err)
		}
	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		log.Fatal(err)
	}

	return buf.String()
}

func scaleImage(src image.Image, w, h int) image.Image {
	// no-op if image already at requested width/height
	if src.Bounds().Max.X == w && src.Bounds().Max.Y == h {
		return src
	}
	return imaging.Resize(src, w, h, imaging.Lanczos)
}

// colorToHex6 converts a color.Color to a 6-digit hexadecimal string.
// If color is nil, the empty string is returned.
func colorToHex6(c color.Color) string {
	if c == nil {
		return ""
	}

	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r), uint8(g), uint8(b))
}
