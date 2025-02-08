package iprange

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/qiuchengxuan/go-types/integer/ranges"
	"github.com/qiuchengxuan/go-types/integer/uint128"
	"github.com/qiuchengxuan/go-types/ip"
)

type IPRanges struct {
	V4 IPv4Ranges
	V6 IPv6Ranges
}

func (r IPRanges) IsEmpty() bool {
	return r.V4.IsEmpty() && r.V6.IsEmpty()
}

func (r IPRanges) IsDualStack() bool {
	return !r.V4.IsEmpty() && !r.V6.IsEmpty()
}

func (r IPRanges) Equal(other IPRanges) bool {
	return r.V4.Equal(other.V4) && r.V6.Equal(other.V6)
}

func (r IPRanges) Contains(value net.IP) bool {
	if value.To4() != nil {
		return r.V4.Contains(value)
	}
	return r.V6.Contains(ip.From(value))
}

func (r IPRanges) Foreach(consumer func(net.IP)) {
	r.V4.Foreach(consumer)
	r.V6.Foreach(consumer)
}

func (r IPRanges) Add(rhs IPRanges) IPRanges {
	return IPRanges{r.V4.Add(rhs.V4), r.V6.Add(rhs.V6)}
}

func (r IPRanges) Sub(rhs IPRanges) IPRanges {
	return IPRanges{r.V4.Sub(rhs.V4), r.V6.Sub(rhs.V6)}
}

func (r IPRanges) AddIP(value net.IP) IPRanges {
	if value.To4() != nil {
		return IPRanges{r.V4.AddIP(value), r.V6}
	}
	return IPRanges{r.V4, r.V6.AddIP(ip.From(value))}
}

func (r IPRanges) SubIP(value net.IP) IPRanges {
	if value.To4() != nil {
		return IPRanges{r.V4.SubIP(value), r.V6}
	}
	return IPRanges{r.V4, r.V6.SubIP(ip.From(value))}
}

func (r IPRanges) Intersect(other IPRanges) IPRanges {
	return IPRanges{r.V4.Intersect(other.V4), r.V6.Intersect(other.V6)}
}

func (r IPRanges) WriteTo(buf *strings.Builder) {
	if !r.V4.IsEmpty() {
		r.V4.WriteTo(buf)
		if !r.V6.IsEmpty() {
			buf.WriteRune(',')
		}
	}
	if !r.V6.IsEmpty() {
		r.V6.WriteTo(buf)
	}
}

func (r IPRanges) String() string {
	var buf strings.Builder
	r.WriteTo(&buf)
	return buf.String()
}

func (r IPRanges) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

func (r *IPRanges) UnmarshalText(text []byte) error {
	*r = FromStr(string(text))
	if r.IsEmpty() && len(text) > 0 {
		return errors.New("Malformed ip ranges")
	}
	return nil
}

func (r IPRanges) MarshalBinary() ([]byte, error) {
	if r.V4.IsEmpty() && r.V6.IsEmpty() {
		return nil, nil
	}
	data, _ := r.V6.MarshalBinary()
	v4data, _ := r.V4.MarshalBinary()
	var meta [8]byte
	binary.BigEndian.PutUint32(meta[:], uint32(len(data)))
	binary.BigEndian.PutUint32(meta[4:], uint32(len(data)+len(v4data)))
	return append(append(data, v4data...), meta[:]...), nil
}

func (r *IPRanges) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) < 16 {
		return errors.New("Not ip ranges")
	}
	total := binary.BigEndian.Uint32(data[len(data)-4:])
	if len(data)-8 != int(total) {
		return errors.New("Not ip ranges")
	}
	v6Size := binary.BigEndian.Uint32(data[len(data)-8:])
	data = data[:len(data)-8]
	if err := r.V6.UnmarshalBinary(data[:v6Size]); err != nil {
		return err
	}
	return r.V4.UnmarshalBinary(data[v6Size:])
}

func FromIPNet(ipNet net.IPNet) IPRanges {
	if ipNet.IP == nil {
		return IPRanges{}
	}
	if ipNet.IP.To4() != nil {
		prefix, length := ipNet.Mask.Size()
		if prefix >= 31 {
			return Empty()
		}
		start := binary.BigEndian.Uint32(ipNet.IP.To4())
		stop := start + 1<<int(length-prefix) - 1
		return IPRanges{V4: IPv4Ranges{ranges.FromTo(start, stop).Plural()}}
	}
	prefix, length := ipNet.Mask.Size()
	if prefix >= 128 {
		return Empty()
	}
	start := ip.FromBytes(ipNet.IP)
	stop := start.Add(uint128.FromPrimitive(1).Shl(uint(length - prefix)).SubU64(1))
	return IPRanges{V6: IPv6Ranges{{start, stop}}}
}

func Empty() IPRanges {
	return IPRanges{}
}

func FromStr(ipRanges string) IPRanges {
	if ipRanges == "" {
		return IPRanges{}
	}
	index := strings.IndexRune(ipRanges, ':')
	if index < 0 {
		return IPRanges{V4: v4FromStr(ipRanges)}
	}
	index = strings.LastIndexByte(ipRanges[:index], ',')
	if index < 0 {
		return IPRanges{V6: v6FromStr(ipRanges)}
	}
	return IPRanges{V4: v4FromStr(ipRanges[:index]), V6: v6FromStr(ipRanges[index+1:])}
}

func FromNetIPs(addrs []net.IP) IPRanges {
	v4 := make([]uint32, 0, len(addrs))
	v6 := make([]ip.IP, 0, len(addrs))
	for _, addr := range addrs {
		if ip4 := addr.To4(); ip4 != nil {
			v4 = append(v4, binary.BigEndian.Uint32(ip4))
		} else {
			v6 = append(v6, ip.FromBytes(addr))
		}
	}
	slices.Sort(v4)
	slices.SortFunc(v6, ip.Compare)
	return IPRanges{IPv4Ranges{ranges.FromSlice(v4)}, v6FromSlice(v6)}
}
