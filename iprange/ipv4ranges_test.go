package iprange

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPv4Ranges(t *testing.T) {
	ipranges := FromStr("0.0.0.1-0.0.0.4").V4
	sum := 0
	ipranges.Foreach(func(ip net.IP) { sum += int(ip[len(ip)-1]) })
	assert.Equal(t, sum, 10)

	ipranges = ipranges.AddIP(net.ParseIP("0.0.0.5"))
	assert.Equal(t, "0.0.0.1-0.0.0.5", ipranges.String())
	assert.True(t, ipranges.Contains(net.ParseIP("0.0.0.2")))
	assert.Equal(t, "0.0.0.2", ipranges.Index(1).String())
	assert.Nil(t, ipranges.Index(100))
	assert.Equal(t, "0.0.0.1", ipranges.Pop().String())
	assert.Equal(t, "0.0.0.2-0.0.0.5", ipranges.String())
}
