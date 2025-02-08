package ranges

import (
	"fmt"
	"strconv"

	"golang.org/x/exp/constraints"
)

type Range[I constraints.Integer] struct{ start, end I } // end is closed

func (r Range[I]) has(x I) bool {
	return r.start <= x && x <= r.end
}

func (r Range[I]) Contains(other Range[I]) bool {
	return r.start <= other.start && other.end <= r.end
}

func (r Range[I]) Intersect(other Range[I]) Range[I] {
	if r.start > other.start {
		r, other = other, r
	}
	switch {
	case r.end < other.start:
		return Range[I]{1, 0}
	case other.start <= r.end && r.end <= other.end:
		return Range[I]{other.start, r.end}
	default:
		return Range[I]{other.start, other.end}
	}
}

func (r Range[I]) Len() uint64 {
	if r.end < r.start {
		return uint64(0)
	}
	return uint64(r.end) - uint64(r.start) + 1
}

func (r Range[I]) String() string {
	switch {
	case r.end < r.start:
		return ""
	case r.end-r.start == 0:
		return strconv.Itoa(int(r.start))
	default:
		return fmt.Sprintf("%d-%d", r.start, r.end)
	}
}

func (r Range[I]) Start() I {
	return r.start
}

func (r Range[I]) End() I {
	return r.end
}

func (r *Range[I]) expandLower(size I) {
	r.start -= size
}

func (r *Range[I]) expandUpper(size I) {
	r.end += size
}

func (r *Range[I]) shrinkLower(size I) {
	r.start += size
}

func (r *Range[I]) shrinkUpper(size I) {
	r.end -= size
}

func (r Range[I]) StartOf(start I) Range[I] {
	return Range[I]{start: start, end: r.end}
}

func (r Range[I]) EndOf(end I) Range[I] {
	return Range[I]{start: r.start, end: end}
}

func FromTo[I constraints.Integer](from, to I) Range[I] {
	return Range[I]{start: from, end: to}
}

func Of[I constraints.Integer](v I) Range[I] {
	return Range[I]{start: v, end: v}
}

func Empty[I constraints.Integer]() Range[I] {
	return Range[I]{start: I(1), end: I(0)}
}
