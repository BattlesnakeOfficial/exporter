package render

import (
	"fmt"
	"image"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadImageFile(t *testing.T) {
	i, err := loadImageFile("testdata/sample1.png")
	require.NoError(t, err)
	require.NotNil(t, i)

	i, err = loadImageFile("testdata/notexistingimage.png")
	require.Error(t, err)
	require.Nil(t, i)
}

func TestLoadImageAsset(t *testing.T) {
	// happy paths for assets that should always exist
	i, err := loadLocalImageAsset(fmt.Sprintf("assets/heads/%s.png", AssetFallbackHeadName), 20, 20, 0)
	require.NoError(t, err)
	require.NotNil(t, i)
	// ensure caching works
	_, ok := assetImageCache.Get(imageCacheKey(fmt.Sprintf("assets/heads/%s.png", AssetFallbackHeadName), 20, 20, 0))
	require.True(t, ok, "image should get cached")

	i, err = loadLocalImageAsset(fmt.Sprintf("assets/tails/%s.png", AssetFallbackTailName), 20, 20, 0)
	require.NoError(t, err)
	require.NotNil(t, i)
	i, err = loadLocalImageAsset("assets/watermark.png", 100, 100, 0)
	require.NoError(t, err)
	require.NotNil(t, i)

	// ensure non-existing is handled gracefully
	_, err = loadLocalImageAsset("assets/doesnotexistfooblah.png", 100, 100, 0)
	require.Error(t, err, "this image doesnt exist, so it should error when loading")
}

func TestEnsureSubdirExists(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), randStr(12))
	mgr := svgManager{
		baseDir: baseDir,
	}
	require.NoError(t, mgr.ensureSubdirExists("thing"))
	require.DirExists(t, filepath.Join(baseDir, "thing"))
}

func TestLoadSVGImageAsset(t *testing.T) {
	mgr := svgManager{
		baseDir: filepath.Join(os.TempDir(), randStr(12)),
	}
	var err error
	var img image.Image
	// try to load SVG with a few retries in case of network flake
	for i := 0; i < 3; i++ {
		img, err = mgr.loadSVGImageAsset("default", AssetHead, 20, 20, rotate180)
		if err == nil {
			break
		}
		t.Logf("encountered error '%v', retrying", err)
	}
	require.NoError(t, err)
	require.NotNil(t, img)
}

func randStr(n int) string {
	var letters = []rune("346789BCDFGHJKMPQRSTVWXYbcdfghjkmpqrtvwxy")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
