package iprange

import (
	"math"

	"github.com/qiuchengxuan/go-datastructures/integer/ranges"
	"github.com/qiuchengxuan/go-datastructures/ip"
)

func (r IPRanges) MustCastV4() ranges.Ranges[uint32] {
	if len(r) == 0 {
		return nil
	}
	if !r.V4Only() {
		panic("Integer overflow")
	}
	intRanges := make(ranges.Ranges[uint32], len(r))
	for i, chunk := range r {
		intRanges[i] = ranges.FromTo(chunk.start.U32(), chunk.end.U32())
	}
	return intRanges
}

func (r IPRanges) Cast(base ip.IP) (ranges.Ranges[uint64], bool) {
	if len(r) == 0 {
		return nil, false
	}
	substract := u128(base)
	intRanges := ranges.Empty[uint64]().Ranges()
	index, found := r.Binsearch(base)
	overflow := index > 0
	if found {
		overflow = overflow || r[index].start != base
		end := u128(r[index].end.SubExt(substract)).UnsafeCast()
		intRanges = intRanges.Add(ranges.FromTo(0, end).Ranges())
		index++
	}
	ipRanges := r[index:]
	if len(ipRanges) == 0 {
		return intRanges, overflow
	}
	maximum := base.Add(math.MaxUint64)
	index, found = ipRanges.Binsearch(maximum)
	overflow = overflow || index < len(ipRanges)
	remain := uint64(0)
	if found {
		remain = u128(ipRanges[index].start.SubExt(substract)).UnsafeCast()
	}
	for _, r := range ipRanges[:index] {
		start := u128(r.start.SubExt(substract)).UnsafeCast()
		end := u128(r.end.SubExt(substract)).UnsafeCast()
		intRanges = intRanges.Add(ranges.FromTo(start, end).Ranges())
	}
	if found {
		intRanges = intRanges.Add(ranges.FromTo(remain, math.MaxUint64).Ranges())
	}
	return intRanges, overflow
}
