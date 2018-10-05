package main

import (
	"fmt"
	"os"

	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
	"github.com/macroblock/imed/pkg/zlog/zlogger"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

func doProcess(path string, schema string) bool {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	tn, err := tagname.NewFromFilename(path)
	retif.Error(err, "cannot parse filename")

	if schema == "" {
		schema = tn.Schema()
	}
	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, "cannot convert to '"+schema+"'")

	err = os.Rename(path, newPath)
	retif.Error(err, "cannot rename file")

	log.Notice(schema, " > ", newPath)

	return true
}

func main() {
	log.Add(
		zlogger.Build().
			LevelFilter(logFilter).
			Styler(zlogger.AnsiStyler).
			Done(),
		zlogger.Build().
			LevelFilter(loglevel.Info.Only().Include(loglevel.Notice.Only())).
			Format("~x\n").
			Styler(zlogger.AnsiStyler).
			Done())

	args := os.Args
	if len(args) <= 1 {
		log.Info(`
Error: not enougth arguments

Usage:
    tnrename [-rt|-old] filename {filename}`)
		return
	}

	schema := ""
	switch args[1] {
	case "-rt":
		schema = "rt.normal"
		args = args[2:]
	case "-old":
		schema = "old.normal"
		args = args[2:]
	default:
		args = args[1:]
	}

	wasError := false
	for _, path := range args {
		ok := doProcess(path, schema)
		if !ok {
			wasError = true
		}
	}

	if wasError {
		fmt.Println("Press the <Enter> to continue...")
		var input string
		fmt.Scanln(&input)
	}
}
