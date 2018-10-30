package cli

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

const (
	optInvalid tOptions = ^tOptions(0)
	optNone    tOptions = 1 << iota
	optTerminator
)

type (
	// IElement -
	IElement interface {
		IElementUniquePattern()
	}
	// Interface -
	Interface interface {
		IElement
		Parse(*[]string, string) error
		Do() (bool, error)
		GetKeys() []string
		GetBrief() string
		GetUsage() string
		GetHint() string
		GetDoc() string
		GetOption(tOptions) bool
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
		options  tOptions
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

	tOptions uint
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
	return o.root.Parse(&args, "?")
}

// PrintHelp -
func (o *TFlagSet) PrintHelp() error {
	return o.root.PrintHelp()
}

// Elements -
func (o *TFlagSet) Elements(elements ...IElement) *TFlagSet {
	initElements(&o.root, elements...)
	return o
}

// Usage -
func (o *TFlagSet) Usage(text string) IElement { return Usage(text) }

// Hint -
func (o *TFlagSet) Hint(text string) IElement { return Hint(text) }

// Doc -
func (o *TFlagSet) Doc(text string) IElement { return Doc(text) }

// Usage -
func Usage(text string) IElement { return &tUsage{text: text} }

// Hint -
func Hint(text string) IElement { return &tHint{text: text} }

// Doc -
func Doc(text string) IElement { return &tDoc{text: text} }
