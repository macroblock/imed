package cli

import (
	"fmt"
	"strings"
)

func splitDesc(desc string) ([]string, string) {
	partSep := "\000:|"
	keySep := " \t\n\r"
	spaces := keySep

	// separate the keys and the brief parts first
	partKeys := ""
	partBrief := ""
	for i, r := range desc {
		if strings.IndexRune(partSep, r) != -1 {
			partKeys = desc[:i]
			partBrief = desc[i+1:]
			break
		}
	}

	// split and clean keys
	keys := strings.FieldsFunc(partKeys, func(r rune) bool { return strings.IndexRune(keySep, r) != -1 })
	total := 0
	for i := range keys {
		v := strings.TrimFunc(keys[i], func(r rune) bool { return strings.IndexRune(spaces, r) != -1 })
		if keys[i] == "" {
			continue
		}
		// fmt.Println("key: ", v)
		keys[total] = v
		total++
	}
	keys = keys[:total]
	if len(keys) == 0 {
		keys = append(keys, "")
	}
	// fmt.Printf("splitDesc keys: %v\n", keys)
	// clean brief
	brief := strings.TrimFunc(partBrief, func(r rune) bool { return strings.IndexRune(spaces, r) != -1 })

	return keys, brief
}

func initCommand(o *TCommand, keys []string, brief string, fn func() error, elements ...IElement) {
	if len(keys) > 0 {
		o.name = keys[0]
	}
	o.keys = keys
	o.brief = brief
	o.cmd = fn
	initElements(o, elements...)
}

// Command -
func Command(desc string, fn func() error, elements ...IElement) *TCommand {
	o := TCommand{}
	o.setOption(optTerminator, true)

	keys, brief := splitDesc(desc)
	initCommand(&o, keys, brief, fn, elements...)
	log.Warningf(o.name == "", "command without a key(s)")
	// log.Warning(len(o.elements) == 0, "command without an argument(s)")
	return &o
}

func findKeyHandler(section *TCommand, args []string, stack []Interface) (Interface, string, error) {
	if len(args) == 0 {
		return nil, "", internalError("### what is this? ###")
	}
	key := args[0]
	ret := Interface(nil)
	for _, elem := range section.elements {
		for _, k := range elem.GetKeys() {
			// fmt.Printf("key: %q\n", k)
			if k == key {
				// fmt.Println("found key: ", k)
				return elem, key, nil
			}
			if k == "" && ret == nil {
				ret = elem
			}
		}
	}
	if ret == nil {
		return nil, "", fmt.Errorf("%van unsupported key %q", commandPathPrefix(stack, key), args[0])
	}
	return ret, "", nil
}

func commandPathPrefix(stack []Interface, key string) string {
	// for _, v := range stack {
	// 	fmt.Printf("%v -", v.name)
	// }
	key = normKey(key)
	// fmt.Println(" ")
	if len(stack) == 0 {
		return ""
	}
	for i := 1; i < len(stack); i++ {
		switch t := stack[len(stack)-i].(type) {
		case *TCommand:
			fmt.Printf("name %v\n", t.name)
			if t.name == "" {
				return key + ": "
			}
			return fmt.Sprintf("%v %v: ", t.name, key)
		}
	}
	return key + ": "
}

// Parse -
func (o *TCommand) Parse(args *[]string, key string) (string, error) {
	// fmt.Println("command keys: ", o.keys)
	if len(*args) == 0 {
		return "", internalError("something went wrong")
	}

	cur := o
	stack := []Interface{}
	if key != "" {
		*args = (*args)[1:]
	}
	for len(*args) > 0 {
		elem, key, err := findKeyHandler(cur, *args, stack)
		if err != nil {
			return o.onError.Handle(err)
		}
		stack = append(stack, elem)

		switch t := elem.(type) {
		default:
			hint, err := elem.Parse(args, key)
			if err != nil {
				return hint, fmt.Errorf("%v%v", commandPathPrefix(stack, key), err)
			}
		case *TCommand:
			// fmt.Println("enter command ", t.name)
			*args = (*args)[1:]
			cur = t
		}
	}
	stack = append(stack, cur)
	err := error(nil)
	for _, v := range stack {
		terminate, err := v.Do()
		if err != nil {
			return o.onError.Handle(err)
		}
		if terminate {
			break
		}
	}
	return o.onError.Handle(err)
}
