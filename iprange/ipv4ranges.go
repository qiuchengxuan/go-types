package iprange

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/qiuchengxuan/go-types/integer/ranges"
)

type IPv4Range struct {
	ranges.Range[uint32]
}

func (r IPv4Range) WriteTo(buf *strings.Builder) {
	bytes := [net.IPv4len]byte{}
	binary.BigEndian.PutUint32(bytes[:], uint32(r.Start()))
	start := net.IP(bytes[:]).String()
	buf.WriteString(start)
	if r.Start() != r.End() {
		binary.BigEndian.PutUint32(bytes[:], uint32(r.End()))
		buf.WriteRune('-')
		buf.WriteString(net.IP(bytes[:]).String())
	}
}

type IPv4Ranges struct {
	ranges.Ranges[uint32]
}

func (r IPv4Ranges) Equal(other IPv4Ranges) bool {
	return r.Ranges.Equal(other.Ranges)
}

func (r IPv4Ranges) Binsearch(ip net.IP) int {
	if ip = ip.To4(); ip == nil {
		panic(fmt.Errorf("IP %v must be IPv4", ip))
	}
	return r.Ranges.Binsearch(binary.BigEndian.Uint32(ip))
}

func (r IPv4Ranges) WriteTo(buf *strings.Builder) {
	if r.IsEmpty() {
		return
	}
	for i := 0; i < len(r.Ranges)-1; i++ {
		IPv4Range{r.Ranges[i]}.WriteTo(buf)
		buf.WriteRune(',')
	}
	IPv4Range{r.Ranges[len(r.Ranges)-1]}.WriteTo(buf)
}

func (r IPv4Ranges) String() string {
	var buf strings.Builder
	r.WriteTo(&buf)
	return buf.String()
}

func (r IPv4Ranges) Foreach(consumer func(net.IP)) {
	r.Ranges.Foreach(func(x uint32) {
		ip := make(net.IP, net.IPv4len)
		binary.BigEndian.PutUint32(ip, x)
		consumer(ip)
	})
}

func (r IPv4Ranges) IterChunks(fn func(IPv4Range) bool) IPv4Range {
	for _, r := range r.Ranges {
		if fn(IPv4Range{r}) {
			return IPv4Range{r}
		}
	}
	return IPv4Range{ranges.Empty[uint32]()}
}

func (r IPv4Ranges) AddIP(ip net.IP) IPv4Ranges {
	if ip == nil || ip.To4() == nil {
		return r
	}
	return IPv4Ranges{r.Ranges.AddScalar(binary.BigEndian.Uint32(ip.To4()))}
}

func (r IPv4Ranges) SubIP(ip net.IP) IPv4Ranges {
	if ip == nil || ip.To4() == nil {
		return r
	}
	return IPv4Ranges{r.Ranges.RemoveScalar(binary.BigEndian.Uint32(ip.To4()))}
}

func (r IPv4Ranges) Contains(ip net.IP) bool {
	if ip == nil || r.IsEmpty() || ip.To4() == nil {
		return false
	}
	return r.Ranges.Contains(binary.BigEndian.Uint32(ip.To4()))
}

func (r IPv4Ranges) Index(index uint32) net.IP {
	v, ok := (&r.Ranges).Index(uint64(index))
	if !ok {
		return nil
	}
	ip := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(ip, uint32(v))
	return ip
}

func (r *IPv4Ranges) Pop() net.IP {
	if v, ok := r.Ranges.Pop(); ok {
		ip := make(net.IP, net.IPv4len)
		binary.BigEndian.PutUint32(ip, uint32(v))
		return ip
	}
	return nil
}

func (r IPv4Ranges) Intersect(other IPv4Ranges) IPv4Ranges {
	return IPv4Ranges{r.Ranges.Intersect(other.Ranges)}
}

func (r IPv4Ranges) Add(rhs IPv4Ranges) IPv4Ranges {
	return IPv4Ranges{r.Ranges.Add(rhs.Ranges)}
}

func (r IPv4Ranges) Sub(other IPv4Ranges) IPv4Ranges {
	return IPv4Ranges{r.Ranges.Sub(other.Ranges)}
}

func (r IPv4Ranges) Clone() IPv4Ranges {
	return IPv4Ranges{r.Ranges.Clone()}
}

func (r IPv4Ranges) Find(pattern func(ip net.IP) bool) net.IP {
	ip := make(net.IP, net.IPv4len)
	v, ok := r.Ranges.Iter(func(x uint32) bool {
		binary.BigEndian.PutUint32(ip, x)
		return pattern(ip)
	})
	if ok {
		binary.BigEndian.PutUint32(ip, uint32(v))
		return ip
	}
	return nil
}

func (r IPv4Ranges) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

func (r *IPv4Ranges) UnmarshalText(text []byte) error {
	*r = v4FromStr(string(text))
	if r.IsEmpty() && len(text) > 0 {
		return errors.New("Malformed ipv4 ranges")
	}
	return nil
}

func v4FromStr(ipRanges string) IPv4Ranges {
	if ipRanges == "" {
		return IPv4Ranges{nil}
	}

	splitted := strings.Split(ipRanges, ",")
	intRanges := make([]ranges.Range[uint32], 0, len(splitted))

	for _, ipRange := range splitted {
		before, after, ok := strings.Cut(ipRange, "-")
		if ok && after == "" {
			return IPv4Ranges{nil}
		}
		startIP := net.ParseIP(before).To4()
		if startIP == nil {
			return IPv4Ranges{nil}
		}
		start := binary.BigEndian.Uint32(startIP)
		if after == "" {
			intRanges = append(intRanges, ranges.Of(start))
			continue
		}
		stopIP := net.ParseIP(after).To4()
		if stopIP == nil {
			return IPv4Ranges{nil}
		}
		stop := binary.BigEndian.Uint32(stopIP)
		if start > stop {
			return IPv4Ranges{nil}
		}
		intRanges = append(intRanges, ranges.FromTo(start, stop))
	}
	return IPv4Ranges{ranges.FromRanges(intRanges)}
}
