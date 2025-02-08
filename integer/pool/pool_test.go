package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	var pool Pool[int]
	pool = FromStr[int]("1")
	assert.Equal(t, int(pool.Capacity().Len()), 1)
	pool = FromStr[int]("1-2")
	assert.Equal(t, pool.Available().String(), "1-2")
	pool = FromStr[int]("1-2,4-5")
	assert.Equal(t, pool.Available().String(), "1-2,4-5")
	pool = FromStr[int]("1,3-4")
	assert.Equal(t, int(pool.Capacity().Len()), 3)
	assert.Equal(t, pool.Capacity(), pool.Available())
	pool = FromRange(1, 2)
	assert.Equal(t, pool.Available().String(), "1-2")
	pool = FromRange(1, 1)
	assert.Equal(t, pool.Available().String(), "1")
	pool = FromRange(2, 1)
	assert.Equal(t, pool.Available().String(), "")
	pool = Empty[int]()
	assert.Equal(t, pool.Available().String(), "")
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
	assert.Equal(t, x, 2)
	assert.Equal(t, pool.Available().String(), "3-5,7-8")
	x, _ = pool.Allocate(4)
	assert.Equal(t, x, 4)
	assert.Equal(t, pool.Available().String(), "3,5,7-8")
	x, _ = pool.Allocate()
	assert.Equal(t, x, 3)
	assert.Equal(t, pool.Available().String(), "5,7-8")
	x, _ = pool.Allocate()
	assert.Equal(t, x, 5)
	assert.Equal(t, pool.Available().String(), "7-8")
	x, _ = pool.Allocate(8)
	assert.Equal(t, x, 8)
	assert.Equal(t, pool.Available().String(), "7")
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
	assert.Equal(t, pool.Available().String(), "2,5-6")
	pool.Release(1)
	assert.Equal(t, pool.Available().String(), "1-2,5-6")
	pool.Release(3)
	assert.Equal(t, pool.Available().String(), "1-3,5-6")
	pool.Release(4)
	assert.Equal(t, pool.Available().String(), "1-6")
	pool.Release(7)
	assert.Equal(t, pool.Available().String(), "1-7")
	pool.Release(7)
	assert.Equal(t, pool.Available().String(), "1-7")
	pool.Release(8)
	assert.Equal(t, pool.Available().String(), "1-7")
	for i := 0; i < int(pool.Capacity().Len()); i++ {
		pool.Allocate()
	}
	assert.Equal(t, pool.Available().String(), "")
	pool.Release(2)
	pool.Allocate(2)
	pool.Release(2)
	assert.Equal(t, pool.Available().String(), "2")
	pool.Release(6)
	pool.Release(4)
	pool.Release(3)
	assert.Equal(t, pool.Available().String(), "2-4,6")
}
