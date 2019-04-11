package ptool

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// TProgMaker -
type TProgMaker struct {
	labels []*TLabel
	// entries map[string]*tLutItem
	lut  *tLut
	prog *TParser
	pin  *TParser
	err  error
}

// func newEntry(id int, name string, ip TOffset) *TEntry {
// 	return &TEntry{id: id, name: name, ip: ip}
// }

func newProgMaker(pin *TParser) *TProgMaker {
	if pin == nil {
		return nil
	}
	o := &TProgMaker{prog: &TParser{}, pin: pin}
	return o
}

// Emit -
func (o *TProgMaker) Emit(opcode TOpCode, data interface{}) {
	if o.err != nil {
		return
	}
	o.prog.code = append(o.prog.code, TInstruction{opcode, data})
	switch opcode {
	case opJMP, opJNZ, opJZ:
		label, ok := data.(*TLabel)
		if !ok {
			o.err = fmt.Errorf("while trying cast data to *TLabel at #[%v] instr %v", o.IP(), opcode)
			return
		}
		// if o == nil {
		// 	fmt.Println("##### ", nil)
		// }
		// fmt.Println("##### ", 1)
		if label != nil {
			label.addJmp(o.IP())
		}
	}
}

// // AddEntry -
// func (o *TProgMaker) AddEntry(id int, name string, ip TOffset) {
// 	o.entries = append(o.entries, *newEntry(id, name, ip))
// }

// AddLabel -
func (o *TProgMaker) AddLabel(name string) *TLabel {
	ret := newLabel(name)
	o.labels = append(o.labels, ret)
	return ret
}

// IP -
func (o *TProgMaker) IP() TOffset {
	return TOffset(len(o.prog.code) - 1)
}

// NextIP -
func (o *TProgMaker) NextIP() TOffset {
	return TOffset(len(o.prog.code))
}

func (o *TProgMaker) buildLUT(root *TNode) {
	// fmt.Println(root)
	// fmt.Println("===============================================\n", o.pin.entries)
	// idCounter := 1
	lut := newLut()
	// lut := map[string]*tLutItem{}
	err := inspectPreOrder(root, func(node *TNode) (bool, error) {
		// fmt.Println("---", o.pin.ByID(node.Type))
		switch o.pin.ByID(node.Type) {
		case cStmt:
			if len(node.Links) != 2 || (o.pin.ByID(node.Links[0].Type) != cIdent && o.pin.ByID(node.Links[0].Type) != cLVal) {
				fmt.Println(node)
				fmt.Println(node.Type)
				// fmt.Println(cStmt)
				return false, fmt.Errorf("optimizeZBNFTree:buildLUT: incorrect statement in a tree")
			}
			name := node.Links[0].Value
			// fmt.Println("---", name)
			// _, ok := lut[name]
			err := lut.addItem(name, node)
			if err != nil {
				return false, fmt.Errorf("optimizeZBNFTree:buildLUT: duplicate identifier %q in a tree", name)
			}
			// fmt.Println("-", name, "-", node.Links[0].Type)
			// lut[name] = &tLutItem{node: node.Links[1]}
			// lut[name] = &tLutItem{node: node, ip: -1, id: idCounter, name: name}
			// idCounter++
		}
		return true, nil
	})
	if err != nil {
		fmt.Printf("error %v", err)
	}
	// entries := make([]tProgEntry, idCounter)
	// entries[0].name = cErrorNode
	// entries[0].ip = -1
	// for _, val := range lut {
	// 	entries[val.id].name = val.name
	// 	entries[val.id].ip = val.ip
	// }
	// o.prog.entries = entries
	// o.entries = lut
	// fmt.Println("out:\n", lut)
	o.lut = lut
}

