package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagFiles  []string
	flagClean  bool
	flagFormat bool
)

var langTable = map[string]string{
	"rus": "русская",
	"eng": "английская",
	"chn": "китайская",
	"tur": "турецкая",
	"ger": "немецкая",
}

func numToText(n int) string {
	s := ""
	switch n {
	default:
		s = fmt.Sprintf("%v звуковых дорожек", n)
	case 1:
		s = fmt.Sprintf("одна звуковая дорожка")
	case 2:
		s = fmt.Sprintf("две звуковые дорожки")
	case 3:
		s = fmt.Sprintf("три звуковые дорожки")
	case 4:
		s = fmt.Sprintf("четыре звуковые дорожки")
	}
	return s
}

func doProcess(path string, schema string, checkLevel int) string {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	tn, err := tagname.NewFromString("", path, checkLevel)
	retif.Error(err, "cannot parse filename")

	tn.AddHash()

	if schema == "" {
		schema = tn.Schema()
	}

	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, "cannot convert to '"+schema+"'")

	// err = os.Rename(path, newPath)
	// retif.Error(err, "cannot rename file")

	if flagFormat {
		options := []string{}
		a, err := tn.GetAudio()
		retif.Error(err, "cannot infer audio tag")
		switch len(a) {
		case 0:
			retif.Error(true, "0 audio: that shouldn't be happened")
		case 1:
			if a[0].Language == "rus" {
				break
			}
			text, ok := langTable[a[0].Language]
			log.Errorf(!ok, "unknown audio language: %v", a[0].Language)
			if !ok {
				text = "ОШИБКА!"
			} else {
				text += " звуковая дорожка"
			}
			options = append(options, text)
		default:
			options = append(options, numToText(len(a)))
		}

		subs := tn.GetTags("stag")
		if len(subs) > 0 {
			// ### FIXME !!!
			options = append(options, "русские субтитры")
		}

		hardsub := tn.GetTags("hardsubtag")
		if len(hardsub) > 0 {
			options = append(options, "русский хардсаб")
		}

		if len(options) > 0 {
			newPath += " (" + strings.Join(options, " и ") + ")"
		}
	}

	log.Notice(schema, " > ", newPath)
	return newPath
}

// func doProcess(path string) {
// 	defer retif.Catch()
// 	log.Info("")
// 	log.Info("rename: " + path)
// 	dir, name := filepath.Split(path)
// 	ext := ""

// 	file, err := os.Open(path)
// 	retif.Error(err, "cannot open file: ", path)

// 	stat, err := file.Stat()
// 	retif.Error(err, "cannot get filestat: ", path)

// 	err = file.Close()
// 	retif.Error(err, "cannot close file: ", path)

// 	if !stat.IsDir() {
// 		ext = filepath.Ext(path)
// 	}
// 	name = strings.TrimSuffix(name, ext)
// 	name, _ = translit.Do(name)
// 	name = upper(name)
// 	err = os.Rename(path, dir+name+ext)
// 	retif.Error(err, "cannot rename file")

// 	log.Notice("result: " + dir + name + ext)
// }

func invalidRune(r rune) bool {
	if 'a' <= r && r <= 'z' ||
		'0' <= r && r <= '9' ||
		r == '_' ||
		'A' <= r && r <= 'Z' {
		return false
	}
	return true
}

func cleanLine(str string) string {
	return strings.TrimFunc(str, invalidRune)
}

func mainFunc() error {
	// if len(flagFiles) == 0 && !flagClipboard {
	// 	return cli.ErrorNotEnoughArguments()
	// }

	checkLevel := tagname.CheckNormal
	schemaName := "rt"

	if clipboard.Unsupported {
		return fmt.Errorf("%s", "clipboard unsupported on this OS")
	}
	text, err := clipboard.ReadAll()
	if err != nil {
		return err
	}
	lines := strings.Split(text, "\n")
	outLines := []string{}
	for i := range lines {
		s := strings.TrimSpace(lines[i])
		if flagClean {
			s = cleanLine(s)
		}
		if len(s) == 0 {
			continue
		}
		s = doProcess(s, schemaName, checkLevel)
		outLines = append(outLines, s)
	}
	// if flagD == "" {
	// 	flagD = "\n"
	// }
	text = strings.Join(outLines, "\n")
	// if flagFormat {
	// 	text = "Заливаются следующие мастер-копии:\n\n" + text
	// }
	clipboard.WriteAll(text)

	// for _, path := range flagFiles {
	// 	doProcess(path)
	// }
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
	cmdLine := cli.New("!PROG! the program that converts tags to RT Form before post.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help      : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		// cli.Flag("-c -clipboard : transtlit clipboard data.", &flagClipboard),
		// cli.Flag("-u            : upper case first letter.", &flagU),
		// cli.Flag("-d -delimiter : delimiter to separate multiple files. CR by default.", &flagD),
		// cli.Flag(": files to be processed", &flagFiles),
		cli.Flag("-c --clean          :  'clean' lines", &flagClean),
		cli.Flag("-m --format-message : format output result.", &flagFormat),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
