package cli

import (
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
	// IValue -
	IValue interface {
		String() string
		Set(string) error
	}
	// IElement -
	IElement interface {
		IElementUniquePattern()
	}
	// Interface -
	Interface interface {
		IElement
		Parse(*[]string, string) (string, error)
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
		onError  *tErrorHandler
	}

	// TCommand -
	TCommand struct {
		TFlag
		name     string
		elements []Interface
		postHint string
	}

	// TFlagSet -
	TFlagSet struct {
		root TCommand
		// options
	}

	tOnError struct {
		IElement
		val interface{}
	}
	tUsage struct {
		IElement
		text string
	}
	tHint tUsage
	tDoc  tUsage

	tOptions uint
)

var programName string

func init() {
	programName = filepath.Base(os.Args[0])
	programName = strings.TrimSuffix(programName, filepath.Ext(programName))
}

// New -
func New(brief string, fn func() error, elements ...IElement) *TFlagSet {
	o := &TFlagSet{}
	initCommand(&o.root, nil, brief, fn, elements...)
	o.root.setOption(optTerminator, false)
	return o
}

// Parse -
func (o *TFlagSet) Parse(args []string) error {
	hint, err := o.root.Parse(&args, "?")
	o.root.postHint = hint
	return err
}

// GetHint -
func (o *TFlagSet) GetHint() string {
	return o.root.postHint
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

// OnError -
func OnError(arg interface{}) IElement {
	o, err := newErrorHandler(arg)
	log.Error(err)
	return o
}
