package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	// "path/filepath"
	"strings"
	// "unicode"

	// "github.com/atotto/clipboard"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/subrip"
	"github.com/macroblock/imed/pkg/types"
	// "github.com/macroblock/imed/pkg/translit"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

type Timecode = types.Timecode

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagCheckOnly  bool
	flagFixIt      bool
	flagPoints     string
	flagMove       string
	flagScale      string
	flagInsertZero bool
	// flagInputOptions string
	// flagOutputOptions string
	// flagListOptions string
	flagBackup bool
	flagK      bool
	flagFiles  []string

	inOpts  = subrip.MildOptions()
	outOpts = subrip.StrictOptions()
	moveBy  = types.NewTimecode(0, 0, 0)
	scaleBy = 1.0

	pointA, pointB struct {
		enabled bool
		id      int
		tc      Timecode
	}
)

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

func getBackupName(path string) string {
	if !fileExists(path + ".bak") {
		return path + ".bak"
	}
	for i := 2; i < 10000; i++ {
		ret := fmt.Sprintf("%v.bak%02v", path, i)
		if !fileExists(ret) {
			return ret
		}
	}
	panic("too many backups")
}

func doProcess(path string) {
	// defer retif.Catch()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return
	}
	srt, err := subrip.Parse(strings.NewReader(string(data)))
	if err != nil {
		log.Errorf(err, "on parse error")
		return
	}
	// err = subrip.CheckOpt(srt, inOpts)
	// if err != nil {
	// // fmt.Print("sss", err.Error())
	// log.Errorf(err, "on input check")
	// return
	// }
	if flagCheckOnly {
		err = subrip.CheckOpt(srt, outOpts)
		if err != nil {
			log.Errorf(err, "on check")
			return
		}
		log.Notice("Ok.")
		return
	}

	zp := types.NewTimecode(0, 0, 0)
	scale := 1.0
	move := types.NewTimecode(0, 0, 0)
	if pointA.enabled {
		zp, scale, move, err = calcTransform(srt)
		if err != nil {
			log.Errorf(err, "on transform error")
			return
		}
	}
	move += moveBy
	scale *= scaleBy
	if zp != move || scale != 1.0 {
		subrip.Transform(&srt, zp, scale, move)
	}

	if flagInsertZero {
		if len(srt) == 0 || srt[0].In.InSeconds() > 0 {
			record := subrip.Record{
				ID:   1,
				In:   types.NewTimecode(0, 0, 0.0),
				Out:  types.NewTimecode(0, 0, 0.0),
				Text: "<i>",
			}
			srt = append([]subrip.Record{record}, srt...)
		}
	}
	if flagFixIt {
		srt = subrip.Fix(srt)
	}

	err = subrip.CheckOpt(srt, outOpts)
	if err != nil {
		log.Errorf(err, "on output check")
		return
	}
	if flagBackup {
		backup := getBackupName(path)
		err := Copy(path, backup)
		if err != nil {
			log.Errorf(err, "on copy file to backup")
			return
		}
	}
	err = ioutil.WriteFile(path, []byte(subrip.ToString(srt)), 0666)
	if err != nil {
		log.Errorf(err, "on write output file")
		return
	}
	log.Notice("Ok.")
}

