package parse_test

import (
	"image/color"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/parse"
	"github.com/stretchr/testify/require"
)

func TestHexColor(t *testing.T) {
	c := parse.HexColor("#000")
	require.Equal(t, color.RGBA{0x00, 0x00, 0x00, 0xff}, c)
	c = parse.HexColor("000")
	require.Equal(t, color.RGBA{0x00, 0x00, 0x00, 0xff}, c)

	c = parse.HexColor("#00CCAA")
	require.Equal(t, color.RGBA{0x00, 0xcc, 0xaa, 0xff}, c)
	c = parse.HexColor("00CCAA")
	require.Equal(t, color.RGBA{0x00, 0xcc, 0xaa, 0xff}, c)

	c = parse.HexColor("#0aff1299")
	require.Equal(t, color.RGBA{0x0a, 0xff, 0x12, 0x99}, c)
	c = parse.HexColor("0aff1299")
	require.Equal(t, color.RGBA{0x0a, 0xff, 0x12, 0x99}, c)

	c = parse.HexColor("")
	require.Equal(t, color.RGBA{0x00, 0x00, 0x00, 0xff}, c)
	c = parse.HexColor(")S*F)fj02fu82f")
	require.Equal(t, color.RGBA{0x00, 0x00, 0x00, 0xff}, c)
}
