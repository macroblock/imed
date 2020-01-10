package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/atotto/clipboard"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

var (
	_ = clipboard.Unsupported
)

var (
	optDontDownload = false
	optPauseAlways  = false
	optPauseOnError = false
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

func prError() {
	misc.CPrintUndo()
	misc.CPrint(misc.ColorRed, lastStr+"\n")
	log.SetState(loglevel.Error.Only())
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

func doDownload(pkgPath string) error {
	if optDontDownload {
		return nil
	}
	dir := path.Dir(pkgPath)
	dir = strings.TrimSuffix(dir, "/cmd")
	if prefixes[dir] || prefixes[pkgPath] {
		return nil
	}
	_, err := misc.RunCommand("go", "get", "-u", "-n", pkgPath)
	if err != nil {
		return err
	}
	prefixes[dir] = true
	prefixes[pkgPath] = true
	return nil
}

func doInstall(pkgPath string) error {
	_, err := misc.RunCommand("go", "install", pkgPath)
	return err
}

func argsLen(args []string) int {
	ret := 0
	for _, s := range args {
		if s == "-n" || s == "-p" || s == "-pe" {
			continue
		}
		ret++
	}
	return ret
}

func cmdInstall(args []string) {
	if argsLen(args) == 0 {
		args = append(args, packageNameList...)
	}
	calcMaxLen(args)
	for _, pkg := range args {
		switch pkg {
		case "-n":
			optDontDownload = true
			continue
		case "-p":
			optPauseAlways = true
			continue
		case "-pe":
			optPauseOnError = true
			continue
		}
		prHead(pkg)
		prProc("describe")
		if pkg = doFindPackage(pkg); pkg == "" {
			prError()
			continue
		}
		prProc("download")
		if doDownload(pkg) != nil {
			prError()
			continue
		}
		prProc("install")
		if doInstall(pkg) != nil {
			prError()
			continue
		}
		prOk("ok")
	}
}

func cmdList() {
	for _, s := range packageNameList {
		fmt.Printf("%v\n", s)
	}
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
		if optPauseAlways ||
			optPauseOnError && log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// process command line arguments
	if len(os.Args) <= 1 {
		log.Warning(true, "not enough parameters")
		log.Info("Usage:\n    imed (install|upgrade|list) {flag|moduleName}\n")
		log.Info("Flags:\n    -n    don't download source code (install binaries only)" +
			"\n    -p    pause at the end of process" +
			"\n    -pe   pause only if an error is occured")
		return
	}

	// main job
	args := os.Args[1:]
	mode := args[0]
	switch mode {
	default:
		log.Warning(true, fmt.Sprintf("unsupported flag %q", mode))
		log.Info("Usage:\n    imed (install|upgrade) {moduleName}\n")
		return
	case "install":
		cmdInstall(args[1:])
	case "upgrade":
		args = args[1:]
		log.Error(true, "not yet supported")
	case "list":
		cmdList()
	}
	// out, err := misc.RunCommand("go", "get", "-u", "-n", "-v", "github.com/macroblock/imed")
	// if err != nil {
	// 	fmt.Println("error:\n", err)
	// }
	// fmt.Println("output:\n", out)
	// misc.PauseTerminal()
}
