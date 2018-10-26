package zflag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log = zlog.Instance("zflag")
)

// TFlag -
type TFlag struct {
	arg      func()
	cmd      func() //, entry tEntry)
	keys     []string
	variable interface{}
	brief    string
	usage    string
	hint     string
}

// TSection -
type TSection struct {
	TFlag

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
func (o *TFlag) Brief() string { return o.brief }

// Usage -
func (o *TFlag) Usage() string { return o.usage }

// Hint -
func (o *TFlag) Hint() string { return o.hint }

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
func NewSection(name string, cmd func(), brief, usage, hint string, elements ...Interface) *TSection {
	section := &TSection{}
	// section.keys = splitKeys(keys)
	section.keys = []string{name}
	section.cmd = cmd
	section.brief = brief
	section.usage = usage
	section.hint = hint
	// for _, item := range elements {
	// 	switch t := item.(type) {
	// 	default:
	// 		log.Panic(true, "unreachable")
	// 	case *TFlag:
	// 		if t != nil {
	// 			section.flags = append(section.flags, t)
	// 		}
	// 	case *TSection:
	// 		if t != nil {
	// 			section.sections = append(section.sections, t)
	// 		}
	// 	}
	// }
	section.elements = elements
	log.Warning(len(section.keys) == 0, "section without key(s)")
	log.Warning(len(section.elements) == 0, "section without flag(s)")

	return section
}

// New -
func New(keys string, variable interface{}, brief, usage, hint string) *TFlag {
	flag := &TFlag{}
	flag.keys = splitKeys(keys)
	flag.variable = variable
	flag.cmd = nil
	flag.brief = brief
	flag.usage = usage
	flag.hint = hint
	log.Warning(len(flag.keys) == 0, "flag without key(s)")
	log.Warning(flag.cmd == nil, "flag without func")
	return flag
}

// ElemByKey -
func ElemByKey(section *TSection, key string) Interface {
	var ret Interface
	for _, elem := range section.elements {
		for _, k := range elem.Keys() {
			if k == key {
				return elem
			}
			if k == "" && ret == nil {
				ret = elem
			}
		}
	}
	return ret
}

// NextElem -
func NextElem(section *TSection, args *[]string) (Interface, error) {
	if len(*args) == 0 {
		return nil, nil
	}
	key := (*args)[0]
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
	if ret == nil {
		return nil, fmt.Errorf("section %q has an unsupported key %q", section.keys[0], key)
	}
	return ret, nil
}

// Parse -
func (o *TSection) Parse(args *[]string) error {
	if len(*args) == 0 {
		return fmt.Errorf("%v", "not enough parameters")
	}
	*args = (*args)[1:]
	elem, err := NextElem(o, args)
	if err != nil {
		return err
	}
	for elem != nil {
		err := elem.Parse(args)
		if err != nil {
			return err
		}
		elem, err = NextElem(o, args)
		if err != nil {
			return err
		}
	}
	return nil
}

// Parse -
func (o *TFlag) Parse(args *[]string) error {
	key := (*args)[0]
	*args = (*args)[1:]
	if t, ok := o.variable.(*bool); ok {
		*t = true
	}
	*args = (*args)[1:]
	switch t := o.variable.(type) {
	case *bool:
		*t = true
		return nil
	case *string:
		*t = arg
	case *[]string:
		*t = append(*t, arg)
	}
	*args = (*args)[1:]
	return nil
}
