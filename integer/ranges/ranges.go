package ranges

import (
	"cmp"
	"iter"
	"slices"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type Ranges[I constraints.Integer] []Range[I]

func (r Ranges[I]) Equal(other Ranges[I]) bool {
	return slices.Equal(r, other)
}

func (r Ranges[I]) Binsearch(value I) (int, bool) {
	return slices.BinarySearchFunc(r, value, func(item Range[I], value I) int {
		if item.has(value) {
			return 0
		}
		return cmp.Compare(item.start, value)
	})
}

func (r Ranges[I]) Contains(other Ranges[I]) bool {
	index := 0
	for _, chunk := range other {
		offset, ok := r[index:].Binsearch(chunk.start)
		if !ok {
			return false
		}
		index += offset
		if r[index].end < chunk.end {
			return false
		}
	}
	return true
}

func (r Ranges[I]) Has(x I) bool {
	if len(r) == 0 {
		return false
	}
	if x < r[0].start || x > r[len(r)-1].end {
		return false
	}

	_, found := r.Binsearch(x)
	return found
}

func (r Ranges[I]) ContainsNaive(naive Range[I]) bool {
	index, ok := r.Binsearch(naive.start)
	return ok && naive.end <= r[index].end
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
	for i := range len(r) - 1 {
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

func (r Ranges[I]) Clone() Ranges[I] { return slices.Clone(r) }

func (r Ranges[I]) Ref() *Ranges[I] { return &r }

func (r *Ranges[I]) Assign() Assigner[I] { return Assigner[I]{r} }

func (r Ranges[I]) Iter() iter.Seq[I] {
	return func(yield func(I) bool) {
		for i := range len(r) {
			for v := r[i].start; v <= r[i].end; v++ {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func (r Ranges[I]) NumChunks() int { return len(r) }

func (r Ranges[I]) First() I { return r[0].start }
func (r Ranges[I]) Last() I  { return r[len(r)-1].end }

func (r Ranges[I]) firstChunk() *Range[I] { return &r[0] }
func (r Ranges[I]) lastChunk() *Range[I]  { return &r[len(r)-1] }

func (r Ranges[I]) Index(index uint64) (I, bool) {
	for i := range len(r) {
		if index < r[i].Len() {
			return r[i].start + I(index), true
		}
		index -= r[i].Len()
	}
	return 0, false
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
	limited := ranges.Limit(Range[I]{minimum, maximum})
	retval := make(Ranges[T], len(limited))
	for i, r := range limited {
		retval[i] = Range[T]{start: T(r.start), end: T(r.end)}
	}
	return retval, !limited.Equal(ranges)
}

// Caller must ensure slice ordered, otherwise result probably invalid
func FromSlice[I constraints.Integer](slice []I) Ranges[I] {
	if len(slice) == 0 {
		return Empty[I]().Ranges()
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
		ranges = ranges.Add(ranging.Ranges())
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
