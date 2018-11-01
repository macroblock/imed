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

// SetState -
func (o *TLog) SetState(level loglevel.TFilter) {
	o.node.state = level
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
func (o *TLog) Log(level loglevel.TLevel, resetFilter loglevel.TFilter, err error, text string) {
	if text == "" {
		if err == nil {
			return
		}
		text = err.Error()
		err = nil
	}
	formatParams := zlogger.TFormatParams{
		Time:       time.Now(),
		LogLevel:   level,
		Text:       text,
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
		panic(text)
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
		o.Log(loglevel.Panic, 0, err, fmt.Sprint(text...))
	}
}

// Panicf -
func (o *TLog) Panicf(condition interface{}, format string, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Panic, 0, err, fmt.Sprintf(format, text...))
	}
}

// Error -
func (o *TLog) Error(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Error, 0, err, fmt.Sprint(text...))
	}
}

// Errorf -
func (o *TLog) Errorf(condition interface{}, format string, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Error, 0, err, fmt.Sprintf(format, text...))
	}
}

// Warning -
func (o *TLog) Warning(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Warning, 0, err, fmt.Sprint(text...))
	}
}

// Warningf -
func (o *TLog) Warningf(condition interface{}, format string, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.Log(loglevel.Warning, 0, err, fmt.Sprintf(format, text...))
	}
}

// Reset -
func (o *TLog) Reset(resetFilter loglevel.TFilter, text ...interface{}) {
	o.Log(loglevel.Reset, resetFilter, nil, fmt.Sprint(text...))
}

// Resetf -
func (o *TLog) Resetf(resetFilter loglevel.TFilter, format string, text ...interface{}) {
	o.Log(loglevel.Reset, resetFilter, nil, fmt.Sprintf(format, text...))
}

// Notice -
func (o *TLog) Notice(text ...interface{}) {
	o.Log(loglevel.Notice, 0, nil, fmt.Sprint(text...))
}

// Noticef -
func (o *TLog) Noticef(format string, text ...interface{}) {
	o.Log(loglevel.Notice, 0, nil, fmt.Sprintf(format, text...))
}

// Info -
func (o *TLog) Info(text ...interface{}) {
	o.Log(loglevel.Info, 0, nil, fmt.Sprint(text...))
}

// Infof -
func (o *TLog) Infof(format string, text ...interface{}) {
	o.Log(loglevel.Info, 0, nil, fmt.Sprintf(format, text...))
}

// Debug -
func (o *TLog) Debug(text ...interface{}) {
	o.Log(loglevel.Debug, 0, nil, fmt.Sprint(text...))
}

// Debugf -
func (o *TLog) Debugf(fomat string, text ...interface{}) {
	o.Log(loglevel.Debug, 0, nil, fmt.Sprintf(fomat, text...))
}

// Catcher -
func (o *TLog) Catcher() *TCatcher {
	return &TCatcher{log: o}
}
