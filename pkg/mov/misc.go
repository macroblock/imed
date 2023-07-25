package mov

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const (
	MaxUint64 uint64 = ^uint64(0)
	MaxInt64  int64  = int64(MaxUint64 >> 1)
	MaxUint32 uint32 = ^uint32(0)
	MaxInt32  int32  = int32(MaxUint32 >> 1)
	MaxUint16 uint16 = ^uint16(0)
	MaxInt16  int16  = int16(MaxUint16 >> 1)
)

const AsciiCopyright = '\xa9' // '©'

func Uint32ToStr(v uint32) string {
	data := [4]byte{}
	binary.BigEndian.PutUint32(data[:], v)
	return string(data[:])
}

type Fcc uint32

func (o Fcc) String() string {
	buf := [4]byte{}
	buf[3] = byte(o)
	o = o >> 8
	buf[2] = byte(o)
	o = o >> 8
	buf[1] = byte(o)
	o = o >> 8
	buf[0] = byte(o)
	return string(buf[:])
}

func (o Fcc) IsSimple() bool {
	return o>>24 != AsciiCopyright
}

func StrToFccOrPanic(str string) Fcc {
	fcc, err := StrToFcc(str)
	if err != nil {
		panic(fcc)
	}
	return fcc
}

func StrToFcc(str string) (Fcc, error) {
	s := str
	if strings.HasPrefix(s, "(c)") {
		b := []byte(s[2:])
		b[0] = AsciiCopyright
		s = string(b)
	}
	if len(s) != 4 {
		return 0, fmt.Errorf("size of FourCC string %q, must be exactly 4 bytes or 6 bytes if begin with '(c)'", str)
	}
	_ = s[3]
	ret := uint32(s[0])
	ret <<= 8
	ret += uint32(s[1])
	ret <<= 8
	ret += uint32(s[2])
	ret <<= 8
	ret += uint32(s[3])
	return Fcc(ret), nil
}

func QuoteStr(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

func StrToLangCode(s string) (LangCode, error) {
	if len(s) != 3 {
		fmt.Errorf("LangCode string must be 3 bytes long to convert")
	}
	lc := uint16(0)
	lc += uint16(s[0] - 0x60)
	lc <<= 5
	lc += uint16(s[1] - 0x60)
	lc <<= 5
	lc += uint16(s[2] - 0x60)
	return LangCode(lc), nil
}

func LangCodeToStr(lc LangCode) string {
	lang := [3]byte{}
	v := lc
	if v < 0x400 {
		//TODO: Macintosh language code
	}
	lang[2] = byte((v & 0x1f) + 0x60)
	v >>= 5
	lang[1] = byte((v & 0x1f) + 0x60)
	v >>= 5
	lang[0] = byte((v & 0x1f) + 0x60)

	return string(lang[:])
}

type SizeLimit struct {
	limits []int64
}

func InitSizeLimitCap(size int) SizeLimit {
	ret := SizeLimit{
		limits: make([]int64, 0, size+1),
	}
	ret.limits = append(ret.limits, MaxInt64)
	return ret
}

func (o *SizeLimit) Remainder(pos int64) int64 {
	last := len(o.limits) - 1
	return o.limits[last] - pos
}

func (o *SizeLimit) Push(pos, limit int64) error {
	last := len(o.limits) - 1
	if limit > o.limits[last]-pos {
		return fmt.Errorf("attempt to set incorrect limit value: (pos: %x, limit: %x, limit to set: %x)",
			pos, o.limits[last]-pos, limit)
	}
	o.limits = append(o.limits, limit+pos)
	return nil
}

func (o *SizeLimit) Pop(pos int64) error {
	last := len(o.limits) - 1
	if last <= 0 {
		return fmt.Errorf("unpaired call of pop limit without push limit")
	}
	//fmt.Printf("pos: %x, limit: %x\n", o.pos, o.limits[last])
	curLimit := o.limits[last]
	if curLimit < pos {
		return fmt.Errorf("limit has been exceeded (pos: %x, limit: %x)",
			pos, curLimit)
	}
	if curLimit > pos {
		return fmt.Errorf("limit has not been reached (pos: %x, limit: %x)",
			pos, curLimit)
	}
	o.limits = o.limits[:last]
	return nil
}

type FilterAtomFn func(path string, atom *Atom) bool

func ShallowCopy(atoms []*Atom, path string, filter FilterAtomFn) []*Atom {
	ret := []*Atom(nil)
	for i := range atoms {
		p := path + "/" + atoms[i].Type().String()
		a := *atoms[i]
		a.atoms = ShallowCopy(a.atoms, p, filter)
		if len(a.atoms) > 0 || filter(p, &a) {
			ret = append(ret, &a)
		}
	}
	return ret
}

type WalkAtomsFn func(path string, atom *Atom)

func WalkAtomsBT(atoms []*Atom, path string, fn WalkAtomsFn) []*Atom {
	ret := []*Atom(nil)
	for _, a := range atoms {
		p := path + "/" + a.Type().String()
		_ = WalkAtomsBT(a.atoms, p, fn)
		fn(p, a)
	}
	return ret
}
