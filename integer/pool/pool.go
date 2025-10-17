package pool

import (
	"strings"

	"golang.org/x/exp/constraints"

	"github.com/qiuchengxuan/go-types/integer/ranges"
)

type Pool[I constraints.Integer] struct{ capacity, available, exceed ranges.Ranges[I] }

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

func (p *Pool[I]) Capacity() ranges.Ranges[I] { return p.capacity }

func (p *Pool[I]) InUse() ranges.Ranges[I] { return p.capacity.Sub(p.available).Add(p.exceed) }

func (p *Pool[I]) Available() ranges.Ranges[I] { return p.available }

func (p *Pool[I]) Exceed() ranges.Ranges[I] { return p.exceed }

func (p *Pool[I]) LazyInUse() ranges.LazyRanges[I] {
	return ranges.LazySub(p.capacity.Add(p.exceed), p.available)
}

func (p *Pool[I]) IsEmpty() bool { return p.available.IsEmpty() }

func (p *Pool[I]) IsFull() bool { return p.available.Equal(p.capacity) }

func (p *Pool[I]) Allocated() bool { return !p.IsFull() || p.exceed.IsEmpty() }

func (p Pool[I]) Clone() Pool[I] {
	return Pool[I]{p.capacity.Clone(), p.available.Clone(), p.exceed.Clone()}
}

func (p *Pool[I]) Allocate(opts ...I) (I, bool) {
	if len(opts) > 0 {
		if p.available.Assign().RemoveScalar(opts[0]) {
			return opts[0], true
		}
		return 0, false
	}
	return p.available.Assign().Pop()
}

func (p *Pool[I]) ExceedAllocate(value I) bool {
	if p.capacity.Has(value) {
		return p.available.Assign().RemoveScalar(value)
	}
	retval := !p.exceed.Has(value)
	p.exceed.Assign().AddScalar(value)
	return retval
}

func (p *Pool[I]) Release(x I) bool {
	if !p.capacity.Has(x) {
		return p.exceed.Assign().RemoveScalar(x)
	}
	p.available.Assign().AddScalar(x)
	return true
}

func (p *Pool[I]) SetAvailable(available ranges.Ranges[I]) {
	p.available = p.capacity.Intersect(available).Clone()
}

func (p *Pool[I]) Reset(capacity ranges.Ranges[I]) {
	inuse := p.InUse()
	p.exceed = inuse.Sub(capacity)
	p.available = capacity.Sub(inuse).Clone()
	p.capacity = capacity.Clone()
}

func FromRanges[I constraints.Integer](intRanges ranges.Ranges[I]) Pool[I] {
	return Pool[I]{intRanges.Clone(), intRanges.Clone(), nil}
}

func Empty[I constraints.Integer]() Pool[I] { return Pool[I]{nil, nil, nil} }

func FromStr[I constraints.Integer](str string) Pool[I] {
	return FromRanges(ranges.FromStr[I](str))
}

func FromRange[I constraints.Integer](start, stop I) Pool[I] {
	if start > stop {
		return Pool[I]{}
	}

	capacity := ranges.FromTo(start, stop).Ranges()
	return Pool[I]{capacity, capacity.Clone(), nil}
}
