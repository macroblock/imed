package main

import (
	"fmt"
	"os"

	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

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
		log.Info("Usage:\n    imed {filename}\n")
		return
	}

	// main job
	// args := os.Args[1:]
	// for _, path := range args {
	// }
	out, err := misc.RunCommand("go", "get", "-u", "-n", "-v", "github.com/macroblock/imed")
	if err != nil {
		fmt.Println("error:\n", err)
	}
	fmt.Println("output:\n", out)
}