func parsePoints() error {
	arr := strings.Split(flagPoints, ",")
	if len(arr) > 2 {
		return fmt.Errorf("too many points defined (maximum allowed 2)")
	}
	for i, v := range arr {
		m := "first"
		if i == 2 {
			m = "second"
		}
		x := strings.Split(v, "=")
		if len(x) != 2 {
			return fmt.Errorf("invalid format of the %v point", m)
		}
		id, err := strconv.Atoi(strings.TrimSpace(x[0]))
		if err != nil {
			return fmt.Errorf("while parsing <ID> of the %v point got: %v", m, err)
		}
		tc, err := types.ParseTimecode(strings.TrimSpace(x[1]))
		if err != nil {
			return fmt.Errorf("while parsing <timecode> of the %v point got: %v", m, err)
		}
		if pointA.enabled {
			pointB.enabled = true
			pointB.id = id
			pointB.tc = tc
			continue
		}
		pointA.enabled = true
		pointA.id = id
		pointA.tc = tc
	}
	if !pointB.enabled {
		return nil
	}
	if pointA.id == pointB.id {
		return fmt.Errorf("<ID>s of the points cannot be equal")
	}
	if pointA.id > pointB.id {
		t := pointB
		pointB = pointA
		pointA = t
	}
	if pointA.tc >= pointB.tc {
		return fmt.Errorf("<timecode> of the first point cannot be greater or equal to the second one")
	}
	return nil
}

func calcTransform(srt []subrip.Record) (Timecode, float64, Timecode, error) {
	check := 0
	if pointA.enabled {
		check++
	}
	if pointB.enabled {
		check++
	}
	if check == 0 {
		return 0, 1.0, 0, nil
	}
	var zp, moveX Timecode
	scaleX := 1.0
	for _, v := range srt {
		if v.ID == pointA.id {
			check--
			zp = v.In
			moveX = pointA.tc
		}
		if pointB.enabled && v.ID == pointB.id {
			check--
			scaleX = float64(pointB.tc-pointA.tc) / float64(v.In-zp)
		}
	}
	if check != 0 {
		return 0, 1.0, 0, fmt.Errorf("something wrong with defined points (check value %v != 0)", check)
	}
	return zp, scaleX, moveX, nil
}

func mainFunc() error {
	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}
	if flagPoints != "" {
		err := parsePoints()
		if err != nil {
			return err
		}
	}
	if flagMove != "" {
		m, err := types.ParseTimecode(flagMove)
		if err != nil {
			return err
		}
		moveBy = m
	}
	if flagScale != "" {
		s, err := strconv.ParseFloat(flagScale, 64)
		if err != nil {
			return err
		}
		scaleBy = s
	}

	for i, path := range flagFiles {
		log.Infof("%2v/%v: %v", i+1, len(flagFiles), path)
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
	cmdLine := cli.New("!PROG! does some work with subrip !!!overwrites input file(s)!!! returns exit codes 0, 1, 2, 3 on normal, wrarning, panic and error exits respectively.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h --help           : help", cmdLine.PrintHelp).Terminator(),
		cli.Flag("-c --check          : check only", &flagCheckOnly),
		cli.Flag("-x --fix            : fix it (process)", &flagFixIt),
		cli.Flag("-p --points         : recalc points (id=hh:mm:ss.ms,)[1,2]", &flagPoints),
		cli.Flag("-m --move           : set offset (move by) hh:mm:ss.ms", &flagMove),
		cli.Flag("-s --scale          : set scale ratio", &flagScale),
		cli.Flag("-z --zero           : insert empty chunk at zero position", &flagInsertZero),
		// cli.Flag("-io --input-options : input options delimited by comma", &flagInputOptions),
		// cli.Flag("-oo --output-options: output options delimited by comma", &flagOutputOptions),
		// cli.Flag("-lo --list-options  : list available options", &flagListOptions),
		cli.Flag("-b --backup         : leave backup file", &flagBackup),
		cli.Flag("-k                  : do not wait keyboard event on errors", &flagK),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
		cli.Hint(`Most common use cases:
    Recalculate timings from 25fps to 24fps:
        subrip [Path/To/Source.srt] --scale 0.96

    Move timings by 5.2 seconds closer to start:
        subrip [Path/To/Source.srt] --move -5.2

    Recalculate timings so that range of entries numbered 1 through 386 match appropriate timecodes. (copypaste from Premiere for fps25 is acceptable):
        subrip [Path/To/Source.srt] --points 1=00:01:32:04,386=01:48:18:07`),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
