package ranges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegerRange(t *testing.T) {
	ranged := FromTo(1, 2)
	assert.Equal(t, "1-2", ranged.String())
	point := Of(1)
	assert.Equal(t, "1", point.String())
	emptyIntegerRange := Empty[int]()
	assert.Equal(t, 0, int(emptyIntegerRange.Len()))
	assert.Equal(t, "", emptyIntegerRange.String())
}

func TestIntersect(t *testing.T) {
	// no intersection
	intersect := FromTo(0, 1).Intersect(FromTo(2, 3))
	assert.Equal(t, "", intersect.String())
	// left intersection
	intersect = FromTo(0, 2).Intersect(FromTo(2, 3))
	assert.Equal(t, "2", intersect.String())
	// right intersection
	intersect = FromTo(2, 3).Intersect(FromTo(0, 2))
	assert.Equal(t, "2", intersect.String())
	// contained intersection
	intersect = FromTo(2, 3).Intersect(FromTo(0, 4))
	assert.Equal(t, "2-3", intersect.String())
}
