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
	log.Warningf(o.cmd == nil && o.variable == nil, "flag %v cannot produce any work", o.keys)
	return &o
}

// Terminator -
func (o *TFlag) Terminator() *TFlag {
	o.setOption(optTerminator, true)
	return o
}

// // Parse -
// func (o *TFlag) Parse(args *[]string, key string) (string, error) {
// 	log.Panic(len(*args) == 0, "something went wrong")
// 	if key != "" {
// 		// key = (*args)[0]
// 		*args = (*args)[1:]
// 	}
// 	// a boolean key doesn't have an argument so it should be attempted to parse first
// 	ok := true
// 	err := error(nil)
// 	switch t := o.variable.(type) {
// 	default:
// 		ok = false
// 	case nil:
// 	case *bool:
// 		*t = true
// 	case func():
// 		t()
// 	case func() error:
// 		err = t()
// 	}
// 	if ok {
// 		return o.onError.Handle(err)
// 	}

// 	// attempt to parse keys with an argument
// 	if len(*args) == 0 {
// 		return o.onError.Handle(fmt.Errorf("a key %q of type %T without a parameter", key, o.variable))
// 	}
// 	err = error(nil)
// 	arg := (*args)[0]
// 	*args = (*args)[1:]
// 	switch t := o.variable.(type) {
// 	default:
// 		return "", internalErrorf("a key %q got an unsupported parameter of type %T", key, o.variable)
// 	case *string:
// 		*t = arg
// 	case *[]string:
// 		*t = append(*t, arg)
// 	case func(string) error:
// 		err = t(arg)
// 	}
// 	return o.onError.Handle(err)
// }

// Parse -
func (o *TFlag) Parse(args *[]string, key string) (string, error) {
	// fmt.Printf("enter flag.parse %v\n", key)
	// defer func() {
	// 	fmt.Printf("leave flag.parse %v\n", key)
	// }()

	log.Panic(len(*args) == 0, "something went wrong")
	if key != "" {
		*args = (*args)[1:]
	}
	nArgs, fn, err := getFunc(key, o.variable)
	if err != nil {
		return o.onError.Handle(err)
	}
	if nArgs > len(*args) {
		return o.onError.Handle(ErrorNotEnoughArguments())
	}
	if nArgs == 0 {
		return o.onError.Handle(fn("???"))
	}
	for i := 0; i < nArgs; i++ {
		arg := (*args)[0]
		err := fn(arg)
		if err != nil {
			return o.onError.Handle(err)
		}
		*args = (*args)[1:]
	}
	return "", nil
}

func getFunc(key string, variable interface{}) (int, func(string) error, error) {
	fn := func(string) error { return nil }
	n := -1
	switch t := variable.(type) {
	default:
		return n, nil, fmt.Errorf("a key %q got an unsupported parameter of type %T", key, variable)
	case nil:
	case *bool:
		n = 0
		fn = func(val string) error {
			*t = true
			return nil
		}
	case func():
		n = 0
		fn = func(val string) error {
			t()
			return nil
		}
	case func() error:
		n = 0
		fn = func(val string) error {
			return t()
		}
	case *string:
		n = 1
		fn = func(val string) error {
			*t = val
			return nil
		}
	case *[]string:
		n = 1
		fn = func(val string) error {
			*t = append(*t, val)
			return nil
		}
	case func(string) error:
		n = 1
		fn = func(val string) error {
			return t(val)
		}
	case IValue:
		n = 1
		fn = func(val string) error {
			return t.Set(val)
		}
	}
	return n, fn, nil
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
