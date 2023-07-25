package mov

import (
	"bytes"
	"fmt"
)

//// unknown /////////////////////////////////////////////////////////////////

type (
	Unknown []byte
)

var _ IAtomData = (*Unknown)(nil)

func (o Unknown) String() string {
	return fmt.Sprintf("<Size:%v>", len(o))
}

func (o Unknown) Size() int64 {
	return int64(len(o))
}

func (o Unknown) Write(wr *StreamWriter) error {
	wr.WriteSlice(o)
	return wr.Err()
}

func ReadUnknown(rd *StreamReader) (Unknown, error) {
	size := rd.LimitRemainder()
	data := make([]byte, size)
	rd.ReadSlice(data)
	if rd.Err() != nil {
		return nil, rd.Err()
	}
	return Unknown(data), nil
}

//// free ////////////////////////////////////////////////////////////////////

type AnyFree struct {
	size int64
}

var _ IAtomData = (*AnyFree)(nil)

func NewAnyFree(size int64) *AnyFree {
	size -= 8 // header size
	if size > MaxInt64 {
		size -= 8 // extended header size
	}

	if size < 0 {
		panic("<free> atom cannot be less then atom header size")
	}
	return &AnyFree{size}
}

func (o *AnyFree) String() string {
	return fmt.Sprintf("free space: %v", o.size)
}

func (o *AnyFree) Size() int64 {
	return o.size
}

func (o *AnyFree) Write(wr *StreamWriter) error {
	wr.Fill(o.size, 0)
	return wr.Err()
}

//// moov/trak/mdia/hdrl /////////////////////////////////////////////////////

type MoovTrakMdiaHdlr struct {
	Version               byte
	Flags                 uint16 // actually 3 bytes
	ComponentType         Fcc
	ComponentSubtype      Fcc
	ComponentManufacturer uint32 // reserved
	ComponentFlags        uint32 // reserved
	ComponentFlagsMask    uint32 // reserved
	ComponentName         string
}

var _ IAtomData = (*MoovTrakMdiaHdlr)(nil)

func (o *MoovTrakMdiaHdlr) String() string {
	return fmt.Sprintf("Type:    %q\nSubtype: %q\nName:    %q",
		o.ComponentType, o.ComponentSubtype, o.ComponentName)
}

func (o *MoovTrakMdiaHdlr) Size() int64 {
	return int64(4 + 4 + 4 + 12 + len(o.ComponentName)) // TODO: +2 or +1
}

func (o *MoovTrakMdiaHdlr) Write(wr *StreamWriter) error {
	wr.WriteU8(o.Version)
	wr.Fill(3, 0)
	wr.WriteU32(uint32(o.ComponentType))
	wr.WriteU32(uint32(o.ComponentSubtype))
	wr.Fill(12, 0)
	wr.WriteSlice([]byte(o.ComponentName))
	return wr.Err()
}

func ReadMoovTrakMdiaHdlr(rd *StreamReader) (*MoovTrakMdiaHdlr, error) {
	var (
		typ, subtype uint32
	)
	ret := &MoovTrakMdiaHdlr{}
	rd.ReadU8(&ret.Version)
	rd.Skip(3)
	rd.ReadU32(&typ)
	rd.ReadU32(&subtype)
	rd.Skip(12)
	if rd.Err() != nil {
		return nil, rd.Err()
	}

	ret.ComponentType = Fcc(typ)
	ret.ComponentSubtype = Fcc(subtype)

	size := rd.LimitRemainder()
	if size == 0 {
		return ret, nil
	}

	data := make([]byte, size)
	rd.ReadSlice(data)
	if rd.Err() != nil {
		return nil, rd.Err()
	}

	l := bytes.IndexByte(data, 0)
	if l < 0 {
		l = len(data)
	}
	// Actually this should be a Pascal (counted) string
	// as it written in QTFF doc
	// but in the real world it is possible to get a C-string instead
	// so we check both
	if int(data[0]) == l-1 {
		ret.ComponentName = string(data[1:l])
	} else {
		ret.ComponentName = string(data[:l])
	}
	return ret, nil
}
