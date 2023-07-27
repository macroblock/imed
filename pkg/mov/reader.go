package mov

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type StreamReader struct {
	err   error
	f     *os.File
	pos   int64
	order binary.ByteOrder
	limit SizeLimit
}

func NewStreamReaderBE(f *os.File) *StreamReader {
	return NewStreamReader(f, binary.BigEndian)
}

func NewStreamReaderLE(f *os.File) *StreamReader {
	return NewStreamReader(f, binary.LittleEndian)
}

func NewStreamReader(f *os.File, byteOrder binary.ByteOrder) *StreamReader {
	return &StreamReader{
		err:   nil,
		f:     f,
		pos:   0,
		order: byteOrder,
		limit: InitSizeLimitCap(16),
	}
}

func (o *StreamReader) Pos() int64 {
	return o.pos
}

func (o *StreamReader) CanRead() bool {
	if o.err != nil || o.limit.Remainder(o.pos) <= 0 {
		return false
	}
	pos, err := o.f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return false
	}
	info, err := o.f.Stat()
	if err != nil {
		return false
	}
	return pos < info.Size()
}

func (o *StreamReader) LimitRemainder() int64 {
	return o.limit.Remainder(o.pos)
}

func (o *StreamReader) PushLimit(limit int64) {
	if o.err != nil {
		return
	}
	o.err = o.limit.Push(o.pos, limit)
}

func (o *StreamReader) PopLimit() {
	if o.err != nil {
		return
	}
	o.err = o.limit.Pop(o.pos)
}

func (o *StreamReader) Err() error {
	return o.err
}

func (o *StreamReader) ReadSlice(buf []byte) {
	if o.err != nil {
		return
	}
	n, err := io.ReadAtLeast(o.f, buf, len(buf))
	o.pos += int64(n)
	o.err = err
}

func (o *StreamReader) u8() uint8 {
	buf := [1]byte{}
	o.ReadSlice(buf[:])
	return buf[0]
}

func (o *StreamReader) u16() uint16 {
	buf := [2]byte{}
	o.ReadSlice(buf[:])
	return o.order.Uint16(buf[:])
}

func (o *StreamReader) u32() uint32 {
	buf := [4]byte{}
	o.ReadSlice(buf[:])
	return o.order.Uint32(buf[:])
}

func (o *StreamReader) u64() uint64 {
	buf := [8]byte{}
	o.ReadSlice(buf[:])
	return o.order.Uint64(buf[:])
}

func (o *StreamReader) ReadU8(v *uint8)   { *v = o.u8() }
func (o *StreamReader) ReadU16(v *uint16) { *v = o.u16() }
func (o *StreamReader) ReadU32(v *uint32) { *v = o.u32() }
func (o *StreamReader) ReadU64(v *uint64) { *v = o.u64() }

func (o *StreamReader) ReadI8(v *int8)   { *v = int8(o.u8()) }
func (o *StreamReader) ReadI16(v *int16) { *v = int16(o.u16()) }
func (o *StreamReader) ReadI32(v *int32) { *v = int32(o.u32()) }
func (o *StreamReader) ReadI64(v *int64) { *v = int64(o.u64()) }

func (o *StreamReader) Skip(size int64) {
	if o.err != nil {
		return
	}
	if size < 0 {
		o.err = fmt.Errorf("negative limit reamainder (%v)", size)
		return
	}
	pos, err := o.f.Seek(size, os.SEEK_CUR)
	//n, err := io.CopyN(io.Discard, o.f, size)
	o.pos = pos
	o.err = err
}

func (o *StreamReader) SkipLimitRemainder() {
	o.Skip(o.limit.Remainder(o.pos))
}
