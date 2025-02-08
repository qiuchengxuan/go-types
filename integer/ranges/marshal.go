package ranges

import (
	"encoding/binary"
	"errors"
	"unsafe"
)

func (r Ranges[I]) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

func (r *Ranges[I]) UnmarshalText(text []byte) error {
	*r = FromStr[I](string(text))
	if *r == nil {
		return errors.New("Malformed ranges")
	}
	return nil
}

func (r Ranges[I]) MarshalBinary() ([]byte, error) {
	if r.IsEmpty() {
		return nil, nil
	}
	sizeOf := int(unsafe.Sizeof(r[0].start))
	data := make([]I, r.NumChunks()*2)
	var buf [8]byte
	for i, chunk := range r {
		binary.BigEndian.PutUint64(buf[:], uint64(chunk.start))
		data[i*2] = *(*I)(unsafe.Pointer(&buf[8-sizeOf]))
		binary.BigEndian.PutUint64(buf[:], uint64(chunk.end))
		data[i*2+1] = *(*I)(unsafe.Pointer(&buf[8-sizeOf]))
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), len(data)*sizeOf), nil
}

func (r *Ranges[I]) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var tmp I
	sizeOf := int(unsafe.Sizeof(tmp))
	if len(data)%(sizeOf*2) != 0 {
		return errors.New("Not ranges data")
	}
	aligned := data
	length := len(aligned) / (sizeOf * 2)
	ranges := make(Ranges[I], length)
	if uintptr(unsafe.Pointer(&data[0]))%uintptr(sizeOf) != 0 {
		aligned := unsafe.Slice((*byte)(unsafe.Pointer(&ranges[0])), len(data))
		copy(aligned, data)
	}
	slice := unsafe.Slice((*I)(unsafe.Pointer(&aligned[0])), len(aligned)/sizeOf)
	var buf [8]byte
	for i := 0; i < length; i++ {
		*(*I)(unsafe.Pointer(&buf[8-sizeOf])) = slice[i*2]
		start := I(binary.BigEndian.Uint64(buf[:]))
		*(*I)(unsafe.Pointer(&buf[8-sizeOf])) = slice[i*2+1]
		end := I(binary.BigEndian.Uint64(buf[:]))
		if start > end {
			return errors.New("Not a valid ranges data")
		}
		ranges[i] = Range[I]{start, end}
	}
	for i := 1; i < length; i++ {
		if ranges[i-1].end >= ranges[i].start {
			return errors.New("Not a valid ranges data")
		}
	}
	*r = ranges
	return nil
}
