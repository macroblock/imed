package mov

/*
import (
	"bytes"
	"fmt"
)

type Options int

const (
	OptEmpty     Options = 0
	OptMandatory Options = 1 << iota
)

type (
	Node struct {
		Type    string
		Options Options
		Fn      func(rd *StreamReader) (IAtomData, error)
		Nodes   []*Node
	}

	SchemeNode struct {
		Parent  *SchemeNode
		Type    string
		Options Options
		Fn      func(rd *StreamReader) (IAtomData, error)
		Nodes   map[string]*SchemeNode
	}
)

func fnUnimplemented(rd *StreamReader) (IAtomData, error) {
	panic("unimplemented")
}

func fnSkip(rd *StreamReader) (IAtomData, error) {
	panic("unimplemented")
}

func fnString(rd *StreamReader) (IAtomData, error) {
	size := rd.LimitRemainder()
	data := make([]byte, size)
	rd.ReadSlice(data)
	if rd.Err() != nil {
		return nil, rd.Err()
	}
	pos := bytes.IndexByte(data, 0)
	if pos < 0 {
		pos = len(data)
	}
	return StringData(data[:pos]), nil
}

type StringData string

func (o StringData) String() string {
	return fmt.Sprintf("%q", string(o))
}

func fnLangCodeString(rd *StreamReader) (IAtomData, error) {
	var (
		size     uint16
		langCode uint16
	)
	rd.ReadU16(&size)
	rd.ReadU16(&langCode)

	data := make([]byte, size)
	rd.ReadSlice(data)
	if rd.Err() != nil {
		return nil, rd.Err()
	}
	pos := bytes.IndexByte(data, 0)
	if pos < 0 {
		pos = len(data)
	}
	return LangCodeStringData{langCode: langCode, data: data[:pos]}, nil
}

type LangCodeStringData struct {
	langCode uint16
	data     []byte
}

func (o LangCodeStringData) String() string {
	lang := [3]byte{'#', '#', '#'}
	v := o.langCode
	lang[2] = byte((v & 0x1f) + 0x60)
	v >>= 5
	lang[1] = byte((v & 0x1f) + 0x60)
	v >>= 5
	lang[0] = byte((v & 0x1f) + 0x60)

	return fmt.Sprintf("%q:%q", string(lang[:]), string(o.data))
}

func fnMoovTrakMdiaHdlr(rd *StreamReader) (IAtomData, error) {
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

	ret.ComponentType = Uint32ToStr(typ)
	ret.ComponentSubtype = Uint32ToStr(subtype)

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
	// There were Pascal-strings in older QTFF.
	// So we must check if it is Pascal or C style string
	if int(data[0]) == l-1 {
		ret.ComponentName = string(data[1:l])
	} else {
		ret.ComponentName = string(data[:l])
	}
	return ret, nil
}

var (
	nodeWide = &Node{Type: "wide"}
	nodeFree = &Node{Type: "free"}
	nodeSkip = &Node{Type: "skip"}
	nodeUdta = &Node{Type: "udta",
		Nodes: []*Node{
			{Type: "name", Fn: fnString},
			{Type: "\xa9nam", Fn: fnLangCodeString},
			{Type: "\xa9swr", Fn: fnLangCodeString},
		},
	}
)

var Scheme *SchemeNode

var schemeNodes = []*Node{
	nodeWide, nodeFree, nodeSkip,
	nodeUdta,
	{
		Type: "moov", Options: OptMandatory,
		Nodes: []*Node{
			nodeUdta,
			{
				Type: "trak",
				Nodes: []*Node{
					nodeUdta,
					{
						Type: "mdia", Options: OptMandatory,
						Nodes: []*Node{
							nodeUdta,
							{Type: "hdlr", Fn: fnMoovTrakMdiaHdlr},
							{Type: "minf",
								Nodes: []*Node{
									{Type: "stbl"},
								},
							},
						},
					},
				},
			},
		},
	},
}

func GetScheme() map[string]*SchemeNode {
	if Scheme != nil {
		return Scheme.Nodes
	}
	node := &SchemeNode{}
	generate(node, nil, schemeNodes)
	Scheme = node

	return Scheme.Nodes
}

func generate(outNodes *SchemeNode, parent *SchemeNode, inNodes []*Node) {
	out := make(map[string]*SchemeNode, len(inNodes))
	for _, in := range inNodes {
		if _, ok := out[in.Type]; ok {
			panic(fmt.Sprintf("duplicated Atom Type %q", in.Type))
		}
		node := &SchemeNode{
			Parent:  parent,
			Type:    in.Type,
			Options: in.Options,
			Fn:      in.Fn,
			Nodes:   nil,
		}

		if in.Nodes != nil {
			generate(node, node, in.Nodes)
		}
		out[in.Type] = node
	}
	outNodes.Nodes = out
}

*/
