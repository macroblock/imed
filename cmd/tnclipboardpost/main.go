package main

import (
	"fmt"
	"os"
	"path/filepath"
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

	flagFiles   []string
	flagClean   bool
	flagFormat  bool
	flagTrimTo  string
	flagUnixSep bool
	flagNoCheck bool
)

var langTable = map[string]string{
	"und": "unknown",
	"rus": "русская",
	"eng": "английская",
	"chn": "китайская",
	"tur": "турецкая",
	"ger": "немецкая",
	"fra": "французская",
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

	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	if flagTrimTo == "" {
		path = filepath.Base(path)
	} else {
		x := "/" + flagTrimTo + "/"
		i := strings.Index(path, x)
		if i >= 0 {
			path = path[i+len(x)-1:]
		}
	}

	// tn, err := tagname.NewFromString("", path, checkLevel)
	newPath := path
	tn := &tagname.TTagname{} // ### ugly hack
	if !flagNoCheck {
		err := error(nil)
		tn, err = tagname.NewFromFilename(path, checkLevel)
		retif.Error(err, "cannot parse filename")

		tn.AddHash()

		if schema == "" {
			schema = tn.Schema()
		}

		newPath, err = tn.ConvertTo(schema)
		retif.Error(err, "cannot convert to '"+schema+"'")
	} else {
		flagFormat = false
	}

	// err = os.Rename(path, newPath)
	// retif.Error(err, "cannot rename file")

	if flagUnixSep {
		newPath = filepath.ToSlash(newPath)
	}

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
		cli.Flag("-c --clean          :  'clean' lines", &flagClean),
		cli.Flag("-m --format-message : format output result.", &flagFormat),
		cli.Flag("-t --trim-path      : trim path up to the value icnlusively (removes whole path if not set).", &flagTrimTo),
		cli.Flag("-u --unix-separator : convert path separator to Unix style '/'.", &flagUnixSep),
		cli.Flag("-n --no-check       : do not do any tag name checks and/or convertions (disables flag -m)", &flagNoCheck),

		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
