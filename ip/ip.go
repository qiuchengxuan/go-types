package ip

import (
	"errors"
	"fmt"
	"math"
	"net"

	"golang.org/x/exp/constraints"

	"github.com/qiuchengxuan/go-types/integer/uint128"
)

type IPv4 uint32

func (ip IPv4) IP() IP { return IP{0, uint64(ip) | (0xFFFF << 32)} }

type IP [2]uint64

var MaxIP = IP{math.MaxUint64, math.MaxUint64}

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

func (ip IP) AddExt(rhs uint128.U128) IP { return IP(u128(ip).Add(rhs)) }
func (ip IP) Add(rhs uint64) IP          { return IP(u128(ip).AddU64(rhs)) }
func (ip IP) SubExt(rhs uint128.U128) IP { return IP(u128(ip).Sub(rhs)) }
func (ip IP) Sub(rhs uint64) IP          { return IP(u128(ip).SubU64(rhs)) }

type Assigner struct{ *IP }

func (ip *IP) Assign() Assigner { return Assigner{ip} }

func (a Assigner) Add(rhs uint64) IP {
	*a.IP = a.IP.Add(rhs)
	return *a.IP
}

func (a Assigner) AddExt(rhs uint128.U128) IP {
	*a.IP = a.IP.AddExt(rhs)
	return *a.IP
}

func (a Assigner) Sub(rhs uint64) IP {
	*a.IP = a.IP.Sub(rhs)
	return *a.IP
}

func (a Assigner) SubExt(rhs uint128.U128) IP {
	*a.IP = a.IP.SubExt(rhs)
	return *a.IP
}

func (ip IP) LessThan(rhs IP) bool           { return u128(ip).LessThan(u128(rhs)) }
func (ip IP) LessOrEqualThan(rhs IP) bool    { return u128(ip).LessOrEqualThan(u128(rhs)) }
func (ip IP) GreaterThan(rhs IP) bool        { return u128(ip).GreaterThan(u128(rhs)) }
func (ip IP) GreaterOrEqualThan(rhs IP) bool { return u128(ip).GreaterOrEqualThan(u128(rhs)) }

func (ip IP) IntoBytes() [16]byte { return u128(ip).IntoBytes() }

func (ip IP) String() string { return ip.Into().String() }

func (ip IP) IsV4() bool { return ip[0] == 0 && uint32(ip[1]>>32) == 0xFFFF }

func (ip IP) U32() uint32        { return uint32(ip[1]) }
func (ip IP) U128() uint128.U128 { return uint128.U128(ip) }

func Compare(a, b IP) int { return uint128.Compare(u128(a), u128(b)) }

func FromPrimitive[T constraints.Integer](value T) IP { return IP{0, uint64(value)} }

func FromBytes(bytes []byte) IP { return IP(uint128.FromBytes(bytes)) }

func (ip IP) MarshalText() ([]byte, error) {
	return ip.Into().MarshalText()
}

func (ip *IP) UnmarshalText(data []byte) error {
	value := net.ParseIP(string(data))
	if value == nil {
		return errors.New("not a valid IP")
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
