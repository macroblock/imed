package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/malashin/ffinfo"

	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
	"github.com/macroblock/imed/pkg/zlog/zlogger"
)

// should be set like "//host/path/etc" even on Windows
const ageLogoPathEnv = "AGELOGOPATH"

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

type tItem struct {
	path    string
	newPath string
	age     string
	sdhd    string
	qtag    string
	msmk    string
}

func doProcess(filePath string) (string, bool) {
	defer retif.Catch()
	log.Info("")
	log.Info("processing: " + filePath)
	tn, err := tagname.NewFromFilename(filePath)
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

	file, err := ffinfo.Probe(filePath)
	retif.Error(err, "ffinfo.Probe() (ffprobe)")

	sar := ""
	logoPostfix := ""
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
		case "0:1", "1:1":
			switch sdhd {
			default:
				retif.Error(fmt.Errorf("inconvenient set of SAR [%v] and sdhd tag %q", sar, sdhd))
			case "3d":
				logoPostfix = "3D"
			case "hd":
				logoPostfix = "HD"
			}
		}
		err = nil
		break
	}
	retif.Error(err)

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
	logo := filepath.Join(ageLogoPath, age+"_"+logoPostfix+".mov")
	// workaround: replace windows backslashes to use it in ffmpeg filter
	logo = strings.Replace(logo, "\\", "/", -1)
	ret := "ffmpeg -i \"" + filePath + "\" -filter_complex \"movie=" + logo + ",setsar=" + sar +
		"[age]; [0:0][age]overlay=0:0:eof_action=pass[v]; [0:1]aresample=48000[a]" + strSmoking + "\" " +
		" -map [v] " + strVCodec + " -map [a] " + strACodec + " " + newPath

	return ret, true
}

var (
	pauseCmd    = ""
	dstFileName = ""
	ageLogoPath = ""
)

func init() {
	ageLogoPath = os.Getenv(ageLogoPathEnv)
	switch runtime.GOOS {
	case "windows":
		pauseCmd = "pause"
		dstFileName = "#run_age.bat"
	default:
		pauseCmd = "read -n1 -r -p \"Press any key to continue...\" key"
		dstFileName = "#run_age.bash"
	}
}

func main() {
	log.Add(
		zlogger.Build().
			LevelFilter(logFilter).
			Styler(zlogger.AnsiStyler).
			Done(),
		zlogger.Build().
			LevelFilter(loglevel.Info.Only().Include(loglevel.Notice.Only())).
			Format("~x\n").
			Styler(zlogger.AnsiStyler).
			Done())

	args := os.Args
	if len(args) <= 1 {
		log.Info(`
Error: not enougth arguments

Usage:
    agelogo {filename}`)
		return
	}

	args = args[1:]

	wasError := false
	list := []string{}
	for _, path := range args {
		cmd, ok := doProcess(path)
		if !ok {
			wasError = true
			continue
		}
		list = append(list, cmd)
	}
	list = append(list, pauseCmd)

	err := misc.SliceToFile(dstFileName, list)
	retif.Error(err, fmt.Sprintf("cannot write to %q", dstFileName))

	if wasError {
		fmt.Println("Press the <Enter> to continue...")
		var input string
		fmt.Scanln(&input)
	}
}
