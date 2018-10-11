package misc

import (
	"strings"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlogger"
)

// NewSimpleLogger -
func NewSimpleLogger(filter loglevel.TFilter, format string) *zlogger.TLogger {
	if strings.TrimSpace(format) == "" {
		format = zlogger.DefaultFormat
	}
	return zlogger.Build().
		Format(format).
		LevelFilter(filter).
		Done()
}

// NewAnsiLogger -
func NewAnsiLogger(filter loglevel.TFilter, format string) *zlogger.TLogger {
	if strings.TrimSpace(format) == "" {
		format = zlogger.DefaultFormat
	}
	return zlogger.Build().
		Writer(ansi.NewAnsiStdout()).
		Styler(zlogger.AnsiStyler).
		Format(format).
		LevelFilter(filter).
		Done()
}
