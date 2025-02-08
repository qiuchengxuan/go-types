package ranges

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCast(t *testing.T) {
	// upcast
	u8ranges := FromTo[uint8](0, math.MaxUint8).Plural()
	u16ranges, overflow := Cast[uint8, uint16](u8ranges)
	assert.False(t, overflow)
	assert.Equal(t, "0-255", u16ranges.String())

	u8ranges = FromTo[uint8](0, math.MaxUint8).Plural()
	i16ranges, overflow := Cast[uint8, int16](u8ranges)
	assert.False(t, overflow)
	assert.Equal(t, "0-255", i16ranges.String())

	i8ranges := FromTo[int8](math.MinInt8, math.MaxInt8).Plural()
	i16ranges, overflow = Cast[int8, int16](i8ranges)
	assert.False(t, overflow)
	assert.Equal(t, "-128-127", i16ranges.String())

	// u8 -> i8
	u8ranges = FromTo[uint8](0, math.MaxUint8).Plural()
	i8ranges, overflow = Cast[uint8, int8](u8ranges)
	assert.True(t, overflow)
	assert.Equal(t, "0-127", i8ranges.String())

	u8ranges = FromTo[uint8](0, math.MaxInt8).Plural()
	i8ranges, overflow = Cast[uint8, int8](u8ranges)
	assert.False(t, overflow)
	assert.Equal(t, "0-127", i8ranges.String())

	u8ranges = FromStr[uint8]("0-127,129-255")
	i8ranges, overflow = Cast[uint8, int8](u8ranges)
	assert.True(t, overflow)
	assert.Equal(t, "0-127", i8ranges.String())

	// i8 -> u8
	i8ranges = FromTo[int8](math.MinInt8, -1).Plural()
	u8ranges, overflow = Cast[int8, uint8](i8ranges)
	assert.True(t, overflow)
	assert.Equal(t, "", u8ranges.String())

	i8ranges = FromTo[int8](math.MinInt8, math.MaxInt8).Plural()
	u8ranges, overflow = Cast[int8, uint8](i8ranges)
	assert.True(t, overflow)
	assert.Equal(t, "0-127", u8ranges.String())

	i8ranges = FromTo[int8](0, math.MaxInt8).Plural()
	u8ranges, overflow = Cast[int8, uint8](i8ranges)
	assert.False(t, overflow)
	assert.Equal(t, "0-127", u8ranges.String())

	// downcast
	// u16 -> u8
	u16ranges = FromTo[uint16](0, math.MaxUint16).Plural()
	u8ranges, overflow = Cast[uint16, uint8](u16ranges)
	assert.True(t, overflow)
	assert.Equal(t, "0-255", u8ranges.String())

	// u16 -> i8
	u16ranges = FromTo[uint16](0, math.MaxUint16).Plural()
	i8ranges, overflow = Cast[uint16, int8](u16ranges)
	assert.True(t, overflow)
	assert.Equal(t, "0-127", i8ranges.String())

	// i16 -> i8
	i16ranges = FromTo[int16](math.MinInt16, math.MaxInt16).Plural()
	i8ranges, overflow = Cast[int16, int8](i16ranges)
	assert.True(t, overflow)
	assert.Equal(t, "-128-127", i8ranges.String())

	// i16 -> u8
	i16ranges = FromTo[int16](math.MinInt16, math.MaxInt16).Plural()
	u8ranges, overflow = Cast[int16, uint8](i16ranges)
	assert.True(t, overflow)
	assert.Equal(t, "0-255", u8ranges.String())
}
