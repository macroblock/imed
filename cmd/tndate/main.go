package main

import (
	"fmt"
	"os"
	"time"

	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

func doProcess(path string, schema string, isDeepCheck bool) {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	tn, err := tagname.NewFromFilename(path, isDeepCheck)
	retif.Error(err, "cannot parse filename")

	tag, err := tn.GetTag("type")
	if tag != "film" && tag != "trailer" {
		log.Info("skip")
		return
	}

	t := time.Now()
	y := t.Year() % 100
	tag = fmt.Sprintf("d%02v%02v%02v%02v%02v", y, int(t.Month()), t.Day(), t.Hour(), t.Minute())
	tn.SetTag("datetag", tag)

	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, "cannot convert to '"+schema+"'")

	err = os.Rename(path, newPath)
	retif.Error(err, "cannot rename file")

	log.Notice(schema, " > ", newPath)
}

func main() {
	// setup log
	newLogger := misc.NewSimpleLogger
	if misc.IsTerminal() {
		newLogger = misc.NewAnsiLogger
	}
	log.Add(
		newLogger(loglevel.Warning.OrLower(), ""),
		newLogger(loglevel.Info.Only().Include(loglevel.Notice.Only()), "~x\n"),
	)

	defer func() {
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// process command line arguments
	if len(os.Args) <= 1 {
		log.Warning(true, "not enough parameters")
		log.Info("Usage:\n    tnrename [-rt|-old] {filename}\n")
		return
	}

	// main job
	args := os.Args[1:]
	schema := ""
	switch args[0] {
	case "-rt":
		schema = "rt"
		args = args[1:]
	case "-old":
		schema = "old"
		args = args[1:]
	}

	// wasError := false
	for _, path := range args {
		doProcess(path, schema, false) //tagname.CheckDeepStrict)
	}
}
