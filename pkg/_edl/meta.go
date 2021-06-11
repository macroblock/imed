package edl

import (
	"fmt"
)

type Meta struct {
	Name string
}

func(o *Meta) String() string {
	if o == nil {
		return fmt.Sprintf("%v", nil)
	}
	return fmt.Sprintf("%+v", *o)
}
