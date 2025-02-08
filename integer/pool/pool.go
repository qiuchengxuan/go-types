package pool

import (
	"strings"

	"golang.org/x/exp/constraints"

	"github.com/qiuchengxuan/go-types/integer/ranges"
)

type Pool[I constraints.Integer] struct{ capacity, available ranges.Ranges[I] }

func (p Pool[I]) String() string {
	if p.IsEmpty() {
		return ""
	}
	var buf strings.Builder
	p.available.WriteTo(&buf)
	buf.WriteRune('/')
	p.capacity.WriteTo(&buf)
	return buf.String()
}

func (p *Pool[I]) Capacity() ranges.Ranges[I] {
	return p.capacity
}

func (p *Pool[I]) Available() ranges.Ranges[I] {
	return p.available
}

func (p *Pool[I]) InUse() ranges.LazyRanges[I] {
	return ranges.LazySub(p.capacity, p.available)
}

func (p *Pool[I]) IsEmpty() bool {
	return p.available.IsEmpty()
}

func (p *Pool[I]) IsFull() bool {
	return p.available.Equal(p.capacity)
}

func (p Pool[I]) Clone() Pool[I] {
	return Pool[I]{capacity: p.capacity.Clone(), available: p.available.Clone()}
}

func (p *Pool[I]) Allocate(opts ...I) (I, bool) {
	if len(opts) > 0 {
		if p.available.Remove(opts[0]) {
			return opts[0], true
		}
		return 0, false
	}
	return p.available.Pop()
}

func (p *Pool[I]) Release(x I) bool {
	if !p.capacity.Contains(x) {
		return false
	}
	p.available = p.available.AddScalar(x)
	return true
}

func (p *Pool[I]) SetAvailable(available ranges.Ranges[I]) {
	p.available = p.capacity.Intersect(available)
}

func FromRanges[I constraints.Integer](intRanges ranges.Ranges[I]) Pool[I] {
	capacity := make(ranges.Ranges[I], len(intRanges))
	available := make(ranges.Ranges[I], len(intRanges))
	copy(capacity, intRanges)
	copy(available, intRanges)
	return Pool[I]{capacity, available}
}

func Empty[I constraints.Integer]() Pool[I] {
	return Pool[I]{nil, nil}
}

func FromStr[I constraints.Integer](str string) Pool[I] {
	return FromRanges(ranges.FromStr[I](str))
}

func FromRange[I constraints.Integer](start, stop I) Pool[I] {
	if start > stop {
		return Pool[I]{nil, nil}
	}

	capacity := ranges.FromTo(start, stop).Plural()
	available := ranges.FromTo(start, stop).Plural()
	return Pool[I]{capacity, available}
}
