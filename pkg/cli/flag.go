package cli

import (
	"fmt"
)

// Do -
func (o *TFlag) Do() (bool, error) {
	if o.cmd != nil {
		return o.GetOption(optTerminator), o.cmd()
	}
	return o.GetOption(optTerminator), nil
}

// Flag -
func Flag(desc string, variable interface{}, elements ...IElement) *TFlag {
	o := TFlag{}
	o.keys, o.brief = splitDesc(desc)
	// fmt.Printf("Flag keys: %v\n", o.keys)
	o.variable = variable
	o.cmd = nil
	initElements(&o, elements...)
	// log.Warning(len(o.keys) == 0, "flag without key(s)")
	log.Warning(o.cmd == nil && o.variable == nil, fmt.Sprintf("flag %v cannot produce any work", o.keys))
	return &o
}

// Terminator -
func (o *TFlag) Terminator() *TFlag {
	o.setOption(optTerminator, true)
	return o
}

// Parse -
func (o *TFlag) Parse(args *[]string, key string) (string, error) {
	if len(*args) == 0 {
		return "", internalErrorf("%v", "something went wrong")
	}
	if key != "" {
		// key = (*args)[0]
		*args = (*args)[1:]
	}
	// a boolean key doesn't have an argument so it should be attempted to parse first
	ok := true
	err := error(nil)
	switch t := o.variable.(type) {
	default:
		ok = false
	case nil:
	case *bool:
		*t = true
	case func():
		t()
	case func() error:
		err = t()
	}
	if ok {
		return o.onError.Handle(err)
	}

	// attempt to parse keys with an argument
	if len(*args) == 0 {
		return o.onError.Handle(fmt.Errorf("a key %q of type %T without a parameter", key, o.variable))
	}
	err = error(nil)
	arg := (*args)[0]
	*args = (*args)[1:]
	switch t := o.variable.(type) {
	default:
		return "", internalErrorf("an unsupported key of type %T", o.variable)
	case *string:
		*t = arg
	case *[]string:
		*t = append(*t, arg)
	case func(string) error:
		err = t(arg)
	}
	return o.onError.Handle(err)
}

// GetKeys -
func (o *TFlag) GetKeys() []string { return o.keys }

// GetBrief -
func (o *TFlag) GetBrief() string { return o.brief }

// GetUsage -
func (o *TFlag) GetUsage() string { return o.usage }

// GetHint -
func (o *TFlag) GetHint() string { return o.hint }

// GetDoc -
func (o *TFlag) GetDoc() string { return o.doc }

// GetOption -
func (o *TFlag) GetOption(opt tOptions) bool {
	return o.options&opt == opt
}

func (o *TFlag) setOption(opt tOptions, val bool) {
	if val {
		o.options |= opt
		return
	}
	o.options &= ^opt
}
