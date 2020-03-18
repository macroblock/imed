package ffmpeg

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IAudioProgress -
type IAudioProgress interface {
	Callback(Time) error
}

/////////////////////////////////////////////////////////////////////////////////////

type tCombineParser struct {
	parsers []IParser
}

// NewCombineParser -
func NewCombineParser(parsers ...IParser) IParser {
	return &tCombineParser{parsers: parsers}
}

// Parse -
func (o *tCombineParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {
	accepted = false
	finished = true
	err = error(nil)
	for i := range o.parsers {
		if o.parsers[i] == nil {
			continue
		}
		if accepted {
			if o.parsers[i] != nil {
				finished = false
			}
			continue
		}
		acc, fin, e := o.parsers[i].Parse(line, eof)
		accepted = acc
		if e != nil {
			return false, true, e
		}
		if fin {
			o.parsers[i] = nil
		} else {
			finished = false
		}
	}
	return accepted, finished, nil
}

/////////////////////////////////////////////////////////////////////////////////////

type tAudioProgressParser struct {
	callback IAudioProgress
}

var reAudioProgress = regexp.MustCompile("size=.+ time=(\\d{2}:\\d{2}:\\d{2}.\\d+) bitrate=.+ speed=.+")

// Parse -
func (o *tAudioProgressParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {
	if o == nil || o.callback == nil {
		return false, false, fmt.Errorf("audio progress parser: either receiver or callback is nil")
	}
	if val := reAudioProgress.FindAllStringSubmatch(line, 1); val != nil {
		t, err := ParseTime(val[0][1])
		if err != nil {
			return true, eof, err
		}
		err = o.callback.Callback(t)
		return true, eof, err
	}
	return false, eof, nil
}

type tDefaultAudioProgressCallback struct {
	lastTime time.Time
	total    Time
}

func (o *tDefaultAudioProgressCallback) Callback(t Time) error {
	ct := time.Now()
	zeroTime := time.Time{}
	if o.lastTime == zeroTime {
		o.lastTime = ct
	}
	fmt.Printf("%v / %v, delta: %v\n", t, o.total, time.Since(o.lastTime))
	return nil
}

// NewAudioProgressParser -
func NewAudioProgressParser(totalLen Time, callback IAudioProgress) IParser {
	// fmt.Printf("@@@@: %q\n", line)
	if callback == nil {
		callback = &tDefaultAudioProgressCallback{total: totalLen}
	}
	return &tAudioProgressParser{
		callback: callback,
	}
}

/////////////////////////////////////////////////////////////////////////////////////

// TEburData -
type TEburData struct {
	I      float64
	Thresh float64

	LRA     float64
	Thresh2 float64
	LRALow  float64
	LRAHigh float64

	TP float64
}

// TEburParser -
type TEburParser struct {
	truePeaks bool
	accepted  bool
	finished  bool
	restLines int
	lines     []string
}

// NewEburParser -
func NewEburParser(truePeaks bool) *TEburParser {
	restLines := 10
	if truePeaks {
		restLines = 13
	}
	o := &TEburParser{truePeaks: truePeaks, restLines: restLines}
	return o
}

var reEbur128 = regexp.MustCompile("\\[Parsed_ebur128_0 @ [^ ]+\\] Summary:.*")

// Parse -
func (o *TEburParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {

	if !o.accepted && !o.finished {
		if reEbur128.MatchString(line) {
			o.accepted = true
		}
		return o.accepted, o.finished, nil
	}

	if !o.finished {
		o.lines = append(o.lines, line)
	}
	if o.restLines <= len(o.lines) {
		o.finished = true
		return o.accepted, o.finished, nil
	}

	return true, false, nil
}

// GetData -
func (o *TEburParser) GetData() (*TEburData, error) {
	if !o.finished {
		return nil, fmt.Errorf("Ebur parser: incomplete data\n %v", strings.Join(o.lines, "\n"))
	}

	data, err := parseEbur128Summary(o.lines, o.truePeaks)
	if err != nil {
		return nil, fmt.Errorf("Ebur parser: %v\n %v", err, strings.Join(o.lines, "\n"))
	}

	return data, nil
}

func skipBlank(list []string) []string {
	for len(list) > 0 && strings.TrimSpace(list[0]) == "" {
		list = list[1:]
	}
	return list
}

func parseVal(list []string, prefix string, trimSuffix string) ([]string, string, error) {
	list = skipBlank(list)
	if len(list) == 0 {
		return nil, "", fmt.Errorf("not enough data")
	}
	s := strings.TrimSpace(list[0])
	if !strings.HasPrefix(s, prefix) {
		return nil, "", fmt.Errorf("does not have prefix %q", prefix)
	}
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, trimSuffix)
	s = strings.TrimSpace(s)
	return list[1:], s, nil
}

func parseValS(list []string, prefix string, trimSuffix string) ([]string, string, error) {
	return parseVal(list, prefix, trimSuffix)
}

func parseValI(list []string, prefix string, trimSuffix string) ([]string, int, error) {
	newList, s, err := parseVal(list, prefix, trimSuffix)
	if err != nil {
		return newList, 0, err
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return newList, 0, err
	}
	return newList, int(val), nil
}

func parseValF(list []string, prefix string, trimSuffix string) ([]string, float64, error) {
	newList, s, err := parseVal(list, prefix, trimSuffix)
	if err != nil {
		return newList, 0.0, err
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return newList, 0.0, err
	}
	return newList, val, nil
}

func parseEbur128Summary(list []string, truePeaks bool) (*TEburData, error) {
	// fmt.Printf("$$$$$$$$\n%q\n", strings.Join(list, "\\n\n"))
	list, _, err := parseValS(list, "Integrated loudness:", "")
	if err != nil {
		return nil, err
	}
	list, I, err := parseValF(list, "I:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, Threshold, err := parseValF(list, "Threshold:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, _, err = parseValS(list, "Loudness range:", "")
	if err != nil {
		return nil, err
	}
	list, LRA, err := parseValF(list, "LRA:", "LU")
	if err != nil {
		return nil, err
	}
	list, Threshold2, err := parseValF(list, "Threshold:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, LRALow, err := parseValF(list, "LRA low:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, LRAHigh, err := parseValF(list, "LRA high:", "LUFS")
	if err != nil {
		return nil, err
	}

	TP := math.NaN()
	if truePeaks {
		list, _, err = parseValS(list, "True peak:", "")
		if err != nil {
			return nil, err
		}

		list, TP, err = parseValF(list, "Peak:", "dBFS")
		if err != nil {
			return nil, err
		}
	}

	ret := &TEburData{
		I:       I,
		Thresh:  Threshold,
		LRA:     LRA,
		Thresh2: Threshold2,
		LRALow:  LRALow,
		LRAHigh: LRAHigh,
		TP:      TP,
	}
	return ret, nil
}
