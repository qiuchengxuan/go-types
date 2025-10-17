package ranges

func (r Ranges[I]) Add(other Ranges[I]) Ranges[I] {
	if len(other) == 0 || r.Contains(other) {
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

func (r Ranges[I]) AddNaive(naive Range[I]) Ranges[I] {
	if naive.Len() == 0 || r.ContainsNaive(naive) {
		return r
	}
	other := [1]Range[I]{naive}
	return r.Add(other[:])
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
		case right.start <= start && left.end <= right.end:
			index++ // right contains left, skip left
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

func (r Ranges[I]) AddScalar(value I) Ranges[I] {
	if r.Has(value) {
		return r
	}
	return r.Clone().Ref().Assign().AddScalar(value)
}

func (r Ranges[I]) SubScalar(value I) Ranges[I] {
	if !r.Has(value) {
		return r
	}
	return r.Clone().Ref().Assign().SubScalar(value)
}
