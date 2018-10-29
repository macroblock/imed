package flagset

import (
	"fmt"
	"strings"

	"github.com/macroblock/exp/pkg/misc"
)

func splitDesc(desc string) ([]string, string) {
	partSep := "\000;|"
	keySep := " \t\n\r"
	spaces := keySep

	// separate the keys and the brief parts first
	partKeys := ""
	partBrief := ""
	for i, r := range desc {
		if strings.IndexRune(partSep, r) != -1 {
			partKeys = desc[:i]
			partBrief = desc[i+1:]
		}
	}

	// split and clean keys
	keys := strings.FieldsFunc(partKeys, func(r rune) bool { return strings.IndexRune(keySep, r) != -1 })
	len := 0
	for i := range keys {
		v := strings.TrimFunc(keys[i], func(r rune) bool { return strings.IndexRune(spaces, r) != -1 })
		if keys[i] == "" {
			continue
		}
		// fmt.Println("key: ", v)
		keys[len] = v
		len++
	}
	keys = keys[:len]

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

	for _, elem := range elements {
		switch t := elem.(type) {
		default:
			log.Error(true, fmt.Sprintf("Command() got an unsupported type of element %T", t))
		case *tUsage:
			o.usage += "\n" + t.text
		case *tHint:
			o.hint += "\n" + t.text
		case *tDoc:
			o.doc += "\n" + t.text
		case Interface:
			// fmt.Println("--- ", t.GetKeys())
			o.elements = append(o.elements, t)
		}
	}
	log.Warning(o.name == "", "command without a name(s)")
	log.Warning(len(o.elements) == 0, "command without an argument(s)")
}

// Command -
func Command(desc string, fn func() error, elements ...IElement) *TCommand {
	o := TCommand{}
	keys, brief := splitDesc(desc)
	initCommand(&o, keys, brief, fn, elements...)
	return &o
}

func findKeyHandler(section *TCommand, args []string, stack []*TCommand) (Interface, error) {
	if len(args) == 0 {
		return nil, nil
	}
	key := args[0]
	ret := Interface(nil)
	for _, elem := range section.elements {
		for _, k := range elem.GetKeys() {
			// fmt.Println("key: ", k)
			if k == key {
				// fmt.Println("found key: ", k)
				return elem, nil
			}
			if k == "" && ret == nil {
				ret = elem
			}
		}
	}
	if ret == nil {
		return nil, fmt.Errorf("%van unsupported key %q",
			commandPathPrefix(stack), args[0])
	}
	return ret, nil
}

func commandPathPrefix(stack []*TCommand) string {
	for _, v := range stack {
		fmt.Printf("%v -", v.name)
	}
	fmt.Println(" ")
	if len(stack) == 0 {
		return ""
	}
	name := stack[len(stack)-1].name
	if name == "" {
		return ""
	}
	return name + ": "
}

// Parse -
func (o *TCommand) Parse(args *[]string) error {
	// fmt.Println("command keys: ", o.keys)
	if len(*args) == 0 {
		return fmt.Errorf("%v", "something went wrong")
	}

	cur := o
	stack := []*TCommand{cur}
	*args = (*args)[1:]
	for len(*args) > 0 {
		elem, err := findKeyHandler(cur, *args, stack)
		if err != nil {
			return err
		}

		if t, ok := elem.(*TCommand); ok {
			cur = t
			stack = append(stack, cur)
			*args = (*args)[1:]
			// fmt.Println("enter command ", cur.name)
			continue
		}

		err = elem.Parse(args)
		if err != nil {
			return fmt.Errorf("%v%v", commandPathPrefix(stack), err)
		}
	}
	return nil
}

func addNewLine(s string) string {
	if s != "" {
		return s + "\n"
	}
	return s
}

func formatFlags(elements []Interface, prefix string) string {
	if len(elements) == 0 {
		return ""
	}
	maxKeyStr := 0
	lines := []string{}
	for _, elem := range elements {
		keys := strings.Join(elem.GetKeys(), ", ")
		maxKeyStr = misc.MaxInt(maxKeyStr, len(keys))
		lines = append(lines, keys)
	}
	for i, elem := range elements {
		lines[i] = fmt.Sprintf("    %v %v  %v", lines[i], strings.Repeat(" ", maxKeyStr-len(lines[i])), elem.GetBrief())
	}
	return fmt.Sprintf("\nThe %s are:\n%v", prefix, strings.Join(lines, ""))
}

// PrintHelp -
func (o *TCommand) PrintHelp() error {

	flags := []Interface{}
	sections := []Interface{}
	for _, elem := range o.elements {
		switch t := elem.(type) {
		default:
			return fmt.Errorf("something went wrong")
		case *TFlag:
			flags = append(flags, t)
		case *TCommand:
			sections = append(sections, t)
		}
	}
	text := fmt.Sprintf("%v%v%v%v%v", o.GetBrief(), o.GetUsage(),
		formatFlags(flags, "flags"),
		formatFlags(sections, "sections"),
		o.GetHint(),
	)
	text = strings.Replace(text, "!PROG!", programName, -1)

	fmt.Print(text)

	return nil
}
