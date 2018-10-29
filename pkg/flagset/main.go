package flagset

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log = zlog.Instance("zflag")
)

type (
	// IElement -
	IElement interface {
		IElementUniquePattern()
	}
	// Interface -
	Interface interface {
		IElement
		Parse(*[]string) error
		Do()
		GetKeys() []string
		GetBrief() string
		GetUsage() string
		GetHint() string
		GetDoc() string
	}

	// TFlag -
	TFlag struct {
		IElement
		arg      func()
		cmd      func() error
		keys     []string
		variable interface{}
		brief    string
		usage    string
		hint     string
		doc      string
	}

	// TCommand -
	TCommand struct {
		TFlag
		name     string
		elements []Interface
	}

	// TFlagSet -
	TFlagSet struct {
		root TCommand
		// options
	}

	tUsage struct {
		IElement
		text string
	}
	tHint tUsage
	tDoc  tUsage
)

// ErrBreakExecutionWithNoError -
var ErrBreakExecutionWithNoError = errors.New("break execution with no error")

var programName string

func init() {
	programName = filepath.Base(os.Args[0])
	programName = strings.TrimSuffix(programName, filepath.Ext(programName))
}

// New -
func New(brief string, fn func() error, elements ...IElement) *TFlagSet {
	o := &TFlagSet{}
	initCommand(&o.root, nil, brief, fn, elements...)
	return o
}

// Parse -
func (o *TFlagSet) Parse(args []string) error {
	return o.root.Parse(&args)
}

// PrintHelp -
func (o *TFlagSet) PrintHelp() error {
	return o.root.PrintHelp()
}

// Elements -
func (o *TFlagSet) Elements(elements ...Interface) *TFlagSet {
	o.root.elements = elements
	return o
}

// Usage -
func (o *TFlagSet) Usage(text string) *TFlagSet {
	o.root.usage = text
	return o
}

// Hint -
func (o *TFlagSet) Hint(text string) *TFlagSet {
	o.root.hint = text
	return o
}

// Usage -
func Usage(text string) IElement {
	return &tUsage{text: text}
}

// Hint -
func Hint(text string) IElement {
	return &tHint{text: text}
}

// Doc -
func Doc(text string) IElement {
	return &tDoc{text: text}
}
