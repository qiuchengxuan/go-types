package iprange

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPv4FromStr(t *testing.T) {
	assert.Equal(t, "", FromStr("").String())
	assert.Equal(t, "0.0.0.0", FromStr("0.0.0.0").String())
	assert.Equal(t, "1.1.1.1", FromStr("1.1.1.1").String())
	assert.Equal(t, "1.1.1.1-1.1.1.2", FromStr("1.1.1.1,1.1.1.2").String())
	expected := "1.1.1.1-1.1.1.255"
	assert.Equal(t, expected, FromStr(expected).String())
	assert.Equal(t, "1.1.1.5", FromStr("1.1.1.5,1.1.1.5").String())
	expected = "1.1.1.1-1.1.1.255,1.1.2.1-1.1.2.255"
	assert.Equal(t, expected, FromStr(expected).String())
	assert.Equal(t, "", FromStr("1.1.1.1-1.1.1.255-1.1.2.1").String())
	assert.Equal(t, "", FromStr("1.1.1-1.1.1.255").String())
	assert.Equal(t, "", FromStr("1.1.1.1-1.1.1.").String())
	assert.Equal(t, "1.1.1.5,1.1.1.10-1.1.1.12", FromStr("1.1.1.10-1.1.1.12,1.1.1.5").String())
}

func TestIPv6FromStr(t *testing.T) {
	assert.Equal(t, "", FromStr("").String())
	assert.Equal(t, "::", FromStr("::").String())
	assert.Equal(t, "fd01::1", FromStr("fd01::1").String())
	expected := "fd01::1-fd01::f"
	assert.Equal(t, expected, FromStr(expected).String())
	assert.Equal(t, "::1-::2", FromStr("::1,::2").String())
	assert.Equal(t, "::1,::3", FromStr("::1,::3").String())
}

func TestMixedFromStr(t *testing.T) {
	expected := "192.168.1.1-192.168.1.255,fd01::1-fd01::f"
	assert.Equal(t, expected, FromStr(expected).String())
}

func TestFromIPNet4(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("192.168.1.0/24")
	assert.Equal(t, "192.168.1.0-192.168.1.255", FromIPNet(*cidr).String())

	_, cidr, _ = net.ParseCIDR("0.0.0.0/0")
	assert.Equal(t, "0.0.0.0-255.255.255.255", FromIPNet(*cidr).String())

	_, cidr, _ = net.ParseCIDR("0.0.0.0/32")
	assert.Equal(t, "", FromIPNet(*cidr).String())
}

func TestFromIPNet6(t *testing.T) {
	_, cidr, _ := net.ParseCIDR("fd01::/64")
	assert.Equal(t, "fd01::-fd01::ffff:ffff:ffff:ffff", FromIPNet(*cidr).String())

	_, cidr, _ = net.ParseCIDR("::/0")
	expected := "::-ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"
	assert.Equal(t, expected, FromIPNet(*cidr).String())

	_, cidr, _ = net.ParseCIDR("::/127")
	assert.Equal(t, "::-::1", FromIPNet(*cidr).String())

	_, cidr, _ = net.ParseCIDR("::/128")
	assert.Equal(t, "", FromIPNet(*cidr).String())
}

func TestFromNetIPs(t *testing.T) {
	expected := FromStr("1.1.1.1")
	data, _ := expected.MarshalBinary()
	var actual IPRanges
	assert.NoError(t, actual.UnmarshalBinary(data))
	assert.Equal(t, expected, actual)

	expected = FromStr("1.1.1.1-1.1.1.4,fd01::1-fd01::4")
	data, _ = expected.MarshalBinary()
	assert.NoError(t, actual.UnmarshalBinary(data))
	assert.Equal(t, expected, actual)
}
