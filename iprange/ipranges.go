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

type IPRange struct{ start, end ip.IP } // end is closed

func (r IPRange) has(ip ip.IP) bool {
	return r.start.LessOrEqualThan(ip) && ip.LessOrEqualThan(r.end)
}

func (r IPRange) IsEmpty() bool {
	return r.end.LessThan(r.start)
}

func (r IPRange) Len() uint128.U128 {
	return u128(r.end).Sub(u128(r.start)).AddU64(1)
}

func (r IPRange) Start() ip.IP {
	return r.start
}

func (r IPRange) End() ip.IP {
	return r.end
}

func (r IPRange) foreach(consumer func(net.IP)) {
	for ip := r.start; ip.LessOrEqualThan(r.end); ip = ip.Add(1) {
		consumer(ip.Into())
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

func empty() IPRange {
	return IPRange{ip.IP(uint128.FromPrimitive(1)), ip.IP(uint128.FromPrimitive(0))}
}

type IPRanges []IPRange

func (r IPRanges) IsEmpty() bool { return len(r) == 0 }

func (r IPRanges) NumChunks() int { return len(r) }

func (r IPRanges) Len() uint128.U128 {
	sum := uint128.Zero()
	for _, r := range r {
		sum = sum.Add(r.Len())
	}
	return sum
}

func (r IPRanges) First() ip.IP { return r[0].start }
func (r IPRanges) Last() ip.IP  { return r[len(r)-1].end }

func (r IPRanges) Equal(other IPRanges) bool { return slices.Equal(r, other) }

func (r IPRanges) V4Only() bool {
	return r.First().GreaterOrEqualThan(ip.V4Min) && r.Last().LessOrEqualThan(ip.V4Max)
}

func (r IPRanges) binsearch(ip ip.IP) (int, bool) {
	index := sort.Search(len(r), func(i int) bool { return ip.LessOrEqualThan(r[i].end) })
	if index == -1 {
		return len(r), false
	} else if index >= len(r) {
		return index, false
	}
	return index, r[index].has(ip)
}

func (r IPRanges) Binsearch(ip ip.IP) (uint128.U128, bool) {
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
			return offset.Add(u128(ip.SubExt(u128(r[index].start)))), true
		}
		return offset, false
	}
}

func (r IPRanges) Contains(ip ip.IP) bool {
	if r.IsEmpty() {
		return false
	}
	if ip.LessThan(r.First()) || ip.GreaterThan(r.Last()) {
		return false
	}
	_, found := r.binsearch(ip)
	return found
}

func (r IPRanges) IndexExt(index uint128.U128) ip.IP {
	if r.IsEmpty() {
		panic("Index out of range")
	}
	for i := 0; i < len(r); i++ {
		if index.LessThan(r[i].Len()) {
			return r[i].start.AddExt(index)
		}
		index = index.Sub(r[i].Len())
	}
	panic("Index out of range")
}

func (r IPRanges) Index(index int) ip.IP {
	return r.IndexExt(uint128.FromPrimitive(index))
}

