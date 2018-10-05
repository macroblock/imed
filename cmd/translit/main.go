package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/macroblock/imed/pkg/translit"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
	"github.com/macroblock/imed/pkg/zlog/zlogger"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
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

	defer func() {
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			cmd := exec.Command("cmd", "/C", "pause")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	}()

	log.Debug("log initialized")
	if len(os.Args) <= 1 {
		log.Warning(true, "not enough parameters")
	}
	for _, path := range os.Args[1:] {
		doProcess(path)
	}
}
