package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-types/integer/ranges"
)

func TestCreate(t *testing.T) {
	var pool Pool[int]
	pool = FromStr[int]("1")
	assert.Equal(t, 1, int(pool.Capacity().Len()))
	pool = FromStr[int]("1-2")
	assert.Equal(t, "1-2", pool.Available().String())
	pool = FromStr[int]("1-2,4-5")
	assert.Equal(t, "1-2,4-5", pool.Available().String())
	pool = FromStr[int]("1,3-4")
	assert.Equal(t, 3, int(pool.Capacity().Len()))
	assert.Equal(t, pool.Capacity(), pool.Available())
	pool = FromRange(1, 2)
	assert.Equal(t, "1-2", pool.Available().String())
	pool = FromRange(1, 1)
	assert.Equal(t, "1", pool.Available().String())
	pool = FromRange(2, 1)
	assert.Equal(t, "", pool.Available().String())
	pool = Empty[int]()
	assert.Equal(t, "", pool.Available().String())
}

func TestMalformed(t *testing.T) {
	var pool Pool[int]
	pool = FromStr[int]("2-1")
	assert.Equal(t, 0, int(pool.Capacity().Len()))
	pool = FromStr[int]("1-2,2-3")
	assert.Equal(t, 3, int(pool.Capacity().Len()))
	pool = FromStr[int]("2-3,1-2")
	assert.Equal(t, 3, int(pool.Capacity().Len()))
	pool = FromStr[int]("3-4,1-2")
	assert.Equal(t, 4, int(pool.Capacity().Len()))
	pool = FromStr[int]("1-2,4-")
	assert.Equal(t, 0, int(pool.Capacity().Len()))
	pool = FromStr[int]("foo,bar")
	assert.Equal(t, 0, int(pool.Capacity().Len()))
	pool = FromStr[int]("test")
	assert.Equal(t, 0, int(pool.Capacity().Len()))
}

func TestAllocate(t *testing.T) {
	var x int
	pool := FromStr[int]("2-5,7-8")
	x, _ = pool.Allocate()
	assert.Equal(t, 2, x)
	assert.Equal(t, "3-5,7-8", pool.Available().String())
	x, _ = pool.Allocate(4)
	assert.Equal(t, 4, x)
	assert.Equal(t, "3,5,7-8", pool.Available().String())
	x, _ = pool.Allocate()
	assert.Equal(t, 3, x)
	assert.Equal(t, "5,7-8", pool.Available().String())
	x, _ = pool.Allocate()
	assert.Equal(t, 5, x)
	assert.Equal(t, "7-8", pool.Available().String())
	x, _ = pool.Allocate(8)
	assert.Equal(t, 8, x)
	assert.Equal(t, "7", pool.Available().String())
	var ok bool
	_, ok = pool.Allocate(8)
	assert.False(t, ok)
	pool.Allocate()
	_, ok = pool.Allocate()
	assert.False(t, ok)
}

func TestRelease(t *testing.T) {
	pool := FromStr[int]("1-7")
	pool.Allocate(1)
	pool.Allocate(3)
	pool.Allocate(4)
	pool.Allocate(7)
	assert.Equal(t, "2,5-6", pool.Available().String())
	pool.Release(1)
	assert.Equal(t, "1-2,5-6", pool.Available().String())
	pool.Release(3)
	assert.Equal(t, "1-3,5-6", pool.Available().String())
	pool.Release(4)
	assert.Equal(t, "1-6", pool.Available().String())
	pool.Release(7)
	assert.Equal(t, "1-7", pool.Available().String())
	pool.Release(7)
	assert.Equal(t, "1-7", pool.Available().String())
	pool.Release(8)
	assert.Equal(t, "1-7", pool.Available().String())
	for range int(pool.Capacity().Len()) {
		pool.Allocate()
	}
	assert.Equal(t, "", pool.Available().String())
	pool.Release(2)
	pool.Allocate(2)
	pool.Release(2)
	assert.Equal(t, "2", pool.Available().String())
	pool.Release(6)
	pool.Release(4)
	pool.Release(3)
	assert.Equal(t, "2-4,6", pool.Available().String())
}

func TestShrink(t *testing.T) {
	pool := FromStr[int]("1-7")
	pool.Allocate(1)
	pool.Allocate(7)
	assert.Equal(t, "1,7", pool.InUse().String())
	pool.Reset(ranges.FromStr[int]("1-4"))
	assert.Equal(t, "1,7", pool.InUse().String())
	pool.Reset(ranges.FromStr[int]("1-7"))
	assert.True(t, pool.exceed.IsEmpty())
	assert.Equal(t, "2-6", pool.Available().String())
}

func TestOutSideModifyOrReset(t *testing.T) {
	capacity := ranges.FromStr[int]("1-7")
	pool := FromRanges(capacity)
	capacity.Assign().AddScalar(8)
	assert.Equal(t, "1-7", pool.Capacity().String())
	_, ok := pool.Allocate()
	assert.True(t, ok)
	assert.Equal(t, "1-7", pool.Capacity().String())

	pool = FromStr[int]("")
	pool.Reset(capacity)
	assert.Equal(t, "1-8", pool.Capacity().String())
	assert.Equal(t, "1-8", pool.Available().String())
	_, ok = pool.Allocate()
	assert.True(t, ok)
	assert.Equal(t, "1-8", pool.Capacity().String())
	assert.Equal(t, "2-8", pool.Available().String())
}
