package media

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
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

var baseDir = "media/assets"
var svgMgr = &svgManager{
	baseDir:  filepath.Join(baseDir, "downloads"),
	inkscape: &inkscape.Client{},
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

func imageCacheKey(path string, w, h int) string {
	return fmt.Sprintf("%s:%d:%d", path, w, h)
}

// loadLocalImageAsset loads the specified media asset from the local filesystem.
// It assumes the "mediaPath" is relative to the base path.
// The base path is the directory where all media assets should be located within.
func loadLocalImageAsset(mediaPath string, w, h int) (image.Image, error) {
	key := imageCacheKey(mediaPath, w, h)
	mediaPath = filepath.Join(baseDir, mediaPath) // file is within the baseDir on disk
	cachedImage, ok := imageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	img, err := loadImageFile(mediaPath)
	if err != nil {
		log.WithField("path", mediaPath).WithError(err).Errorf("Error loading asset from file")
		return nil, err
	}
	img = scaleImage(img, w, h)
	imageCache.Set(key, img, cache.DefaultExpiration)

	return img, nil
}

func getSVGImageWithFallback(path, fallbackPath string, w, h int) (image.Image, error) {
	// first we try to load from the media server SVG's
	img, err := svgMgr.loadSVGImage(path, w, h)
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
		}
	}

	return img, err
}

func (sm svgManager) loadSVGImage(mediaPath string, w, h int) (image.Image, error) {
	key := imageCacheKey(mediaPath, w, h)
	cachedImage, ok := imageCache.Get(key)
	if ok {
		return cachedImage.(image.Image), nil
	}

	// make sure inkscape is available, otherwise we can't create an image from an SVG
	if !sm.inkscape.IsAvailable() {
		return nil, errors.New("inkscape is not available - unable to load SVG")
	}

	err := sm.ensureDownloaded(mediaPath)
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

type svgManager struct {
	baseDir  string
	inkscape *inkscape.Client
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

func (sm svgManager) ensureDownloaded(mediaPath string) error {
	// check if we need to download the SVG from the media server
	_, err := os.Stat(sm.getFullPath(mediaPath))
	if errors.Is(err, fs.ErrNotExist) {
		svg, err := getCachedMediaResource(mediaPath)
		if err != nil {
			return err
		}
		err = sm.writeFile(mediaPath, []byte(svg))
		if err != nil {
			return err
		}
	}

	return nil
}

func scaleImage(src image.Image, w, h int) image.Image {
	// no-op if image already at requested width/height
	if src.Bounds().Max.X == w && src.Bounds().Max.Y == h {
		return src
	}
	return imaging.Resize(src, w, h, imaging.Lanczos)
}
