package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()

	flagStrict    bool
	flagDeep      bool
	flagForce     string
	flagCheckOnly bool
	flagAddHash   bool
	flagReport    bool
	flagDontPause bool
	flagFiles     []string
)

func doProcess(path string, schema string, checkLevel int) {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	tn, err := tagname.NewFromFilename(path, checkLevel)
	retif.Error(err, "cannot parse filename")

	// if flagAddHash {
	// tn.AddHash()
	// } else {
	// tn.RemoveHash()
	// }
	if flagReport {
		tags := tn.ListTags()
		sort.Strings(tags)
		for _, v := range tags {
			list := tn.GetTags(v)
			fmt.Printf("%16v : %v\n", v, list)
		}
	}

	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, "cannot convert to '"+schema+"'")

	if !flagCheckOnly {
		err = os.Rename(path, newPath)
		retif.Error(err, "cannot rename file")
	}

	log.Notice(schema, " > ", newPath)
}

func mainFunc() error {

	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}

	switch flagForce {
	default:
		return fmt.Errorf("Unknown schema %q", flagForce)
	case "old", "rt", "":
	}

	checkLevel := tagname.CheckNormal
	if flagStrict {
		checkLevel |= tagname.CheckStrict
	}
	if flagDeep {
		checkLevel |= tagname.CheckDeep
	}

	// wasError := false
	for _, path := range flagFiles {
		doProcess(path, flagForce, checkLevel) //tagname.CheckDeepNormal) //tagname.CheckDeepStrict)
	}
	return nil
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
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 && !flagDontPause {
			misc.PauseTerminal()
		}
	}()

	// command line interface
	cmdLine := cli.New("!PROG! the program that renames tagged files.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information."),
		cli.Flag("-h --help   : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-s --strict : raise an error on an unknown tag.", &flagStrict),
		cli.Flag("-d --deep   : raise an error on a tag that does not reflect to a real format.", &flagDeep),
		cli.Flag("-f --force  : force to rename to a schema ('old' and 'rt' is supported)", &flagForce),
		cli.Flag("-c --check-only: check only (do not rename files)", &flagCheckOnly),
		cli.Flag("-r --report : print report", &flagReport),
		cli.Flag("-k          : do not wait key press on errors", &flagDontPause),
		// cli.Flag("--add-hash  : add hash to a filename", &flagAddHash),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
