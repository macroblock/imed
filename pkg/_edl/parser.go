package edl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

var rules = `
entry = header {,@clip} ,$;

	header = 'TITLE: ' @title ,'FCM: NON-DROP FRAME';
		title	= {!eol any};

	clip = @number ,@type ,@media@track ,@effect ,@origin ,@timeline [,@meta];
		type	= ident;
		media	= letter{letter};
		track	= [ident];
		effect	= ident;
		origin	= span;

		timeline= span;
			span	= @in, @out;
				in	= @timecode;
				out	= @timecode;
					timecode= dd ':' dd ':' dd ':' dd;
						dd	= digit digit;

		meta	= '* FROM CLIP NAME: ' @m_name;
			m_name	= {!eol any};

,	= (\x00..' '){\x00..' '};
digit	= '0'..'9';
letter	= 'a'..'z'|'A'..'Z';
eol	= \x0d|\x0a;
any	= \x00..\xff;
symbol  = letter|digit;
ident	= symbol{symbol};
number	= digit{digit};
`

var globParser *ptool.TParser
var globFps = 25

type Tree struct {
	Root *ptool.TNode
	ByID func(int) string
}

func (o *Tree) String() string {
	if o == nil {
		return fmt.Sprintf("%v", nil)
	}
	return ptool.TreeToString(o.Root, o.ByID)
}

func getParser() (*ptool.TParser, error) {
	if globParser != nil {
		return globParser, nil
	}
	p, err := ptool.NewBuilder().FromString(rules).Entries("entry").Build()
	if err != nil {
		fmt.Println("\n[old form] parser error: ", err)
		panic("")
	}
	globParser = p
	return p, nil
}

func parse(s string) (*ptool.TNode, error) {
	parser, err := getParser()
	if err != nil {
		return nil, err
	}
	tree, err := parser.Parse(s)
	if err != nil {
		return nil, err
	}
	return tree, err
}

func Parse(s string) (*Edl, *Tree, error) {
	parser, err := getParser()
	if err != nil {
		return nil, nil, err
	}
	tree, err := parser.Parse(s)
	t := &Tree{tree, parser.ByID}
	if err != nil {
		return nil, t, err
	}
	edl, err := newEdl(tree, parser.ByID)
	if err != nil {
		return nil, t, err
	}
	return edl, t, nil
}

func newEdl(tree *ptool.TNode, byID func(int) string) (*Edl, error) {
	ret := &Edl{}
	for _, node := range tree.Links {
		val := node.Value
		typ := byID(node.Type)
		switch typ {
		default:
			return nil, fmt.Errorf("(Edl) unexpected type %q", typ)
		case "title":
			ret.Title = val
		case "clip":
			clip, convErr, err := newClip(node, byID)
			if err != nil {
				return nil, err
			}
			ret.Clips = append(ret.Clips, clip)
			ret.convErr = ret.convErr || convErr
		}
	}
	return ret, nil
}

func newClip(tree *ptool.TNode, byID func(int) string) (*Clip, bool, error) {
	convErr := false
	ret := &Clip{}
	for _, node := range tree.Links {
		val := node.Value
		typ := byID(node.Type)
		err := error(nil)
		switch typ {
		default:
			return nil, false, fmt.Errorf("(Clip) unexpected type %q", typ)
		case "number":
			ret.Number, err = strconv.Atoi(val)
		case "type":
			ret.Type = val
		case "media":
			ret.Media = val
		case "track":
			if val == "" {
				val = "1"
			}
			ret.Track, err = strconv.Atoi(val)
		case "effect":
			ret.Effect = val
		case "origin":
			ret.Origin, convErr, err = newTimespan(node, byID)
		case "timeline":
			ret.Timespan, convErr, err = newTimespan(node, byID)
		case "meta":
			ret.Meta, err = newMeta(node, byID)
		}
		if err != nil {
			return nil, convErr, err
		}
	}
	return ret, convErr, nil
}

func newMeta(tree *ptool.TNode, byID func(int) string) (*Meta, error) {
	ret := &Meta{}
	for _, node := range tree.Links {
		val := node.Value
		typ := byID(node.Type)
		err := error(nil)
		switch typ {
		default:
			return nil, fmt.Errorf("(Meta) unexpected type %q", typ)
		case "m_name":
			ret.Name = val;
		}
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func newTimespan(tree *ptool.TNode, byID func(int) string) (Timespan, bool, error) {
	convErr := false
	ret := Timespan{}
	nul := Timespan{}
	for _, node := range tree.Links {
		typ := byID(node.Type)
		err := error(nil)
		switch typ {
		default:
			return nul, false, fmt.Errorf("(Timespan) unexpected type %q", typ)
		case "in":
			ret.In, convErr, err = newTimecode(node, byID)
		case "out":
			ret.Out, convErr, err = newTimecode(node, byID)
		}
		if err != nil {
			return nul, convErr, err
		}
	}
	return ret, convErr, nil
}

func newTimecode(tree *ptool.TNode, byID func(int) string) (Timecode, bool, error) {
	var ret [4]int
	for _, node := range tree.Links {
		val := node.Value
		typ := byID(node.Type)
		err := error(nil)
		switch typ {
		default:
			return NewTimecode(), false, fmt.Errorf("(Timecode) unexpected type %q", typ)
		case "timecode":
			for i, v := range strings.Split(val, ":") {
				ret[i], err = strconv.Atoi(v)
			}
		}
		if err != nil {
			return NewTimecode(), false, err
		}
	}
	tc, rest := NewTimecodeFromHHMMSSFr(ret[0], ret[1], ret[2], ret[3])
	convErr := rest != 0
	return tc, convErr, nil
}