// Compile -
func (o *TProgMaker) Compile(root *TNode, entries ...string) error {
	deferred := []string{}
	o.buildLUT(root)
	// fmt.Println(o.lut)
	// pm := newProgMaker()
	// err := error(nil)
	for _, entry := range entries {
		// item, ok := o.entries[entry]
		item := o.lut.find(entry)
		// fmt.Println(item)
		if item == nil {
			return fmt.Errorf("Compile: undefined statement %q", entry)
		}
		err := o.lut.addEntry(item)
		if item == nil {
			return fmt.Errorf("Compile: duplicate entry %q", entry)
		}
		o.Emit(opCALL, item.name)
		o.Emit(opEND, nil)
		def, err := o.localCompile(o, nil, item.node, o.err)
		o.err = err
		deferred = append(deferred, def...)
	}
	for len(deferred) > 0 {
		entry := deferred[len(deferred)-1]
		deferred = deferred[:len(deferred)-1]
		// item, ok := o.entries[entry]
		item := o.lut.find(entry)
		if item == nil {
			return fmt.Errorf("Compile: undefined statement %q", entry)
		}
		if item.ip >= 0 {
			continue
		}
		def, err := o.localCompile(o, nil, item.node, o.err)
		o.err = err
		if o.err != nil {
			return err
		}
		deferred = append(deferred, def...)
	}
	if o.err != nil {
		return o.err
	}
	o.prog.items = o.lut.makeProgItems()
	o.prog.entries = o.lut.makeProgEntries()
	return nil
}

func (o *TProgMaker) getRuneFromTerm(node *TNode) (rune, error) {
	if o.err != nil {
		return 0, o.err
	}
	intType := node.Type
	strType := o.pin.ByID(node.Type)
	param := node.Value
	switch strType {
	default:
		return 0, fmt.Errorf("unknown term type [%v %q]", intType, strType)
	case cString:
		if utf8.RuneCountInString(param) != 1 {
			return 0, fmt.Errorf("not a rune %q", param)
		}
		ret, _ := utf8.DecodeRuneInString(param)
		return ret, nil
	case cEOF:
		// fmt.Println("getRune: ", RuneEOF)
		return RuneEOF, nil
	case cHex8:
		ret, _ := strconv.ParseInt(param, 16, 16)
		return rune(ret), nil
	}
}

