package ptool

import (
	"strconv"
)

// TBuilder -
type TBuilder struct {
	// rules []TRule
	// root      *TNode
	tree      *TNode
	err       error
	text      string
	pin, pout *TParser
	entries   []string
}

// NewBuilder -
func NewBuilder() *TBuilder {
	return &TBuilder{}
}

// FromString -
func (o *TBuilder) FromString(src string) *TBuilder {
	o.text = src
	return o
}

// TreeToString -
func (o *TBuilder) TreeToString() string {
	if o == nil || o.tree == nil {
		return "<nil>"
	}
	fn := func(v int) string {
		return strconv.Itoa(v)
	}
	if o.pin != nil {
		fn = o.pin.ByID
	}
	s := TreeToString(o.tree, fn)
	return s
}

// Entries -
func (o *TBuilder) Entries(entries ...string) *TBuilder {
	o.entries = entries
	return o
}

// Build -
func (o *TBuilder) Build() (*TParser, error) {
	if o.err != nil {
		return nil, o.err
	}
	if o.pout != nil {
		return o.pout, nil
	}
	if o.pin == nil {
		o.pin, o.err = NewZBNFParser()
	}
	if o.err != nil {
		return nil, o.err
	}
	o.tree, o.err = o.pin.Parse(o.text)
	if o.err != nil {
		return nil, o.err
	}
	// fmt.Println("xxx")
	// fmt.Println(TreeToString(o.tree, o.pin.ByID))
	// fmt.Println(o.pin.ByName("stmt"), o.pin.ByID(14))
	// fmt.Println("xxx")
	pm := newProgMaker(o.pin)
	//fmt.Println(TreeToString(o.tree, o.pin.ByID))
	o.err = pm.Compile(o.tree, o.entries...)
	if o.err != nil {
		return nil, o.err
	}
	// fmt.Println(pm.labels)
	// fmt.Println(pm.entries)
	// fmt.Println(pm)
	o.err = pm.calcJumps()
	if o.err != nil {
		return nil, o.err
	}
	return pm.prog, nil
}

// Tree -
func (o *TBuilder) Tree() *TNode {
	return o.tree
}
