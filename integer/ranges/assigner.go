package ranges

import (
	"golang.org/x/exp/constraints"
)

// Slightly reduce memory allocation
type Assigner[I constraints.Integer] struct{ *Ranges[I] }

func (a Assigner[I]) Into() Ranges[I] { return Ranges[I](*a.Ranges) }

func (a Assigner[I]) Add(other Ranges[I]) Ranges[I] {
	mayOverflow := !a.IsEmpty() && a.Last()+1 < a.Last()
	switch {
	case other.IsEmpty():
		return *a.Ranges
	case a.IsEmpty():
		*a.Ranges = other
	case !mayOverflow && a.Last()+1 < other.First():
		*a.Ranges = append(*a.Ranges, other...)
	case !mayOverflow && a.Last()+1 == other.First():
		a.lastChunk().end = other[0].end
		*a.Ranges = append(*a.Ranges, other[1:]...)
	default:
		*a.Ranges = a.Ranges.Add(other)
	}
	return *a.Ranges
}

func (a Assigner[I]) Sub(other Ranges[I]) Ranges[I] {
	*a.Ranges = a.Ranges.Sub(other)
	return *a.Ranges
}

func (a Assigner[I]) AddNaive(naive Range[I]) Ranges[I] {
	if naive.Len() == 0 {
		return *a.Ranges
	}
	other := [1]Range[I]{naive}
	return a.Add(other[:])
}

func (a Assigner[I]) addScalar(value I) {
	index, found := a.Binsearch(value)
	if found {
		return
	}
	ranges := *a.Ranges
	switch {
	case 0 < index && ranges[index-1].end+1 == value:
		ranges[index-1].end++
	case index < len(ranges) && ranges[index].start-1 == value:
		ranges[index].start--
	default:
		ranges := append(ranges, Range[I]{})
		copy(ranges[index+1:], ranges[index:len(ranges)-1])
		ranges[index] = Of(value)
		*a.Ranges = ranges
	}
	if 0 < index && index < len(ranges) && ranges[index-1].end+1 == ranges[index].start {
		ranges[index-1].end = ranges[index].end
		copy(ranges[index:len(ranges)-1], ranges[index+1:])
		*a.Ranges = ranges[:len(ranges)-1]
	}
}

func (a Assigner[I]) AddScalar(value I) Ranges[I] {
	mayOverflow := !a.IsEmpty() && a.Last()+1 < a.Last()
	switch {
	case a.Has(value):
		return *a.Ranges
	case a.IsEmpty() || !mayOverflow && a.Last()+1 < value:
		*a.Ranges = append(*a.Ranges, Of(value))
	case !mayOverflow && a.Last()+1 == value:
		a.lastChunk().end = value
	default:
		a.addScalar(value)
	}
	return *a.Ranges
}

func (a Assigner[I]) RemoveScalar(value I) bool {
	if a.IsEmpty() || value < a.First() || a.Last() < value {
		return false
	}
	index, found := a.Binsearch(value)
	if !found {
		return false
	}
	ranges := *a.Ranges
	switch entry := &ranges[index]; {
	case entry.Len() == 1:
		*a.Ranges = append(ranges[:index], ranges[index+1:]...) //nolint: gocritic
		return true
	case value == entry.start:
		entry.start++
	case value == entry.end:
		entry.end--
	default:
		ranges = append(ranges[:index+1], ranges[index:]...)
		ranges[index] = ranges[index].EndOf(value - 1)
		ranges[index+1] = ranges[index+1].StartOf(value + 1)
		*a.Ranges = ranges
	}
	return true
}

func (a Assigner[I]) SubScalar(value I) Ranges[I] {
	a.RemoveScalar(value)
	return *a.Ranges
}

func (a Assigner[I]) Pop() (I, bool) {
	if a.IsEmpty() {
		return 0, false
	}
	retval := a.First()
	a.firstChunk().start++
	if a.firstChunk().Len() == 0 {
		*a.Ranges = (*a.Ranges)[1:]
	}
	return retval, true
}
