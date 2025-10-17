package iprange

import "github.com/qiuchengxuan/go-datastructures/ip"

func (r IPRanges) Add(rhs IPRanges) IPRanges {
	if len(rhs) == 0 {
		return r
	} else if len(r) == 0 {
		return rhs
	}
	lower, upper := r, rhs
	retval := make(IPRanges, 0, len(r)+len(rhs))
	pending := Empty()
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
			pending = Empty()
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

func (r IPRanges) AddIP(addr ip.IP) IPRanges {
	if r.Has(addr) {
		return r
	}
	return r.Clone().Ref().Assign().AddIP(addr)
}

func (r IPRanges) SubIP(addr ip.IP) IPRanges {
	if !r.Has(addr) {
		return r
	}
	return r.Clone().Ref().Assign().SubIP(addr)
}
