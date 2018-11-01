package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/malashin/ffinfo"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

// should be set like "//host/path/etc" even on Windows
const ageLogoPathEnv = "AGELOGOPATH"

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

var (
	flagHelp   bool
	flagStrict bool
	flagDeep   bool
	flagFiles  []string
)

type tItem struct {
	path    string
	newPath string
	age     string
	sdhd    string
	qtag    string
	msmk    string
}

func doProcess(filePath string, checkLevel int) string {
	defer retif.Catch()
	log.Info("")
	log.Info("processing: " + filePath)
	tn, err := tagname.NewFromFilename(filePath, checkLevel)
	retif.Error(err, "cannot parse filename")

	schema := tn.Schema()

	age, err := tn.GetTag("agetag")
	retif.Error(err, "cannot get 'agetag' tag")
	sdhd, err := tn.GetTag("sdhd")
	retif.Error(err, "cannot get 'sdhd' tag")
	qtag, err := tn.GetTag("qtag")
	retif.Error(err, "cannot get 'qtag' tag")
	tn.RemoveTags("agetag")
	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, "cannot convert to '"+schema+"' schema")

	hasSmokingTag := false
	if _, err := tn.GetTag("smktag"); err == nil {
		hasSmokingTag = true
	}
	hasSideBySideTag := false
	if _, err := tn.GetTag("sbstag"); err == nil {
		hasSideBySideTag = true
	}

	file, err := ffinfo.Probe(filePath)
	retif.Error(err, "ffinfo.Probe() (ffprobe)")

	sar := ""
	logoPostfix := ""
	sbsPostfix := ""
	err = fmt.Errorf("%v", "cannot find video stream")
	for i, s := range file.Streams {
		if s.CodecType != "video" {
			continue
		}
		sar = s.SampleAspectRatio
		log.Notice(fmt.Sprintf("stream: %v, sar [%v]", i, sar))
		switch sar {
		default:
			retif.Error(fmt.Errorf("inconvenient SAR [%v]", sar))
		case "64:45":
			logoPostfix = "169"
		case "16:15":
			logoPostfix = "43"
		case "0:1", "1:1", "":
			sar = "1:1"
			switch sdhd {
			default:
				retif.Error(fmt.Errorf("inconvenient set of SAR [%v] and sdhd tag %q", sar, sdhd))
			case "3d":
				logoPostfix = "3D"
				if hasSideBySideTag {
					sbsPostfix = "_SBS"
				}
			case "hd":
				logoPostfix = "HD"
			}
		}
		err = nil
		break
	}
	retif.Error(err)
	strSar := strings.Replace(sar, ":", "/", -1) // x:y -> x/y

	x := qtag[2]
	if x != 'w' && x != 's' ||
		x == 'w' && (logoPostfix != "169" && logoPostfix != "HD" && logoPostfix != "3D") ||
		x == 's' && logoPostfix != "43" {
		retif.Error(fmt.Errorf("wrong qtag %q for video %v", qtag, logoPostfix))
	}

	strVCodec := "#error#"
	strACodec := "#error#"
	switch logoPostfix {
	default:
		retif.Error(fmt.Errorf("unreachable"))
	case "HD", "3D":
		strVCodec = "-vcodec libx264 -preset slow -b:v 32000k -bf 2 -refs 4 -level 4.2"
		strACodec = "-acodec ac3 -ab 320k"
	case "43", "169":
		strVCodec = "-vcodec mpeg2video -b:v 11000k -maxrate 15000k -minrate 0 -bufsize 1835008 -rc_init_occupancy 600000 -g 12 -bf 2 -q:v 1"
		strACodec = "-acodec mp2 -ab 320k"
	}

	strSmoking := ""
	if hasSmokingTag {
		smkImg := filepath.Join(ageLogoPath, "msmoking_"+logoPostfix+sbsPostfix+".mov")
		// workaround: replace windows backslashes to use it in ffmpeg filter
		smkImg = strings.Replace(smkImg, "\\", "/", -1)
		strSmoking = "; movie=" + smkImg + ",setsar=" + strSar + "[smoking]; " +
			" anullsrc=r=48000:cl=2,atrim=end=5[silence]; [smoking][silence][v][a]concat=n=2:v=1:a=1[v][a]; [v]setsar=" + strSar + "[v]"
	}

	logo := filepath.Join(ageLogoPath, age+"_"+logoPostfix+sbsPostfix+".mov")
	// workaround: replace windows backslashes to use it in ffmpeg filter
	logo = strings.Replace(logo, "\\", "/", -1)
	ret := "ffmpeg -i \"" + filePath + "\" -filter_complex \"movie=" + logo + ",setsar=" + strSar +
		"[age]; [0:0][age]overlay=0:0:eof_action=pass[v]; [0:1]aresample=48000[a]" + strSmoking + "\" " +
		" -map [v] " + strVCodec + " -map [a] " + strACodec + " " + newPath

	return ret
}

func mainFunc() error {
	if len(flagFiles) == 0 || flagHelp {
		if !flagHelp {
			return cli.ErrorNotEnoughArguments()
		}
	}

	checkLevel := tagname.CheckNormal
	if flagStrict {
		checkLevel |= tagname.CheckStrict
	}
	if flagDeep {
		checkLevel |= tagname.CheckDeep
	}

	list := []string{}
	for _, path := range flagFiles {
		cmd := doProcess(path, checkLevel)
		list = append(list, cmd)
	}
	list = append(list, pauseCmd)

	err := misc.SliceToFile(dstFileName, list)
	if err != nil {
		err = fmt.Errorf("cannot write to %q because of %v", dstFileName, err)
	}
	return err
}

var (
	pauseCmd    = misc.PauseTerminalStr()
	dstFileName = "#run_age" + misc.BatchFileExt()
	ageLogoPath = os.Getenv(ageLogoPathEnv)
)

func main() {
	defer retif.Catch()
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
	cmdLine := cli.New("!PROG! the program that creates a script that burns agelogo over the specified files.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help   : help", func() { flagHelp = true; cmdLine.PrintHelp() }).Terminator(),
		cli.Flag("-s -strict : will raise an error when meets an unknown tag.", &flagStrict),
		cli.Flag("-d -deep   : will raise an error when a tag do not correspond to a real format.", &flagDeep),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
