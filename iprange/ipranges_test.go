package iprange

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qiuchengxuan/go-types/integer/uint128"
	"github.com/qiuchengxuan/go-types/ip"
)

func TestRanges(t *testing.T) {
	addr := ip.From(net.ParseIP("1.1.1.1"))
	assert.Nil(t, FromTo(addr.Add(2), addr).Ranges())
	assert.Nil(t, Empty().Ranges())
}

func TestIPv4Ranges(t *testing.T) {
	ipranges := FromStr("0.0.0.1-0.0.0.4")
	assert.True(t, ipranges.V4Only())

	sum := 0
	for ip := range ipranges.Iter() {
		sum += int(ip.U32())
	}
	assert.Equal(t, 10, sum)

	ipranges = ipranges.AddIP(ip.MustParse("0.0.0.5"))
	assert.Equal(t, "0.0.0.1-0.0.0.5", ipranges.String())
	assert.True(t, ipranges.Has(ip.MustParse("0.0.0.2")))
	assert.Equal(t, "0.0.0.2", ipranges.Index(1).String())
	ip, _ := ipranges.Assign().Pop()
	assert.Equal(t, "0.0.0.1", ip.String())
	assert.Equal(t, "0.0.0.2-0.0.0.5", ipranges.String())
}

func TestIPv6Ranges(t *testing.T) {
	ipranges := FromStr("::3-::5")
	sum := 0
	for ip := range ipranges.Iter() {
		sum += int(ip[len(ip)-1])
	}
	assert.Equal(t, 12, sum)

	assert.Equal(t, "::1,::3-::5", ipranges.AddIP(ip.MustParse("::1")).String())
	assert.Equal(t, "::2-::5", ipranges.AddIP(ip.MustParse("::2")).String())
	assert.Equal(t, "::3-::5", ipranges.AddIP(ip.MustParse("::3")).String())
	assert.Equal(t, "::3-::5", ipranges.AddIP(ip.MustParse("::4")).String())
	assert.Equal(t, "::3-::6", ipranges.AddIP(ip.MustParse("::6")).String())
	assert.Equal(t, "::3-::5,::7", ipranges.AddIP(ip.MustParse("::7")).String())

	ipranges = FromStr("::3-::5,::7-::8")
	assert.Equal(t, "::3-::8", ipranges.AddIP(ip.MustParse("::6")).String())

	ipranges = FromStr("::3-::5,::9")
	assert.Equal(t, "::3-::5,::7,::9", ipranges.AddIP(ip.MustParse("::7")).String())

	ipranges = FromStr("::1-::3")
	assert.Equal(t, "::1-::3", ipranges.SubIP(ip.MustParse("::")).String())
	assert.Equal(t, "::2-::3", ipranges.SubIP(ip.MustParse("::1")).String())
	assert.Equal(t, "::1,::3", ipranges.SubIP(ip.MustParse("::2")).String())
	assert.Equal(t, "::1-::2", ipranges.SubIP(ip.MustParse("::3")).String())
	assert.Equal(t, "::1-::3", ipranges.SubIP(ip.MustParse("::4")).String())

	assert.Equal(t, "::1,::5", FromStr("::1,::3,::5").SubIP(ip.MustParse("::3")).String())
}

func TestTypeCast(t *testing.T) {
	ipranges := FromStr("8000::1-8000::4")
	cast, overflow := ipranges.Cast(ip.MustParse("::"))
	assert.Equal(t, "", cast.String())
	assert.True(t, overflow)

	cast, overflow = ipranges.Cast(ip.MustParse("::1"))
	assert.Equal(t, "", cast.String())
	assert.True(t, overflow)

	base := ip.IP(uint128.FromPrimitive(1).Shl(127))
	cast, overflow = ipranges.Cast(base.Sub(1))
	assert.Equal(t, "2-5", cast.String())
	assert.False(t, overflow)

	cast, overflow = ipranges.Cast(base.Sub(math.MaxUint64).Add(3))
	expected := fmt.Sprintf("%d-%d", uint64(math.MaxUint64)-2, uint64(math.MaxUint64))
	assert.Equal(t, expected, cast.String())
	assert.True(t, overflow)

	cast, overflow = ipranges.Cast(base)
	assert.Equal(t, "1-4", cast.String())
	assert.False(t, overflow)

	cast, overflow = ipranges.Cast(base.Add(1))
	assert.Equal(t, "0-3", cast.String())
	assert.False(t, overflow)

	cast, overflow = ipranges.Cast(ip.IP(uint128.FromPrimitive(3).Shl(126)))
	assert.Equal(t, "", cast.String())
	assert.True(t, overflow)

	ipranges = FromStr("::1-::3,::5-::7")
	cast, overflow = ipranges.Cast(ip.FromPrimitive(1))
	assert.Equal(t, "0-2,4-6", cast.String())
	assert.False(t, overflow)

	cast, overflow = ipranges.Cast(ip.FromPrimitive(2))
	assert.Equal(t, "0-1,3-5", cast.String())
	assert.True(t, overflow)

	cast, overflow = ipranges.Cast(ip.FromPrimitive(4))
	assert.Equal(t, "1-3", cast.String())
	assert.True(t, overflow)

	cast, overflow = ipranges.Cast(ip.FromPrimitive(6))
	assert.Equal(t, "0-1", cast.String())
	assert.True(t, overflow)

	cast, overflow = ipranges.Cast(ip.FromPrimitive(8))
	assert.Equal(t, "", cast.String())
	assert.True(t, overflow)
}

func TestBinsearch(t *testing.T) {
	value, ok := FromStr("").Binsearch(ip.MustParse("::"))
	assert.Equal(t, 0, value)
	assert.False(t, ok)
	ranges := FromStr("::1-::2,::5-::6,::8")

	value, ok = ranges.Binsearch(ip.MustParse("::"))
	assert.Equal(t, 0, value)
	assert.False(t, ok)

	value, ok = ranges.Binsearch(ip.MustParse("::1"))
	assert.Equal(t, 0, value)
	assert.True(t, ok)

	value, ok = ranges.Binsearch(ip.MustParse("::2"))
	assert.Equal(t, 0, value)
	assert.True(t, ok)

	value, ok = ranges.Binsearch(ip.MustParse("::3"))
	assert.Equal(t, 1, value)
	assert.False(t, ok)

	value, ok = ranges.Binsearch(ip.MustParse("::4"))
	assert.Equal(t, 1, value)
	assert.False(t, ok)

	value, ok = ranges.Binsearch(ip.MustParse("::8"))
	assert.Equal(t, 2, value)
	assert.True(t, ok)

	value, ok = ranges.Binsearch(ip.MustParse("::9"))
	assert.Equal(t, 3, value)
	assert.False(t, ok)
}

func TestBinaryMarshal(t *testing.T) {
	expected := FromStr("::1-::3,::5-::7")
	data, _ := expected.MarshalBinary()
	var actual IPRanges
	assert.Error(t, actual.UnmarshalBinary(data[:len(data)-1]))
	assert.NoError(t, actual.UnmarshalBinary(data))
	assert.Equal(t, expected, actual)
}
