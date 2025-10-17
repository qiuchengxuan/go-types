package ranges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
