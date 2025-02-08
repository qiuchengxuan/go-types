package iprange

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
	"sort"
	"strings"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"

	"github.com/qiuchengxuan/go-types/integer/ranges"
	"github.com/qiuchengxuan/go-types/integer/uint128"
	"github.com/qiuchengxuan/go-types/ip"
)

type u128 = uint128.U128

type IPv6Range struct{ start, end ip.IP } // end is closed

func (r IPv6Range) has(ip ip.IP) bool {
	return r.start.LessOrEqualThan(ip) && ip.LessOrEqualThan(r.end)
}

func (r IPv6Range) IsEmpty() bool {
	return r.end.LessThan(r.start)
}

func (r IPv6Range) Len() uint128.U128 {
	return u128(r.end).Sub(u128(r.start)).AddU64(1)
}

func (r IPv6Range) Start() ip.IP {
	return r.start
}

func (r IPv6Range) End() ip.IP {
	return r.end
}

func (r IPv6Range) foreach(consumer func(net.IP)) {
	for ip := r.start; ip.LessOrEqualThan(r.end); ip = ip.AddU64(1) {
		consumer(ip.Into())
	}
}

func (r IPv6Range) intersect(other IPv6Range) (IPv6Range, bool) {
	if r.start.GreaterThan(other.start) {
		r, other = other, r
	}
	switch {
	case r.end.LessThan(other.start):
		return IPv6Range{}, false
	case other.start.LessOrEqualThan(r.end) && r.end.LessOrEqualThan(other.end):
		return IPv6Range{other.start, r.end}, true
	default:
		return IPv6Range{other.start, other.end}, true
	}
}

func (r IPv6Range) WriteTo(buf *strings.Builder) {
	bytes := r.start.IntoBytes()
	buf.WriteString(net.IP(bytes[:]).String())
	if r.start != r.end {
		buf.WriteRune('-')
		bytes := r.end.IntoBytes()
		buf.WriteString(net.IP(bytes[:]).String())
	}
}

func empty() IPv6Range {
	return IPv6Range{ip.IP(uint128.FromPrimitive(1)), ip.IP(uint128.FromPrimitive(0))}
}

type IPv6Ranges []IPv6Range

func (r IPv6Ranges) IsEmpty() bool {
	return len(r) == 0
}

func (r IPv6Ranges) NumChunks() int { return len(r) }

func (r IPv6Ranges) Len() uint128.U128 {
	sum := uint128.Zero()
	for _, r := range r {
		sum = sum.Add(r.Len())
	}
	return sum
}

func (r IPv6Ranges) Equal(other IPv6Ranges) bool {
	return slices.Equal(r, other)
}

func (r IPv6Ranges) binsearch(ip ip.IP) (int, bool) {
	index := sort.Search(len(r), func(i int) bool { return ip.LessOrEqualThan(r[i].end) })
	if index == -1 {
		return len(r), false
	} else if index >= len(r) {
		return index, false
	}
	return index, r[index].has(ip)
}

func (r IPv6Ranges) Binsearch(ip ip.IP) (uint128.U128, bool) {
	if len(r) == 0 || ip.LessThan(r[0].start) {
		return uint128.Zero(), false
	}
	index := sort.Search(len(r), func(i int) bool { return ip.LessOrEqualThan(r[i].end) })
	switch {
	case index == -1:
		return uint128.Zero(), false
	case index >= len(r):
		return r.Len(), false
	default:
		offset := r[:index].Len()
		if r[index].has(ip) {
			return offset.Add(u128(ip.Sub(u128(r[index].start)))), true
		}
		return offset, false
	}
}

func (r IPv6Ranges) Contains(ip ip.IP) bool {
	if r.IsEmpty() {
		return false
	}
	if ip.LessThan(r[0].start) || ip.GreaterThan(r[len(r)-1].end) {
		return false
	}
	_, found := r.binsearch(ip)
	return found
}

func (r IPv6Ranges) Index(index uint128.U128) ip.IP {
	if r.IsEmpty() {
		panic("Index out of range")
	}
	for i := 0; i < len(r); i++ {
		if index.LessThan(r[i].Len()) {
			return r[i].start.Add(index)
		}
		index = index.Sub(r[i].Len())
	}
	panic("Index out of range")
}

