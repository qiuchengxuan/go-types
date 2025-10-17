package iprange

import (
	"github.com/qiuchengxuan/go-datastructures/ip"
)

// Slightly reduct memory allocation
type IPAssigner struct{ *IPRanges }

func (a IPAssigner) Into() IPRanges { return *a.IPRanges }

func (a IPAssigner) Add(other IPRanges) IPRanges {
	mayOverflow := !a.IsEmpty() && a.Last() == ip.MaxIP
	switch {
	case other.IsEmpty():
		return *a.IPRanges
	case a.IsEmpty():
		*a.IPRanges = other
	case !mayOverflow && a.Last().Add(1).LessThan(other.First()):
		*a.IPRanges = append(*a.IPRanges, other...)
	case !mayOverflow && a.Last().Add(1) == other.First():
		a.lastChunk().end = other[0].end
		*a.IPRanges = append(*a.IPRanges, other[1:]...)
	default:
		*a.IPRanges = a.IPRanges.Add(other)
	}
	return *a.IPRanges
}

func (a IPAssigner) Sub(other IPRanges) IPRanges {
	*a.IPRanges = a.IPRanges.Sub(other)
	return *a.IPRanges
}

func (a IPAssigner) addIP(addr ip.IP) {
	index, found := a.Binsearch(addr)
	if found {
		return
	}
	ranges := *a.IPRanges
	switch {
	case 0 < index && ranges[index-1].end.Add(1) == addr:
		ranges[index-1].end.Assign().Add(1)
	case index < len(ranges) && ranges[index].start.Sub(1) == addr:
		ranges[index].start.Assign().Sub(1)
	default:
		ranges := append(ranges, IPRange{})
		copy(ranges[index+1:], ranges[index:len(ranges)-1])
		ranges[index] = Of(addr)
		*a.IPRanges = ranges
	}
	if 0 < index && index < len(ranges) && ranges[index-1].end.Add(1) == ranges[index].start {
		ranges[index-1].end = ranges[index].end
		copy(ranges[index:len(ranges)-1], ranges[index+1:])
		*a.IPRanges = ranges[:len(ranges)-1]
	}
}

func (a IPAssigner) AddIP(addr ip.IP) IPRanges {
	mayOverflow := !a.IsEmpty() && a.Last() == ip.MaxIP
	switch {
	case a.Has(addr):
		return *a.IPRanges
	case a.IsEmpty() || !mayOverflow && a.Last().Add(1).LessThan(addr):
		*a.IPRanges = append(*a.IPRanges, Of(addr))
	case !mayOverflow && a.Last().Add(1) == addr:
		a.lastChunk().end = addr
	default:
		a.addIP(addr)
	}
	return *a.IPRanges
}

func (a IPAssigner) RemoveIP(addr ip.IP) bool {
	if a.IsEmpty() || addr.LessThan(a.First()) || a.Last().LessThan(addr) {
		return false
	}
	index, found := a.Binsearch(addr)
	if !found {
		return false
	}
	ranges := *a.IPRanges
	switch entry := &ranges[index]; {
	case entry.Len() == 1:
		*a.IPRanges = append(ranges[:index], ranges[index+1:]...) //nolint: gocritic
		return true
	case addr == entry.start:
		entry.start.Assign().Add(1)
	case addr == entry.end:
		entry.end.Assign().Sub(1)
	default:
		ranges = append(ranges[:index+1], ranges[index:]...)
		ranges[index] = ranges[index].EndOf(addr.Sub(1))
		ranges[index+1] = ranges[index+1].StartOf(addr.Add(1))
		*a.IPRanges = ranges
	}
	return true
}

func (a IPAssigner) SubIP(addr ip.IP) IPRanges {
	a.RemoveIP(addr)
	return *a.IPRanges
}

func (a IPAssigner) Pop() (ip.IP, bool) {
	if a.IsEmpty() {
		return ip.IP{}, false
	}
	ip := a.First()
	a.firstChunk().start.Assign().Add(1)
	if a.firstChunk().IsEmpty() {
		*a.IPRanges = (*a.IPRanges)[1:]
	}
	return ip, true
}
