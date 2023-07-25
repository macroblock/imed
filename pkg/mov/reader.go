package mov

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type StreamReader struct {
	err   error
	rd    *bufio.Reader
	pos   int64
	order binary.ByteOrder
	limit SizeLimit
}

func NewStreamReaderBE(rd io.Reader) *StreamReader {
	return NewStreamReader(rd, binary.BigEndian)
}

func NewStreamReaderLE(rd io.Reader) *StreamReader {
	return NewStreamReader(rd, binary.LittleEndian)
}

func NewStreamReader(rd io.Reader, byteOrder binary.ByteOrder) *StreamReader {
	return &StreamReader{
		err:   nil,
		rd:    bufio.NewReader(rd),
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
	_, err := o.rd.Peek(1)
	return err == nil
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
	n, err := io.ReadAtLeast(o.rd, buf, len(buf))
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
	n, err := io.CopyN(io.Discard, o.rd, size)
	o.pos += n
	o.err = err
}

func (o *StreamReader) SkipLimitRemainder() {
	if o.err != nil {
		return
	}
	size := o.limit.Remainder(o.pos)
	if size < 0 {
		o.err = fmt.Errorf("negative limit reamainder (%v)", size)
		return
	}
	n, err := io.CopyN(io.Discard, o.rd, size)
	o.pos += n
	o.err = err
}
