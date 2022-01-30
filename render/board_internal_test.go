package render

import (
	"testing"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/stretchr/testify/assert"
)

func TestGetDirection(t *testing.T) {

	cases := []struct {
		p1   engine.Point // initial point
		p2   engine.Point // point moved to
		want string       // direction of movement
		desc string       // description of test case
	}{
		// easy cases
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 3, Y: 4}, want: "up", desc: "non-wrapped up"},
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 3, Y: 2}, want: "down", desc: "non-wrapped down"},
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 4, Y: 3}, want: "right", desc: "non-wrapped right"},
		{p1: engine.Point{X: 3, Y: 3}, p2: engine.Point{X: 2, Y: 3}, want: "left", desc: "non-wrapped left"},

		// wrapped cases
		{p1: engine.Point{X: 1, Y: 10}, p2: engine.Point{X: 1, Y: 0}, want: "up", desc: "wrapped up"},
		{p1: engine.Point{X: 1, Y: 0}, p2: engine.Point{X: 1, Y: 10}, want: "down", desc: "wrapped down"},
		{p1: engine.Point{X: 0, Y: 4}, p2: engine.Point{X: 10, Y: 4}, want: "left", desc: "wrapped left"},
		{p1: engine.Point{X: 10, Y: 0}, p2: engine.Point{X: 0, Y: 0}, want: "right", desc: "wrapped right"},
	}

	b := NewBoard(11, 11)
	for _, tc := range cases {
		got := b.getDirection(tc.p1, tc.p2)
		assert.Equal(t, tc.want, got, tc.desc)
	}
}
