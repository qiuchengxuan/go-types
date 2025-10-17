package iprange

import "slices"

func (r IPRanges) NumChunks() int { return len(r) }

func (r IPRanges) Limit(bound IPRange) IPRanges {
	lower, upper := bound.start, bound.end
	if r.IsEmpty() || lower.LessOrEqualThan(r.First()) && r.Last().LessOrEqualThan(upper) {
		return r
	}
	lowerIndex, lowerFound := r.Binsearch(lower)
	upperIndex, upperFound := r[lowerIndex:].Binsearch(upper)
	upperIndex += lowerIndex
	if !lowerFound && !upperFound {
		return r[lowerIndex:upperIndex]
	}
	if upperFound {
		upperIndex++
	}
	retval := slices.Clone(r[lowerIndex:upperIndex])
	if lowerFound {
		retval[0].start = lower
	}
	if upperFound {
		retval[len(retval)-1].end = upper
	}
	return retval
}

func (r IPRanges) Intersect(other IPRanges) IPRanges {
	if r.IsEmpty() || other.IsEmpty() {
		return nil
	}
	if r.Contains(other) {
		return other
	}
	if other.Contains(r) {
		return r
	}

	index, otherIndex := 0, 0
	var retval IPRanges
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
