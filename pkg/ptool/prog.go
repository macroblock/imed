package ptool

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// TParser -
type TParser struct {
	src  *bytes.Reader
	cpos TPos
	// crune rune
	code    []TInstruction
	items   []tProgItem
	entries []int
}

// TInstruction -
type TInstruction struct {
	opcode TOpCode
	data   interface{}
}
type (
	// TOpCode -
	TOpCode int
	// TOffset -
	TOffset int
)

type tProgItem struct {
	name string
	ip   TOffset
}

// TPos -
type TPos struct {
	offs      TOffset
	line, col int
	r         rune
	w         int
}

// String -
func (o TPos) String() string {
	return fmt.Sprintf("[0x%04x] (%v,%v)", o.offs, o.line+1, o.col)
}

// RuneEOF -
const RuneEOF = rune(0x7fffffff) //rune(^0) //'\U0010ffff'

//go:generate stringer -type=TOpCode
// -
const (
	opERROR TOpCode = iota
	opNOP
	opLABEL
	opEND
	opJMP
	opJZ
	opJNZ
	opCALL
	opRET
	opTRUE
	opFALSE
	opMARK
	opRESTORE
	opRELEASE
	opREPEAT
	opACCEPT
	opPUSHNODE
	opPOPNODE
	opCHECKRUNE
	opCHECKRANGE
	opCHECKSTR
	opSETERROR
	opMAXINSTRUCTION
)

var stepCounter = 0

// newParser -
func newParser() *TParser {
	return &TParser{}
}

// Reset -
func (o *TParser) Reset(src []byte) {
	o.src = bytes.NewReader(src)
}

func (o *TParser) readRune() error {
	r, w, err := o.src.ReadRune()
	if err != nil {
		o.cpos.r = RuneEOF
		o.cpos.w = 0
		return err
	}
	o.cpos.r = r
	o.cpos.w = w
	o.cpos.offs += TOffset(w)
	switch r {
	default:
		o.cpos.col += w
	case '\x0a': // newline
		o.cpos.line++
		o.cpos.col = 0
	case '\x0d': // linefeed
		o.cpos.col = 0
	}
	return nil
}

