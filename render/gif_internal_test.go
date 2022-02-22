package render

import (
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetColorCounts(t *testing.T) {
	cases := []struct {
		name string
		want usageList
	}{
		{
			name: "sample1.png",
			want: usageList{
				{
					Key:   parse.HexColor("#fff"),
					Value: 400,
				},
			},
		},
		{
			name: "sample2.png",
			want: usageList{
				{
					Key:   parse.HexColor("#fff"),
					Value: 300,
				},
				{
					Key:   parse.HexColor("#000"),
					Value: 100,
				},
			},
		},
		{
			name: "sample3.png",
			want: usageList{
				{
					Key:   parse.HexColor("#fff"),
					Value: 50,
				},
				{
					Key:   parse.HexColor("#000"),
					Value: 100,
				},
				{
					Key:   parse.HexColor("#00ff00"),
					Value: 50,
				},
				{
					Key:   parse.HexColor("#f00"),
					Value: 100,
				},
				{
					Key:   parse.HexColor("#0011ff"),
					Value: 100,
				},
			},
		},
	}

	for _, tc := range cases {
		assert.ElementsMatch(t, tc.want, getColorCounts(loadSample(tc.name)), tc.name)
	}
}

func TestBuildGIFPallete(t *testing.T) {
	// How this test works...
	// I made a special image that has more than 256 colours in it to ensure the pallete building caps it.
	// I also made a square of black in the middle of the image which should be the most dominant colour by far.
	// So this test should validate that the pallete caps max colours to the GIF limit and properly orders colours.
	img := loadSample("sample4.png")
	pallete := buildGIFPallete(img)
	require.NotEmpty(t, pallete, "the pallete should not be empty")
	assert.Len(t, pallete, GIFMaxColorsPerFrame, "the pallete should not be larger than a GIF can support")
	require.Equal(t, parse.HexColor("#000"), pallete[0], "the black square should be the most dominant colour and be first in the pallete")
}

func loadSample(name string) image.Image {
	f, err := os.Open(fmt.Sprintf("testdata/%s", name))
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