func (r IPv6Ranges) IndexU64(index uint64) ip.IP {
	return r.Index(uint128.FromPrimitive(index))
}

func (r IPv6Ranges) Intersect(other IPv6Ranges) IPv6Ranges {
	if r.IsEmpty() || other.IsEmpty() {
		return nil
	}

	index, otherIndex := 0, 0
	retval := make(IPv6Ranges, 0, len(r))
	for index < len(r) && otherIndex < len(other) {
		left, right := r[index], other[otherIndex]
		switch {
		case right.end.LessThan(left.start):
			otherIndex++
		case left.end.LessThan(right.start):
			index++
		default:
			if intersect, ok := r[index].intersect(other[otherIndex]); ok {
				retval = append(retval, intersect)
			}
			if left.end.LessThan(right.end) {
				index++
			} else {
				otherIndex++
			}
		}
	}
	return retval
}

func (r IPv6Ranges) Add(rhs IPv6Ranges) IPv6Ranges {
	if len(rhs) == 0 {
		return r
	} else if len(r) == 0 {
		return rhs
	}
	lower, upper := r, rhs
	retval := make(IPv6Ranges, 0, len(r)+len(rhs))
	pending := empty()
	for len(lower) > 0 && len(upper) > 0 {
		if lower[0].start.GreaterThan(upper[0].start) {
			lower, upper = upper, lower
		}
		switch {
		case pending.IsEmpty():
			pending = lower[0]
			lower = lower[1:]
		case pending.end.AddU64(1).GreaterOrEqualThan(lower[0].start):
			if pending.end.LessThan(lower[0].end) {
				pending.end = lower[0].end
			}
			lower = lower[1:]
		default:
			retval = append(retval, pending)
			pending = empty()
		}
	}
	remain := lower
	if len(upper) > 0 {
		remain = upper
	}
	if !pending.IsEmpty() {
		for len(remain) > 0 && pending.end.AddU64(1).GreaterOrEqualThan(remain[0].start) {
			if pending.end.LessThan(remain[0].end) {
				pending.end = remain[0].end
			}
			remain = remain[1:]
		}
		retval = append(retval, pending)
	}
	return append(retval, remain...)
}

func (r IPv6Ranges) Sub(other IPv6Ranges) IPv6Ranges {
	if r.IsEmpty() || other.IsEmpty() {
		return r
	}
	index, otherIndex := 0, 0
	start := r[0].start
	retval := make(IPv6Ranges, 0, len(r)+len(other)*2)
	for index < len(r) && otherIndex < len(other) {
		left, right := r[index], other[otherIndex]
		if start.LessThan(left.start) || left.end.LessThan(start) {
			start = left.start
		}
		switch {
		case left.end.LessThan(right.start): // orthangonal and left lower, add directly
			retval = append(retval, IPv6Range{start, left.end})
			index++
		case right.end.LessThan(start): // orthangognal and right lower, skip right
			otherIndex++
		case right.start.LessOrEqualThan(start) && left.end.LessOrEqualThan(right.end):
			index++ // right contains left, skip left
		default:
			if right.end.LessThan(left.end) { // break apart or lower part has intersection
				if start.LessThan(right.start) {
					r := IPv6Range{start, right.start.SubU64(1)}
					retval = append(retval, r)
				}
				otherIndex++
			} else { // higher part has intersection
				retval = append(retval, IPv6Range{start, right.start.SubU64(1)})
				index++
			}
			start = right.end.AddU64(1)
		}
	}
	if index >= len(r) {
		return retval
	}
	remain := r[index]
	if start != remain.start && start.LessOrEqualThan(remain.end) {
		retval = append(retval, IPv6Range{start, remain.end})
		index++
	}
	return append(retval, r[index:]...)
}

func (r IPv6Ranges) Find(pattern func(ip ip.IP) bool) (ip.IP, bool) {
	for _, r := range r {
		for ip := r.start; ip.LessOrEqualThan(r.end); ip = ip.AddU64(1) {
			if pattern(ip) {
				return ip, true
			}
		}
	}
	return ip.IP{}, false
}

