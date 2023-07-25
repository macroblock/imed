package mov

import (
	//"encoding/binary"
	//"os"
	"errors"
	"fmt"
)

var (
	//Exit     = errors.New("terminate")
	SkipAtom = errors.New("skip this atom")
	//ParseContainer = errors.New("parse this atom as container")
)

//// AtomHeader32 ////////////////////////////////////////////////////////////

type AtomHeader struct {
	Pos        int64
	TotalSize  int64
	HeaderSize int64
	Type       Fcc
}

func (o AtomHeader) NewAtom() *Atom {
	return NewAtom(o.Type)
}

func (o AtomHeader) String() string {
	return fmt.Sprintf(
		"<AtomHeader(%v, pos 0x%08x, size 0x%016x)>",
		o.Type, o.Pos, o.TotalSize,
	)
}

func (o AtomHeader) DataSize() int64 {
	return o.TotalSize - o.HeaderSize
}

func ReadAtomHeader(rd *StreamReader, header *AtomHeader) {
	var (
		sz  uint32
		typ uint32
	)

	header.Pos = rd.Pos()

	rd.ReadU32(&sz)
	rd.ReadU32(&typ)

	header.TotalSize = int64(sz)
	header.Type = Fcc(typ)

	if sz == 1 {

		rd.ReadI64(&header.TotalSize)

		if header.TotalSize < 0 {
			panic("atom size overflow (int64)")
		}
	}

	header.HeaderSize = rd.Pos() - header.Pos

	/*
		// for mp4 only
		if typ == boxUUID {
			flags |= FlagUUID
			data := [16]byte{}
			rd.ReadSlice(data[:16])
			typ = string(data[:16])
		}
	*/
}

//// AtomHeader16 ////////////////////////////////////////////////////////////

type AtomHeader16 struct {
	Pos       int64
	TotalSize int
	DataSize  int
	Type      uint16
}

func ReadAtomHeader16(rd *StreamReader, header *AtomHeader16) {
	var size int16
	header.Pos = rd.Pos()

	rd.ReadI16(&size)
	rd.ReadU16(&header.Type)

	header.TotalSize = int(size)
	header.DataSize = header.TotalSize - int(rd.Pos()-header.Pos)
}

//// Walkers /////////////////////////////////////////////////////////////////

type WalkStreamFn func(rd *StreamReader, path string, info *AtomHeader, atom *Atom, fn WalkStreamFn) error

func WalkStream(rd *StreamReader, path string, fn WalkStreamFn) ([]*Atom, error) {
	ret := []*Atom(nil)
	err := error(nil)

	for err == nil && rd.CanRead() {
		header := AtomHeader{}

		ReadAtomHeader(rd, &header)

		rd.PushLimit(header.DataSize())

		//fmt.Printf("++ %v\n", header)

		p := path + "/" + header.Type.String()

		atom := header.NewAtom()

		if fn != nil {
			err = fn(rd, p, &header, atom, fn)
		}

		rd.SkipLimitRemainder()
		rd.PopLimit()

		skip := false
		if err == SkipAtom {
			skip = true
			err = nil
		}

		if !skip {
			ret = append(ret, atom)
		}
	} // for err == nil && rd.CanRead()

	if err == nil {
		err = rd.Err()
	}

	return ret, err
}

/*
type readContainer16Fn func(rd *StreamReader, header *AtomHeader16) error

func readContainer16(rd *StreamReader, fn readContainer16Fn) error {
	err := error(nil)

	for err == nil && rd.CanRead() {
		header := AtomHeader16{}

		ReadAtomHeader16(rd, &header)

		rd.PushLimit(int64(header.DataSize))

		err = fn(rd, &header)

		rd.SkipLimitRemainder()
		rd.PopLimit()
	}

	if err == nil {
		err = rd.Err()
	}

	return err
}
*/
