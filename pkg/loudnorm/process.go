package loudnorm

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/macroblock/imed/pkg/ffmpeg"
	"github.com/macroblock/imed/pkg/misc"
)

// GlobalDebug -
var GlobalDebug = false

var (
	ac3Params2 = []string{
		"-c:a", "ac3",
		// "-b:a", "256k", // dvd quality
		"-b:a", "320k", // bluray quality
		"-ac:a", "2",
	}
	ac3Params6 = []string{
		"-c:a", "ac3",
		// "-b:a", "448k", // dvd quality
		"-b:a", "640k", // bluray quality
		"-ac:a", "6",
	}
	mp2Params2 = []string{
		"-c:a", "mp2",
		"-b:a", "320k", // !!!! FIXME: is it correct?
		"-ac:a", "2",
	}
	mp2Params6 = []string{
		"-c:a", "mp2",
		"-b:a", "640k", // !!!! FIXME: is it correct?
		"-ac:a", "6",
	}
	alacParams = []string{
		"-c:a", "alac",
		"-compression_level", "0",
		"-sample_fmt", "s16p",
	}
)

// ScanAudio -
func ScanAudio(fi *TFileInfo) error {
	streams := []*TStreamInfo{}
	for _, stream := range fi.Streams {
		if stream.Type != "audio" {
			continue
		}
		err := error(nil)
		stream.AudioParams, err = inferAudioParams(stream)
		if err != nil {
			return err
		}
		if ValidLoudness(stream.LoudnessInfo) {
			if GlobalDebug {
				fmt.Printf("stream %v has valid loudness (%v)\n", stream.Index, stream.LoudnessInfo)
			}
			stream.validLoudness = true
			continue
		}
		streams = append(streams, stream)
	}
	err := Scan(streams)
	if err != nil {
		return err
	}
	for _, stream := range streams {
		li := stream.LoudnessInfo
		filename := stream.Parent.Filename
		index := stream.Index
		if GlobalDebug {
			fmt.Printf("ebur128 %v:%v:\n  input: I: %v, LRA: %v, TP: %v, TH: %v, MP: %v\n",
				filepath.Base(filename), index,
				li.I, li.RA, li.TP, li.TH, li.MP,
			)
		}
	}
	return nil
}

func renderParameters(fi *TFileInfo) error {
	streams := []*TStreamInfo{}
	for _, stream := range fi.Streams {
		if stream.Type != "audio" {
			continue
		}
		// if stream.done || stream.validLoudness {
		// 	if GlobalDebug {
		// 		fmt.Printf("stream %v has valid loudness (%v)\n", stream.Index, stream.LoudnessInfo)
		// 	}
		// 	continue
		// }
		streams = append(streams, stream)
	}
	err := RenderParameters(streams)
	if err != nil {
		return err
	}
	return nil
}

// CheckIfReadyToCompile -
func CheckIfReadyToCompile(fi *TFileInfo) error {
	for _, stream := range fi.Streams {
		if !(stream.done || stream.validLoudness) {
			filename := fi.Filename
			index := stream.Index
			return fmt.Errorf("%v:%v is not ready to compile (%v)", filename, index, stream.LoudnessInfo)
		}
	}
	return nil
}

