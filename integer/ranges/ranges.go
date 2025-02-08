package ranges

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type Ranges[I constraints.Integer] []Range[I]

func (r Range[I]) Plural() Ranges[I] {
	if r.Len() == 0 {
		return nil
	}
	return Ranges[I]{r}
}

func (r Ranges[I]) Equal(other Ranges[I]) bool {
	return slices.Equal(r, other)
}

func (r Ranges[I]) binsearch(x I) (int, bool) {
	index := sort.Search(len(r), func(i int) bool { return x <= r[i].end })
	switch {
	case index == -1:
		return len(r), false
	case index >= len(r):
		return index, false
	default:
		return index, r[index].has(x)
	}
}

func (r Ranges[I]) Binsearch(x I) int {
	if len(r) == 0 || x < r[0].start {
		return -1
	}
	index := sort.Search(len(r), func(i int) bool { return x <= r[i].end })
	switch {
	case index == -1:
		return index
	case index >= len(r):
		return -int(r.Len())
	default:
		offset := int(r[:index].Len())
		if r[index].has(x) {
			return offset + int(x-r[index].start)
		}
		return -offset - 1
	}
}

func (r Ranges[I]) Contains(x I) bool {
	if len(r) == 0 {
		return false
	}
	if x < r[0].start || x > r[len(r)-1].end {
		return false
	}

	_, found := r.binsearch(x)
	return found
}

func (r Ranges[I]) Len() uint64 {
	sum := uint64(0)
	for _, r := range r {
		sum += r.Len()
	}
	return sum
}

func (r Ranges[I]) IsEmpty() bool {
	return len(r) == 0
}

func (r Ranges[I]) WriteTo(buf *strings.Builder) {
	if len(r) == 0 {
		return
	}
	for i := 0; i < len(r)-1; i++ {
		buf.WriteString(r[i].String())
		buf.WriteRune(',')
	}
	buf.WriteString(r[len(r)-1].String())
}

func (r Ranges[I]) String() string {
	if len(r) == 0 {
		return ""
	}
	var buf strings.Builder
	r.WriteTo(&buf)
	return buf.String()
}

func (r Ranges[I]) Clone() Ranges[I] {
	ranges := make(Ranges[I], len(r))
	copy(ranges, r)
	return ranges
}

func (r Ranges[I]) Iter(fn func(x I) bool) (I, bool) {
	for i := 0; i < len(r); i++ {
		for v := r[i].start; v <= r[i].end; v++ {
			if fn(v) {
				return v, true
			}
		}
	}
	return 0, false
}

func (r Ranges[I]) NumChunks() int { return len(r) }

func (r Ranges[I]) IterChunks(fn func(Range[I]) bool) Range[I] {
	for _, chunk := range r {
		if fn(chunk) {
			return chunk
		}
	}
	return Range[I]{1, 0}
}

func (r Ranges[I]) Foreach(consumer func(x I)) {
	for i := 0; i < len(r); i++ {
		for v := r[i].start; v <= r[i].end; v++ {
			consumer(v)
		}
	}
}

func (r Ranges[I]) First() I {
	return r[0].start
}

func (r Ranges[I]) Last() I {
	return r[len(r)-1].end
}

func (r *Ranges[I]) Pop() (I, bool) {
	if len(*r) == 0 {
		return 0, false
	}
	firstInterval := (*r)[0]
	if firstInterval.Len() == 1 {
		*r = (*r)[1:]
	} else {
		(*r)[0].shrinkLower(1)
	}
	return firstInterval.start, true
}

func (r Ranges[I]) Index(index uint64) (I, bool) {
	for i := 0; i < len(r); i++ {
		if index < r[i].Len() {
			return r[i].start + I(index), true
		}
		index -= r[i].Len()
	}
	return 0, false
}

