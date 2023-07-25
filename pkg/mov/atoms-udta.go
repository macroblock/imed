package mov

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

//// udta ////////////////////////////////////////////////////////////////////

type (
	Udta struct {
		//List []UdtaItem
		Data map[Fcc]*UdtaItem
	}

	UdtaItem struct {
		//Type Fcc
		//Strings []LangString
		Map map[LangCode]string
	}

	LangCode uint16
)

var _ IAtomData = (*Udta)(nil)

func NewUdta() *Udta {
	return &Udta{Data: map[Fcc]*UdtaItem{}}
}

func (o UdtaItem) String(typ Fcc) string {
	ret := ""
	t := typ.String()
	justString := t[0] != AsciiCopyright
	if justString {
		ret += t
	} else {
		ret += "(c)" + t[1:]
	}

	if len(o.Map) == 0 {
		return ret + ": <nil>"
	}

	list := make([]string, 0, len(o.Map))

	if justString {
		list = append(list, fmt.Sprintf("%q", o.Map[0]))
	} else {
		keys := make([]LangCode, 0, len(o.Map))
		for k := range o.Map {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, k := range keys {
			v := o.Map[k]
			list = append(list,
				fmt.Sprintf("%v, %q", QuoteStr(k.String()), QuoteStr(v)))
		}
	}

	return ret + ": " + strings.Join(list, "\n        ")
}

func (o LangCode) String() string {
	return LangCodeToStr(o)
}

func (o Udta) String() string {
	keys := make([]Fcc, 0, len(o.Data))
	for k := range o.Data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	ret := make([]string, 0, len(keys))
	for _, k := range keys {
		v := o.Data[k]
		ret = append(ret, v.String(k))
	}
	return strings.Join(ret, "\n")
}

func (o Udta) Size() int64 {
	size := int64(0)
	for k, v := range o.Data {
		size += v.Size(k)
	}
	return size
}

func (o UdtaItem) Size(typ Fcc) int64 {
	size := int64(8) // size32 + type32
	if typ.IsSimple() {
		size += int64(len(o.Map[0]))
	} else {
		for k, v := range o.Map {
			size += sizeOfLangString(k, v)
		}
	}
	if size > int64(MaxInt32) {
		size += 8
	}
	return size
}

func sizeOfLangString(lang LangCode, str string) int64 {
	size := int64(4) // size16 + langCode16
	size += int64(len(str))
	// TODO: check twice size-4
	if size-4 > int64(MaxUint16) {
		panic("size field (16bit) overflow in LangString (" + lang.String() + ")")
	}
	return size
}

func (o UdtaItem) Write(wr *StreamWriter, typ Fcc) error {
	wr.WriteU32(uint32(o.Size(typ)))
	wr.WriteU32(uint32(typ))
	if typ.IsSimple() {
		wr.WriteSlice([]byte(o.Map[0]))
	} else {
		keys := make([]LangCode, 0, len(o.Map))
		for k := range o.Map {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, k := range keys {
			v := o.Map[k]
			wr.WriteU16(uint16(sizeOfLangString(k, v) - 4)) // TODO: -4
			wr.WriteU16(uint16(k))
			wr.WriteSlice([]byte(v))
		}
	}
	return wr.Err()
}

func (o Udta) Write(wr *StreamWriter) error {
	keys := make([]Fcc, 0, len(o.Data))
	for k := range o.Data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		v := o.Data[k]
		_ = v.Write(wr, k)
	}
	return wr.Err()
}

func ReadAnyUdta(rd *StreamReader) (*Udta, error) {
	ret := map[Fcc]*UdtaItem{}
	err := error(nil)

	_, err = WalkStream(rd, "",
		func(rd *StreamReader, path string, header *AtomHeader, atom *Atom, fn WalkStreamFn) error {
			m := map[LangCode]string{}

			if header.Type.IsSimple() {
				m[0] = ReadString(rd)
			} else {
				m = ReadLangStrings(rd)
			}

			ret[header.Type] = &UdtaItem{Map: m}

			return SkipAtom
		})

	return &Udta{Data: ret}, err
}

func ReadString(rd *StreamReader) string {
	size := rd.LimitRemainder()
	data := make([]byte, size)

	rd.ReadSlice(data)

	if rd.Err() != nil {
		return ""
	}
	pos := bytes.IndexByte(data, 0)
	if pos < 0 {
		pos = len(data)
	}
	return string(data[:pos])
}

func ReadLangStrings(rd *StreamReader) map[LangCode]string {
	// list of small(16bit) data atoms:
	//   2 bytes - size (including size langcode and data)
	//   2 bytes - langcode
	//   size-4 (??? this shouldn't be right) - text
	ret := map[LangCode]string{}

	for rd.CanRead() {
		var (
			size     uint16
			langCode uint16
		)

		rd.ReadU16(&size)
		rd.ReadU16(&langCode)
		//size -= 4

		data := make([]byte, size)
		rd.ReadSlice(data)

		// trim trailing zero if needed
		pos := bytes.IndexByte(data, 0)
		if pos < 0 {
			pos = len(data)
		}

		ret[LangCode(langCode)] = string(data[:pos])
	}

	return ret
}

func (o *Udta) Set(typ Fcc, lc LangCode, s string) {
	item, ok := o.Data[typ]
	if !ok {
		item = &UdtaItem{Map: map[LangCode]string{}}
		o.Data[typ] = item
	}
	if typ.IsSimple() {
		lc = 0
	}
	item.Map[lc] = s
}

func (o *Udta) Remove(typ Fcc, lc LangCode) {
	switch typ.IsSimple() {
	case true:
		delete(o.Data, typ)
	case false:
		item := o.Data[typ]
		if item != nil {
			delete(item.Map, lc)
			if len(item.Map) == 0 {
				delete(o.Data, typ)
			}
		}
	}
}

func (o *Udta) Merge(src *Udta) {
	for typ, item := range src.Data {
		switch typ.IsSimple() {
		case true:
			o.Set(typ, 0, item.Map[0])
		case false:
			for lc, s := range item.Map {
				o.Set(typ, lc, s)
			}
		}
	}
}
