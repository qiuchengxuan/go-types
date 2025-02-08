package ranges

import "slices"

func (r Ranges[I]) Add(other Ranges[I]) Ranges[I] {
	if len(other) == 0 {
		return r
	} else if len(r) == 0 {
		return other
	}
	lower, upper := r, other
	if lower.Last() > other.First() {
		lower, upper = upper, lower
	}
	if lower.Last()+1 < other.First() {
		return append(lower, other...)
	}
	retval := make(Ranges[I], 0, len(r)+len(other))
	pending := Empty[I]()
	for len(lower) > 0 && len(upper) > 0 {
		if lower[0].start > upper[0].start {
			lower, upper = upper, lower
		}
		switch {
		case pending.Len() == 0:
			pending = lower[0]
			lower = lower[1:]
		case pending.end+1 >= lower[0].start:
			if pending.end < lower[0].end {
				pending.end = lower[0].end
			}
			lower = lower[1:]
		default:
			retval = append(retval, pending)
			pending = Empty[I]()
		}
	}
	remain := lower
	if len(upper) > 0 {
		remain = upper
	}
	if pending.Len() > 0 {
		if pending.end+1 < pending.end {
			return append(retval, pending)
		}
		for len(remain) > 0 && pending.end+1 >= remain[0].start {
			if pending.end < remain[0].end {
				pending.end = remain[0].end
			}
			remain = remain[1:]
		}
		retval = append(retval, pending)
	}
	return append(retval, remain...)
}

func (r Ranges[I]) Intersect(other Ranges[I]) Ranges[I] {
	if len(r) == 0 || len(other) == 0 {
		return nil
	}

	index, otherIndex := 0, 0
	retval := make(Ranges[I], 0, len(r))
	for index < len(r) && otherIndex < len(other) {
		left, right := r[index], other[otherIndex]
		switch {
		case right.end < left.start:
			otherIndex++
		case left.end < right.start:
			index++
		default:
			if intersect := r[index].Intersect(other[otherIndex]); intersect.Len() > 0 {
				retval = append(retval, intersect)
			}
			if left.end < right.end {
				index++
			} else {
				otherIndex++
			}
		}
	}
	return retval
}

// time complexity O(N + M)
func (r Ranges[I]) Sub(other Ranges[I]) Ranges[I] {
	if r.IsEmpty() || other.IsEmpty() {
		return r
	}

	index, otherIndex := 0, 0
	start := r[0].start
	retval := make(Ranges[I], 0, len(r)+len(other)*2)
	for index < len(r) && otherIndex < len(other) {
		left, right := r[index], other[otherIndex]
		if start < left.start || left.end < start {
			start = left.start
		}
		switch {
		case left.end < right.start: // orthangonal and left lower, add directly
			retval = append(retval, FromTo(start, left.end))
			index++
		case right.end < start: // orthangognal and right lower, skip right
			otherIndex++
		case right.start <= start && left.end <= right.end: // right contains left, skip left
			index++
		default:
			if right.end < left.end { // break apart or lower part has intersection
				if start < right.start {
					retval = append(retval, FromTo(start, right.start-1))
				}
				otherIndex++
			} else { // higher part has intersection
				retval = append(retval, FromTo(start, right.start-1))
				index++
			}
			start = right.end + 1
		}
	}
	if index >= len(r) {
		return retval
	}
	remain := r[index]
	if start != remain.start && start <= remain.end {
		retval = append(retval, FromTo(start, remain.end))
		index++
	}
	return append(retval, r[index:]...)
}

func (r Ranges[I]) Add1(value I) Ranges[I] {
	if len(r) == 0 {
		return FromTo(value, value).Plural()
	}
	index, found := r.binsearch(value)
	if found {
		return r
	}
	if 0 < index && index < len(r) {
		if r[index-1].end+1 == value && value == r[index].start-1 {
			ranges := slices.Clone(r[:index-1])
			ranges = append(ranges, FromTo(r[index-1].start, r[index].end))
			return append(ranges, r[index+1:]...)
		}
	}
	ranges := slices.Clone(r)
	switch {
	case 0 < index && r[index-1].end+1 == value:
		ranges[index-1].expandUpper(1)
	case index < len(r) && r[index].start-1 == value:
		ranges[index].expandLower(1)
	case index == len(r):
		return append(ranges, Of(value))
	default:
		ranges = append(ranges[:index], Of(value))
		return append(ranges, r[index:]...)
	}
	return ranges
}

func (r *Ranges[I]) Delete(value I) bool {
	if len(*r) == 0 {
		return false
	}
	if value < r.First() || r.Last() < value {
		return false
	}
	index, found := r.binsearch(value)
	if !found {
		return false
	}
	switch entry := &(*r)[index]; {
	case entry.Len() == 1:
		*r = append((*r)[:index], (*r)[index+1:]...)
		return true
	case value == entry.start:
		entry.shrinkLower(1)
	case value == entry.end:
		entry.shrinkUpper(1)
	default:
		(*r) = append((*r)[:index+1], (*r)[index:]...)
		(*r)[index] = (*r)[index].EndOf(value - 1)
		(*r)[index+1] = (*r)[index+1].StartOf(value + 1)
	}
	return true
}

func (r Ranges[I]) Sub1(value I) Ranges[I] {
	ranges := slices.Clone(r)
	ranges.Delete(value)
	return ranges
}
