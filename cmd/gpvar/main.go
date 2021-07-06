package main

import (
	// "bufio"
	// "io/ioutil"
	"fmt"
	"os"
	// "sort"
	"path/filepath"
	"strings"

	"github.com/malashin/ffinfo"

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

	// flagStrict     bool
	// flagDeep       bool
	// flagForce      string
	// flagScriptFile string
	// flagFileList   string
	// flagDoRename   bool
	// flagAddHash    bool
	// flagReport     bool
	flagDontPause bool
	// flagSilent     bool
	flagGlobFilter string = "*"
	flagFiles      []string

	// globalTags map[string]map[string]bool
)

func doProcess(filePath string, deepCheck bool) (string, error) {
	fileName := filepath.Base(filePath)
	path := strings.TrimSuffix(filePath, fileName)
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	parts := strings.Split(name, "__")
	if len(parts) != 2 {
		return "", fmt.Errorf("inappropriate number of '__'")
	}
	flags := parts[1]

	parts = strings.Split(flags, "_")
	if len(parts) == 0 {
		return "", fmt.Errorf("has no 'flags'")
	}
	size := parts[len(parts)-1]

	if deepCheck {
		info, err := ffinfo.Probe(filePath)
		if err != nil {
			return "", err
		}
		sizeTag, err := tagname.GatherSizeTag(info)
		if err != nil {
			return "", err
		}
		if size != sizeTag {
			return "", fmt.Errorf("flag [%v] != real size [%v]", size, sizeTag)
		}
	}

	flags2 := strings.TrimSuffix(flags, "_"+size)
	outName := ""
	switch {
	default:
		return "", fmt.Errorf("unsupported flag set")
	case flags == "logo_600x600":
		outName = "haslogo"
	case flags == "logo_1800x1000":
		outName = "hastitle_logo"
	case flags2 == "background":
		outName = "iconic_background"
		if !in(size, "1000x1500", "3840x2160") {
			return "", fmt.Errorf("%v is an invalid size for %q", size, flags2)
		}
	case flags2 == "poster":
		outName = "iconic_poster"
		if !in(size, "600x600", "600x800", "800x600", "1000x1500", "3840x2160") {
			return "", fmt.Errorf("%v is an invalid size for %q", size, flags2)
		}
	}

	outName = "g_" + outName + "_" + size
	outPath := filepath.Join(path, outName+ext)

	return outPath, nil
}

func in(item string, vals ...string) bool {
	for _, s := range vals {
		if item == s {
			return true
		}
	}
	return false
}

func makeGlobFilter(patterns string) (func(string) (bool, error), error) {
	patList := strings.Split(patterns, ";")
	// fake run through all patterns to check on errors before real job
	for _, pat := range patList {
		_, err := filepath.Match(pat, "anything")
		if err != nil {
			return nil, err
		}
	}
	return func(s string) (bool, error) {
		for _, pat := range patList {
			ok, err := filepath.Match(pat, s)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}, nil
}

func mainFunc() error {
	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}

	filter, err := makeGlobFilter(flagGlobFilter)
	if err != nil {
		return err
	}

	for _, path := range flagFiles {
		ok, err := filter(filepath.Base(path))
		log.Error(err, path)
		if err != nil {
			panic("unreachable")
		}
		if !ok {
			continue
		}
		outPath, err := doProcess(path, true)
		log.Error(err, path)
		if err != nil {
			continue
		}
		err = os.Rename(path, outPath)
		log.Error(err)
		if err != nil {
			continue
		}
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
	cmdLine := cli.New("!PROG! that checks some sort of things for some reason", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information."),
		cli.Flag("-h --help   : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		// cli.Flag("-s --strict : raise an error on an unknown tag.", &flagStrict),
		// cli.Flag("-d --deep   : raise an error on a tag that does not reflect to a real format.", &flagDeep),
		// cli.Flag("-f --force  : force to rename to a schema ('old' and 'rt' is supported)", &flagForce),
		// cli.Flag("-n --do-rename: do rename files)", &flagDoRename),
		// cli.Flag("-r --report : print cumulative report", &flagReport),
		cli.Flag("-k          : do not wait key press on errors", &flagDontPause),
		// cli.Flag("-q --quiet  : quiet mode (display errors only)", &flagSilent),
		// cli.Flag("-t --script : a script file path to run", &flagScriptFile),
		// cli.Flag("-l --filelist   : specifies a file that contains list of files to process", &flagFileList),
		cli.Flag("-f --filter   : glob filters separated by ';' (default: '*')", &flagGlobFilter),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
