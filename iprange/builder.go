package iprange

import (
	"net"
	"strings"

	"golang.org/x/exp/constraints"

	"github.com/qiuchengxuan/go-datastructures/integer/ranges"
	"github.com/qiuchengxuan/go-datastructures/integer/uint128"
	"github.com/qiuchengxuan/go-datastructures/ip"
)

func FromIntRanges[I constraints.Integer](base ip.IP, ranges ranges.Ranges[I]) IPRanges {
	ipv6Ranges := make(IPRanges, len(ranges))
	for i, r := range ranges {
		r := IPRange{base.Add(uint64(r.Start())), base.Add(uint64(r.End()))}
		ipv6Ranges[i] = r
	}
	return ipv6Ranges
}

func FromIPNet(ipNet net.IPNet, subnet bool) IPRanges {
	if ipNet.IP == nil {
		return IPRanges{}
	}
	first := ip.From(ipNet.IP)
	ones, bits := ipNet.Mask.Size()
	if subnet && ones == bits {
		return IPRanges{}
	}
	last := first.AddExt(uint128.FromPrimitive(1).Shl(uint(bits - ones)).SubU64(1))
	if subnet {
		first = first.Add(1)
		if last.IsV4() {
			last = last.Sub(1)
		}
	}
	return FromTo(first, last).Ranges()
}

func FromSlice(slice []ip.IP) IPRanges {
	if len(slice) == 0 {
		return nil
	}
	var retval IPRanges
	pending := IPRange{slice[0], slice[0]}
	for _, value := range slice[1:] {
		if value.LessOrEqualThan(pending.end) {
			continue
		}
		if value == pending.end.Add(1) {
			pending.end = value
			continue
		}
		retval = append(retval, pending)
		pending = IPRange{value, value}
	}
	return append(retval, pending)
}

func FromStr(ipRanges string) IPRanges {
	if ipRanges == "" {
		return IPRanges{}
	}

	splitted := strings.Split(ipRanges, ",")
	ranges := make([]IPRange, len(splitted))

	for i, ipRange := range splitted {
		before, after, ok := strings.Cut(ipRange, "-")
		if ok && after == "" {
			return nil
		}
		startIP := net.ParseIP(before)
		if startIP == nil {
			return nil
		}
		start := ip.From(startIP)
		if after == "" {
			ranges[i] = IPRange{start, start}
			continue
		}
		stopIP := net.ParseIP(after)
		if stopIP == nil {
			return nil
		}
		stop := ip.From(stopIP)
		if start.GreaterThan(stop) {
			return nil
		}
		ranges[i] = IPRange{start, stop}
	}

	index := 0
	for _, ipRange := range ranges[1:] {
		last := &ranges[index]
		switch {
		case ipRange.start.LessThan(last.end):
			continue
		case ipRange.start == last.end:
			fallthrough
		case ipRange.start == last.end.Add(1):
			last.end = ipRange.end
		default:
			index++
			ranges[index] = ipRange
		}
	}
	return ranges[:index+1]
}
