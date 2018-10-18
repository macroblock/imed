package main

import (
	"fmt"
	"os"
	"path"
	"strings"

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
	packagePathList = []string{
		"github.com/macroblock/imed",
		"github.com/macroblock/imed/cmd/agelogo",
		"github.com/macroblock/imed/cmd/tagname",
		"github.com/macroblock/imed/cmd/tnrename",
		"github.com/macroblock/imed/cmd/translit",
		"github.com/malashin/fflite",
	}

	packageNameList      []string
	maxPackageNameLength int
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

func doFindPackage(pkgName string) string {
	pkgName = "/" + pkgName
	for _, pkgPath := range packagePathList {
		if strings.HasSuffix(pkgPath, pkgName) {
			return pkgPath
		}
	}
	return ""
}

func prHead(s string) {
	repeat := maxPackageNameLength - len(s) + 3
	ansi.Printf("%v %v ", s, strings.Repeat(".", repeat))
	misc.CPrint(misc.ColorReset, "")
}

func prProc(s string) {
	misc.CPrintUndo()
	misc.CPrint(misc.ColorYellow, s)
}

func prOk(s string) {
	misc.CPrintUndo()
	misc.CPrint(misc.ColorGreen, s+"\n")
}

func prError(s string) {
	misc.CPrintUndo()
	misc.CPrint(misc.ColorRed, s+"\n")
}

func doInstall(pkgPath string) error {
	_, err := misc.RunCommand("go", "install", pkgPath)
	return err
}

func doDownload(pkgPath string) error {
	_, err := misc.RunCommand("go", "get", "-u", pkgPath)
	return err
}

func cmdInstall(names []string) {
	if len(names) == 0 {
		names = packageNameList
	}
	calcMaxLen(names)
	for _, pkg := range names {
		prHead(pkg)
		prProc("describe")
		if pkg = doFindPackage(pkg); pkg == "" {
			prError("describe")
			continue
		}
		prProc("download")
		if doDownload(pkg) != nil {
			prError("download")
			continue
		}
		prProc("install")
		if doInstall(pkg) != nil {
			prError("install")
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
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// process command line arguments
	if len(os.Args) <= 1 {
		log.Warning(true, "not enough parameters")
		log.Info("Usage:\n    imed (install|upgrade|list) {moduleName}\n")
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
