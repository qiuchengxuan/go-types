package ranges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	ranges := fromStr("1-9")
	assert.Equal(t, "1-9", ranges.Limit(FromTo(0, 10)).String())
	assert.Equal(t, "1-9", ranges.Limit(FromTo(1, 9)).String())
	assert.Equal(t, "2-9", ranges.Limit(FromTo(2, 9)).String())
	assert.Equal(t, "1-8", ranges.Limit(FromTo(1, 8)).String())
	assert.Equal(t, "2-8", ranges.Limit(FromTo(2, 8)).String())

	ranges = fromStr("1-2,5-6,8-9")
	assert.Equal(t, "5-6,8-9", ranges.Limit(FromTo(3, 10)).String())
	assert.Equal(t, "1-2,5-6", ranges.Limit(FromTo(0, 7)).String())
	assert.Equal(t, "2,5-6,8-9", ranges.Limit(FromTo(2, 10)).String())
	assert.Equal(t, "1-2,5-6,8", ranges.Limit(FromTo(0, 8)).String())

	ranges = Empty[int]().Ranges()
	assert.Equal(t, "", ranges.Limit(FromTo(0, 10)).String())
}
