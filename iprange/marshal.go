package iprange

import (
	"encoding/binary"
	"errors"
)

func (r IPRanges) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *IPRanges) UnmarshalText(text []byte) error {
	*r = FromStr(string(text))
	if r.IsEmpty() && len(text) > 0 {
		return errors.New("malformed ip ranges")
	}
	return nil
}

func (r IPRanges) MarshalBinary() ([]byte, error) {
	if r.IsEmpty() {
		return nil, nil
	}
	sizeOf := 8 // size of uint64
	data := make([]byte, r.NumChunks()*4*sizeOf)
	for i, chunk := range r {
		binary.BigEndian.PutUint64(data[i*4*sizeOf:], chunk.start[0])
		binary.BigEndian.PutUint64(data[(i*4+1)*sizeOf:], chunk.start[1])
		binary.BigEndian.PutUint64(data[(i*4+2)*sizeOf:], chunk.end[0])
		binary.BigEndian.PutUint64(data[(i*4+3)*sizeOf:], chunk.end[1])
	}
	return data, nil
}

func (r *IPRanges) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	sizeOf := 8 // size of uint64
	if len(data)%(4*sizeOf) != 0 {
		return errors.New("not IPv6 ranges")
	}
	length := len(data) / (4 * sizeOf)
	ranges := make(IPRanges, length)
	for i := range length {
		chunk := &ranges[i]
		chunk.start[0] = binary.BigEndian.Uint64(data[i*4*sizeOf:])
		chunk.start[1] = binary.BigEndian.Uint64(data[(i*4+1)*sizeOf:])
		chunk.end[0] = binary.BigEndian.Uint64(data[(i*4+2)*sizeOf:])
		chunk.end[1] = binary.BigEndian.Uint64(data[(i*4+3)*sizeOf:])
		if chunk.start.GreaterOrEqualThan(chunk.end) {
			return errors.New("not a valid IPv6 ranges")
		}
	}
	for i := 1; i < length; i++ {
		if ranges[i-1].start.GreaterOrEqualThan(ranges[i].end) {
			return errors.New("not a valid IPv6 ranges")
		}
	}
	*r = ranges
	return nil
}
