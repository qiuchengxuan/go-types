package ranges

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	assert.Equal(t, "1-4", FromSlice([]int{1, 2, 3, 4}).String())
	assert.Equal(t, "4", FromSlice([]int{4, 3, 2, 1}).String())
	assert.Equal(t, "1,3-4", FromSlice([]int{1, 3, 2, 2, 4}).String())
	assert.Equal(t, "", FromSlice[int](nil).String())
}

func fromStr(str string) Ranges[int] {
	return FromStr[int](str)
}

func TestFromStr(t *testing.T) {
	ranges := FromStr[int]("")
	assert.Equal(t, "", ranges.String())
}

func TestMisorder(t *testing.T) {
	assert.Equal(t, "2-4,6", FromStr[int]("3,4,6,2").String())
}

func TestLength(t *testing.T) {
	u8range := FromTo[uint8](1, 0)
	assert.Equal(t, 0, int(u8range.Len()))
	u8range = FromTo[uint8](0, 0)
	assert.Equal(t, 1, int(u8range.Len()))
	u8range = FromTo[uint8](0, math.MaxUint8)
	assert.Equal(t, 256, int(u8range.Len()))
}

func TestInnerBinsearch(t *testing.T) {
	r := fromStr("1,3-5,9-10")
	index, found := r.Binsearch(0)
	assert.Equal(t, 0, index)
	assert.False(t, found)
	index, found = r.Binsearch(1)
	assert.Equal(t, 0, index)
	assert.True(t, found)
	index, found = r.Binsearch(2)
	assert.Equal(t, 1, index)
	assert.False(t, found)
	index, found = r.Binsearch(3)
	assert.Equal(t, 1, index)
	assert.True(t, found)
	index, found = r.Binsearch(7)
	assert.Equal(t, 2, index)
	assert.False(t, found)
}

func TestNoIntersection(t *testing.T) {
	assert.Equal(t, "", fromStr("2-3").Intersect(fromStr("4-5")).String())
	assert.Equal(t, "", fromStr("4-5").Intersect(fromStr("2-3")).String())
	assert.Equal(t, "", fromStr("2-3,6-7").Intersect(fromStr("4-5")).String())
	assert.Equal(t, "", fromStr("").Intersect(fromStr("4-5")).String())
	assert.Equal(t, "", fromStr("4-5").Intersect(fromStr("")).String())
	assert.Equal(t, "", fromStr("").Intersect(fromStr("")).String())
}

func TestIntersection(t *testing.T) {
	assert.Equal(t, "2,4", fromStr("2-4").Intersect(fromStr("0-2,4-5")).String())
	assert.Equal(t, "2", fromStr("2-4").Intersect(fromStr("2")).String())
	assert.Equal(t, "3", fromStr("2-4").Intersect(fromStr("3")).String())
	assert.Equal(t, "4", fromStr("2-4").Intersect(fromStr("4")).String())
	assert.Equal(t, "2-3", fromStr("0-4,6-7").Intersect(fromStr("2-3")).String())
	assert.Equal(t, "6-7", fromStr("0-4,6-8").Intersect(fromStr("6-7")).String())
	intersection := fromStr("0-4,6-8").Intersect(fromStr("0-1,6-7"))
	assert.Equal(t, "0-1,6-7", intersection.String())
	intersection = fromStr("0-1,3-4,6-7").Intersect(fromStr("0,6"))
	assert.Equal(t, "0,6", intersection.String())
	intersection = fromStr("0-1,3-4,6-7").Intersect(fromStr("1-6"))
	assert.Equal(t, "1,3-4,6", intersection.String())
}

func TestBinsearch(t *testing.T) {
	conv := func(value int, ok bool) int {
		if !ok {
			return -value - 1
		}
		return value
	}
	assert.Equal(t, -1, conv(fromStr("").Binsearch(0)))
	ranges := fromStr("1-2,5-6,8")
	assert.Equal(t, -1, conv(ranges.Binsearch(0)))
	assert.Equal(t, 0, conv(ranges.Binsearch(1)))
	assert.Equal(t, 0, conv(ranges.Binsearch(2)))
	assert.Equal(t, -2, conv(ranges.Binsearch(3)))
	assert.Equal(t, -2, conv(ranges.Binsearch(4)))
	assert.Equal(t, 2, conv(ranges.Binsearch(8)))
	assert.Equal(t, -4, conv(ranges.Binsearch(9)))
}

func TestContains(t *testing.T) {
	ranges := fromStr("1-3,5-7")
	assert.True(t, ranges.Contains(fromStr("1")))
	assert.True(t, ranges.Contains(fromStr("1-2")))
	assert.True(t, ranges.Contains(fromStr("2-3")))
	assert.True(t, ranges.Contains(fromStr("1-3")))
	assert.True(t, ranges.Contains(fromStr("5-7")))
	assert.True(t, ranges.Contains(ranges))
	assert.False(t, ranges.Contains(fromStr("0")))
	assert.False(t, ranges.Contains(fromStr("4")))
	assert.False(t, ranges.Contains(fromStr("8")))
	assert.False(t, ranges.Contains(fromStr("3-4")))
	assert.False(t, ranges.Contains(fromStr("3-5")))
	assert.False(t, ranges.Contains(fromStr("4-5")))
	assert.False(t, ranges.Contains(fromStr("1-7")))
	assert.False(t, ranges.Contains(fromStr("0-8")))
	assert.False(t, ranges.Contains(fromStr("1-3,5-7,9-10")))
	assert.False(t, ranges.Contains(fromStr("9-10")))
}

func TestIndex(t *testing.T) {
	ranges := fromStr("0-1,3-5,7")
	v, ok := ranges.Index(0)
	assert.True(t, ok)
	assert.Equal(t, 0, v)
	v, _ = ranges.Index(1)
	assert.Equal(t, 1, v)
	v, _ = ranges.Index(2)
	assert.Equal(t, 3, v)
	v, _ = ranges.Index(3)
	assert.Equal(t, 4, v)
	v, _ = ranges.Index(4)
	assert.Equal(t, 5, v)
	v, _ = ranges.Index(5)
	assert.Equal(t, 7, v)
	_, ok = ranges.Index(6)
	assert.False(t, ok)
}
