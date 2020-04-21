package main

import (
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/loudnorm"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagFiles     []string
	flagVerbosity bool

	flagLI,
	flagLRA,
	flagTP string
)

func doProcess(path string) {
	log.Notice("result: ", path)
}

func doScan() error {
	// if len(flagFiles) == 0 {
	// 	return cli.ErrorNotEnoughArguments()
	// }
	// for _, path := range flagFiles {
	// 	li, err := loudnorm.Scan(path, 0)
	// 	if err != nil {
	// 		log.Errorf("FAIL %v %q", fmt.Sprint(err), path)
	// 		continue
	// 	}
	// 	log.Infof("DONE (%v) %q:", li, filepath.Base(path))
	// }
	return nil
}

func mainFunc() error {
	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}

	loudnorm.GlobalDebug = flagVerbosity

	if flagLI != "" {
		val, err := strconv.ParseFloat(flagLI, 64)
		if err != nil {
			return err
		}
		loudnorm.SetTargetLI(val)
	}

	if flagLRA != "" {
		val := math.NaN()
		if strings.ToUpper(flagLRA) != "OFF" {
			err := error(nil)
			val, err = strconv.ParseFloat(flagLRA, 64)
			if err != nil {
				return err
			}
		}
		loudnorm.SetTargetLRA(val)
	}

	if flagTP != "" {
		val := math.NaN()
		if strings.ToUpper(flagTP) != "OFF" {
			err := error(nil)
			val, err = strconv.ParseFloat(flagTP, 64)
			if err != nil {
				return err
			}
		}
		loudnorm.SetTargetTP(val)
	}

	// t := time.Now()
	err := loudnorm.Process(flagFiles[0])
	if err != nil {
		return err
	}
	// fmt.Printf("%v", time.Since(t))
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
		// if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
		// 	misc.PauseTerminal()
		// }
		misc.PauseTerminal()
	}()

	// command line interface
	cmdLine := cli.New("!PROG! loudness normalization tool.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help      : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-v            : verbosity", &flagVerbosity),
		cli.Flag("-li           : targeted integrated loudness (LUFS)", &flagLI),
		cli.Flag("-li           : targeted integrated loudness (LUFS)", &flagLI),
		cli.Flag("-lra          : max allowed loudness range (LU) or 'off' to disable LRA check", &flagLRA),
		cli.Flag("-tp           : max allowed true peaks (dBFS) or 'off' to disable TP calculation", &flagTP),
		cli.Flag(": files to be processed", &flagFiles),
		cli.Command("scan       : scan loudnes parameters", doScan,
			// cli.Flag("-l --light: light mode (whithout TP)", &flagLight),
			cli.Flag(": files to be processed", &flagFiles),
		),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
