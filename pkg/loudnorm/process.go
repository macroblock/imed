package loudnorm

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/macroblock/imed/pkg/ffmpeg"
)

// GlobalDebug -
var GlobalDebug = false

var (
	ac3Params2 = []string{
		"-c:a", "ac3",
		"-b:a", "256k",
		"-ac:a", "2",
	}
	ac3Params6 = []string{
		"-c:a", "ac3",
		"-b:a", "448k",
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

// TMode -
type TMode int

// -
const (
	ModeUnknown = TMode(iota)
	ModeSD
	ModeHD
)

func (o TMode) String() string {
	switch o {
	default:
		return "<error>"
	case ModeUnknown:
		return "unknown"
	case ModeSD:
		return "SD"
	case ModeHD:
		return "HD"
	}
}

func callFFMPEG(parse func([]byte) (interface{}, error), args ...string) error {
	if GlobalDebug {
		fmt.Println("### params: ", args)
	}

	c := exec.Command("ffmpeg", args...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err := c.Run()
	if err != nil {
		return errors.New(string(e.Bytes()))
	}

	if parse == nil {
		return nil
	}
	result, err := parse(e.Bytes())
	_ = result
	if err != nil {
		return err
	}
	return nil
}

// func fmtErrors(errors []error) error {
// 	if len(errors) == 0 {
// 		return nil
// 	}
// 	ret := "errors occured:\n"
// 	for i, err := range errors {
// 		ret = fmt.Sprintf("%v%2d: %v\n", ret, i, err)
// 	}
// 	return fmt.Errorf("%s", ret)
// }

// ScanAudio -
func ScanAudio(fi *TFileInfo) error {
	streams := []*TStreamInfo{}
	for _, stream := range fi.Streams {
		if stream.Type != "audio" {
			continue
		}
		stream.AudioParams = inferAudioParams(stream)
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
		// stream.AudioParams = generateAudioParams(fi, stream)
		if stream.done || stream.validLoudness {
			if GlobalDebug {
				fmt.Printf("stream %v has valid loudness (%v)\n", stream.Index, stream.LoudnessInfo)
			}
			continue
		}
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
			// if stream.ExtName != "" {
			// 	filename = stream.ExtName
			// 	index = 0
			// }
			return fmt.Errorf("%v:%v is not ready to compile (%v)", filename, index, stream.LoudnessInfo)
		}
	}
	return nil
}

// ProcessTo -
func ProcessTo(fi *TFileInfo) error {

	params := []string{"-y", "-hide_banner"}
	params = append(params, getGloblaFlags()...)
	params = append(params, "-i", fi.Filename)

	filters := []string{}
	outputs := []string{}
	ffmpeg.UniqueReset()

	time := ffmpeg.FloatToTime(fi.Duration)
	combParser := ffmpeg.NewCombineParser(
		ffmpeg.NewAudioProgressParser(time, nil),
	)

	// inputIndex := 0
	// for _, stream := range fi.Streams {
	// 	switch stream.Type {
	// 	case "video":
	// 	case "subtitle":
	// 	case "audio":
	// 		if stream.ExtName != "" {
	// 			inputIndex++
	// 			stream.extInputIndex = inputIndex
	// 			params = append(params, "-i", stream.ExtName)
	// 		}
	// 	}
	// }
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
				"[0:~idx~]~compressor~,asplit[s~u~][o~idx~];"+
					"[s~u~]~vd~,~ebur~,anullsink")
			outputs = appendPattern(outputs, stream, nil,
				"-map", "[o~idx~]")
			outputs = append(outputs, stream.AudioParams...)
			outputs = append(outputs,
				"-metadata:s:a:"+strconv.Itoa(audioIndex), "language="+stream.Lang,
				"-disposition:s:a:"+strconv.Itoa(audioIndex), def,
			)
		}
	}
	params = append(params, "-filter_complex")
	params = append(params, strings.Join(filters, ";"))
	params = append(params, outputs...)
	params = append(params, "-metadata")
	params = append(params, "comment="+PackTargetLoudnessInfo(fi)) //+strings.Join(metadata, "\n"))
	// params = append(params, "description="+strings.Join(metadata, "\n"))
	params = append(params, generateOutputName(fi))
	if GlobalDebug {
		fmt.Println("### params: ", params)
	}
	err := ffmpeg.Run(combParser, params...)
	if err != nil {
		return err
	}

	errStrs := []string{}
	for i, stream := range fi.Streams {
		stream.LoudnessInfo = &TLoudnessInfo{
			I:  stream.eburInfo.I,
			RA: stream.eburInfo.LRA,
			TP: stream.eburInfo.TP,
			TH: stream.eburInfo.Thresh,
			MP: stream.volumeInfo.MaxVolume,
			CR: stream.TargetLI.CR, //-1.0,
		}
		if GlobalDebug {
			fmt.Println("##### stream:", i,
				"\n  ebur >", stream.eburInfo,
				"\n  vol  >", stream.volumeInfo)
		}
		if !LoudnessIsEqual(stream.LoudnessInfo, stream.TargetLI) {
			errStrs = append(errStrs, fmt.Sprintf("stream #%v: loudness info is not equal to the planned one", i))
		}
	}
	if len(errStrs) != 0 {
		err := os.Remove(generateOutputName(fi))
		if err != nil {
			fmt.Printf("!!! error while removing file %v: %v", generateOutputName(fi), err)
		}
		return fmt.Errorf("%v", strings.Join(errStrs, "\n"))
	}

	return nil
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
	s = strings.Replace(s, ":", ": ", -1)
	return s
}

// Process -
func Process(filename string) error {
	if GlobalDebug {
		fmt.Printf("%+v\n\n", formatSettings())
	}
	fmt.Println("getting info...")
	fi, err := LoadFile(filename, 0)
	if err != nil {
		return err
	}
	if GlobalDebug {
		fmt.Printf("loaded:\n%v", fi)
	}

	gt := time.Now()
	t := time.Now()

	fmt.Println("scanning...")
	err = ScanAudio(fi)
	if err != nil {
		return err
	}
	if GlobalDebug {
		fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))
	}

	t = time.Now()
	fmt.Println("calculating parameters...")
	err = renderParameters(fi)
	for _, stream := range fi.Streams {
		fmt.Printf("        #%v: %v\n", stream.Index, stream.CompParams)
	}
	if err != nil {
		return err
	}

	err = CheckIfReadyToCompile(fi)
	if err != nil {
		return err
	}

	t = time.Now()
	fmt.Println("processing...")
	err = ProcessTo(fi)
	if err != nil {
		return err
	}
	if GlobalDebug {
		fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))
	}

	fmt.Printf("Ok. Elapsed time: %v\n", time.Since(gt))

	return nil
}

func alignStr(w int, s string) string {
	if len(s) == 0 {
		return s
	}
	for _, r := range s {
		if r == '.' {
			break
		}
		w--
	}
	if w < 0 {
		w = 0
	}
	return strings.Repeat(" ", w) + s
}
