package zlog

import (
	"fmt"
	"strings"
	"time"

	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlogger"
)

// Default -
var defaultLog *TLog

type tNode struct {
	state   loglevel.TFilter
	loggers []*zlogger.TLogger
}

// TLog -
type TLog struct {
	node *tNode
	name string
}

// New -
func New(name string) *TLog {
	return &TLog{name: name, node: &tNode{}}
}

// Instance -
func Instance(name string) *TLog {
	if defaultLog == nil {
		defaultLog = New(name)
		return defaultLog
	}
	ret := *defaultLog
	ret.name = name
	return &ret
}

// Instance -
func (o *TLog) Instance(name string) *TLog {
	if o == nil {
		o = New(name)
		return o
	}
	ret := *o
	ret.name = name
	return &ret
}

// State -
func (o *TLog) State() loglevel.TFilter {
	return o.node.state
}

// String -
func (o *TLog) String() string {
	sl := []string{}
	for _, l := range o.node.loggers {
		sl = append(sl, l.LevelFilter().String()) //+": "+strings.Join(l.prefixes, ","))
	}
	return strings.Join(sl, "\n")
}

// Add -
func (o *TLog) Add(logger ...*zlogger.TLogger) {
	// TODO: check on nil
	o.node.loggers = append(o.node.loggers, logger...)
}

// Log -
func (o *TLog) Log(level loglevel.TLevel, resetFilter loglevel.TFilter, err error, text ...interface{}) {
	formatParams := zlogger.TFormatParams{
		Time:       time.Now(),
		LogLevel:   level,
		Text:       fmt.Sprint(text...),
		Error:      err,
		State:      o.node.state,
		ModuleName: o.name,
	}
	o.node.state &^= resetFilter
	o.node.state |= level.Only()

	for _, logger := range o.node.loggers {
		if level.NotIn(logger.LevelFilter()) {
			continue
		}
		if !logger.CanHandle(o.name) {
			continue
		}
		formatParams.Format = logger.Format()
		msg := logger.Formatter(formatParams)
		if _, err := logger.Writer().Write([]byte(msg)); err != nil {
			// TODO: smarter
			fmt.Println(err)
		}
	}
	if level == loglevel.Panic {
		panic(fmt.Sprint(text...))
	}
}

func getErrorCondition(condition interface{}) (bool, error) {
	err := error(nil)
	ok := false
	switch v := condition.(type) {
	case nil:
	case bool:
		ok = v
	case error:
		err = v
	default:
		ok = true
	}
	return ok, err
}

// Panic -
func (o *TLog) Panic(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Panic, 0, err, text...)
	}
}

// Error -
func (o *TLog) Error(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Error, 0, err, text...)
	}
}

// Warning -
func (o *TLog) Warning(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Warning, 0, err, text...)
	}
}

// Reset -
func (o *TLog) Reset(resetFilter loglevel.TFilter, text ...interface{}) {
	o.Log(loglevel.Reset, resetFilter, nil, text...)
}

// Notice -
func (o *TLog) Notice(text ...interface{}) {
	o.Log(loglevel.Notice, 0, nil, text...)
}

// Info -
func (o *TLog) Info(text ...interface{}) {
	o.Log(loglevel.Info, 0, nil, text...)
}

// Debug -
func (o *TLog) Debug(text ...interface{}) {
	o.Log(loglevel.Debug, 0, nil, text...)
}

// Catcher -
func (o *TLog) Catcher() *TCatcher {
	return &TCatcher{log: o}
}
