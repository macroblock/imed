package zlogger

import (
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/macroblock/imed/pkg/zlog/loglevel"
)

// DefaultFormat -
// ~d - date when a message occurred
// ~t - time when a message occurred
// ~l - loglevel
// ~m - module name
// ~s - log state
// ~e - error message
// ~x - text message
// example: "~d ~t (~m) ~l: ~x~e\n"
const DefaultFormat = "(~m) ~l~s~x~e\n"

// TStyler -
type TStyler func(key rune, params *TFormatParams) (string, bool)

// Supported stylers
var (
	DefaultStyler TStyler = defaultStyler
	AnsiStyler    TStyler = ansiStyler
)

// TLogger -
type TLogger struct {
	name         string
	writer       io.Writer
	styler       TStyler
	levelFilter  loglevel.TFilter
	moduleFilter []string
	format       string
}

// LevelFilter -
func (o *TLogger) LevelFilter() loglevel.TFilter {
	return o.levelFilter
}

// ModuleFilter -
func (o *TLogger) ModuleFilter() []string {
	return append([]string(nil), o.moduleFilter...)
}

// Format -
func (o *TLogger) Format() string {
	return o.format
}

// Writer -
func (o *TLogger) Writer() io.Writer {
	return o.writer
}

// Styler -
func (o *TLogger) Styler() TStyler {
	return o.styler
}

// CanHandle -
func (o *TLogger) CanHandle(moduleName string) bool {
	if len(o.moduleFilter) == 0 {
		return true
	}
	for _, name := range o.moduleFilter {
		if name == moduleName {
			return true
		}
	}
	return false
}

// CannotHandle -
func (o *TLogger) CannotHandle(moduleName string) bool {
	return !o.CanHandle(moduleName)
}

// TFormatParams -
type TFormatParams struct {
	Format     string
	Time       time.Time
	LogLevel   loglevel.TLevel
	Text       string
	Error      error
	HasError   bool
	State      loglevel.TFilter
	ModuleName string
}

// Formatter -
func (o *TLogger) Formatter(params TFormatParams) string {
	if len(params.Format) == 0 {
		return ""
	}
	ret := ""

	if str, ok := o.styler('~', &params); ok {
		params.Format = str
	}
	format := params.Format
	for len(format) > 0 {
		ch, offs := utf8.DecodeRuneInString(format)
		format = format[offs:]
		//if ch == utf8.RuneError {
		//}
		if ch != '~' {
			ret += string(ch)
			continue
		}
		if len(format) > 0 {
			nextCh, offs := utf8.DecodeRuneInString(format)
			//if ch == utf8.RuneError {
			//}
			if nextCh == '~' {
				ret += "~"
			} else if str, ok := o.styler(nextCh, &params); ok {
				ret += str
			} else {
				ret += string(ch)
				continue
			}
			format = format[offs:]
		}
	} // for len(format) > 0
	return ret
}

func defaultStyler(key rune, params *TFormatParams) (string, bool) {
	switch key {
	case '~':
		format := params.Format
		replace := ""
		if params.Error != nil {
			replace = "\n    +cause: ~e"
		}
		format = strings.Replace(format, "~e", replace, -1)
		return format, true
	case 'd':
		return params.Time.Format("2006-01-02"), true
	case 't':
		return params.Time.Format("15:04:05"), true //params.Time().Format("2006-01-02 15:04:05"), true
	case 'l':
		return params.LogLevel.String(), true
	case 'm':
		return params.ModuleName, true
	case 'e':
		ret := ""
		if params.Error != nil {
			ret = params.Error.Error()
		}
		return ret, true
	case 'x':
		return params.Text, true
	case 's':
		if params.LogLevel == loglevel.Reset && params.State != 0 {
			return " " + params.State.String() + " ", true
		}
		return " ", true
	} // end of switch
	return "", false
}

func loglevelColor(level loglevel.TLevel) string {
	color := "\x1b[0m"
	switch level {
	case loglevel.Debug:
		color = "\x1b[1;30m" // bright black
	case loglevel.Info:
		color = "\x1b[0m" // reset //white (lightgrey)
	case loglevel.Notice:
		color = "\x1b[1;32m" // bright green
	case loglevel.Reset:
		color = "\x1b[1;36m" // bright cyan
	case loglevel.Warning:
		color = "\x1b[1;33m" // bright yellow
	case loglevel.Error:
		color = "\x1b[1;31m" // bright red
	case loglevel.Panic:
		color = "\x1b[1;31m" // bright red
	}
	return color
}

func ansiStyler(key rune, params *TFormatParams) (string, bool) {
	ret, ok := defaultStyler(key, params)
	if key == '~' {
		ret = loglevelColor(params.LogLevel) + ret + "\x1b[0m"
		ok = true
	}
	return ret, ok
}
