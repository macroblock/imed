package loudnorm

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GlobalDebug -
const GlobalDebug = true

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

func fmtErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	ret := "errors occured:\n"
	for i, err := range errors {
		ret = fmt.Sprintf("%v%2d: %v\n", ret, i, err)
	}
	return fmt.Errorf("%s", ret)
}

func printInfo(name string, si []*TStreamInfo) {
	fmt.Printf("    %s:\n", name)
	for _, stream := range si {
		if stream.Type != "audio" {
			continue
		}
		fmt.Printf("        %2d: %v\n", stream.Index, stream.LoudnessInfo)
	}
}

// ScanAudio -
func ScanAudio(fi *TFileInfo) error {
	streams := []*TStreamInfo{}
	for _, stream := range fi.Streams {
		if stream.Type != "audio" {
			continue
		}
		stream.AudioParams = generateAudioParams(fi, stream)
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
		// stream.validLoudness = true
		// if !ValidLoudness(stream.LoudnessInfo) {
		// 	stream.validLoudness = false
		// }
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
			if stream.ExtName != "" {
				filename = stream.ExtName
				index = 0
			}
			return fmt.Errorf("%v:%v is not ready to compile (%v)", filename, index, stream.LoudnessInfo)
		}
	}
	return nil
}

// ProcessTo -
func ProcessTo(fi *TFileInfo) error {
	muxParams := []string{
		"-y",
		"-hide_banner",
		"-i", fi.Filename,
	}
	filters := []string{}
	inputIndex := 0
	for _, stream := range fi.Streams {
		switch stream.Type {
		case "video":
		case "subtitle":
		case "audio":
			if stream.ExtName != "" {
				inputIndex++
				stream.extInputIndex = inputIndex
				muxParams = append(muxParams, "-i", stream.ExtName)
			}
		}
	}
	muxParams = append(muxParams,
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
			muxParams = append(muxParams,
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
			muxParams = append(muxParams,
				"-map", "0:"+strconv.Itoa(stream.Index),
				"-c:s", "mov_text",
				"-metadata:s:s:"+strconv.Itoa(subtitleIndex), "language="+stream.Lang,
				// "-disposition:s:"+strconv.Itoa(subtitleIndex), def,
			)
		case "audio":
			audioIndex++
			_ = filters
			def := "none"
			if isFirstAudio {
				isFirstAudio = false
				def = "default"
			}
			mapParam := "0:" + strconv.Itoa(stream.Index)
			if stream.ExtName != "" {
				mapParam = strconv.Itoa(stream.extInputIndex) + ":0"
			}
			muxParams = append(muxParams,
				"-map", mapParam, //strconv.Itoa(stream.tempInputIndex)+":0",
				"-c:a", "copy",
				"-metadata:s:a:"+strconv.Itoa(audioIndex), "language="+stream.Lang,
				"-disposition:s:a:"+strconv.Itoa(audioIndex), def,
			)
		}
	}
	muxParams = append(muxParams, "-metadata")
	muxParams = append(muxParams, "comment="+PackLoudnessInfo(fi)) //+strings.Join(metadata, "\n"))
	// muxParams = append(muxParams, "description="+strings.Join(metadata, "\n"))
	muxParams = append(muxParams, generateOutputName(fi.Filename))
	err := callFFMPEG(nil, muxParams...)
	if err != nil {
		return err
	}
	return nil
}

// Process -
func Process(filename string) error {
	fmt.Println("getting info...")
	fi, err := LoadFile(filename, 0)
	if err != nil {
		return err
	}
	fmt.Printf("loaded:\n%v", fi)

	gt := time.Now()
	t := time.Now()

	fmt.Println("scanning...")
	err = ScanAudio(fi)
	if err != nil {
		return err
	}
	fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))

	t = time.Now()
	fmt.Println("render parameters...")
	err = renderParameters(fi)
	if err != nil {
		return err
	}
	t = time.Now()
	fmt.Println("muxing...")
	err = ProcessTo(fi)
	if err != nil {
		return err
	}
	fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))

	// fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))
	fmt.Println("Ok.")

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
