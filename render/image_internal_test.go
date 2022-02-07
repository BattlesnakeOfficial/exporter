package render

import (
	"fmt"
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
