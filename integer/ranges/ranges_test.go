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

func TestAddScalar(t *testing.T) {
	// Add lower adjacency
	assert.Equal(t, "0-3", fromStr("1-3").Add1(0).String())
	// Add lower no adjacency
	assert.Equal(t, "0,2-3", fromStr("2-3").Add1(0).String())
	// Add existing
	assert.Equal(t, "1-3", fromStr("1-3").Add1(2).String())
	// Add upper adjacency
	assert.Equal(t, "1-4", fromStr("1-3").Add1(4).String())
	// Add upper no adjacency
	assert.Equal(t, "1-3,5", fromStr("1-3").Add1(5).String())
	// Add adjacency on both sides
	assert.Equal(t, "1-7", fromStr("1-3,5-7").Add1(4).String())
	// Add adjacency on both sides
	assert.Equal(t, "1-7,9-10", fromStr("1-3,5-7,9-10").Add1(4).String())
	// Add in middle no adjacency
	assert.Equal(t, "1-3,5,7-9", fromStr("1-3,7-9").Add1(5).String())
	// Add to empty
	assert.Equal(t, "1", Empty[int]().Plural().Add1(1).String())
	// Add to maximum
	assert.Equal(t, "1-255", FromStr[uint8]("1-255").Add1(1).String())
}

func TestRemove(t *testing.T) {
	assert.Equal(t, "", fromStr("").Sub1(0).String())
	assert.Equal(t, "1-3", fromStr("0-3").Sub1(0).String())
	assert.Equal(t, "0-2", fromStr("0-3").Sub1(3).String())
	assert.Equal(t, "0,2-3", fromStr("0-3").Sub1(1).String())
	assert.Equal(t, "0-2,5-7", fromStr("0-3,5-7").Sub1(3).String())
	assert.Equal(t, "0-3,6-7", fromStr("0-3,5-7").Sub1(5).String())
	assert.Equal(t, "0-3,5-7", fromStr("0-3,5-7").Sub1(4).String())
	assert.Equal(t, "0-3,5-7", fromStr("0-3,5-7").Sub1(-1).String())
	assert.Equal(t, "0-3,5-7", fromStr("0-3,5-7").Sub1(8).String())
	assert.Equal(t, "0-3,7-9", fromStr("0-3,5,7-9").Sub1(5).String())
}

func TestInnerBinsearch(t *testing.T) {
	r := fromStr("1,3-5,9-10")
	index, found := r.binsearch(0)
	assert.Equal(t, 0, index)
	assert.False(t, found)
	index, found = r.binsearch(1)
	assert.Equal(t, 0, index)
	assert.True(t, found)
	index, found = r.binsearch(2)
	assert.Equal(t, 1, index)
	assert.False(t, found)
	index, found = r.binsearch(3)
	assert.Equal(t, 1, index)
	assert.True(t, found)
	index, found = r.binsearch(7)
	assert.Equal(t, 2, index)
	assert.False(t, found)
}

func TestAdd(t *testing.T) {
	// Add adjacency
	assert.Equal(t, "0-3", fromStr("0-1").Add(fromStr("2-3")).String())
	// Add interlock
	assert.Equal(t, "0-9", fromStr("0-1,4-5,8-9").Add(fromStr("2-3,6-7")).String())
	// Add orthognonal
	assert.Equal(t, "0-1,3-6,8-9", fromStr("0-1,8-9").Add(fromStr("3-6")).String())
	// Add intersection
	assert.Equal(t, "0-7", fromStr("0-5").Add(fromStr("4-7")).String())
	assert.Equal(t, "0-1,3,5-9", fromStr("3,5-9").Add(fromStr("0-1,3")).String())
	// Add contains
	assert.Equal(t, "0-9", fromStr("0-9").Add(fromStr("2-3")).String())
	assert.Equal(t, "1-255", FromStr[uint8]("1-255").Add(FromStr[uint8]("10-20")).String())
	// Add contained
	assert.Equal(t, "0-9", fromStr("2-3").Add(fromStr("0-9")).String())
	// Add identical
	assert.Equal(t, "1", fromStr("1").Add(fromStr("1")).String())
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

func TestSub(t *testing.T) {
	remain := fromStr("0-9").Sub(fromStr("1")).Sub(fromStr("4")).Sub(fromStr("0"))
	assert.Equal(t, "2-3,5-9", remain.String())
	// Sub seperated
	assert.Equal(t, "0-3,7-9", fromStr("0-4,6-9").Sub(fromStr("4-6")).String())
	// Sub middle
	assert.Equal(t, "1,7", fromStr("1-7").Sub(fromStr("2-6")).String())
	// Remain overlap
	assert.Equal(t, "", fromStr("2-6").Sub(fromStr("1-7")).String())
	// Upper overlap
	assert.Equal(t, "0-1,4-5", fromStr("0-65535").Sub(fromStr("2-3,6-65535")).String())
	// Lower parital intersection
	assert.Equal(t, "6-7", fromStr("4-7").Sub(fromStr("1-5")).String())
	// Parital intersection on both sides
	assert.Equal(t, "4-5,11", fromStr("4-7,9-11").Sub(fromStr("6-10")).String())
	// Upper no intersection
	assert.Equal(t, "1-3", fromStr("1-3").Sub(fromStr("4-6")).String())
	// Lower no intersection
	assert.Equal(t, "4-6", fromStr("4-6").Sub(fromStr("1-3")).String())
}

func TestBinsearch(t *testing.T) {
	assert.Equal(t, -1, fromStr("").Binsearch(0))
	ranges := fromStr("1-2,5-6,8")
	assert.Equal(t, -1, ranges.Binsearch(0))
	assert.Equal(t, 0, ranges.Binsearch(1))
	assert.Equal(t, 1, ranges.Binsearch(2))
	assert.Equal(t, -3, ranges.Binsearch(3))
	assert.Equal(t, -3, ranges.Binsearch(4))
	assert.Equal(t, 4, ranges.Binsearch(8))
	assert.Equal(t, -5, ranges.Binsearch(9))
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

func TestLimit(t *testing.T) {
	ranges := fromStr("1-9")
	assert.Equal(t, "1-9", ranges.Limit(0, 10).String())
	assert.Equal(t, "1-9", ranges.Limit(1, 9).String())
	assert.Equal(t, "2-9", ranges.Limit(2, 9).String())
	assert.Equal(t, "1-8", ranges.Limit(1, 8).String())
	assert.Equal(t, "2-8", ranges.Limit(2, 8).String())

	ranges = fromStr("1-2,5-6,8-9")
	assert.Equal(t, "5-6,8-9", ranges.Limit(3, 10).String())
	assert.Equal(t, "1-2,5-6", ranges.Limit(0, 7).String())
	assert.Equal(t, "2,5-6,8-9", ranges.Limit(2, 10).String())
	assert.Equal(t, "1-2,5-6,8", ranges.Limit(0, 8).String())

	ranges = Empty[int]().Plural()
	assert.Equal(t, "", ranges.Limit(0, 10).String())
}