func (o *TParser) readStringFrom(pos *TPos) (string, error) {
	// fmt.Printf("fromto : %v-%v\n", pos.offs, o.cpos.offs)
	from := pos.offs - TOffset(pos.w)
	to := o.cpos.offs - TOffset(o.cpos.w)
	if to == from {
		return "", nil
	}
	buf := make([]byte, to-from)
	_, err := o.src.Seek(int64(from), io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	_, err = o.src.Read(buf)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	// o.ReadRune()

	_, _ = o.src.Seek(int64(o.cpos.offs), io.SeekStart)
	return string(buf), nil
}

func (o *TParser) restorePos() error {
	_, err := o.src.Seek(int64(o.cpos.offs), io.SeekStart)
	return err
}

// Parse -
func (o *TParser) Parse(src string, entry ...string) (*TNode, error) {
	id := -1
	l := len(entry)
	switch l {
	default:
		return nil, fmt.Errorf("Too many entries %v", l)
	case 0:
		id = o.entries[0]
	case 1:
		id = o.EntryByName(entry[0])
		if id < 1 {
			return nil, fmt.Errorf("entry not found %q", entry[0])
		}
	}

	o.src = bytes.NewReader([]byte(src))
	stepCounter = 0
	type tfs struct {
		pos TPos
		r   rune
	}
	expected := false
	errpos := TPos{}
	// fmt.Println("start")
	ps := []TOffset{}
	fs := []TPos{}
	ls := []int{}
	ns := []*TNode{}
	ip := TOffset(0)
	res := false
	i64, err := o.src.Seek(0, io.SeekCurrent)
	o.cpos.col = 0
	o.cpos.line = 1
	o.cpos.offs = TOffset(i64)
	if err != nil {
		return nil, err
	}
	cnode := &TNode{Type: -1}
	tree := cnode
	o.readRune()
	// fmt.Println("loop")
	for {
		instr := o.code[ip]
		switch instr.opcode {
		default:
			o.log("illegal", ip, "", "")
			return tree, fmtError("illegal instruction")
		case opNOP:
			o.log("nop", ip, "", "")
		case opEND:
			o.log("end", ip, instr.data, "")
			// fmt.Printf("ps: %v fs: %v ls: %v ns %v\n", len(ps), len(fs), len(ls), len(ns))
			// res = instr.data.(bool)
			if !res {
				prefix := "expected"
				if expected {
					prefix = "unexpected"
				}
				str := "end of file"
				if errpos.r != RuneEOF {
					str = fmt.Sprintf("%q", errpos.r)
				}
				return tree, fmt.Errorf("[%06x] (l:%v, c:%v): %v %v", errpos.offs, errpos.line, errpos.col, prefix, str)
			}
			return tree, nil
		case opJMP:
			o.log("jmp", ip, instr.data, "")
			ip = instr.data.(TOffset)
			continue
		case opJZ:
			if !res {
				o.log("jz", ip, instr.data, "")
				ip = instr.data.(TOffset)
				ip--
			}
		case opJNZ:
			if res {
				o.log("jnz", ip, instr.data, "")
				ip = instr.data.(TOffset)
				ip--
			}
		case opCALL:
			o.log("call", ip, instr.data, "")
			ps = append(ps, ip)
			ip = instr.data.(TOffset)
			ip--
		case opRET:
			o.log("ret", ip, instr.data, "")
			res = instr.data.(bool)
			ip = ps[len(ps)-1]
			ps = ps[:len(ps)-1]
		case opTRUE:
			o.log("true", ip, "", "")
			res = true
		case opFALSE:
			o.log("false", ip, "", "")
			res = false
		case opSETERROR:
			o.log("seterror", ip, instr.data, "")
			expected = instr.data.(bool)
			if o.cpos.offs >= errpos.offs {
				errpos = o.cpos
			}
		case opMARK:
			o.log("mark", ip, "", "")
			fs = append(fs, o.cpos)
			ls = append(ls, len(cnode.Links))
		case opRESTORE:
			o.log("restore", ip, "", "")
			l := ls[len(ls)-1]
			// fmt.Println("length: ", l, " fslen: ", len(fs))
			cnode.Links = cnode.Links[:l]
			o.cpos = fs[len(fs)-1]
			ls = ls[:len(ls)-1]
			fs = fs[:len(fs)-1]
			err := o.restorePos()
			if err != nil {
				return tree, fmtError(err)
			}
		case opRELEASE:
			o.log("release", ip, "", "")
			ls = ls[:len(ls)-1]
			fs = fs[:len(fs)-1]
		case opREPEAT:
			o.log("repeat", ip, "", "")
			l := ls[len(ls)-1]
			cnode.Links = cnode.Links[:l]
			o.cpos = fs[len(fs)-1]
			err := o.restorePos()
			if err != nil {
				return tree, fmtError(err)
			}
		case opPUSHNODE:
			o.log("pushnode", ip, instr.data, "")
			// ls = ls[:len(ls)-1]
			ns = append(ns, cnode)
			cnode = &TNode{Type: instr.data.(int)}
		case opPOPNODE:
			o.log("popnode", ip, "", "")
			cnode = ns[len(ns)-1]
			ns = ns[:len(ns)-1]
		case opACCEPT:
			o.log("accept", ip, "", "")
			ls = ls[:len(ls)-1]
			pos := fs[len(fs)-1]
			fs = fs[:len(fs)-1]
			if len(cnode.Links) == 0 {
				cnode.Value, err = o.readStringFrom(&pos)
				if err != nil {
					return tree, fmtError(err)
				}
			}
			x := cnode
			cnode = ns[len(ns)-1]
			ns = ns[:len(ns)-1]
			cnode.Links = append(cnode.Links, x)
		case opCHECKRUNE:
			o.log("checkrune", ip, instr.data, "")
			res = false
			if o.cpos.r == instr.data.(rune) {
				res = true
				o.readRune()
			}
		case opCHECKRANGE:
			o.log("checkrange", ip, instr.data, "")
			res = false
			data := instr.data.([2]rune)
			if o.cpos.r >= data[0] && o.cpos.r <= data[1] {
				res = true
				o.readRune()
			}
		case opCHECKSTR:
			o.log("checkstr", ip, instr.data, "")
			res = true
			s := instr.data.(string)
			for _, r := range s {
				if r != o.cpos.r {
					res = false
					break
				}
				o.readRune()
			}
		} // switch o.code[ip]
		ip++
	} // for
}

func fmtError(err interface{}) error {
	return fmt.Errorf("(internal) %v", err)
}

func (o *TParser) log(cmd string, ip TOffset, p1, p2 interface{}) {
	stepCounter++
	s := ""
	switch p1.(type) {
	default:
		s = fmt.Sprintf("%6v %-4v %4v %6v", ip, cmd, p1, p2)
	case rune:
		s = fmt.Sprintf("%6v %-12v %q %q : %q", ip, cmd, p1, p2, o.cpos.r)
	case string:
		s = fmt.Sprintf("%6v %-12v %q = %q", ip, cmd, p1, o.cpos.r)
	}
	_ = s
	// fmt.Printf("%6v %v\n", stepCounter, s)
}

// ByID -
func (o *TParser) ByID(id int) string {
	if id >= 0 && id < len(o.items) {
		return o.items[id].name
	}
	return strconv.Itoa(id)
}

// ByName -
func (o *TParser) ByName(name string) int {
	for i := range o.items {
		if o.items[i].name == name {
			return i
		}
	}
	return -1
}

// EntryByName -
func (o *TParser) EntryByName(name string) int {
	for id := range o.entries {
		if name == o.items[id].name {
			return id
		}
	}
	return -1
}
