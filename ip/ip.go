package ip

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/exp/constraints"

	"github.com/qiuchengxuan/go-types/integer/uint128"
)

type IP [2]uint64

func From(ip net.IP) IP {
	if ip == nil {
		return IP{}
	}
	return IP(uint128.FromBytes(ip.To16()))
}

func (ip IP) Into() net.IP {
	bytes := uint128.U128(ip).IntoBytes()
	return bytes[:]
}

type u128 = uint128.U128

func (ip IP) Add(rhs uint128.U128) IP { return IP(u128(ip).Add(rhs)) }
func (ip IP) AddU64(rhs uint64) IP    { return IP(u128(ip).AddU64(rhs)) }
func (ip IP) Sub(rhs uint128.U128) IP { return IP(u128(ip).Sub(rhs)) }
func (ip IP) SubU64(rhs uint64) IP    { return IP(u128(ip).SubU64(rhs)) }

func (ip IP) LessThan(rhs IP) bool           { return u128(ip).LessThan(u128(rhs)) }
func (ip IP) LessOrEqualThan(rhs IP) bool    { return u128(ip).LessOrEqualThan(u128(rhs)) }
func (ip IP) GreaterThan(rhs IP) bool        { return u128(ip).GreaterThan(u128(rhs)) }
func (ip IP) GreaterOrEqualThan(rhs IP) bool { return u128(ip).GreaterOrEqualThan(u128(rhs)) }

func (ip IP) IntoBytes() [16]byte { return u128(ip).IntoBytes() }

func Compare(a, b IP) int { return uint128.Compare(u128(a), u128(b)) }

func FromPrimitive[T constraints.Integer](value T) IP { return IP{0, uint64(value)} }

func FromBytes(bytes []byte) IP { return IP(uint128.FromBytes(bytes)) }

func (ip IP) MarshalText() ([]byte, error) {
	return ip.Into().MarshalText()
}

func (ip *IP) UnmarshalText(data []byte) error {
	value := net.ParseIP(string(data))
	if value == nil {
		return errors.New("Not a valid IP")
	}
	*ip = IP(uint128.FromBytes(value))
	return nil
}

func MustParse(text string) IP {
	ip := net.ParseIP(text)
	if ip == nil {
		panic(fmt.Sprintf("Text %s is not an IP", text))
	}
	return From(ip)
}
