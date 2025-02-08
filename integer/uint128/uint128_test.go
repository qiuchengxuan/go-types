package uint128

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	u := FromPrimitive(uint64(math.MaxUint64))
	assert.Equal(t, U128{1, 0}, u.AddU64(1))
	assert.Equal(t, U128{1, 0}, u.Add(FromPrimitive(uint64(1))))

	u = U128{uint64(math.MaxUint64), uint64(math.MaxUint64)}
	assert.Equal(t, U128{0, 0}, u.AddU64(1))
	assert.Equal(t, U128{0, 0}, u.Add(FromPrimitive(uint64(1))))
}
