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
	Prepare() error
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

// Prepare -
func (o *TFlag) Prepare() error { return nil }

// Parse -
func (o *TFlag) Parse(args *[]string) error { return nil }

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

// Prepare -
func (o *TSection) Prepare() error {
	if o.keyMap != nil {
		return nil
	}
	o.keyMap = map[string]Interface{}
	for _, elem := range o.elements {
		err := elem.Prepare()
		if err != nil {
			return err
		}
		for _, key := range elem.Keys() {
			_, ok := o.keyMap[key]
			log.Warning(ok, fmt.Sprintf("section %q has duplicated key %q", o.keys[0], key))
			o.keyMap[key] = elem
		}
	}
	return nil
}

// Parse -
func (o *TSection) Parse(args *[]string) error {
	err := o.Prepare()
	if err != nil {
		return err
	}
	for len(*args) > 0 {
		elem, ok := o.keyMap[(*args)[0]]
		if !ok {
			return fmt.Errorf("section %q unsupported key %q", o.keys[0], (*args)[0])
		}
		args = (*args)[1:]
		err := elem.Parse(args)
		if err != nil {
			return err
		}
	}

	return nil
}
