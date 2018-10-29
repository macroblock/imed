package flagset

import (
	"fmt"
)

// Do -
func (o *TFlag) Do() {
	if o.cmd != nil {
		o.cmd()
	}
}

// Flag -
func Flag(desc string, variable interface{}, elements ...IElement) *TFlag {
	o := &TFlag{}
	o.keys, o.brief = splitDesc(desc)
	o.variable = variable
	o.cmd = nil
	for _, elem := range elements {
		switch t := elem.(type) {
		default:
			log.Error(true, fmt.Sprintf("Flag() got unsupported type of element %T", t))
		case *tUsage:
			o.usage += "\n" + t.text
		case *tHint:
			o.hint += "\n" + t.text
		case *tDoc:
			o.doc += "\n" + t.text
		}
	}
	log.Warning(len(o.keys) == 0, "flag without key(s)")
	log.Warning(o.cmd == nil, "flag without func")
	return o
}

// Parse -
func (o *TFlag) Parse(args *[]string) error {
	if len(*args) == 0 {
		return fmt.Errorf("%v", "something went wrong")
	}
	key := (*args)[0]
	*args = (*args)[1:]
	// a boolean key doesn't have an argument so it should be attempted to parse first
	switch t := o.variable.(type) {
	case *bool:
		*t = true
		return nil
	case func():
		t()
		return nil
	case func() error:
		return t()
	}

	// attempt to parse keys with an argument
	if len(*args) == 0 {
		return fmt.Errorf("a key %q of type %T without a parameter", key, o.variable)
	}
	err := error(nil)
	arg := (*args)[0]
	*args = (*args)[1:]
	switch t := o.variable.(type) {
	default:
		return fmt.Errorf("an unsupported key of type %T", o.variable)
	case *string:
		*t = arg
	case *[]string:
		*t = append(*t, arg)
	case func(string) error:
		err = t(arg)
	}
	return err
}

// GetKeys -
func (o *TFlag) GetKeys() []string { return o.keys }

// GetBrief -
func (o *TFlag) GetBrief() string {
	if o.brief == "" {
		return ""
	}
	return fmt.Sprintf("%v\n", o.brief)
}

// GetUsage -
func (o *TFlag) GetUsage() string {
	if o.usage == "" {
		return ""
	}
	return fmt.Sprintf("\nUsage:\n    %v\n", o.usage)
}

// GetHint -
func (o *TFlag) GetHint() string {
	if o.hint == "" {
		return ""
	}
	return fmt.Sprintf("\n%v\n", o.hint)
}

// GetDoc -
func (o *TFlag) GetDoc() string {
	if o.doc == "" {
		return ""
	}
	return fmt.Sprintf("\n%v\n", o.doc)
}
