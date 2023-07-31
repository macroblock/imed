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
	// flagHelp   bool
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

func in(what string, where ...string) bool {
	for _, s := range where {
		if what == s {
			return true
		}
	}
	return false
}

func doProcess(filePath string, isDeepCheck bool) string {
	defer retif.Catch()
	log.Info(" ")
	log.Info("processing: " + filePath)
	tn, err := tagname.NewFromFilename(filePath, isDeepCheck)
	retif.Error(err, "cannot parse filename")

	// schema := "" //tn.Schema()

	/*
		_, err = tn.GetTag("alreadyagedtag")
		doAge := true
		if err != nil {
			doAge = false
		}

		age := ""
		if !doAge {
			age, err = tn.GetTag("agetag")
			retif.Error(err, "cannot get 'agetag' tag")
		}
	*/
	audTag := ""
	audTag, err = tn.GetTag("atag")
	retif.Error(err, "cannot get 'atag' tag")

	subTag := ""
	subTag, err = tn.GetTag("stag")

	audsubPostfix := audTag
	if subTag != "" {
		audsubPostfix += "_" + subTag
	}

	sdhd, err := tn.GetTag("sdhd")
	retif.Error(err, "cannot get 'sdhd' tag")

	// qtag, err := tn.GetTag("qtag")

	hasSmokingTag := false
	if _, err := tn.GetTag("smktag"); err == nil {
		hasSmokingTag = true
	}
	hasAlcoholTag := false
	if _, err := tn.GetTag("alcotag"); err == nil {
		hasAlcoholTag = true
	}
	hasSideBySideTag := false
	if _, err := tn.GetTag("sbstag"); err == nil {
		hasSideBySideTag = true
	}

	retif.Error(!hasSmokingTag && !hasAlcoholTag, "nothing to do")

	// tn.RemoveTags("agetag")
	// tn.RemoveTags("alreadyagedtag")
	tn.RemoveTags("smktag")
	tn.RemoveTags("alcotag")

	newPath, err := tn.ConvertTo("")
	retif.Error(err, "cannot convert to '"+tn.Schema()+"' schema")

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
			retif.Error(fmt.Errorf("unsupported SAR [%v]", sar))
		case "64:45":
			logoPostfix = "SD169"
		case "16:15":
			logoPostfix = "SD43"
		case "0:1", "1:1", "":
			sar = "1:1"
			switch sdhd {
			default:
				retif.Error(fmt.Errorf("SAR [%v] does not match sdhd tag %q", sar, sdhd))
			case "3d":
				logoPostfix = "3D"
				if hasSideBySideTag {
					logoPostfix += "SBS"
				}
			case "hd":
				logoPostfix = "HD"
			case "4k":
				logoPostfix = "4K"
			}
		}
		err = nil
		break
	}
	retif.Error(err)

	ret := ""
	redir := ">"
	if hasAlcoholTag {
		path := filepath.Join(ageLogoPath, "alcohol",
			"alcohol_"+logoPostfix+"_"+audsubPostfix+".mp4")
		// workaround: replace windows backslashes to use it in ffmpeg filter
		path = strings.Replace(path, "\\", "/", -1)
		ret += fmt.Sprintf("echo file %v %v #fflist.txt\n", path, redir)
		redir = ">>"
	}

	if hasSmokingTag {
		path := filepath.Join(ageLogoPath, "smk",
			"msmoking_"+logoPostfix+"_"+audsubPostfix+".mp4")
		// workaround: replace windows backslashes to use it in ffmpeg filter
		path = strings.Replace(path, "\\", "/", -1)
		ret += fmt.Sprintf("echo file %v %v #fflist.txt\n", path, redir)
		redir = ">>"
	}
	ret += fmt.Sprintf("echo file %v %v #fflist.txt\n", filePath, redir)

	exportMetaStr := fmt.Sprintf("movmeta -i %v -export #meta", filePath)
	processStr := fmt.Sprintf("ffmpeg -f concat -safe 0 -i #fflist.txt -map 0:v? -map 0:a? -map 0:s? -c copy -codec:s mov_text %v", newPath)
	importMetaStr := fmt.Sprintf("movmeta -i %v -merge #meta -write", newPath)

	ret += fmt.Sprintf("%v && ^\n%v && ^\n%v\n", exportMetaStr, processStr, importMetaStr)
	ret += "\n"

	return ret
}

