package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

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
	flagFileList  string
	flagDoRename  bool
	flagAddHash   bool
	flagReport    bool
	flagDontPause bool
	flagSilent    bool
	flagFiles     []string

	globalTags map[string]map[string]bool
)

func doProcess(path string, schema string, checkLevel int) {
	defer retif.Catch()
	errPrefix := ""
	if flagSilent {
		errPrefix = "\n"+path+"\n"
	}
	if !flagSilent {
		log.Info("")
		log.Info("rename: " + path)
	}
	tn, err := tagname.NewFromFilename(path, checkLevel)
	if flagReport && tn != nil {
		if globalTags == nil {
			globalTags = map[string]map[string]bool{}
		}
		tags := tn.ListTags()
		for _, t := range tags {
			if _, ok := globalTags[t]; !ok {
				globalTags[t] = map[string]bool{}
			}
			dst := globalTags[t]
			list := tn.GetTags(t)
			for  _, v := range list {
				dst[v] = true
			}
			// fmt.Printf("%16v : %v\n", v, list)
		}
	}
	retif.Error(err, errPrefix + "cannot parse filename")


	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, errPrefix + "cannot convert to '"+schema+"'")

	if flagDoRename {
		err = os.Rename(path, newPath)
		retif.Error(err, errPrefix + "cannot rename file")
	}

	if !flagSilent {
		log.Notice(schema, " > ", newPath)
	}
}

func printTags() {
	if globalTags == nil {
		return
	}
	taglist := []string{}
	for tag, _ := range globalTags {
		taglist = append(taglist, tag)
	}
	sort.Strings(taglist)
	for _, tag := range taglist {
		valmap, ok := globalTags[tag]
		if !ok {
			panic("unreachable")
		}
		vallist := []string{}
		for key, _ := range valmap {
			vallist = append(vallist, key)
		}
		sort.Strings(vallist)
		fmt.Printf("%16v : %q\n", tag, vallist)
	}
}

func mainFunc() error {

	if len(flagFiles) == 0 && flagFileList == "" {
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

	if flagFileList != "" {
		file, err := os.Open(flagFileList)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			doProcess(line, flagForce, checkLevel)
		}
	}

	if flagReport {
		printTags()
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
		cli.Flag("-n --do-rename: do rename files)", &flagDoRename),
		cli.Flag("-r --report : print cumulative report", &flagReport),
		cli.Flag("-k          : do not wait key press on errors", &flagDontPause),
		cli.Flag("-q --quiet  : quiet mode (display errors only)", &flagSilent),
		cli.Flag("-l --filelist   : specifies the file that contains list of files to process", &flagFileList),
		// cli.Flag("--add-hash  : add hash to a filename", &flagAddHash),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
