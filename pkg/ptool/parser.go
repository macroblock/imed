package ptool

// // TParser -
// type TParser struct {
// 	src  *bytes.Reader
// 	prog *TProg
// }

// // NewParser -
// func NewParser() *TParser {
// 	return &TParser{}
// }

// // Reset -
// func (o *TParser) Reset(src []byte) {
// 	o.src = bytes.NewReader(src)
// 	o.prog.Reset(src)
// }

// NewZBNFParser -
func NewZBNFParser() (*TParser, error) {
	// init
	zbnf := makeZBNFRules()
	// fmt.Println("---------------------------\n", zbnf.pin.entries)
	// optimizeZBNFTree(zbnf.root, "ident")
	// fmt.Println(zbnf.TreeToString())

	// compile
	pm := newProgMaker(zbnf.pin)
	err := pm.Compile(zbnf.tree, "entry")
	// fmt.Println(TreeToString(zbnf.tree, zbnf.pin.ByID))
	if err != nil {
		return nil, err
	}
	// fmt.Println(pm.labels)
	// fmt.Println(pm.entries)
	// fmt.Println(pm)
	err = pm.calcJumps()
	if err != nil {
		return nil, err
	}
	return pm.prog, nil
}
