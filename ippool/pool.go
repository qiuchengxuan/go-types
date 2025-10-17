package ippool

import (
	"strings"

	"github.com/qiuchengxuan/go-datastructures/ip"
	"github.com/qiuchengxuan/go-datastructures/iprange"
)

type Pool struct{ capacity, available, exceed iprange.IPRanges }

func (p Pool) String() string {
	if p.capacity.IsEmpty() {
		return ""
	}
	var buf strings.Builder
	p.available.WriteTo(&buf)
	buf.WriteRune('/')
	p.capacity.WriteTo(&buf)
	return buf.String()
}

func (p *Pool) Capacity() iprange.IPRanges { return p.capacity }

func (p *Pool) InUse() iprange.IPRanges { return p.capacity.Sub(p.available).Add(p.exceed) }

func (p *Pool) Available() iprange.IPRanges { return p.available }

func (p *Pool) Exeed() iprange.IPRanges { return p.exceed }

func (p *Pool) IsEmpty() bool { return p.available.IsEmpty() }

func (p *Pool) IsFull() bool { return p.available.Equal(p.capacity) }

func (p *Pool) Allocated() bool { return !p.IsFull() || p.exceed.IsEmpty() }

func (p Pool) Clone() Pool {
	return Pool{p.capacity.Clone(), p.available.Clone(), p.exceed.Clone()}
}

func (p *Pool) Allocate(opts ...ip.IP) (ip.IP, bool) {
	if len(opts) > 0 {
		if p.available.Assign().RemoveIP(opts[0]) {
			return opts[0], true
		}
		return ip.IP{}, false
	}
	return p.available.Assign().Pop()
}

func (p *Pool) ExceedAllocate(value ip.IP) bool {
	if p.capacity.Has(value) {
		return p.available.Assign().RemoveIP(value)
	}
	retval := !p.exceed.Has(value)
	p.exceed.Assign().AddIP(value)
	return retval
}

func (p *Pool) Release(x ip.IP) bool {
	if !p.capacity.Has(x) {
		return p.exceed.Assign().RemoveIP(x)
	}
	p.available.Assign().AddIP(x)
	return true
}

func (p *Pool) SetAvailable(available iprange.IPRanges) {
	p.available = p.capacity.Intersect(available)
}

func (p *Pool) Reset(capacity iprange.IPRanges) {
	inuse := p.InUse()
	p.exceed = inuse.Sub(capacity)
	p.available = capacity.Sub(inuse).Clone()
	p.capacity = capacity.Clone()
}

func FromRanges(ipRanges iprange.IPRanges) Pool {
	return Pool{ipRanges.Clone(), ipRanges.Clone(), nil}
}

func Empty() Pool { return Pool{} }

func FromStr(str string) Pool { return FromRanges(iprange.FromStr(str)) }

func FromRange(start, stop ip.IP) Pool {
	if start.GreaterThan(stop) {
		return Pool{}
	}

	capacity := iprange.FromTo(start, stop).Ranges()
	return Pool{capacity, capacity.Clone(), nil}
}
