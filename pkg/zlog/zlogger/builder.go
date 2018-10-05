package zlogger

import (
	"fmt"
	"io"
	"os"

	"github.com/macroblock/imed/pkg/zlog/loglevel"
)

var loggerNumber int

// tBuild -
type tBuild struct {
	logger *TLogger
}

func autoName() string {
	ret := fmt.Sprintf("logger%v", loggerNumber)
	loggerNumber++
	return ret
}

// Default -
func Default() *TLogger {
	return &TLogger{writer: os.Stdout, styler: DefaultStyler, name: autoName(), levelFilter: loglevel.All, format: defaultFormat}
}

// Build -
func Build() tBuild {
	return tBuild{logger: Default()}
}

// Done -
func (o tBuild) Done() *TLogger {
	ret := &TLogger{}
	*ret = *o.logger
	ret.moduleFilter = append([]string(nil), o.logger.moduleFilter...)
	o.logger.name = autoName()
	return ret
}

// Name -
func (o tBuild) Name(name string) tBuild {
	o.logger.name = name
	return o
}

// LevelFilter -
func (o tBuild) LevelFilter(filter loglevel.TFilter) tBuild {
	o.logger.levelFilter = filter
	return o
}

// ModuleFilter -
func (o tBuild) ModuleFilter(filter []string) tBuild {
	o.logger.moduleFilter = nil
	if filter != nil {
		o.logger.moduleFilter = append([]string(nil), filter...)
		//sort.Strings(o.logger.moduleFilter)
	}
	return o
}

// Writer -
func (o tBuild) Writer(writer io.Writer) tBuild {
	o.logger.writer = writer
	return o
}

// Styler -
func (o tBuild) Styler(styler TStyler) tBuild {
	o.logger.styler = styler
	return o
}

// Format -
func (o tBuild) Format(format string) tBuild {
	o.logger.format = format
	return o
}
