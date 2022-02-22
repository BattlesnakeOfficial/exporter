package imagetest

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func Equal(t *testing.T, a, b image.Image) {
	require.Equal(t, a.Bounds().Min.X, b.Bounds().Min.X)
	require.Equal(t, a.Bounds().Min.Y, b.Bounds().Min.Y)
	require.Equal(t, a.Bounds().Max.X, b.Bounds().Max.X)
	require.Equal(t, a.Bounds().Max.Y, b.Bounds().Max.Y)

	for x := 0; x < a.Bounds().Max.X; x++ {
		for y := 0; y < a.Bounds().Max.Y; y++ {
			c1 := a.At(x, y)
			c2 := b.At(x, y)
			sameColor(t, c1, c2, x, y)
		}
	}
}

func sameColor(t *testing.T, c1, c2 color.Color, x, y int) {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	require.True(t, r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2, fmt.Sprintf("%v != %v at pixel (%d,%d)", c1, c2, x, y))
}
