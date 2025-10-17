package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	ip := From(net.ParseIP("1.1.1.1"))
	assert.True(t, ip.IsV4())
	assert.Equal(t, "1.1.1.1", IPv4(ip.U32()).IP().String())

	assert.True(t, From(net.ParseIP("0.0.0.0")).IsV4())
	assert.False(t, From(net.ParseIP("::")).IsV4())
}
