package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	// "unicode"

	// "github.com/atotto/clipboard"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	//"github.com/macroblock/imed/pkg/types"
	// "github.com/macroblock/imed/pkg/translit"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagDoCheck    bool
	flagK          bool
	flagGenPause   bool
	flagFilterFile string = "./mux.cfg"
	flagNamesFile  string
	flagOutPath    string
	flagFiles      []string
)

var lineSep = "\\"
var scriptExt = ".sh"

type (
	Filter struct {
		re    *regexp.Regexp
		in    string
		typ   string
		lang  string
		title string
		grout string
		out   string
	}
)

const replsep = "\x00sep\x00"

func escapeSep(s string) string {
	return strings.Replace(s, "\\:", replsep, -1)
}

func unescapeSep(s string) string {
	return strings.Replace(s, replsep, ":", -1)
}

func readFilters(fname string) ([]Filter, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	ret := []Filter{}

	for i, line := range strings.Split(string(data), "\n") {
		l := strings.TrimSpace(line)
		if l == "" || strings.HasPrefix(l, "//") {
			continue
		}
		tuple := strings.Split(escapeSep(l), ":")
		if len(tuple) != 6 {
			return nil, fmt.Errorf("%q:%v: something wrong with ':'", fname, i+1)
		}
		clean := func(s string) string {
			return strings.TrimSpace(unescapeSep(s))
		}
		in := clean(tuple[0])
		typ := clean(tuple[1])
		lang := clean(tuple[2])
		title := clean(tuple[3])
		grout := clean(tuple[4])
		out := clean(tuple[5])
		re, err := regexp.Compile(in)
		if err != nil {
			return nil, err
		}
		if typ != "v" && typ != "a" && typ != "s" {
			return nil, fmt.Errorf("%q:%v: not valid type of stream '%v'", fname, i+1, typ)
		}

		ret = append(ret, Filter{re: re, in: in, typ: typ, lang: lang, title: title, grout: grout, out: out})
	}
	return ret, nil
}

func checkFilters(fname string, filters []Filter) (string, int, Filter, error) {
	//name := strings.TrimSuffix(fname, filepath.Ext(fname))
	name := fname
	for i, flt := range filters {
		res := flt.re.FindStringSubmatch(name)
		switch len(res) {
		default:
			return "", -1, Filter{}, fmt.Errorf("re [%q] does have more than one capture group", flt.in)
		case 0:
			continue
		case 1:
			return "", -1, Filter{}, fmt.Errorf("re [%q] does not have a capture group", flt.in)
		case 2:
			return res[1], i, flt, nil
			//return strings.TrimSuffix(name, flt.in), i, flt, nil
		}
	}
	return "", -1, Filter{}, fmt.Errorf("there are no patterns to satisfy %q", fname)
}

func genmux(name string, item []Filter) (string, error) {
	ins := []string{}
	maps := []string{}
	idx := 0
	aidx := 0
	vidx := 0
	sidx := 0
	vout := ""
	vgr := ""
	aout := ""
	agr := ""
	sout := ""
	sgr := ""
	for _, v := range item {
		if v.in == "" {
			continue
		}
		ins = append(ins, fmt.Sprintf("-i \"%v\"", v.in))

		out := fmt.Sprintf("    -map %v:%v", idx, v.typ)
		useIdx := -1
		switch v.typ {
		default:
			panic("unreachable")
		case "v":
			vout += v.out
			vgr = v.grout
			useIdx = vidx
			vidx++
		case "a":
			aout += v.out
			agr = v.grout
			useIdx = aidx
			aidx++
		case "s":
			sout += v.out
			sgr = v.grout
			useIdx = sidx
			sidx++
		}
		idx++

		metadata := fmt.Sprintf("-metadata:s:%v:%v", v.typ, useIdx)
		if len(v.lang) > 0 {
			out += fmt.Sprintf(" %v language=%v", metadata, v.lang)
		}
		if len(v.title) > 0 {
			out += fmt.Sprintf(" %v title=%v", metadata, v.title)
		}
		maps = append(maps, out)
	}
	outname := ""
	if vout != "" {
		outname += vgr + vout
	}
	if aout != "" {
		outname += agr + aout
	}
	if sout != "" {
		outname += sgr + sout
	}
	sep := " " + lineSep + "\n"
	outname = filepath.Join(flagOutPath, name+outname)
	ret := "chcp 65001\n" +
		strings.Join([]string{
			"ffmpeg",
			strings.Join(ins, sep),
			"-codec:v copy -codec:a copy -codec:s mov_text",
			strings.Join(maps, sep),
			fmt.Sprintf("\"%v.mp4\"", outname),
		}, sep)
	return ret, nil
}

