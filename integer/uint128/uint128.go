package uint128

import (
	"encoding/binary"
	"math/bits"

	"golang.org/x/exp/constraints"
)

type U128 [2]uint64

func Zero() U128 {
	return U128{}
}

func (u U128) Add(rhs U128) U128 {
	lower, carry := bits.Add64(u[1], rhs[1], 0)
	return U128{u[0] + rhs[0] + carry, lower}
}

func (u U128) AddU64(value uint64) U128 {
	lower, carry := bits.Add64(u[1], value, 0)
	return U128{u[0] + carry, lower}
}

func (u U128) Sub(rhs U128) U128 {
	lower, borrow := bits.Sub64(u[1], rhs[1], 0)
	return U128{u[0] - rhs[0] - borrow, lower}
}

func (u U128) SubU64(value uint64) U128 {
	lower, borrow := bits.Sub64(u[1], value, 0)
	return U128{u[0] - borrow, lower}
}

func (u U128) Shl(bits uint) U128 {
	switch {
	case bits >= 128:
		return U128{}
	case bits >= 64:
		return U128{u[1] << (bits - 64), 0}
	case bits == 0:
		return u
	default:
		u[0] <<= bits
		u[1] |= u[1] >> (64 - bits)
		u[1] <<= bits
		return u
	}
}

func (u U128) And(rhs U128) U128 {
	return U128{u[0] & rhs[0], u[1] & rhs[1]}
}

func (u U128) Or(rhs U128) U128 {
	return U128{u[0] | rhs[0], u[1] | rhs[1]}
}

func (u U128) Flip() U128 {
	return U128{^u[0], ^u[1]}
}

func (u U128) LessThan(rhs U128) bool {
	return u[0] < rhs[0] || (u[0] == rhs[0] && u[1] < rhs[1])
}

func (u U128) LessOrEqualThan(rhs U128) bool {
	return u[0] < rhs[0] || (u[0] == rhs[0] && u[1] <= rhs[1])
}

func (u U128) GreaterThan(rhs U128) bool {
	return u[0] > rhs[0] || (u[0] == rhs[0] && u[1] > rhs[1])
}

func (u U128) GreaterOrEqualThan(rhs U128) bool {
	return u[0] > rhs[0] || (u[0] == rhs[0] && u[1] >= rhs[1])
}

func (u U128) IntoBytes() [16]byte {
	bytes := [16]byte{}
	binary.BigEndian.PutUint64(bytes[0:], u[0])
	binary.BigEndian.PutUint64(bytes[8:], u[1])
	return bytes
}

func (u U128) Cast() (uint64, bool) { // bool means overflow
	return u[1], u[0] > 0
}

func (u U128) UnsafeCast() uint64 {
	return u[1]
}

func Compare(a, b U128) int {
	if a[0] < b[0] {
		return -1
	}
	if a[0] > b[0] {
		return 1
	}
	if a[1] < b[1] {
		return -1
	}
	if a[1] > b[1] {
		return 1
	}
	return 0
}

func FromPrimitive[T constraints.Integer](value T) U128 {
	return U128{0, uint64(value)}
}

func FromBytes(bytes []byte) U128 {
	return U128{
		binary.BigEndian.Uint64(bytes[:8]),
		binary.BigEndian.Uint64(bytes[8:]),
	}
}
