package iprange

import (
	"iter"
	"math"
	"net"
	"strings"

	"github.com/qiuchengxuan/go-datastructures/integer/uint128"
	"github.com/qiuchengxuan/go-datastructures/ip"
)

type u128 = uint128.U128

type IPRange struct{ start, end ip.IP } // end is closed

func (r IPRange) has(ip ip.IP) bool {
	return r.start.LessOrEqualThan(ip) && ip.LessOrEqualThan(r.end)
}

func (r IPRange) IsEmpty() bool { return r.end.LessThan(r.start) }

func (r IPRange) LenExt() uint128.U128 { return u128(r.end).Sub(u128(r.start)).AddU64(1) }

func (r IPRange) Len() uint64 {
	value, overflow := r.LenExt().Cast()
	if overflow {
		return math.MaxUint64
	}
	return value
}

func (r IPRange) Start() ip.IP { return r.start }
func (r IPRange) End() ip.IP   { return r.end }

func (r IPRange) StartOf(start ip.IP) IPRange { return IPRange{start: start, end: r.end} }
func (r IPRange) EndOf(end ip.IP) IPRange     { return IPRange{start: r.start, end: end} }

func (r IPRange) Iter() iter.Seq[ip.IP] {
	return func(yield func(ip.IP) bool) {
		for ip := r.start; ip.LessOrEqualThan(r.end); ip = ip.Add(1) {
			if !yield(ip) {
				return
			}
		}
	}
}

func (r IPRange) intersect(other IPRange) (IPRange, bool) {
	if r.start.GreaterThan(other.start) {
		r, other = other, r
	}
	switch {
	case r.end.LessThan(other.start):
		return IPRange{}, false
	case other.start.LessOrEqualThan(r.end) && r.end.LessOrEqualThan(other.end):
		return IPRange{other.start, r.end}, true
	default:
		return IPRange{other.start, other.end}, true
	}
}

func (r IPRange) WriteTo(buf *strings.Builder) {
	bytes := r.start.IntoBytes()
	buf.WriteString(net.IP(bytes[:]).String())
	if r.start != r.end {
		buf.WriteRune('-')
		bytes := r.end.IntoBytes()
		buf.WriteString(net.IP(bytes[:]).String())
	}
}

func Of(addr ip.IP) IPRange { return FromTo(addr, addr) }

func Empty() IPRange {
	return IPRange{ip.IP(uint128.FromPrimitive(1)), ip.IP(uint128.FromPrimitive(0))}
}

func FromTo(from, to ip.IP) IPRange {
	if to.LessThan(from) {
		return Empty()
	}
	return IPRange{from, to}
}

func (r IPRange) Ranges() IPRanges {
	if r.IsEmpty() {
		return nil
	}
	return IPRanges{r}
}