func (o *TProgMaker) localCompile(pm *TProgMaker, toFail *TLabel, root *TNode, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	deferred := []string{}
	def := []string{}
	switch o.pin.ByID(root.Type) {
	default:
		fmt.Println("#Unknown!!!")
		pm.Emit(opERROR, nil)
		return nil, fmt.Errorf("compile: unsupported Type %v", root.Type)
	case cStmt:
		name := root.Links[0].Value
		node := root.Links[1]
		// fmt.Println("#statement: ", name)
		fail := pm.AddLabel(name + "-Fail")
		fail.SetTo(pm.NextIP())
		pm.Emit(opSETERROR, true)
		pm.Emit(opRET, false)
		// pm.AddEntry(-1, name, pm.NextIP())
		// entry, ok := pm.entries[name]
		item := o.lut.find(name)
		if item == nil {
			return nil, fmt.Errorf("compile: cStmt undefined statement %q", name)
		}
		item.ip = pm.NextIP()
		def, err = o.localCompile(pm, fail, node, err)
		deferred = append(deferred, def...)
		pm.Emit(opRET, true)
	case cAnd:
		// fmt.Println("#and")
		skipSpace := false
		for i := 0; i < len(root.Links); i++ {
			node := root.Links[i]
			if o.pin.ByID(node.Type) == cNoSpace {
				skipSpace = false
				continue
			}
			if skipSpace {
				// entry, ok := pm.entries[""]
				// if ok {
				item := o.lut.find("")
				if item != nil {
					pm.Emit(opCALL, "")
					if item.ip < 0 {
						deferred = append(deferred, "")
					}
				}
			}
			def, err = o.localCompile(pm, toFail, node, err)
			deferred = append(deferred, def...)
			skipSpace = true
			_ = skipSpace
		}
	case cOr:
		// fmt.Println("#or")
		lbl := pm.AddLabel("Or")
		pm.Emit(opMARK, nil)
		for i := range root.Links {
			fail := pm.AddLabel("Or-Fail")
			node := root.Links[i]
			def, err = o.localCompile(pm, fail, node, err)
			deferred = append(deferred, def...)
			pm.Emit(opJMP, lbl)
			fail.SetTo(pm.NextIP())
			if i == len(root.Links)-1 {
				break
			}
			pm.Emit(opSETERROR, true)
			pm.Emit(opREPEAT, nil)
		}
		pm.Emit(opSETERROR, true)
		pm.Emit(opRELEASE, nil)
		pm.Emit(opJMP, toFail)
		lbl.SetTo(pm.NextIP())
		pm.Emit(opRELEASE, nil)
	case cStar:
		// fmt.Println("#{}: ")
		node := root.Links[0]
		skipSpace := o.pin.ByID(node.Type) != cNoSpace
		if !skipSpace {
			node = root.Links[1]
		}
		fail := pm.AddLabel("{}-Fail")
		loop := pm.AddLabel("loop")
		loop.SetTo(pm.NextIP())
		pm.Emit(opMARK, nil)
		def, err = o.localCompile(pm, fail, node, err)
		deferred = append(deferred, def...)
		pm.Emit(opRELEASE, nil)
		if skipSpace {
			// entry, ok := pm.entries[""]
			// if ok {
			item := o.lut.find("")
			if item != nil {
				pm.Emit(opCALL, "")
				if item.ip < 0 {
					deferred = append(deferred, "")
				}
			}
		}
		pm.Emit(opJMP, loop)
		fail.SetTo(pm.NextIP())
		pm.Emit(opRESTORE, nil)
		pm.Emit(opTRUE, nil)
	case cMaybe:
		// fmt.Println("#[]: ")
		node := root.Links[0]
		fail := pm.AddLabel("[]-Fail")
		label := pm.AddLabel("[]-Ok")
		pm.Emit(opMARK, nil)
		def, err = o.localCompile(pm, fail, node, err)
		deferred = append(deferred, def...)
		pm.Emit(opRELEASE, nil)
		pm.Emit(opJMP, label)
		fail.SetTo(pm.NextIP())
		pm.Emit(opRESTORE, nil)
		pm.Emit(opTRUE, nil)
		label.SetTo(pm.NextIP())
	case cNegative:
		// fmt.Println("#!: ")
		node := root.Links[0]
		fail := pm.AddLabel("!-Fail")
		pm.Emit(opMARK, nil)
		def, err = o.localCompile(pm, fail, node, err)
		deferred = append(deferred, def...)
		pm.Emit(opFALSE, nil)
		pm.Emit(opRESTORE, nil)
		pm.Emit(opJMP, toFail)
		fail.SetTo(pm.NextIP())
		pm.Emit(opSETERROR, false)
		pm.Emit(opRESTORE, nil)
		pm.Emit(opTRUE, nil)
	case cKeep:
		node := root.Links[0]
		name := root.Links[0].Value
		// fmt.Println("#@:", name)
		// entry, ok := pm.entries[name]
		// if !ok {
		item := o.lut.find(name)
		if item == nil {
			return nil, fmt.Errorf("compile: cKeep undefined statement %q", name)
		}
		label := pm.AddLabel("@-Ok")
		fail := pm.AddLabel("@-Fail")
		pm.Emit(opPUSHNODE, item.id)
		pm.Emit(opMARK, nil)
		def, err = o.localCompile(pm, fail, node, err)
		deferred = append(deferred, def...)
		pm.Emit(opACCEPT, nil)
		pm.Emit(opJMP, label)
		fail.SetTo(pm.NextIP())
		pm.Emit(opSETERROR, true)
		pm.Emit(opRELEASE, nil)
		pm.Emit(opPOPNODE, nil)
		pm.Emit(opJMP, toFail)
		label.SetTo(pm.NextIP())
	case cIdent:
		name := root.Value
		//fmt.Println("#ident:", name)
		// entry, ok := pm.entries[name]
		// if !ok {
		item := o.lut.find(name)
		if item == nil {
			return nil, fmt.Errorf("compile: cIdent undefined statement %q", name)
		}
		if item.ip < 0 {
			deferred = append(deferred, name)
		}
		pm.Emit(opCALL, name)
		pm.Emit(opJZ, toFail)
	case cString:
		s := root.Value
		// fmt.Println("#string: ", s)
		switch len(s) {
		default:
			pm.Emit(opCHECKSTR, s)
			pm.Emit(opJZ, toFail)
		case 1:
			r, _ := utf8.DecodeRuneInString(root.Value)
			pm.Emit(opCHECKRUNE, r)
			pm.Emit(opJZ, toFail)
		case 0:
		}
	case cRange:
		a, err := o.getRuneFromTerm(root.Links[0])
		if err != nil {
			return nil, fmt.Errorf("first range parameter: %v", err)
		}
		b, err := o.getRuneFromTerm(root.Links[2])
		if err != nil {
			return nil, fmt.Errorf("second range parameter: %v", err)
		}
		// fmt.Println("#range: ", a, b)
		if root.Links[1].Value == "-" {
			b--
			// fmt.Println("b: ", b)
		}
		pm.Emit(opCHECKRANGE, [2]rune{a, b})
		pm.Emit(opJZ, toFail)
	case cHex8:
		a, err := o.getRuneFromTerm(root)
		if err != nil {
			return nil, fmt.Errorf("escaped parameter: %v", err)
		}
		// fmt.Println("#range: ", a, b)
		pm.Emit(opCHECKRUNE, a)
		pm.Emit(opJZ, toFail)
	case cEOF:
		// fmt.Println("#EOF")
		// fmt.Println("#EOF ", RuneEOF)
		pm.Emit(opCHECKRUNE, RuneEOF)
		pm.Emit(opJZ, toFail)
	}
	return deferred, err
}