func (r Ranges[I]) Limit(lower, upper I) Ranges[I] {
	if lower > upper {
		panic(fmt.Sprintf("Lower %d > upper %d", lower, upper))
	}
	if r.IsEmpty() || lower <= r.First() && r.Last() <= upper {
		return r
	}
	lowerIndex, lowerFound := r.binsearch(lower)
	upperIndex, upperFound := r.binsearch(upper)
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

func Cast[I, T constraints.Integer](ranges Ranges[I]) (Ranges[T], bool) {
	if ranges.IsEmpty() {
		return nil, false
	}
	downcast := unsafe.Sizeof(I(0)) > unsafe.Sizeof(T(0))
	minimum, maximum := I(0), ^I(0)
	iSigned, signed := ^I(0) < 0, ^T(0) < 0
	if iSigned == signed && signed {
		if downcast {
			minimum = I(T(1 << (unsafe.Sizeof(T(0))*8 - 1)))
		} else {
			minimum = 1 << (unsafe.Sizeof(I(0))*8 - 1)
		}
	}
	if upcast := unsafe.Sizeof(I(0)) < unsafe.Sizeof(T(0)); upcast {
		if iSigned {
			maximum = I(1<<(unsafe.Sizeof(I(0))*8-1)) - 1
		}
	} else {
		switch {
		case signed:
			maximum = I(T(1<<(unsafe.Sizeof(T(0))*8-1)) - 1)
		case unsafe.Sizeof(I(0)) == unsafe.Sizeof(T(0)) && iSigned:
			maximum = I(1<<(unsafe.Sizeof(I(0))*8-1)) - 1
		case downcast && !signed:
			maximum = I(^T(0))
		}
	}
	limited := ranges.Limit(minimum, maximum)
	retval := make(Ranges[T], len(limited))
	for i, r := range limited {
		retval[i] = Range[T]{start: T(r.start), end: T(r.end)}
	}
	return retval, !limited.Equal(ranges)
}

// Caller must ensure slice ordered, otherwise result probably invalid
func FromSlice[I constraints.Integer](slice []I) Ranges[I] {
	if len(slice) == 0 {
		return Empty[I]().Plural()
	}
	var retval Ranges[I]
	pending := Of(slice[0])
	for _, value := range slice[1:] {
		if value <= pending.end {
			continue
		}
		if value == pending.end+1 {
			pending.end = value
			continue
		}
		retval = append(retval, pending)
		pending = Of(value)
	}
	return append(retval, pending)
}

func mergeRanges[I constraints.Integer](rangeList []Range[I]) Ranges[I] {
	var ranges Ranges[I]
	for _, ranging := range rangeList {
		ranges = ranges.Add(ranging.Plural())
	}
	return ranges
}

func FromRanges[I constraints.Integer](ranges []Range[I]) Ranges[I] {
	if len(ranges) == 0 {
		return nil
	}
	index := 0
	for _, intRange := range ranges[1:] {
		last := &ranges[index]
		switch {
		case intRange.start < last.end:
			return mergeRanges(ranges) // Won't be affected by ranges change
		case intRange.start == last.end:
			return mergeRanges(ranges)
		case intRange.start == last.end+1:
			last.end = intRange.end
		default:
			index++
			ranges[index] = intRange
		}
	}
	return ranges[:index+1]
}

func FromStr[I constraints.Integer](str string) Ranges[I] {
	splitted := strings.Split(str, ",")
	ranges := make([]Range[I], 0, len(splitted))
	for _, text := range splitted {
		before, after, ok := strings.Cut(text, "-")
		if ok && after == "" {
			return nil
		}
		start, err := strconv.Atoi(before)
		if err != nil {
			return nil
		}

		if int(I(start)) != start {
			return nil
		}

		if after == "" {
			ranges = append(ranges, Of(I(start)))
			continue
		}

		end, err := strconv.Atoi(after)
		if err != nil {
			return nil
		}
		if start > end || int(I(end)) != end {
			return nil
		}
		ranges = append(ranges, FromTo(I(start), I(end)))
	}
	return FromRanges(ranges)
}
