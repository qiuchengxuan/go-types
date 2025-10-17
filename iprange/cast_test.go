package iprange

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustCastV4(t *testing.T) {
	v4Ranges := FromStr("0.0.0.0-0.0.0.255,255.255.255.255")
	u32Ranges := v4Ranges.MustCastV4()
	assert.Equal(t, 0, int(u32Ranges.First()))
	assert.Equal(t, math.MaxUint32, int(u32Ranges.Last()))
	assert.Equal(t, 257, int(u32Ranges.Len()))
}
