package iprange

import (
	"iter"
	"slices"
	"strings"

	"github.com/qiuchengxuan/go-types/integer/uint128"
	"github.com/qiuchengxuan/go-types/ip"
)

type IPRanges []IPRange

func (r IPRanges) IsEmpty() bool { return len(r) == 0 }

func (r IPRanges) Len() uint128.U128 {
	sum := uint128.Zero()
	for _, r := range r {
		sum = sum.Add(r.LenExt())
	}
	return sum
}

func (r IPRanges) Clone() IPRanges { return slices.Clone(r) }

func (r IPRanges) Ref() *IPRanges { return &r }

func (r IPRanges) First() ip.IP { return r[0].start }
func (r IPRanges) Last() ip.IP  { return r[len(r)-1].end }

func (r IPRanges) Equal(other IPRanges) bool { return slices.Equal(r, other) }

func (r IPRanges) V4Only() bool { return r.First().IsV4() && r.Last().IsV4() }

func (r IPRanges) Binsearch(addr ip.IP) (int, bool) {
	return slices.BinarySearchFunc(r, addr, func(item IPRange, value ip.IP) int {
		if item.has(value) {
			return 0
		}
		return ip.Compare(item.start, value)
	})
}

func (r IPRanges) Contains(other IPRanges) bool {
	index := 0
	for _, chunk := range other {
		offset, ok := r[index:].Binsearch(chunk.start)
		if !ok {
			return false
		}
		index += offset
		if r[index].end.LessThan(chunk.end) {
			return false
		}
	}
	return true
}

func (r IPRanges) Has(ip ip.IP) bool {
	if r.IsEmpty() {
		return false
	}
	if ip.LessThan(r.First()) || ip.GreaterThan(r.Last()) {
		return false
	}
	_, found := r.Binsearch(ip)
	return found
}

func (r IPRanges) IndexExt(index uint128.U128) ip.IP {
	if r.IsEmpty() {
		panic("Index out of range")
	}
	for i := range len(r) {
		if index.LessThan(r[i].LenExt()) {
			return r[i].start.AddExt(index)
		}
		index = index.Sub(r[i].LenExt())
	}
	panic("Index out of range")
}

func (r IPRanges) Index(index int) ip.IP {
	return r.IndexExt(uint128.FromPrimitive(index))
}

func (r IPRanges) firstChunk() *IPRange { return &r[0] }
func (r IPRanges) lastChunk() *IPRange  { return &r[len(r)-1] }

func (r IPRanges) Find(pattern func(ip ip.IP) bool) (ip.IP, bool) {
	for _, r := range r {
		for ip := r.start; ip.LessOrEqualThan(r.end); ip = ip.Add(1) {
			if pattern(ip) {
				return ip, true
			}
		}
	}
	return ip.IP{}, false
}

func (r IPRanges) Iter() iter.Seq[ip.IP] {
	return func(yield func(ip.IP) bool) {
		for _, chunk := range r {
			for ip := chunk.start; ip.LessOrEqualThan(chunk.end); ip = ip.Add(1) {
				if !yield(ip) {
					return
				}
			}
		}
	}
}

func (r IPRanges) WriteTo(buf *strings.Builder) {
	if len(r) == 0 {
		return
	}
	for i := range len(r) - 1 {
		r[i].WriteTo(buf)
		buf.WriteRune(',')
	}
	r[len(r)-1].WriteTo(buf)
}

func (r *IPRanges) Assign() IPAssigner {
	return IPAssigner{r}
}

func (r IPRanges) String() string {
	var buf strings.Builder
	r.WriteTo(&buf)
	return buf.String()
}
