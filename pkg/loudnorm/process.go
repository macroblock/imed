package loudnorm

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
		"-c:a", "ac3",
		"-b:a", "320k", // !!!! FIXME: is it correct?
		"-ac:a", "2",
	}
	mp2Params6 = []string{
		"-c:a", "ac3",
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
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	errors := []error{}
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
		wg.Add(1)
		go func(stream *TStreamInfo) {
			defer wg.Done()
			filename := stream.Parent.Filename
			index := stream.Index
			if stream.ExtName != "" {
				filename = stream.ExtName
				index = 0
			}
			li, err := Scan(filename, index)
			defer mtx.Unlock()
			mtx.Lock()
			if err != nil {
				errors = append(errors, err)
				return
			}
			if GlobalDebug {
				fmt.Printf("ebur128 %v:%v:\n  input: I: %v, LRA: %v, TP: %v, TH: %v, MP: %v\n",
					filepath.Base(filename), index,
					li.I, li.RA, li.TP, li.TH, li.MP,
				)
			}
			stream.LoudnessInfo = li

			stream.validLoudness = true
			if !ValidLoudness(stream.LoudnessInfo) {
				stream.ExtName = generateExtFilename(fi, stream)
				stream.validLoudness = false
			}
		}(stream)
	}

	wg.Wait()
	if len(errors) != 0 {
		return fmtErrors(errors)
	}
	return nil
}

func calculateParameters(fi *TFileInfo) error {
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	errors := []error{}
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
		wg.Add(1)
		go func(stream *TStreamInfo) {
			defer wg.Done()
			filename := stream.Parent.Filename
			index := stream.Index
			if stream.ExtName != "" {
				filename = stream.ExtName
				index = 0
			}
			li, err := Normalize(filename, index, false, stream.LoudnessInfo)
			defer mtx.Unlock()
			mtx.Lock()
			if err != nil {
				errors = append(errors, err)
				return
			}
			if GlobalDebug {
				fmt.Printf("ebur128 %v:%v:\n  input: I: %v, LRA: %v, TP: %v, TH: %v, MP: %v\n",
					filepath.Base(filename), index,
					li.I, li.RA, li.TP, li.TH, li.MP,
				)
			}
			stream.LoudnessInfo = li

			stream.validLoudness = true
			if !ValidLoudness(stream.LoudnessInfo) {
				stream.ExtName = generateExtFilename(fi, stream)
				stream.validLoudness = false
			}
		}(stream)
	}

	wg.Wait()
	if len(errors) != 0 {
		return fmtErrors(errors)
	}
	return nil
}

// CheckIfReadyToCompile -
func CheckIfReadyToCompile(fi *TFileInfo) error {
	for _, stream := range fi.Streams {
		if /*!stream.validFormat ||*/ !stream.validLoudness {
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

// DemuxAndNormalize -
func DemuxAndNormalize(fi *TFileInfo) error {
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	errors := []error{}
	for _, stream := range fi.Streams {
		if stream.Type != "audio" {
			continue
		}
		if ValidLoudness(stream.LoudnessInfo) {
			if GlobalDebug {
				fmt.Printf("stream %v has valid loudness (%v)\n", stream.Index, stream.LoudnessInfo)
			}
			continue
		}
		wg.Add(1)
		go func(stream *TStreamInfo) {
			defer wg.Done()
			// li, err := NormalizeTo(fi.Filename, stream.Index, stream.ExtName, stream.AudioParams,
			// 	stream.LoudnessInfo.I, stream.LoudnessInfo.RA, stream.LoudnessInfo.TP, stream.LoudnessInfo.TH)
			// li := &TLoudnessInfo{}
			err := fmt.Errorf("debug error")
			mtx.Lock()
			defer mtx.Unlock()
			if err != nil {
				errors = append(errors, err)
				return
			}
			if GlobalDebug {
				// fmt.Printf("loudnorm %v:%v:\n  %v\n  input: I: %v, LRA: %v, TP: %v, Thresh: %v, Offs: %v\n  output: I: %v, TP: %v, Thresh: %v\n",
				// 	filepath.Base(fi.Filename), stream.Index,
				// 	li.NormalizationType,
				// 	li.InputI, li.InputLRA, li.InputTP, li.InputThresh, li.TargetOffset,
				// 	li.OutputI, li.OutputTP, li.OutputThresh,
				// )
			}
		}(stream)
	}

	wg.Wait()
	if len(errors) != 0 {
		return fmtErrors(errors)
	}
	return nil
}

// MuxTo -
func MuxTo(fi *TFileInfo) error {
	muxParams := []string{
		"-y",
		"-hide_banner",
		"-i", fi.Filename,
	}
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
				// "-metadata:s:a:"+strconv.Itoa(audioIndex), "handler_name"+
				// 	"=AudioHandler\nL_I:test_string"+stream.LoudInfo.InputI+
				// 	"\nL_RA:"+stream.LoudInfo.InputLRA+
				// 	"\nL_TP:"+stream.LoudInfo.InputTP+
				// 	"\nL_TH:"+stream.LoudInfo.InputThresh,
			)
			// metadata = append(metadata,
			// 	fmt.Sprintf("[Stream #:%v]\nL_I  % 5.2f\nL_RA % 5.2f\nL_TP % 5.2f\nL_TH % 5.2f",
			// 		inputIndex,
			// 		stream.LoudnessInfo.I,
			// 		stream.LoudnessInfo.RA,
			// 		stream.LoudnessInfo.TP,
			// 		stream.LoudnessInfo.TH,
			// 	),
			// )
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
	fi, err := LoadFile(filename)
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
	fmt.Println("calulating parameters...")
	err = calculateParameters(fi)
	if err != nil {
		return err
	}

	fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))
	t = time.Now()
	fmt.Println("demuxing and normalizing...")
	err = DemuxAndNormalize(fi)
	if err != nil {
		return err
	}
	fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))

	t = time.Now()
	fmt.Println("checking...")
	err = ScanAudio(fi)
	if err != nil {
		return err
	}
	fmt.Printf("local %v, global %v\n", time.Since(t), time.Since(gt))

	err = CheckIfReadyToCompile(fi)
	if err != nil {
		return err
	}

	t = time.Now()
	fmt.Println("muxing...")
	err = MuxTo(fi)
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
