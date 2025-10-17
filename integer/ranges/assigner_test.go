package ranges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func refFromStr(text string) *Ranges[int] { return FromStr[int](text).Ref() }

func TestAssignAdd(t *testing.T) {
	// Add to empty
	assert.Equal(t, "2-3", Empty[int]().Ranges().Ref().Assign().Add(fromStr("2-3")).String())
	// Add empty
	assert.Equal(t, "2-3", refFromStr("2-3").Assign().Add(nil).String())
	// Add adjacency
	assert.Equal(t, "0-3", refFromStr("0-1").Assign().Add(fromStr("2-3")).String())
	// Add last intersection
	assert.Equal(t, "0-1,3-5", refFromStr("0-1,3-4").Assign().Add(fromStr("4-5")).String())
	// Add greater
	assert.Equal(t, "0-1,3-4,6-7", refFromStr("0-1,3-4").Assign().Add(fromStr("6-7")).String())
	// Add contains
	assert.Equal(t, "0-9", refFromStr("0-9").Assign().Add(fromStr("2-3")).String())
	// Add contained
	assert.Equal(t, "0-9", refFromStr("2-3").Assign().Add(fromStr("0-9")).String())
	// Add overflow
	one := Of[uint8](1).Ranges()
	assert.Equal(t, "1-255", FromStr[uint8]("2-255").Ref().Assign().Add(one).String())
}

func TestAddScalar(t *testing.T) {
	// Add lower adjacency
	assert.Equal(t, "0-3", fromStr("1-3").AddScalar(0).String())
	// Add lower no adjacency
	assert.Equal(t, "0,2-3", fromStr("2-3").AddScalar(0).String())
	// Add existing
	assert.Equal(t, "1-3", fromStr("1-3").AddScalar(2).String())
	// Add upper adjacency
	assert.Equal(t, "1-4", fromStr("1-3").AddScalar(4).String())
	// Add upper no adjacency
	assert.Equal(t, "1-3,5", fromStr("1-3").AddScalar(5).String())
	// Add adjacency on both sides
	assert.Equal(t, "1-7", fromStr("1-3,5-7").AddScalar(4).String())
	// Add adjacency on both sides
	assert.Equal(t, "1-7,9-10", fromStr("1-3,5-7,9-10").AddScalar(4).String())
	// Add in middle no adjacency
	assert.Equal(t, "1-3,5,7-9", fromStr("1-3,7-9").AddScalar(5).String())
	// Add to empty
	assert.Equal(t, "1", Empty[int]().Ranges().AddScalar(1).String())
	// Add to maximum
	assert.Equal(t, "1-255", FromStr[uint8]("1-255").AddScalar(1).String())
}

func TestSubScalar(t *testing.T) {
	assert.Equal(t, "", fromStr("").SubScalar(0).String())
	assert.Equal(t, "1-3", fromStr("0-3").SubScalar(0).String())
	assert.Equal(t, "0-2", fromStr("0-3").SubScalar(3).String())
	assert.Equal(t, "0,2-3", fromStr("0-3").SubScalar(1).String())
	assert.Equal(t, "0-2,5-7", fromStr("0-3,5-7").SubScalar(3).String())
	assert.Equal(t, "0-3,6-7", fromStr("0-3,5-7").SubScalar(5).String())
	assert.Equal(t, "0-3,5-7", fromStr("0-3,5-7").SubScalar(4).String())
	assert.Equal(t, "0-3,5-7", fromStr("0-3,5-7").SubScalar(-1).String())
	assert.Equal(t, "0-3,5-7", fromStr("0-3,5-7").SubScalar(8).String())
	assert.Equal(t, "0-3,7-9", fromStr("0-3,5,7-9").SubScalar(5).String())
}
