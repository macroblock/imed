package edl

import (
	"fmt"
)

type Clip struct {
	Timespan
	Number int
	Type string  // BL AX
	Media string // V A
	Track int
	Effect string
	Origin Timespan
	Meta *Meta
}

func (o *Clip) String() string {
	if o == nil {
		return fmt.Sprintf("%v", nil)
	}
	return fmt.Sprintf("%v-%v-%v:%v", o.Media, o.Type, o.Effect, o.Timespan)
}
