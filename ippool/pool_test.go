package ippool

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qiuchengxuan/go-datastructures/integer/uint128"
	"github.com/qiuchengxuan/go-datastructures/ip"
	"github.com/qiuchengxuan/go-datastructures/iprange"
)

func TestCreate(t *testing.T) {
	var pool Pool
	pool = FromStr("::1")
	assert.Equal(t, uint128.FromPrimitive(1), pool.Capacity().Len())
	pool = FromStr("::1-::2")
	assert.Equal(t, "::1-::2", pool.Available().String())
	pool = FromStr("::1-::2,::4-::5")
	assert.Equal(t, "::1-::2,::4-::5", pool.Available().String())
	pool = FromStr("::1,::3-::4")
	assert.Equal(t, pool.Capacity().Len(), uint128.FromPrimitive(3))
	assert.Equal(t, pool.Capacity(), pool.Available())
	pool = FromRange(ip.MustParse("::1"), ip.MustParse("::2"))
	assert.Equal(t, "::1-::2", pool.Available().String())
	pool = FromRange(ip.MustParse("::1"), ip.MustParse("::1"))
	assert.Equal(t, "::1", pool.Available().String())
	pool = FromRange(ip.MustParse("::2"), ip.MustParse("::1"))
	assert.Equal(t, "", pool.Available().String())
	pool = Empty()
	assert.Equal(t, "", pool.Available().String())
}

func TestMalformed(t *testing.T) {
	var pool Pool
	pool = FromStr("::2-::1")
	assert.Equal(t, uint128.Zero(), pool.Capacity().Len())
	pool = FromStr("::1-::2,::2-::3")
	assert.Equal(t, uint128.FromPrimitive(3), pool.Capacity().Len())
	pool = FromStr("::1-::2,::4-")
	assert.Equal(t, uint128.FromPrimitive(0), pool.Capacity().Len())
	pool = FromStr("foo,bar")
	assert.Equal(t, uint128.FromPrimitive(0), pool.Capacity().Len())
	pool = FromStr("test")
	assert.Equal(t, uint128.FromPrimitive(0), pool.Capacity().Len())
}

func TestAllocate(t *testing.T) {
	var x ip.IP
	pool := FromStr("::2-::5,::7-::8")
	x, _ = pool.Allocate()
	assert.Equal(t, x, ip.MustParse("::2"))
	assert.Equal(t, "::3-::5,::7-::8", pool.Available().String())
	x, _ = pool.Allocate(ip.MustParse("::4"))
	assert.Equal(t, x, ip.MustParse("::4"))
	assert.Equal(t, "::3,::5,::7-::8", pool.Available().String())
	x, _ = pool.Allocate()
	assert.Equal(t, x, ip.MustParse("::3"))
	assert.Equal(t, "::5,::7-::8", pool.Available().String())
	x, _ = pool.Allocate()
	assert.Equal(t, x, ip.MustParse("::5"))
	assert.Equal(t, "::7-::8", pool.Available().String())
	x, _ = pool.Allocate(ip.MustParse("::8"))
	assert.Equal(t, x, ip.MustParse("::8"))
	assert.Equal(t, "::7", pool.Available().String())
	var ok bool
	_, ok = pool.Allocate(ip.MustParse("::8"))
	assert.False(t, ok)
	pool.Allocate()
	_, ok = pool.Allocate()
	assert.False(t, ok)
}

func TestRelease(t *testing.T) {
	pool := FromStr("::1-::7")
	pool.Allocate(ip.MustParse("::1"))
	pool.Allocate(ip.MustParse("::3"))
	pool.Allocate(ip.MustParse("::4"))
	pool.Allocate(ip.MustParse("::7"))
	assert.Equal(t, "::2,::5-::6", pool.Available().String())
	pool.Release(ip.MustParse("::1"))
	assert.Equal(t, "::1-::2,::5-::6", pool.Available().String())
	pool.Release(ip.MustParse("::3"))
	assert.Equal(t, "::1-::3,::5-::6", pool.Available().String())
	pool.Release(ip.MustParse("::4"))
	assert.Equal(t, "::1-::6", pool.Available().String())
	pool.Release(ip.MustParse("::7"))
	assert.Equal(t, "::1-::7", pool.Available().String())
	pool.Release(ip.MustParse("::7"))
	assert.Equal(t, "::1-::7", pool.Available().String())
	pool.Release(ip.MustParse("::8"))
	assert.Equal(t, "::1-::7", pool.Available().String())
	for i := uint128.Zero(); i.LessThan(pool.Capacity().Len()); i = i.AddU64(1) {
		pool.Allocate()
	}
	assert.Equal(t, "", pool.Available().String())
	pool.Release(ip.MustParse("::2"))
	pool.Allocate(ip.MustParse("::2"))
	pool.Release(ip.MustParse("::2"))
	assert.Equal(t, "::2", pool.Available().String())
	pool.Release(ip.MustParse("::6"))
	pool.Release(ip.MustParse("::4"))
	pool.Release(ip.MustParse("::3"))
	assert.Equal(t, "::2-::4,::6", pool.Available().String())
}

func TestShrink(t *testing.T) {
	pool := FromStr("::1-::7")
	pool.Allocate(ip.MustParse("::1"))
	pool.Allocate(ip.MustParse("::7"))
	assert.Equal(t, "::1,::7", pool.InUse().String())
	pool.Reset(iprange.FromStr("::1-::4"))
	assert.Equal(t, "::1,::7", pool.InUse().String())
	pool.Reset(iprange.FromStr("::1-::7"))
	assert.True(t, pool.exceed.IsEmpty())
	assert.Equal(t, "::2-::6", pool.Available().String())
}