func compact(ranges IPv6Ranges) IPv6Ranges {
	compacted := make(IPv6Ranges, 1, len(ranges))
	compacted[0] = ranges[0]
	for _, r := range ranges[1:] {
		last := compacted[len(compacted)-1]
		if last.end.AddU64(1).GreaterOrEqualThan(r.start) { // intersection or adjacency
			if r.end.GreaterThan(last.end) { // r not in last
				compacted[len(compacted)-1].end = r.end
			}
		} else {
			compacted = append(compacted, r)
		}
	}
	return compacted
}

func (r IPv6Ranges) AddIP(ip ip.IP) IPv6Ranges {
	if len(r) == 0 {
		return IPv6Ranges{{start: ip, end: ip}}
	}
	index, found := r.binsearch(ip)
	if found {
		return r
	}
	if index >= len(r) {
		return compact(append(r, IPv6Range{ip, ip}))
	}
	ranges := slices.Clone(r)
	pRange := &ranges[index]
	switch {
	case pRange.end.AddU64(1) == ip:
		pRange.end = ip
	case pRange.start.SubU64(1) == ip:
		pRange.start = ip
	default:
		ranges = append(ranges[:index], IPv6Range{ip, ip})
		ranges = append(ranges, r[index:]...)
	}
	return compact(ranges)
}

func (r IPv6Ranges) SubIP(ip ip.IP) IPv6Ranges {
	if len(r) == 0 {
		return r
	}
	index, found := r.binsearch(ip)
	if !found {
		return r
	}
	var ranges IPv6Ranges
	switch {
	case r[index].Len() == uint128.FromPrimitive(1):
		return append(r[:index], r[index+1:]...)
	case r[index].start == ip:
		ranges = slices.Clone(r)
		ranges[index].start = ip.AddU64(1)
	case r[index].end == ip:
		ranges = slices.Clone(r)
		ranges[index].end = ip.SubU64(1)
	default:
		ranges = make(IPv6Ranges, len(r)+1)
		copy(ranges[:index], r[:index])
		copy(ranges[index+1:], r[index:])
		ranges[index] = IPv6Range{r[index].start, ip.SubU64(1)}
		ranges[index+1].start = ip.AddU64(1)
	}
	return ranges
}

func (r IPv6Ranges) Foreach(consumer func(net.IP)) {
	for _, r := range r {
		r.foreach(consumer)
	}
}

func (r IPv6Ranges) IterChunks(fn func(IPv6Range) bool) IPv6Range {
	for _, r := range r {
		if fn(r) {
			return r
		}
	}
	return empty()
}

func (r IPv6Ranges) WriteTo(buf *strings.Builder) {
	if len(r) == 0 {
		return
	}
	for i := 0; i < len(r)-1; i++ {
		r[i].WriteTo(buf)
		buf.WriteRune(',')
	}
	r[len(r)-1].WriteTo(buf)
}

func (r IPv6Ranges) String() string {
	var buf strings.Builder
	r.WriteTo(&buf)
	return buf.String()
}

func (r IPv6Ranges) Cast(base ip.IP) (ranges.Ranges[uint64], bool) {
	if len(r) == 0 {
		return nil, false
	}
	substract := u128(base)
	intRanges := ranges.Empty[uint64]().Plural()
	index, found := r.binsearch(base)
	overflow := index > 0
	if found {
		overflow = overflow || r[index].start != base
		end := u128(r[index].end.Sub(substract)).UnsafeCast()
		intRanges = intRanges.Add(ranges.FromTo(0, end).Plural())
		index++
	}
	ipRanges := r[index:]
	if len(ipRanges) == 0 {
		return intRanges, overflow
	}
	maximum := base.Add(uint128.FromPrimitive(uint64(math.MaxUint64)))
	index, found = ipRanges.binsearch(maximum)
	overflow = overflow || index < len(ipRanges)
	remain := uint64(0)
	if found {
		remain = u128(ipRanges[index].start.Sub(substract)).UnsafeCast()
	}
	for _, r := range ipRanges[:index] {
		start := u128(r.start.Sub(substract)).UnsafeCast()
		end := u128(r.end.Sub(substract)).UnsafeCast()
		intRanges = intRanges.Add(ranges.FromTo(start, end).Plural())
	}
	if found {
		intRanges = intRanges.Add(ranges.FromTo(remain, math.MaxUint64).Plural())
	}
	return intRanges, overflow
}

