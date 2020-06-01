package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/ffmpeg"
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

	flagScanOnly bool
	flagStereo   bool

	flagLI,
	flagLRA,
	flagTP,
	flagMP,
	flagPrecision string

	flagSTStatsTHBelow,
	flagSTStatsTHAbove string

	flagAttack,
	flagRelease,
	flagTries,
	flagStep string

	flagT,
	flagSS string
)

func adobeTimeToFFMPEG(s string) (ffmpeg.Time, error) {
	x := strings.Split(s, ":")
	val, err := strconv.Atoi(x[len(x)-1])
	if err != nil {
		return 0, fmt.Errorf("while converting timecode %v: %v", s, err)
	}
	x = x[:len(x)-1]
	if len(x) == 0 {
		x = []string{"0"}
	}
	str := fmt.Sprintf("%v.%v", strings.Join(x, ":"), strconv.Itoa(val*40))
	ret, err := ffmpeg.ParseTime(str)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

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

func colorizedPrintf(color misc.TTerminalColor, format string, args ...interface{}) {
	printf := ansi.Printf
	c := misc.Color(color)
	r := misc.Color(misc.ColorReset)
	if !misc.IsTerminal() {
		printf = fmt.Printf
		c = ""
		r = ""
	}
	printf(c+format+r, args...)
}

type tuples map[string]float64
type tErrorGroup struct {
	err error
}

func (o *tErrorGroup) adobeTime(flag string, val **ffmpeg.Time) bool {
	if o.err != nil {
		return false
	}
	if flag != "" {
		ret, err := ffmpeg.ParseHHMMSSFr(flag, 40)
		if err != nil {
			o.err = err
			return false
		}
		*val = &ret
		return true
	}
	return false
}

func (o *tErrorGroup) float(flag string, val *float64, keyVal map[string]float64) bool {
	if o.err != nil {
		return false
	}
	if flag != "" {
		v, ok := keyVal[strings.ToUpper(flag)]
		if ok {
			*val = v
			return true
		}
		ret, err := strconv.ParseFloat(flag, 64)
		if err != nil {
			o.err = err
			return false
		}
		*val = ret
		return true
	}
	return false
}

func (o *tErrorGroup) int(flag string, val *int, keyVal map[string]float64) bool {
	if o.err != nil {
		return false
	}
	if flag != "" {
		ret, err := strconv.ParseInt(flag, 10, 64)
		if err != nil {
			o.err = err
			return false
		}
		*val = int(ret)
		return true
	}
	return false
}

func mainFunc() error {

	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}

	loudnorm.GlobalDebug = flagVerbosity

	settings := loudnorm.GetSettings()

	settings.Behavior.ScanOnly = flagScanOnly
	settings.Behavior.ForceStereo = flagStereo

	parse := &tErrorGroup{}
	parse.adobeTime(flagSS, &settings.Edit.ClipPoint)
	parse.adobeTime(flagT, &settings.Edit.ClipDuration)

	parse.float(flagAttack, &settings.Compressor.Attack, nil)
	parse.float(flagRelease, &settings.Compressor.Release, nil)
	parse.int(flagTries, &settings.Compressor.NumTries, nil)
	parse.float(flagStep, &settings.Compressor.CorrectionStep, nil)

	parse.float(flagLI, &settings.Loudness.I, nil)
	parse.float(flagLRA, &settings.Loudness.RA, tuples{"OFF": math.NaN()})
	parse.float(flagTP, &settings.Loudness.TP, tuples{"OFF": math.NaN()})
	parse.float(flagMP, &settings.Loudness.MP, nil)
	parse.float(flagPrecision, &settings.Loudness.Precision, nil)

	parse.float(flagSTStatsTHBelow, &settings.Loudness.STStatTHBelow, nil)
	parse.float(flagSTStatsTHAbove, &settings.Loudness.STStatTHAbove, nil)

	if parse.err != nil {
		return parse.err
	}

	loudnorm.SetSettings(settings)

	errors := []string{}
	for n, filename := range flagFiles {
		fmt.Println()
		colorizedPrintf(misc.ColorBold, "==== ")
		fmt.Printf("[%v/%v] %q\n", n+1, len(flagFiles), filename)
		err := loudnorm.Process(filename)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%q: %v", filename, err.Error()))
		}
	}
	if len(errors) > 0 {
		fmt.Println()
		colorizedPrintf(misc.ColorBold, "==== ")
		return fmt.Errorf("Error(s):\n    %v", strings.Join(errors, "\n    "))
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
		// if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
		// 	misc.PauseTerminal()
		// }
		// misc.PauseTerminal()
	}()

	// command line interface
	cmdLine := cli.New("!PROG! loudness normalization tool.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help      : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-v            : verbosity", &flagVerbosity),
		cli.Flag("-scan         : do not process files (scan only)", &flagScanOnly),
		cli.Flag("-li           : targeted integrated loudness (LUFS)", &flagLI),
		cli.Flag("-lra          : max allowed loudness range (LU) or 'off' to disable LRA check", &flagLRA),
		cli.Flag("-tp           : max allowed true peaks (dBFS) or 'off' to disable TP calculation", &flagTP),
		cli.Flag("-mp           : max allowed sample peaks (dBFS)", &flagMP),
		cli.Flag("-lprec        : integrated loudness precision", &flagLI),
		cli.Flag("-a            : compressor attack time (seconds)", &flagAttack),
		cli.Flag("-r            : compressor release time (seconds)", &flagRelease),
		cli.Flag("-tries        : number of tries (default 5)", &flagTries),
		cli.Flag("-step         : compress correction step (default = 0.1)", &flagStep),
		cli.Flag("-t            : same meaning as in ffmpeg but has different format (hh:mm:ss:fr)", &flagT),
		cli.Flag("-ss           : same meaning as in ffmpeg but has different format (hh:mm:ss:fr)", &flagSS),
		cli.Flag("-stereo       : force stereo", &flagStereo),
		cli.Flag("-stthb        : ST stats below < I + this (default = -4.0)", &flagSTStatsTHBelow),
		cli.Flag("-sttha        : ST stats above >= I + this(default = +4.0)", &flagSTStatsTHAbove),
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