/*
func doProcess(filePath string, isDeepCheck bool) string {
	defer retif.Catch()
	log.Info(" ")
	log.Info("processing: " + filePath)
	tn, err := tagname.NewFromFilename(filePath, isDeepCheck)
	retif.Error(err, "cannot parse filename")

	// schema := "" //tn.Schema()

	_, err = tn.GetTag("alreadyagedtag")
	doAge := true
	if err != nil {
		doAge = false
	}

	age := ""
	if !doAge {
		age, err = tn.GetTag("agetag")
		retif.Error(err, "cannot get 'agetag' tag")
	}

	sdhd, err := tn.GetTag("sdhd")
	retif.Error(err, "cannot get 'sdhd' tag")

	qtag, err := tn.GetTag("qtag")

	hasSmokingTag := false
	if _, err := tn.GetTag("smktag"); err == nil {
		hasSmokingTag = true
	}
	hasSideBySideTag := false
	if _, err := tn.GetTag("sbstag"); err == nil {
		hasSideBySideTag = true
	}

	retif.Error(!hasSmokingTag, "nothing to do")

	// tn.RemoveTags("agetag")
	// tn.RemoveTags("alreadyagedtag")
	tn.RemoveTags("smktag")

	newPath, err := tn.ConvertTo("")
	retif.Error(err, "cannot convert to '"+tn.Schema()+"' schema")

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

	x := "-"
	if qtag != "" {
		x = string(qtag[2])
	}
	if !(x == "-" ||
		x == "s" && in(logoPostfix, "43") ||
		x == "w" && in(logoPostfix, "169", "HD", "3D")) {
		retif.Error(fmt.Errorf("wrong qtag %q for %v video", qtag, logoPostfix))
	}

	strVCodec := "#error#"
	strACodec := "#error#"
	switch logoPostfix {
	default:
		retif.Error(fmt.Errorf("unreachable"))
	case "HD", "3D":
		strVCodec = "-vcodec libx264 -preset medium -crf 16 -pix_fmt yuv420p -g 0"
		strACodec = "-acodec ac3 -ab 320k"
	case "43", "169":
		strVCodec = "-vcodec mpeg2video -b:v 11000k -maxrate 15000k -minrate 0 -bufsize 1835008 -rc_init_occupancy 600000 -g 12 -bf 2 -q:v 1 -pix_fmt yuv420p"
		strACodec = "-acodec mp2 -ab 320k"
	}

	strSmoking := ""
	if hasSmokingTag {
		smkImg := filepath.Join(ageLogoPath, "msmoking_"+logoPostfix+sbsPostfix+".mov")
		// workaround: replace windows backslashes to use it in ffmpeg filter
		smkImg = strings.Replace(smkImg, "\\", "/", -1)
		strSmoking = "; movie=" + smkImg + ",setsar=" + strSar + "[smoking]; " +
			" anullsrc=r=48000:cl=mono,atrim=end=5,pan=2c|c0=c0|c1=c0[silence]; [smoking][silence][v][a]concat=n=2:v=1:a=1[v][a]; [v]setsar=" + strSar + "[v]"
	}

	logo := filepath.Join(ageLogoPath, age+"_"+logoPostfix+sbsPostfix+".mov")
	// workaround: replace windows backslashes to use it in ffmpeg filter
	logo = strings.Replace(logo, "\\", "/", -1)
	// ret := "fflite -i \"" + filePath + "\" -filter_complex \"movie=" + logo + ",setsar=" + strSar +
	// "[age]; [0:0][age]overlay=0:0:eof_action=pass[v]; [0:1]aresample=48000[a]" + strSmoking + "\" " +
	// " -map [v] " + strVCodec + " -map [a] " + strACodec + " " + newPath

	// strAgeLogo := "; movie=" + logo + ",setsar=" + strSar +
	// "[age]; [v][age]overlay=0:0:eof_action=pass[v]; [a]aresample=48000[a]"

	strAgeLogo := ""
	ret := "fflite -i \"" + filePath + "\" -filter_complex \"[0:0]null[v]; [0:1]anull[a]" +
		strAgeLogo + strSmoking + "\" " +
		" -map [v] " + strVCodec + " -map [a] " + strACodec + " " + newPath

	return ret
}
*/

func mainFunc() error {
	if len(flagFiles) == 0 {
		return cli.ErrorNotEnoughArguments()
	}

	// checkLevel := tagname.CheckNormal
	// if flagStrict {
	// checkLevel |= tagname.CheckStrict
	// }
	// if flagDeep {
	// checkLevel |= tagname.CheckDeep
	// }

	list := []string{}
	for _, path := range flagFiles {
		cmd := doProcess(path, flagDeep)
		list = append(list, cmd)
	}
	list = append(list, pauseCmd)

	err := misc.SliceToFile(dstFileName, 0775, list)
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
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help   : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-s -strict : raise an error on an unknown tag.", &flagStrict),
		cli.Flag("-d -deep   : raise an error on a tag that does not correspond to a real format.", &flagDeep),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
