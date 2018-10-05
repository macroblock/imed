package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
	"github.com/macroblock/imed/pkg/tagname"
)

var (
	progName     string
	osWindows    bool
	filterParser *ptool.TParser
)

const (
	constConfigEnvVar = "TOOLSDIR"
	constConfigDir    = "config"
	constListExt      = ".list"
)

type tEntry struct {
	arg     func()
	cmd     func(command string, entry tEntry)
	brief   string
	usage   string
	argList string
	hint    string
}

func dummy() {}

var entries = map[string]tEntry{}

var cfgPath, cwdPath string

var (
	flagFromSchemas string
	flagToSchema    string
	flagDataSource  string
	flagDataDest    string
	flagFilter      tStringSlice
	flagSave        string
)

type tStringSlice []string

func (o *tStringSlice) String() string {
	return fmt.Sprintf("%v", *o)
}

func (o *tStringSlice) Set(val string) error {
	*o = append(*o, val)
	return nil
}

func initEntries() {
	command := ""
	entries[command] = tEntry{
		brief: fmt.Sprintf("%v is a tool for internal use.", progName),
		usage: fmt.Sprintf(
			"%v command [arguments...]"+
				"\n\n  The commands are:", progName),
		argList: fmt.Sprintf("... to be here ..."),
		hint:    fmt.Sprintf("Use '%v %v [command]' for information about a command.", progName, command),
		cmd:     cmdBasicHelp,
	}
	command = "help"
	entries[command] = tEntry{
		brief: fmt.Sprintf("Help topic."),
		usage: fmt.Sprintf("%v %v [command]", progName, command),
		cmd:   cmdHelp,
	}
	command = "rename"
	entries[command] = tEntry{
		brief: fmt.Sprintf("Rename file(s)."),
		usage: fmt.Sprintf("%v %v [flags]", progName, command),
		hint:  fmt.Sprintf("Use '%v help schemas' for information about registered schemas.", progName),
		arg: func() {
			flag.StringVar(&flagFromSchemas, "from", "", "List of [,;|]-separated names of schemas that will be used to attempt read source file(s).")
			flag.StringVar(&flagToSchema, "to", "", "Name of the schema that will be used before rename file(s).")
			flag.StringVar(&flagDataSource, "src", "", "Data source.")
			flag.StringVar(&flagDataDest, "dst", "", "Destination of the result.")
			flag.Var(&flagFilter, "filter", "Each consequtive filter flag unites with AND logic.")
		},
		cmd: cmdRename,
	}
	command = "search"
	entries[command] = tEntry{
		brief: fmt.Sprintf("Search file(s)."),
		usage: fmt.Sprintf("%v %v [flags] [filenames...]", progName, command),
		arg: func() {
			flag.StringVar(&flagFromSchemas, "from", "", "List of [,;|]-separated names of schemas that will be used to attempt read source file(s).")
			flag.StringVar(&flagDataSource, "src", "", "Data source.")
			flag.StringVar(&flagDataDest, "dst", "", "Destination of the result.")
			flag.Var(&flagFilter, "filter", "Each consequtive filter flag gets united with OR logic.")
			flag.StringVar(&flagSave, "save", "", "Destination of the result.")
		},
		cmd: cmdSearch,
	}
	command = "schemas"
	entries[command] = tEntry{
		brief: fmt.Sprintf("List of registered schemas."),
		usage: fmt.Sprintf("%v %v", progName, command),
		cmd:   cmdSchemas,
	}
	command = "dummy"
	entries[command] = tEntry{
		brief: fmt.Sprintf("--------"),
		usage: fmt.Sprintf("%v %v [command]", progName, command),
		hint:  fmt.Sprintf("Use '%v %v [command]' for information about a command.", progName, command),
		cmd:   cmdHelp,
	}

	initBasicHelpArgs()
}

func cmdSchemas(command string, entry tEntry) {
	flag.Parse()
	if len(flag.Args()) != 0 {
		fmt.Printf("Too many arguments given.\n")
		return
	}
	schemas := tagname.Schemas()
	fmt.Printf("\nregistered schemas:\n\n  %v\n", strings.Join(schemas, "\n  "))
}

func getDeGlobRestOfArgs() ([]string, error) {
	var err error
	ret := []string{}
	for _, arg := range flag.Args() {
		list := []string{arg}
		if strings.IndexRune(arg, '*') > -1 {
			list, err = filepath.Glob(arg)
			if err != nil {
				return nil, err
			}
		}
		ret = append(ret, list...)
	}
	return ret, nil
}

// IsPathExists -
func IsPathExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		return false
	}
	return true
}

// IsFile -
func IsFile(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return true
}

// IsItDir -
func IsItDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}
	return true
}

func checkPath(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	if !info.IsDir() {
		return ""
	}
	return path
}

func init() {
	if runtime.GOOS == "windows" {
		osWindows = true
	}
	progName = filepath.Base(os.Args[0])
	progName = strings.TrimSuffix(progName, filepath.Ext(progName))
	fmt.Printf("prog name: %q\n", progName)

	cfgPath = strings.TrimSpace(os.Getenv(constConfigEnvVar))
	cwdPath, _ = os.Getwd()
	cwdPath = checkPath(cwdPath)
	cfgPath, _ = filepath.Abs(cfgPath)
	if cfgPath == "" {
		cfgPath = cwdPath
	}
	cfgPath = checkPath(cfgPath)

	// init parser
	p, err := ptool.NewBuilder().FromString(filterParserSource).Entries("entry").Build()
	if err != nil {
		fmt.Println("\n[filterParser] parser error: ", err)
		panic("")
	}
	filterParser = p

	// print stat
	fmt.Printf("config: %q\n", cfgPath)
	fmt.Printf("cwd   : %q\n", cwdPath)

	fmt.Printf("numCPU      : %v\n", runtime.NumCPU())
	fmt.Printf("numGoroutine: %v\n", runtime.NumGoroutine())
	initEntries()
}

func main() {
	command := ""
	if len(os.Args) > 1 {
		command = os.Args[1]
		os.Args = os.Args[1:]
	}

	entry, ok := entries[command]
	if !ok {
		fmt.Printf("Unknown command %q. Run \"%v help\".\n", command, progName)
		os.Exit(1)
	}
	callArg(entry)
	callCmd(command, entry)
}
