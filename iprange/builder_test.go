package iprange

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromIPNet(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("1.1.1.0/24")
	assert.Equal(t, "1.1.1.1-1.1.1.254", FromIPNet(*ipNet, true).String())
	assert.Equal(t, "1.1.1.0-1.1.1.255", FromIPNet(*ipNet, false).String())
	_, ipNet, _ = net.ParseCIDR("1.1.1.1/32")
	assert.Equal(t, "", FromIPNet(*ipNet, true).String())
	assert.Equal(t, "1.1.1.1", FromIPNet(*ipNet, false).String())
	_, ipNet, _ = net.ParseCIDR("1.1.1.0/31")
	assert.Equal(t, "", FromIPNet(*ipNet, true).String())
	assert.Equal(t, "1.1.1.0-1.1.1.1", FromIPNet(*ipNet, false).String())
	_, ipNet, _ = net.ParseCIDR("1.1.1.0/30")
	assert.Equal(t, "1.1.1.1-1.1.1.2", FromIPNet(*ipNet, true).String())
	_, ipNet, _ = net.ParseCIDR("::/124")
	assert.Equal(t, "::1-::f", FromIPNet(*ipNet, true).String())
	_, ipNet, _ = net.ParseCIDR("::/127")
	assert.Equal(t, "::1", FromIPNet(*ipNet, true).String())
	_, ipNet, _ = net.ParseCIDR("::/128")
	assert.Equal(t, "", FromIPNet(*ipNet, true).String())
	assert.Equal(t, "::", FromIPNet(*ipNet, false).String())
}
