package media

import (
	"fmt"
	"image"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/inkscape"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// set up a mock battlesnake media server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "notfound") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.Contains(r.RequestURI, "heads") {
			fmt.Fprint(w, headSVG)
		}

		if strings.Contains(r.RequestURI, "tails") {
			fmt.Fprint(w, tailSVG)
		}
	}))
	mediaServerURL = svr.URL
	defer svr.Close()

	// need to override these directories because the paths aren't right when run by unit tests
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err) // something is pretty wrong if we can't make a temp dir
	}
	baseDir = "assets"
	svgMgr.baseDir = tmpDir
	os.Exit(m.Run())
}

func TestGetHeadSVG(t *testing.T) {
	svg, err := GetHeadSVG("default")
	require.NoError(t, err)
	require.Equal(t, headSVG, svg)
}

func TestGetTailSVG(t *testing.T) {
	svg, err := GetTailSVG("default")
	require.NoError(t, err)
	require.Equal(t, tailSVG, svg)
}

func TestGetTailPNG(t *testing.T) {
	img, err := GetTailPNG("default", 20, 20)
	require.NoError(t, err)
	assertImg(t, img, 20, 20)
}

func TestGetHeadPNG(t *testing.T) {
	img, err := GetHeadPNG("default", 20, 20)
	require.NoError(t, err)
	assertImg(t, img, 20, 20)
}

func TestSVGManager(t *testing.T) {
	baseDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	mgr := svgManager{
		baseDir:  baseDir,
		inkscape: &inkscape.Client{},
	}

	require.Equal(t, mgr.getFullPath("things/foo.svg"), filepath.Join(baseDir, "things/foo.svg"))
	require.NoDirExists(t, filepath.Join(baseDir, "things"))
	require.NoFileExists(t, mgr.getFullPath("things/foo.svg"))

	require.NoError(t, mgr.writeFile("things/foo.svg", []byte(tailSVG)))
	require.DirExists(t, filepath.Join(baseDir, "things"))
	require.FileExists(t, mgr.getFullPath("things/foo.svg"))
	err = mgr.ensureDownloaded("things/foo.svg")
	require.NoError(t, err)

	require.NoError(t, mgr.ensureSubdirExists("some/subdir"))
	require.DirExists(t, mgr.getFullPath("some/subdir"))

	img, err := mgr.loadSVGImage(headSVGPath("default"), 20, 20)
	require.NoError(t, err)
	assertImg(t, img, 20, 20)
}

func TestGetSVGImageWithFallback(t *testing.T) {

	// these shouldn't require a fallback
	img, err := getSVGImageWithFallback(tailSVGPath("default"), "nofallback.png", 20, 20)
	require.NoError(t, err)
	require.NotNil(t, img)
	assertImg(t, img, 20, 20)
	img, err = getSVGImageWithFallback(headSVGPath("default"), "nofallback.png", 20, 20)
	require.NoError(t, err)
	require.NotNil(t, img)
	assertImg(t, img, 20, 20)

	// test head/tail fallbacks
	img, err = getSVGImageWithFallback(tailSVGPath("notfound"), fallbackTail, 20, 20)
	require.NoError(t, err)
	require.NotNil(t, img)
	assertImg(t, img, 20, 20)
	img, err = getSVGImageWithFallback(headSVGPath("notfound"), fallbackHead, 20, 20)
	require.NoError(t, err)
	require.NotNil(t, img)
	assertImg(t, img, 20, 20)

	// this should just error
	img, err = getSVGImageWithFallback(tailSVGPath("notfound"), "404/notfound.png", 20, 20)
	require.Error(t, err)
	require.Nil(t, img)
}

func TestGetWatermarkPNG(t *testing.T) {
	img, err := GetWatermarkPNG(100, 100)
	require.NoError(t, err)
	assertImg(t, img, 100, 100)
}

func assertImg(t *testing.T, img image.Image, w, h int) {
	require.NotNil(t, img)
	assert.Equal(t, img.Bounds().Max.X, w)
	assert.Equal(t, img.Bounds().Max.Y, h)
}

func TestLoadImageFile(t *testing.T) {
	i, err := loadImageFile("assets/watermark.png")
	require.NoError(t, err)
	require.NotNil(t, i)

	i, err = loadImageFile("testdata/notexistingimage.png")
	require.Error(t, err)
	require.Nil(t, i)
}

func TestLoadLocalImageAsset(t *testing.T) {
	// happy paths for assets that should always exist
	i, err := loadLocalImageAsset(fallbackHead, 20, 20)
	require.NoError(t, err)
	require.NotNil(t, i)
	// ensure caching works
	_, ok := imageCache.Get(imageCacheKey(fallbackHead, 20, 20))
	require.True(t, ok, "image should get cached")

	i, err = loadLocalImageAsset(fallbackTail, 20, 20)
	require.NoError(t, err)
	require.NotNil(t, i)
	i, err = loadLocalImageAsset("watermark.png", 100, 100)
	require.NoError(t, err)
	require.NotNil(t, i)

	// ensure non-existing is handled gracefully
	_, err = loadLocalImageAsset("assets/notfound.png", 100, 100)
	require.Error(t, err, "this image doesnt exist, so it should error when loading")
}

func TestScaleImage(t *testing.T) {
	i, err := loadImageFile("assets/watermark.png")
	require.NoError(t, err)
	require.NotNil(t, i)
	newX := i.Bounds().Max.X * 2
	newY := i.Bounds().Max.Y * 2
	si := scaleImage(i, newX, newY)
	require.NotNil(t, si)
	assert.Equal(t, si.Bounds().Max.X, newX)
	assert.Equal(t, si.Bounds().Max.Y, newY)
}

const headSVG = `<svg id="root" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
<circle fill="none" cx="12.52" cy="28.55" r="9.26"/>
<path d="M0 100h100L56 55.39l44-39.89V.11L0 0zm12.52-80.71a9.26 9.26 0 1 1-9.26 9.26 9.26 9.26 0 0 1 9.26-9.26z"/>
</svg>`
const tailSVG = `<svg id="root" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
<path d="M50 0H0v100h50l50-50L50 0z"/>
</svg>`
