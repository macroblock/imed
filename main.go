package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log       = zlog.Instance("main")
	logFilter = loglevel.Warning.OrLower()
)

var (
	argsInstall     = []string{"get"}
	argsUpdate      = []string{"get", "-u"}
	argsJustInstall = []string{"install"}
)

var (
	optPauseAlways  = false
	optPauseOnError = false

	flagVerbose bool
	flagList    bool
	flagInstall bool
	flagUpgrade bool
	flagAll     bool
	flagFiles   []string
	flagSort    bool

	flagDontDownload = false
)

var (
	packagePathList = []string{
		"github.com/macroblock/imed",
		"github.com/macroblock/imed/cmd/agelogo",
		// "github.com/macroblock/imed/cmd/tagname",
		"github.com/macroblock/imed/cmd/tnclipboardpost",
		"github.com/macroblock/imed/cmd/tndate",
		"github.com/macroblock/imed/cmd/tnrename",
		"github.com/macroblock/imed/cmd/translit",
		"github.com/malashin/shuher",
		"github.com/malashin/rtimg",
		"github.com/malashin/fflite",
	}

	packageNameList      []string
	maxPackageNameLength int
	prefixes             = map[string]bool{}
)

func init() {
	for _, s := range packagePathList {
		s := path.Base(s)
		packageNameList = append(packageNameList, s)
	}
}

func calcMaxLen(pkgs []string) {
	for _, s := range pkgs {
		maxPackageNameLength = misc.MaxInt(len(s), maxPackageNameLength)
	}
}

var lastStr string

func prHead(s string) {
	repeat := maxPackageNameLength - len(s) + 3
	ansi.Printf("%v %v ", s, strings.Repeat(".", repeat))
	misc.CPrint(misc.ColorReset, "")
}

func prProc(s string) {
	lastStr = s
	misc.CPrintUndo()
	misc.CPrint(misc.ColorYellow, s)
}

func prOk(s string) {
	misc.CPrintUndo()
	misc.CPrint(misc.ColorGreen, s+"\n")
}

func prError(err error) {
	misc.CPrintUndo()
	misc.CPrint(misc.ColorRed, lastStr+"\n")
	log.SetState(loglevel.Error.Only())
	if flagVerbose {
		fmt.Printf("%v\n", err)
	}
}

func doFindPackage(pkgName string) string {
	pkgName = "/" + pkgName
	for _, pkgPath := range packagePathList {
		if strings.HasSuffix(pkgPath, pkgName) {
			return pkgPath
		}
	}
	return ""
}

func goDownload(pkgPath string, args ...string) error {
	if flagDontDownload {
		return nil
	}
	dir := path.Dir(pkgPath)
	dir = strings.TrimSuffix(dir, "/cmd")
	if prefixes[dir] || prefixes[pkgPath] {
		return nil
	}
	args = append(args, pkgPath)
	info, err := misc.RunCommand("go", args...)
	if err != nil {
		return fmt.Errorf("%v", info)
	}
	prefixes[dir] = true
	prefixes[pkgPath] = true
	return nil
}

func goInstall(pkgPath string) error {
	info, err := misc.RunCommand("go", "install", pkgPath)
	if err != nil {
		return fmt.Errorf("%v", info)
	}
	return err
}

func process(args ...string) error {
	pkgList := flagFiles
	// if len(pkgList) == 0 {
	if flagAll {
		pkgList = packageNameList
	}
	calcMaxLen(pkgList)
	for _, pkg := range pkgList {
		err := error(nil)
		prHead(pkg)
		prProc("describe")
		if pkg = doFindPackage(pkg); pkg == "" {
			prError(fmt.Errorf("unknown package %q", pkg))
			continue
		}
		prProc("download")
		err = goDownload(pkg, args...)
		if err != nil {
			prError(err)
			continue
		}
		prProc("install")
		err = goInstall(pkg)
		if err != nil {
			prError(err)
			continue
		}
		prOk("ok")
	}
	return nil
}

func doList() error {
	for _, s := range packageNameList {
		fmt.Printf("%v\n", s)
	}
	return nil
}

func doInstall() error {
	return process("get")
}

func doUpdate() error {
	return process("get", "-u")
}

func doHelp() error {
	return fmt.Errorf("command 'help' is not yet supported")
}

func mainFunc() error {
	return fmt.Errorf("not enough arguments")
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

	// command line interface
	cmdLine := cli.New("!PROG! tools manager for internal use.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! [flags] <command> [arguments]"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h --help        : help", cmdLine.PrintHelp).Terminator(), // Why does this work ?
		cli.Flag("-v --verbose     : verbose mode", &flagVerbose),
		cli.Command("help          : for more information about command", doHelp,
			cli.Flag(": help topics", &flagFiles),
		),
		cli.Command("list          : list packages", doList,
			cli.Flag("-s --sort    : do sort.", &flagSort),
		),
		cli.Command("install       : install package (installs all packages if no arguments passed)", doInstall,
			cli.Flag("-d           : do not download (rebuild only)", &flagDontDownload),
			cli.Flag("all -a --all : ", &flagAll),
			cli.Flag(": packages to be installed", &flagFiles),
		),
		cli.Command("update        : update packages", doUpdate,
			cli.Flag("-d           : do not download (rebuild only)", &flagDontDownload),
			cli.Flag("all -a --all : ", &flagAll),
			cli.Flag(": packages to be updated", &flagFiles),
		),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