// ProcessTo -
func ProcessTo(fi *TFileInfo) (canBeFixed bool, err error) {

	params := []string{"-y", "-hide_banner"}
	params = append(params, GetSettings().getGlobalFlags()...)
	params = append(params, "-i", fi.Filename)

	filters := []string{}
	outputs := []string{}
	ffmpeg.UniqueReset()

	time := ffmpeg.FloatToTime(fi.Duration)
	combParser := ffmpeg.NewCombineParser(
		ffmpeg.NewAudioProgressParser(time, nil),
	)

	params = append(params,
		"-map_metadata", "-1",
		"-map_chapters", "-1",
		"-id3v2_version", "3",
		"-write_id3v1", "1",
	)
	videoIndex := -1
	audioIndex := -1
	subtitleIndex := -1
	isFirstVideo := true
	isFirstAudio := true
	isFirstSubtitle := true
	// metadata := []string{}
	for _, stream := range fi.Streams {
		switch stream.Type {
		case "video":
			videoIndex++
			if isFirstVideo {
			}
			outputs = append(outputs,
				"-map", "0:"+strconv.Itoa(videoIndex),
				"-c:v", "copy",
			)
		case "subtitle":
			subtitleIndex++
			def := "none"
			_ = def
			if isFirstSubtitle {
				isFirstSubtitle = false
				def = "default"
			}
			outputs = append(outputs,
				"-map", "0:"+strconv.Itoa(stream.Index),
				"-c:s", "mov_text",
				"-metadata:s:s:"+strconv.Itoa(subtitleIndex), "language="+stream.Lang,
				// "-disposition:s:"+strconv.Itoa(subtitleIndex), def,
			)
		case "audio":
			audioIndex++

			def := "none"
			if isFirstAudio {
				isFirstAudio = false
				def = "default"
			}

			filters = appendPattern(filters, stream, combParser,
				"[0:~idx~]~header~,~compressor~,asplit=3[s~u0~][s~u1~][o~idx~];"+
					"[s~u0~]~astats~,anullsink;"+
					"[s~u1~]~ebur~,anullsink")
			outputs = appendPattern(outputs, stream, nil,
				"-map", "[o~idx~]")
			outputs = append(outputs, stream.AudioParams...)
			outputs = append(outputs,
				"-metadata:s:a:"+strconv.Itoa(audioIndex), "language="+stream.Lang,
				"-disposition:s:a:"+strconv.Itoa(audioIndex), def,
			)
		}
		// printStreamParams(stream)
	}
	params = append(params, "-filter_complex")
	params = append(params, strings.Join(filters, ";"))
	params = append(params, outputs...)

	// TODO!!!: something with metadata
	// params = append(params, "-metadata")
	// params = append(params, "comment="+PackTargetLoudnessInfo(fi))

	params = append(params, generateOutputName(fi))
	if GlobalDebug {
		fmt.Println("### params: ", params)
	}
	err = ffmpeg.Run(combParser, params...)
	if err != nil {
		return false, err
	}

	canBeFixed = true
	errStrs := []string{}
	for i, stream := range fi.Streams {
		if stream.Type != "audio" {
			continue
		}
		stream.LoudnessInfo, stream.MiscInfo = initInfo(stream.eburInfo, stream.astatsInfo)
		// stream.LoudnessInfo.CR = stream.TargetLI.CR

		if GlobalDebug {
			fmt.Println("##### stream:", i,
				"\n  ebur >", stream.eburInfo,
				"\n  vol  >", stream.volumeInfo,
				"\n  stats>", stream.astatsInfo)
		}
		// if !LoudnessIsEqual(stream.LoudnessInfo, stream.TargetLI) {
		// 	errStrs = append(errStrs, fmt.Sprintf("stream #%v: actual loudness info is not equal to the planned one"+
		// 		"\n    planned: %v"+
		// 		"\n    actual : %v", i, stream.TargetLI, stream.LoudnessInfo))
		// }
		if !ValidLoudness(stream.LoudnessInfo) {

			errStrs = append(errStrs, fmt.Sprintf("stream #%v: invalid loudness"+
				"\n    planned: %v"+
				"\n    actual : %v", i, stream.TargetLI, stream.LoudnessInfo))
			stream.TargetLI = stream.LoudnessInfo
			printStreamParams(stream, true)

			stream.done = FixLoudnessPostAmp(stream.TargetLI, stream.CompParams)
			if GlobalDebug && stream.done {
				fmt.Println("##### fixed stream:", i,
					"\n  li   >", stream.TargetLI,
					"\n  comp >", stream.CompParams)
			}
			canBeFixed = canBeFixed && stream.done
		} else {
			printStreamParams(stream, true)
		}
	}
	if len(errStrs) != 0 {
		// err := os.Remove(generateOutputName(fi))
		if err != nil {
			fmt.Printf("!!! error while removing file %v: %v", generateOutputName(fi), err)
		}
		return canBeFixed, fmt.Errorf("%v", strings.Join(errStrs, "\n"))
	}

	return true, nil
}

func formatSettings() string {
	const pad = "####"
	s := fmt.Sprintf("%+v", settings)
	s = pad + s[1:len(s)-1]
	s = strings.Replace(s, "} ", "\n"+pad, -1)
	s = strings.Replace(s, "}", "", -1)
	s = strings.Replace(s, "{", "\n"+pad+pad, -1)
	s = strings.Replace(s, " ", "\n"+pad+pad, -1)
	s = strings.Replace(s, "#", " ", -1)
	x := strings.Split(s, "\n")
	for i := range x {
		x[i] = strings.Replace(x[i], ":", ": ", 1)
	}
	s = strings.Join(x, "\n")
	return s
}

func printOnError() {
	colorizedPrintf(misc.ColorRed, "error. \n")
}

func printOnOk() {
	colorizedPrintf(misc.ColorGreen, "Ok. \n")
}

// Process -
func Process(filename string) error {
	const statusColor = misc.ColorFaint

	debugPrintf("%+v\n\n", formatSettings())

	// fmt.Println("getting info...")
	fi, err := LoadFile(filename, 0)
	if err != nil {
		printOnError()
		return err
	}
	debugPrintf("loaded:\n%v", fi)

	gt := time.Now()
	t := time.Now()

	colorizedPrintf(statusColor, "scanning...\n")
	err = ScanAudio(fi)
	if err != nil {
		printOnError()
		return err
	}
	debugPrintf("local %v, global %v\n", time.Since(t), time.Since(gt))

	if settings.Behavior.ScanOnly {
		printOnOk()
		debugPrintf("Elapsed time: %v\n", time.Since(gt))
		return nil
	}

	t = time.Now()
	colorizedPrintf(statusColor, "calculating parameters...\n")
	err = renderParameters(fi)
	if err != nil || settings.Behavior.ScanOnly {
		printOnError()
		return err
	}

	err = CheckIfReadyToCompile(fi)
	if err != nil {
		printOnError()
		return err
	}

	t = time.Now()
	colorizedPrintf(statusColor, "processing...\n")
	canBeFixed, err := ProcessTo(fi)
	if err != nil {
		if !canBeFixed {
			printOnError()
			return err
		}
		colorizedPrintf(misc.ColorYellow, "fixing...\n")
		canBeFixed, err = ProcessTo(fi)
		if err != nil {
			printOnError()
			return err
		}
	}
	debugPrintf("local %v, global %v\n", time.Since(t), time.Since(gt))

	printOnOk()
	debugPrintf("Elapsed time: %v\n", time.Since(gt))

	return nil
}
