package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/translit"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagFiles     []string
	flagClipboard bool
	flagU         bool
	flagD         string
	flagN         string
	flagRA        string
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
	name = upper(name)
	err = os.Rename(path, dir+name+ext)
	retif.Error(err, "cannot rename file")

	log.Notice("result: " + dir + name + ext)
}

func upper(str string) string {
	if !flagU {
		return str
	}
	s := ""
	for i, r := range str {
		if i == 0 {
			r = unicode.ToUpper(r)
		}
		s += string(r)
	}
	return s
}

func mainFunc() error {
	if len(flagFiles) == 0 && !flagClipboard {
		return cli.ErrorNotEnoughArguments()
	}

	if flagClipboard {
		if clipboard.Unsupported {
			return fmt.Errorf("%s", "clipboard unsupported on this OS")
		}
		origText, err := clipboard.ReadAll()
		if err != nil {
			return err
		}
		origText = strings.TrimSpace(origText)

		if flagRA != "" {
			lines := strings.Split(origText, flagRA)
			origText = lines[0]
		}

		lines := strings.Split(origText, "\n")
		lastNonEmpty := -1
		for i := range lines {
			s := lines[i]
			s, _ = translit.Do(s)
			s = strings.Trim(s, "_")
			lines[i] = upper(s)
			if lines[i] != "" {
				lastNonEmpty = i
			}
		}
		lines = lines[:lastNonEmpty+1]

		if flagD == "" {
			flagD = "\n"
		}
		text := strings.Join(lines, flagD)
		text = strings.TrimSpace(text)

		if flagN != "" && text != "" && origText != "" {
			filename := flagN
			// fmt.Printf("%q\n", flagN)
			// fmt.Printf("%q\n", origText)
			// fmt.Printf("%q\n", text)
			filename = strings.ReplaceAll(filename, "${translit}", text)
			filename = strings.ReplaceAll(filename, "${orig}", origText)
			f, err := os.Create(filename)
			if err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
		}

		clipboard.WriteAll(text)
	}

	for _, path := range flagFiles {
		doProcess(path)
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
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// command line interface
	cmdLine := cli.New("!PROG! the program that translit text in filenames or|and clipboard.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help      : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-c -clipboard : transtlit clipboard data.", &flagClipboard),
		cli.Flag("-u            : upper case first letter.", &flagU),
		cli.Flag("-d -delimiter : delimiter to separate multiple files. CR by default.", &flagD),
		cli.Flag("-n            : template to save the name through ${orig}, ${translit} (does not work with files) ", &flagN),
		cli.Flag("-ra           : remove after TEXT", &flagRA),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
