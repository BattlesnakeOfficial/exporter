package inkscape_test

import (
	"errors"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/inkscape"
	"github.com/stretchr/testify/require"
)

func TestSVGToPNG(t *testing.T) {
	// happy path
	client := inkscape.Client{}
	got, err := client.SVGToPNG("testdata/example.svg", 100, 100)
	require.NoError(t, err)
	require.Equal(t, 100, got.Bounds().Max.X)
	require.Equal(t, 100, got.Bounds().Max.Y)
	want := loadTestImage(t)
	same(t, want, got)

	// client should validate width/height
	_, err = client.SVGToPNG("testdata/example.svg", 0, 100)
	require.Equal(t, errors.New("invalid width"), err)
	_, err = client.SVGToPNG("testdata/example.svg", 100, -1)
	require.Equal(t, errors.New("invalid height"), err)

	// should get an error when file doesn't exist
	_, err = client.SVGToPNG("testdata/filedoesntexist.svg", 100, 100)
	require.Equal(t, errors.New("SVG not found"), err)

	// should get an error when the command is wrong
	client = inkscape.Client{
		Command: "invalidinkscape",
	}
	_, err = client.SVGToPNG("testdata/example.svg", 100, 100)
	require.Error(t, err)
}

func TestIsAvailable(t *testing.T) {
	t.Log("Default command")
	client := inkscape.Client{}
	require.True(t, client.IsAvailable())

	t.Log("Invalid command")
	client = inkscape.Client{
		Command: "invalidinkscape",
	}
	require.False(t, client.IsAvailable())
}

func same(t *testing.T, a, b image.Image) {
	require.Equal(t, a.Bounds().Max.X, b.Bounds().Max.X)
	require.Equal(t, a.Bounds().Max.Y, b.Bounds().Max.Y)

	for x := 0; x < a.Bounds().Max.X; x++ {
		for y := 0; y < a.Bounds().Min.Y; y++ {
			c1 := a.At(x, y)
			c2 := b.At(x, y)
			sameColor(t, c1, c2)
		}
	}
}

func sameColor(t *testing.T, c1, c2 color.Color) {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c1.RGBA()
	require.True(t, r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2)
}

func loadTestImage(t *testing.T) image.Image {
	f, err := os.Open("testdata/example.png")
	require.NoError(t, err)
	defer f.Close()

	assetImage, _, err := image.Decode(f)
	require.NoError(t, err)

	return assetImage
}
