package inkscape_test

import (
	"errors"
	"image"
	"io/fs"
	"os"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/imagetest"
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
	imagetest.Equal(t, want, got)

	// client should validate width/height
	_, err = client.SVGToPNG("testdata/example.svg", 0, 100)
	require.Equal(t, errors.New("invalid width"), err)
	_, err = client.SVGToPNG("testdata/example.svg", 100, -1)
	require.Equal(t, errors.New("invalid height"), err)

	// should get an error when file doesn't exist
	_, err = client.SVGToPNG("testdata/filedoesntexist.svg", 100, 100)
	require.ErrorIs(t, err, fs.ErrNotExist)

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

func loadTestImage(t *testing.T) image.Image {
	f, err := os.Open("testdata/example.png")
	require.NoError(t, err)
	defer f.Close()

	assetImage, _, err := image.Decode(f)
	require.NoError(t, err)

	return assetImage
}
