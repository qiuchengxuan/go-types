package ranges

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestBinaryMarshal(t *testing.T) {
	sizeOf := int(unsafe.Sizeof(uint16(0)))
	expected := FromStr[uint16]("1-3,5-7,9-10")
	data, _ := expected.MarshalBinary()
	assert.Equal(t, expected.NumChunks()*2*sizeOf, len(data))
	var actual Ranges[uint16]
	assert.Error(t, actual.UnmarshalBinary(data[:len(data)-1]))
	assert.Error(t, actual.UnmarshalBinary(data[:len(data)-sizeOf]))
	assert.NoError(t, actual.UnmarshalBinary(data))
	assert.Equal(t, expected, actual)

	unaligned := make([]byte, len(data)+1)
	copy(unaligned[1:], data)
	assert.NoError(t, actual.UnmarshalBinary(unaligned[1:]))
	assert.Equal(t, expected, actual)

	expected = FromStr[uint16]("1")
	data, _ = expected.MarshalBinary()
	assert.Equal(t, 2*sizeOf, len(data))
	assert.NoError(t, actual.UnmarshalBinary(data))
	assert.Equal(t, expected, actual)
}