func (r IPv6Ranges) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *IPv6Ranges) UnmarshalText(text []byte) error {
	*r = v6FromStr(string(text))
	if r.IsEmpty() && len(text) > 0 {
		return errors.New("Malformed ipv6 ranges")
	}
	return nil
}

func (r IPv6Ranges) MarshalBinary() ([]byte, error) {
	if r.IsEmpty() {
		return nil, nil
	}
	sizeOf := 8 // size of uint64
	data := make([]byte, r.NumChunks()*4*sizeOf)
	for i, chunk := range r {
		binary.BigEndian.PutUint64(data[i*4*sizeOf:], chunk.start[0])
		binary.BigEndian.PutUint64(data[(i*4+1)*sizeOf:], chunk.start[1])
		binary.BigEndian.PutUint64(data[(i*4+2)*sizeOf:], chunk.end[0])
		binary.BigEndian.PutUint64(data[(i*4+3)*sizeOf:], chunk.end[1])
	}
	return data, nil
}

func (r *IPv6Ranges) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	sizeOf := 8 // size of uint64
	if len(data)%(4*sizeOf) != 0 {
		return errors.New("Not IPv6 ranges")
	}
	length := len(data) / (4 * sizeOf)
	ranges := make(IPv6Ranges, length)
	for i := 0; i < length; i++ {
		chunk := &ranges[i]
		chunk.start[0] = binary.BigEndian.Uint64(data[i*4*sizeOf:])
		chunk.start[1] = binary.BigEndian.Uint64(data[(i*4+1)*sizeOf:])
		chunk.end[0] = binary.BigEndian.Uint64(data[(i*4+2)*sizeOf:])
		chunk.end[1] = binary.BigEndian.Uint64(data[(i*4+3)*sizeOf:])
		if chunk.start.GreaterOrEqualThan(chunk.end) {
			return errors.New("Not a valid IPv6 ranges")
		}
	}
	for i := 1; i < length; i++ {
		if ranges[i-1].start.GreaterOrEqualThan(ranges[i].end) {
			return errors.New("Not a valid IPv6 ranges")
		}
	}
	*r = ranges
	return nil
}

func FromIntRanges[I constraints.Integer](base ip.IP, ranges ranges.Ranges[I]) IPv6Ranges {
	ipv6Ranges := make(IPv6Ranges, len(ranges))
	for i, r := range ranges {
		r := IPv6Range{base.AddU64(uint64(r.Start())), base.AddU64(uint64(r.End()))}
		ipv6Ranges[i] = r
	}
	return ipv6Ranges
}

func v6FromSlice(slice []ip.IP) IPv6Ranges {
	if len(slice) == 0 {
		return nil
	}
	var retval IPv6Ranges
	pending := IPv6Range{slice[0], slice[0]}
	for _, value := range slice[1:] {
		if value.LessOrEqualThan(pending.end) {
			continue
		}
		if value == pending.end.AddU64(1) {
			pending.end = value
			continue
		}
		retval = append(retval, pending)
		pending = IPv6Range{value, value}
	}
	return append(retval, pending)
}

func v6FromStr(ipRanges string) IPv6Ranges {
	if ipRanges == "" {
		return IPv6Ranges{}
	}

	splitted := strings.Split(ipRanges, ",")
	ranges := make([]IPv6Range, 0, len(splitted))

	for _, ipRange := range splitted {
		before, after, ok := strings.Cut(ipRange, "-")
		if ok && after == "" {
			return nil
		}
		startIP := net.ParseIP(before)
		start := ip.FromBytes(startIP)
		if after == "" {
			ranges = append(ranges, IPv6Range{start, start})
			continue
		}
		stopIP := net.ParseIP(after)
		stop := ip.FromBytes(stopIP)
		if start.GreaterThan(stop) {
			return nil
		}
		ranges = append(ranges, IPv6Range{start, stop})
	}
	if len(ranges) == 0 {
		return nil
	}

	index := 0
	for _, ipRange := range ranges[1:] {
		last := &ranges[index]
		switch {
		case ipRange.start.LessThan(last.end):
			continue
		case ipRange.start == last.end:
			fallthrough
		case ipRange.start == last.end.AddU64(1):
			last.end = ipRange.end
		default:
			index++
			ranges[index] = ipRange
		}
	}
	return ranges[:index+1]
}
