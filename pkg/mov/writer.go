package mov

import (
	"bufio"
	"encoding/binary"
	"io"
)

type StreamWriter struct {
	err   error
	wr    *bufio.Writer
	pos   int64
	order binary.ByteOrder
	limit SizeLimit
}

func NewStreamWriterBE(wr io.Writer) *StreamWriter {
	return NewStreamWriter(wr, binary.BigEndian)
}

func NewStreamWriterLE(wr io.Writer) *StreamWriter {
	return NewStreamWriter(wr, binary.LittleEndian)
}

func NewStreamWriter(wr io.Writer, byteOrder binary.ByteOrder) *StreamWriter {
	return &StreamWriter{
		err:   nil,
		wr:    bufio.NewWriter(wr),
		pos:   0,
		order: byteOrder,
		limit: InitSizeLimitCap(16),
	}
}

func (o *StreamWriter) Flush() {
	err := o.wr.Flush()
	if err != nil {
		panic(err.Error())
	}
}

func (o *StreamWriter) Pos() int64 {
	return o.pos
}

func (o *StreamWriter) LimitRemainder() int64 {
	return o.limit.Remainder(o.pos)
}

func (o *StreamWriter) PushLimit(limit int64) {
	if o.err != nil {
		return
	}
	o.err = o.limit.Push(o.pos, limit)
}

func (o *StreamWriter) PopLimit() {
	if o.err != nil {
		return
	}
	o.err = o.limit.Pop(o.pos)
}

func (o *StreamWriter) Err() error {
	return o.err
}

func (o *StreamWriter) WriteSlice(buf []byte) {
	if o.err != nil {
		return
	}
	n, err := o.wr.Write(buf)
	o.pos += int64(n)
	o.err = err
}

func (o *StreamWriter) WriteU8(v uint8) {
	buf := [1]byte{v}
	o.WriteSlice(buf[:])
}

func (o *StreamWriter) WriteU16(v uint16) {
	buf := [2]byte{}
	o.order.PutUint16(buf[:], v)
	o.WriteSlice(buf[:])
}

func (o *StreamWriter) WriteU32(v uint32) {
	buf := [4]byte{}
	o.order.PutUint32(buf[:], v)
	o.WriteSlice(buf[:])
}

func (o *StreamWriter) WriteU64(v uint64) {
	buf := [8]byte{}
	o.order.PutUint64(buf[:], v)
	o.WriteSlice(buf[:])
}

func (o *StreamWriter) Fill(size int64, val byte) {
	if o.err != nil {
		return
	}
	const bufSize = 512
	buf := [bufSize]byte{}
	if val != 0 {
		for i := range buf {
			buf[i] = val
		}
	}
	for size > bufSize {
		o.WriteSlice(buf[:bufSize])
		size -= bufSize
	}
	if size != 0 {
		o.WriteSlice(buf[:size])
	}
}

func (o *StreamWriter) FillLimitRemainder(val byte) {
	size := o.limit.Remainder(o.pos)
	if size < 0 {
		panic("int64 overflow")
	}
	o.Fill(size, val)
}
