package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/loudnorm"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagFiles []string
	flagLight bool

	flagLI,
	flagLRA,
	flagTP string
)

func doProcess(path string) {
	log.Notice("result: ", path)
}

func doScan() error {
	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}
	for _, path := range flagFiles {
		var I, LRA, Thresh float64
		switch flagLight {
		default:
			opts, err := loudnorm.Scan(path, 0)
			if err != nil {
				log.Errorf("FAIL %v %q", fmt.Sprint(err), path)
				continue
			}
			I = opts.InputI
			LRA = opts.InputLRA
			Thresh = opts.InputThresh
		case true:
			opts, err := loudnorm.ScanLight(path, 0)
			if err != nil {
				log.Errorf("FAIL %v %q", fmt.Sprint(err), path)
				continue
			}
			I = opts.I
			LRA = opts.LRA
			Thresh = opts.Thresh
		}
		log.Infof("DONE I: %v LRA: %v Thresh: %v %q:", I, LRA, Thresh, filepath.Base(path))
	}
	return nil
}

func mainFunc() error {
	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}

	if flagLI != "" {
		val, err := strconv.ParseFloat(flagLI, 64)
		if err != nil {
			return err
		}
		loudnorm.SetTargetLI(val)
	}

	if flagLRA != "" {
		val, err := strconv.ParseFloat(flagLRA, 64)
		if err != nil {
			return err
		}
		loudnorm.SetTargetLRA(val)
	}

	if flagTP != "" {
		val, err := strconv.ParseFloat(flagTP, 64)
		if err != nil {
			return err
		}
		loudnorm.SetTargetTP(val)
	}

	t := time.Now()
	// fmt.Println("scanning...")
	// opts, err := loudnorm.ScanLight(flagFiles[0], 0)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("delta time ", time.Since(t))

	// // fmt.Println("opts: ", opts)
	// fmt.Printf("(%v, LRA: %v, Thresh: %v, TP: %v) time: %v\n",
	// 	opts.InputI, opts.InputLRA, opts.InputThresh, opts.InputTP, time.Since(t))

	// fmt.Println("processing...")
	// err = loudnorm.Process(flagFiles[0], flagFiles[0]+".flac", 0,
	// 	opts.InputI, opts.InputLRA, opts.InputThresh, opts.InputTP)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("(%v, LRA: %v, Thresh: %v, TP: %v) time: %v\n",
	// 	opts.InputI, opts.InputLRA, opts.InputThresh, opts.InputTP, time.Since(t))

	err := loudnorm.Process(flagFiles[0])
	if err != nil {
		return err
	}
	fmt.Printf("%v", time.Since(t))
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
	cmdLine := cli.New("!PROG! the program that translit text in filenames or|and clipboard.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help      : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-li           : light mode (whithout TP)", &flagLI),
		cli.Flag("-lra           : light mode (whithout TP)", &flagLRA),
		cli.Flag("-tp           : light mode (whithout TP)", &flagTP),
		cli.Flag(": files to be processed", &flagFiles),
		cli.Command("scan       : scan loudnes parameters", doScan,
			cli.Flag("-l --light: light mode (whithout TP)", &flagLight),
			cli.Flag(": files to be processed", &flagFiles),
		),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