func (r IPRanges) Intersect(other IPRanges) IPRanges {
	if r.IsEmpty() || other.IsEmpty() {
		return nil
	}

	index, otherIndex := 0, 0
	retval := make(IPRanges, 0, len(r))
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

func (r IPRanges) Add(rhs IPRanges) IPRanges {
	if len(rhs) == 0 {
		return r
	} else if len(r) == 0 {
		return rhs
	}
	lower, upper := r, rhs
	retval := make(IPRanges, 0, len(r)+len(rhs))
	pending := empty()
	for len(lower) > 0 && len(upper) > 0 {
		if lower[0].start.GreaterThan(upper[0].start) {
			lower, upper = upper, lower
		}
		switch {
		case pending.IsEmpty():
			pending = lower[0]
			lower = lower[1:]
		case pending.end.Add(1).GreaterOrEqualThan(lower[0].start):
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
		for len(remain) > 0 && pending.end.Add(1).GreaterOrEqualThan(remain[0].start) {
			if pending.end.LessThan(remain[0].end) {
				pending.end = remain[0].end
			}
			remain = remain[1:]
		}
		retval = append(retval, pending)
	}
	return append(retval, remain...)
}

func (r IPRanges) Sub(other IPRanges) IPRanges {
	if r.IsEmpty() || other.IsEmpty() {
		return r
	}
	index, otherIndex := 0, 0
	start := r[0].start
	retval := make(IPRanges, 0, len(r)+len(other)*2)
	for index < len(r) && otherIndex < len(other) {
		left, right := r[index], other[otherIndex]
		if start.LessThan(left.start) || left.end.LessThan(start) {
			start = left.start
		}
		switch {
		case left.end.LessThan(right.start): // orthangonal and left lower, add directly
			retval = append(retval, IPRange{start, left.end})
			index++
		case right.end.LessThan(start): // orthangognal and right lower, skip right
			otherIndex++
		case right.start.LessOrEqualThan(start) && left.end.LessOrEqualThan(right.end):
			index++ // right contains left, skip left
		default:
			if right.end.LessThan(left.end) { // break apart or lower part intersection
				if start.LessThan(right.start) {
					r := IPRange{start, right.start.Sub(1)}
					retval = append(retval, r)
				}
				otherIndex++
			} else { // higher part has intersection
				retval = append(retval, IPRange{start, right.start.Sub(1)})
				index++
			}
			start = right.end.Add(1)
		}
	}
	if index >= len(r) {
		return retval
	}
	remain := r[index]
	if start != remain.start && start.LessOrEqualThan(remain.end) {
		retval = append(retval, IPRange{start, remain.end})
		index++
	}
	return append(retval, r[index:]...)
}

func (r *IPRanges) Pop() (ip.IP, bool) {
	if r.IsEmpty() {
		return ip.IP{}, false
	}
	ranges := *r
	ip := ranges[0].start
	ranges[0].start = ranges[0].start.Add(1)
	if ranges[0].IsEmpty() {
		copy(ranges, ranges[1:])
		*r = ranges[:len(ranges)-1]
	}
	return ip, true
}

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

func compact(ranges IPRanges) IPRanges {
	compacted := make(IPRanges, 1, len(ranges))
	compacted[0] = ranges[0]
	for _, r := range ranges[1:] {
		last := compacted[len(compacted)-1]
		if last.end.Add(1).GreaterOrEqualThan(r.start) { // intersection or adjacency
			if r.end.GreaterThan(last.end) { // r not in last
				compacted[len(compacted)-1].end = r.end
			}
		} else {
			compacted = append(compacted, r)
		}
	}
	return compacted
}

func (r IPRanges) AddIP(ip ip.IP) IPRanges {
	if len(r) == 0 {
		return IPRanges{{start: ip, end: ip}}
	}
	index, found := r.binsearch(ip)
	if found {
		return r
	}
	if index >= len(r) {
		return compact(append(r, IPRange{ip, ip}))
	}
	ranges := slices.Clone(r)
	pRange := &ranges[index]
	switch {
	case pRange.end.Add(1) == ip:
		pRange.end = ip
	case pRange.start.Sub(1) == ip:
		pRange.start = ip
	default:
		ranges = append(ranges[:index], IPRange{ip, ip})
		ranges = append(ranges, r[index:]...)
	}
	return compact(ranges)
}

func (r IPRanges) SubIP(ip ip.IP) IPRanges {
	if len(r) == 0 {
		return r
	}
	index, found := r.binsearch(ip)
	if !found {
		return r
	}
	var ranges IPRanges
	switch {
	case r[index].Len() == uint128.FromPrimitive(1):
		return append(r[:index], r[index+1:]...)
	case r[index].start == ip:
		ranges = slices.Clone(r)
		ranges[index].start = ip.Add(1)
	case r[index].end == ip:
		ranges = slices.Clone(r)
		ranges[index].end = ip.Sub(1)
	default:
		ranges = make(IPRanges, len(r)+1)
		copy(ranges[:index], r[:index])
		copy(ranges[index+1:], r[index:])
		ranges[index] = IPRange{r[index].start, ip.Sub(1)}
		ranges[index+1].start = ip.Add(1)
	}
	return ranges
}

func (r IPRanges) Foreach(consumer func(net.IP)) {
	for _, r := range r {
		r.foreach(consumer)
	}
}

func (r IPRanges) IterChunks(fn func(IPRange) bool) IPRange {
	for _, r := range r {
		if fn(r) {
			return r
		}
	}
	return empty()
}

func (r IPRanges) WriteTo(buf *strings.Builder) {
	if len(r) == 0 {
		return
	}
	for i := 0; i < len(r)-1; i++ {
		r[i].WriteTo(buf)
		buf.WriteRune(',')
	}
	r[len(r)-1].WriteTo(buf)
}

func (r IPRanges) String() string {
	var buf strings.Builder
	r.WriteTo(&buf)
	return buf.String()
}

func (r IPRanges) Cast(base ip.IP) (ranges.Ranges[uint64], bool) {
	if len(r) == 0 {
		return nil, false
	}
	substract := u128(base)
	intRanges := ranges.Empty[uint64]().Plural()
	index, found := r.binsearch(base)
	overflow := index > 0
	if found {
		overflow = overflow || r[index].start != base
		end := u128(r[index].end.SubExt(substract)).UnsafeCast()
		intRanges = intRanges.Add(ranges.FromTo(0, end).Plural())
		index++
	}
	ipRanges := r[index:]
	if len(ipRanges) == 0 {
		return intRanges, overflow
	}
	maximum := base.Add(math.MaxUint64)
	index, found = ipRanges.binsearch(maximum)
	overflow = overflow || index < len(ipRanges)
	remain := uint64(0)
	if found {
		remain = u128(ipRanges[index].start.SubExt(substract)).UnsafeCast()
	}
	for _, r := range ipRanges[:index] {
		start := u128(r.start.SubExt(substract)).UnsafeCast()
		end := u128(r.end.SubExt(substract)).UnsafeCast()
		intRanges = intRanges.Add(ranges.FromTo(start, end).Plural())
	}
	if found {
		intRanges = intRanges.Add(ranges.FromTo(remain, math.MaxUint64).Plural())
	}
	return intRanges, overflow
}

func (r IPRanges) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *IPRanges) UnmarshalText(text []byte) error {
	*r = FromStr(string(text))
	if r.IsEmpty() && len(text) > 0 {
		return errors.New("Malformed ipv6 ranges")
	}
	return nil
}

func (r IPRanges) MarshalBinary() ([]byte, error) {
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

func (r *IPRanges) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	sizeOf := 8 // size of uint64
	if len(data)%(4*sizeOf) != 0 {
		return errors.New("Not IPv6 ranges")
	}
	length := len(data) / (4 * sizeOf)
	ranges := make(IPRanges, length)
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

func FromIntRanges[I constraints.Integer](base ip.IP, ranges ranges.Ranges[I]) IPRanges {
	ipv6Ranges := make(IPRanges, len(ranges))
	for i, r := range ranges {
		r := IPRange{base.Add(uint64(r.Start())), base.Add(uint64(r.End()))}
		ipv6Ranges[i] = r
	}
	return ipv6Ranges
}

func FromSlice(slice []ip.IP) IPRanges {
	if len(slice) == 0 {
		return nil
	}
	var retval IPRanges
	pending := IPRange{slice[0], slice[0]}
	for _, value := range slice[1:] {
		if value.LessOrEqualThan(pending.end) {
			continue
		}
		if value == pending.end.Add(1) {
			pending.end = value
			continue
		}
		retval = append(retval, pending)
		pending = IPRange{value, value}
	}
	return append(retval, pending)
}

func FromStr(ipRanges string) IPRanges {
	if ipRanges == "" {
		return IPRanges{}
	}

	splitted := strings.Split(ipRanges, ",")
	ranges := make([]IPRange, 0, len(splitted))

	for _, ipRange := range splitted {
		before, after, ok := strings.Cut(ipRange, "-")
		if ok && after == "" {
			return nil
		}
		startIP := net.ParseIP(before)
		start := ip.FromBytes(startIP)
		if after == "" {
			ranges = append(ranges, IPRange{start, start})
			continue
		}
		stopIP := net.ParseIP(after)
		stop := ip.FromBytes(stopIP)
		if start.GreaterThan(stop) {
			return nil
		}
		ranges = append(ranges, IPRange{start, stop})
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
		case ipRange.start == last.end.Add(1):
			last.end = ipRange.end
		default:
			index++
			ranges[index] = ipRange
		}
	}
	return ranges[:index+1]
}
