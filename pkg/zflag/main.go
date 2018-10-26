package zflag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/macroblock/imed/pkg/misc"

	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log = zlog.Instance("zflag")
)

// TFlag -
type TFlag struct {
	arg      func()
	cmd      func() error
	keys     []string
	variable interface{}
	brief    string
	usage    string
	hint     string
	fullDoc  string
}

// TSection -
type TSection struct {
	TFlag

	name string
	// flags    []*TFlag
	// sections []*TSection
	elements []Interface

	keyMap map[string]Interface
}

// Interface -
type Interface interface {
	Parse(*[]string) error
	Keys() []string
	Do()
	Brief() string
	Usage() string
	Hint() string
}

var programName string

func init() {
	programName = filepath.Base(os.Args[0])
	programName = strings.TrimSuffix(programName, filepath.Ext(programName))
}

// Do -
func (o *TFlag) Do() {
	if o.cmd != nil {
		o.cmd()
	}
}

// Keys -
func (o *TFlag) Keys() []string { return o.keys }

// Brief -
func (o *TFlag) Brief() string {
	if o.brief == "" {
		return ""
	}
	return fmt.Sprintf("%v\n", o.brief)
}

// Usage -
func (o *TFlag) Usage() string {
	if o.usage == "" {
		return ""
	}
	return fmt.Sprintf("\nUsage:\n    %v\n", o.usage)
}

// Hint -
func (o *TFlag) Hint() string {
	if o.hint == "" {
		return ""
	}
	return fmt.Sprintf("\n%v\n", o.hint)
}

func splitKeys(keys string) []string {
	ret := []string{}
	for _, key := range strings.Split(keys, " ") {
		key = strings.TrimSpace(key)
		if key != "" {
			ret = append(ret, key)
		}
	}
	return ret
}

// NewSection -
func NewSection(name string, fn func() error, brief, usage, hint, fullDoc string, elements ...Interface) *TSection {
	section := &TSection{}
	section.name = name
	if name != "" {
		section.keys = []string{name}
	}
	section.cmd = fn
	section.brief = brief
	section.usage = usage
	section.hint = hint
	section.fullDoc = fullDoc
	section.elements = elements
	log.Warning(len(section.keys) == 0, "section without key(s)")
	log.Warning(len(section.elements) == 0, "section without flag(s)")

	return section
}

// New -
func New(keys string, variable interface{}, fn func() error, brief, usage, hint, fullDoc string) *TFlag {
	flag := &TFlag{}
	flag.keys = splitKeys(keys)
	flag.variable = variable
	flag.cmd = nil
	flag.brief = brief
	flag.usage = usage
	flag.hint = hint
	flag.fullDoc = fullDoc
	log.Warning(len(flag.keys) == 0, "flag without key(s)")
	log.Warning(flag.cmd == nil, "flag without func")
	return flag
}

// FindKeyHandler -
func FindKeyHandler(section *TSection, args []string) (Interface, error) {
	if len(args) == 0 {
		return nil, nil
	}
	key := args[0]
	ret := Interface(nil)
	for _, elem := range section.elements {
		for _, k := range elem.Keys() {
			if k == key {
				return elem, nil
			}
			if k == "" && ret == nil {
				ret = elem
			}
		}
	}
	return ret, nil
}

// Parse -
func (o *TSection) Parse(args *[]string) error {
	if len(*args) == 0 {
		return fmt.Errorf("%v", "something went wrong")
	}
	// get the correct name prefix for an error reporting
	errPrefix := o.name
	if errPrefix != "" {
		errPrefix = fmt.Sprintf("<%v> has an ", errPrefix)
	}

	*args = (*args)[1:]
	for len(*args) > 0 {
		elem, err := FindKeyHandler(o, *args)
		if err != nil {
			return err
		}
		if elem == nil {
			return fmt.Errorf("%vunsupported key %q or a parameter without a key", errPrefix, (*args)[0])
		}
		err = elem.Parse(args)
		if err != nil {
			switch elem.(type) {
			default:
				return fmt.Errorf("%v%v", errPrefix, err)
			case *TSection:
				return fmt.Errorf("%v", err)
			}
		}
	}
	return nil
}

// Parse -
func (o *TFlag) Parse(args *[]string) error {
	if len(*args) == 0 {
		return fmt.Errorf("%v", "something went wrong")
	}
	key := (*args)[0]
	*args = (*args)[1:]
	// a boolean key doesn't have an argument so it should be processing first
	if t, ok := o.variable.(*bool); ok {
		*t = true
		return nil
	}
	if len(*args) == 0 {
		return fmt.Errorf("a key %q of type %T without a parameter", key, o.variable)
	}
	arg := (*args)[0]
	switch t := o.variable.(type) {
	default:
		return fmt.Errorf("an unsupported key of %T type", o.variable)
	case *string:
		*t = arg
	case *[]string:
		*t = append(*t, arg)
	}
	*args = (*args)[1:]
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
		keys := strings.Join(elem.Keys(), ", ")
		maxKeyStr = misc.MaxInt(maxKeyStr, len(keys))
		lines = append(lines, keys)
	}
	for i, elem := range elements {
		lines[i] = fmt.Sprintf("    %v %v  %v", lines[i], strings.Repeat(" ", maxKeyStr-len(lines[i])), elem.Brief())
	}
	return fmt.Sprintf("\nThe %s are:\n%v", prefix, strings.Join(lines, ""))
}

// PrintHelp -
func (o *TSection) PrintHelp() error {

	flags := []Interface{}
	sections := []Interface{}
	for _, elem := range o.elements {
		switch t := elem.(type) {
		default:
			return fmt.Errorf("something went wrong")
		case *TFlag:
			flags = append(flags, t)
		case *TSection:
			sections = append(sections, t)
		}
	}
	text := fmt.Sprintf("%v%v%v%v%v", o.Brief(), o.Usage(),
		formatFlags(flags, "flags"),
		formatFlags(sections, "sections"),
		o.Hint(),
	)
	text = strings.Replace(text, "!progname!", programName, -1)

	fmt.Print(text)

	return nil
}