func doProcess(files []string) error {

	filters, err := readFilters(flagFilterFile)
	if err != nil {
		return err
	}
	order := []string{}
	items := map[string][]Filter{}

	for _, fileX := range files {
		fname := fileX
		if runtime.GOOS == "windows" {
			fname = strings.ToLower(fname)
		}

		basename, n, filter, err := checkFilters(fname, filters)
		if err != nil {
			return err
		}

		if _, ok := items[basename]; !ok {
			items[basename] = make([]Filter, len(filters), len(filters))
			order = append(order, basename)
		}
		item := items[basename]
		if len(item) < n {
			return fmt.Errorf("unreachable")
		}
		if item[n].in != "" {
			return fmt.Errorf("%q has the same suffix [%v] as %q", fname, filter.in, item[n].in)
		}
		filter.in = fileX
		item[n] = filter
	}

	/*
		for i, v := range order {
			fmt.Printf("%2v: %v\n\n", i, items[v])
		}
	*/

	out := ""
	for _, name := range order {
		item := items[name]
		data, err := genmux(name, item)
		if err != nil {
			return err
		}
		//fmt.Printf("%2v: %v", i, data)
		out += data + "\n\n"
	}

	if flagGenPause {
		out += misc.PauseTerminalStr() + "\n"
	}

	err = ioutil.WriteFile("mux"+scriptExt, []byte(out), 0777)
	if err != nil {
		//log.Errorf(err, "on write output file")
		return err
	}

	log.Notice("Ok.")
	return nil
}

func mainFunc() error {
	if len(flagFiles) == 0 && flagNamesFile == "" {
		return cli.ErrorNotEnoughArguments()
	}

	if flagFilterFile == "" {
		return fmt.Errorf("flag -f must be sepcified")
	}

	files := flagFiles
	if flagNamesFile != "" {
		// load names file
		data, err := ioutil.ReadFile(flagNamesFile)
		if err != nil {
			return err
		}
		files = []string{}
		for _, line := range strings.Split(string(data), "\n") {
			files = append(files, strings.TrimSpace(line))
		}
	}

	return doProcess(files)
}

func main() {
	switch runtime.GOOS {
	default:
		lineSep = "\\"
		scriptExt = ".sh"
	case "windows":
		lineSep = "^"
		scriptExt = ".bat"
	}

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
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 && !flagK {
			misc.PauseTerminal()
		}
		if log.State().Intersect(loglevel.Error.Only()) != 0 {
			os.Exit(3)
		}
		if log.State().Intersect(loglevel.Warning.Only()) != 0 {
			os.Exit(1)
		}
	}()

	// command line interface
	cmdLine := cli.New("!PROG! generates mux file for windows", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h --help           : help", cmdLine.PrintHelp).Terminator(),
		//cli.Flag("-c --check          : do internal check", &flagDoCheck),
		cli.Flag("-o --out-path       : output path", &flagOutPath),
		cli.Flag("-f --config         : mux config file (default './mux.cfg')", &flagFilterFile),
		cli.Flag("-n --names          : file names file", &flagNamesFile),
		cli.Flag("-k                  : do not wait keyboard event on errors", &flagK),
		cli.Flag("-p --gen-pause          : gen pause at the end of the output file", &flagGenPause),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
