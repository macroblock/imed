package ptool

import "fmt"

// TLabel -
type TLabel struct {
	name     string
	addr     TOffset
	jmpTable []TOffset
}

func newLabel(name string) *TLabel {
	return &TLabel{name: name, addr: -1}
}

// SetTo -
func (o *TLabel) SetTo(addr TOffset) {
	o.addr = addr
}

func (o *TLabel) addJmp(addr TOffset) {
	o.jmpTable = append(o.jmpTable, addr)
}

// String -
func (o *TLabel) String() string {
	if o == nil {
		return fmt.Sprintf("%v", nil)
	}
	return fmt.Sprintf("%v-%v", o.name, o.addr)
}
