package ranges

import "slices"

func (r Ranges[I]) Limit(bound Range[I]) Ranges[I] {
	lower, upper := bound.start, bound.end
	if r.IsEmpty() || lower <= r.First() && r.Last() <= upper {
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

func (r Ranges[I]) Intersect(other Ranges[I]) Ranges[I] {
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
	var retval Ranges[I]
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
