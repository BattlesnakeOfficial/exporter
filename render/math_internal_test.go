package render

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbs(t *testing.T) {
	assert.Equal(t, 1, abs(-1))
	assert.Equal(t, 1, abs(1))
	assert.Equal(t, 0, abs(0))
	assert.Equal(t, 100, abs(-100))
	assert.Equal(t, math.MaxInt, abs(-math.MaxInt))
	assert.Equal(t, math.MaxInt, abs(math.MaxInt))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 0, min(0, 1))
	assert.Equal(t, 1, min(1, 1))
	assert.Equal(t, 1, min(2, 1))
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, -1, min(-1, 1))
}
