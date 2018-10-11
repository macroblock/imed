package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/translit"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()
)

func doProcess(path string) {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	dir, name := filepath.Split(path)
	ext := ""

	file, err := os.Open(path)
	retif.Error(err, "cannot open file: ", path)

	stat, err := file.Stat()
	retif.Error(err, "cannot get filestat: ", path)

	err = file.Close()
	retif.Error(err, "cannot close file: ", path)

	if !stat.IsDir() {
		ext = filepath.Ext(path)
	}
	name = strings.TrimSuffix(name, ext)
	name, _ = translit.Do(name)
	err = os.Rename(path, dir+name+ext)
	retif.Error(err, "cannot rename file")

	log.Notice("result: " + dir + name + ext)
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
		log.Info("Usage:\n    translit {filename}\n")
		return
	}

	// main job
	args := os.Args[1:]
	for _, path := range args {
		doProcess(path)
	}
}