func (o *TProgMaker) calcJumps() error {
	if o.err != nil {
		return o.err
	}
	for ip := range o.prog.code {
		instr := &o.prog.code[ip]
		switch instr.opcode {
		case opJMP, opJNZ, opJZ:
			label, ok := instr.data.(*TLabel)
			if !ok {
				return fmt.Errorf("renderJumps: when cast to *TLabel [%v] %v", ip, instr)
			}
			instr.data = label.addr
		case opCALL:
			name, ok := instr.data.(string)
			if !ok {
				return fmt.Errorf("renderJumps: when cast to string [%v] %v", ip, instr)
			}
			entry := o.entryByName(name)
			if entry == nil {
				return fmt.Errorf("renderJumps: unknown entry %q [%v] %v", name, ip, instr)
			}
			instr.data = entry.ip
		}
	}
	return nil
}

// LabelsByIP -
func (o *TProgMaker) LabelsByIP(ip TOffset) []*TLabel {
	ret := []*TLabel(nil)
	for _, label := range o.labels {
		if label.addr == ip {
			ret = append(ret, label)
		}
	}
	return ret
}

func (o *TProgMaker) entryByName(name string) *tLutItem {
	// ret, ok := o.entries[name]
	// if !ok {
	return o.lut.find(name)
	// if item == nil {
	// 	return nil
	// }
	// return ret
	// for i := range o.entries {
	// 	if o.entries[i].name == name {
	// 		return &o.entries[i]
	// 	}
	// }
	// return nil
}

func (o *TProgMaker) String() string {
	ret := ""
	line := ""
	indent := "                "
	ret = fmt.Sprintf("%v\n", len(o.prog.code))
	for ip, instr := range o.prog.code {
		labels := o.LabelsByIP(TOffset(ip))
		prefix := fmt.Sprintf("%04v%v", ip, indent)
		for _, label := range labels {
			ret += fmt.Sprintf("      %v:\n", label)
		}
		for _, entry := range o.lut.items {
			if entry.ip == TOffset(ip) {
				ret += fmt.Sprintf("      +++%v:\n", entry.name)
			}
		}
		switch instr.opcode {
		default:
			line = fmt.Sprintf("%v%v", prefix, instr.opcode)
			if instr.data != nil {
				line = fmt.Sprintf("%v %v", line, instr.data)
			}
		case opJMP, opJNZ, opJZ, opCALL:
			form := "%v%v %v"
			if _, ok := instr.data.(TOffset); ok {
				form = "%v%v %4v"
			}
			line = fmt.Sprintf(form, prefix, instr.opcode, instr.data)
		case opCHECKRUNE, opCHECKSTR:
			line = fmt.Sprintf("%v%v %q", prefix, instr.opcode, instr.data)
		case opCHECKRANGE:
			data := instr.data.([2]rune)
			line = fmt.Sprintf("%v%v %q..%q", prefix, instr.opcode, data[0], data[1])
		case opNOP:
			line = fmt.Sprintf("    %v", instr.opcode)
		}
		ret += line + "\n"
	}
	if o.err != nil {
		ret = fmt.Sprintf("%v\n#Error: %v\n", ret, o.err)
	}
	return ret
}
