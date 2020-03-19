package ffmpeg

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// TEburInfo -
type TEburInfo struct {
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
	truePeaks   bool
	accepted    bool
	finished    bool
	linesToRead int
	lines       []string
}

// NewEburParser -
func NewEburParser(truePeaks bool) *TEburParser {
	linesToRead := 10
	if truePeaks {
		linesToRead = 13
	}
	o := &TEburParser{truePeaks: truePeaks, linesToRead: linesToRead}
	return o
}

var reEbur128 = regexp.MustCompile("\\[Parsed_ebur128_\\d+ @ [^ ]+\\] Summary:.*")

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
	if len(o.lines) >= o.linesToRead {
		o.finished = true
		return o.accepted, o.finished, nil
	}

	return true, false, nil
}

// GetData -
func (o *TEburParser) GetData() (*TEburInfo, error) {
	if !o.finished {
		return nil, fmt.Errorf("Ebur parser: uncompleted\n %v", strings.Join(o.lines, "\n"))
	}

	data, err := parseEbur128Summary(o.lines, o.truePeaks)
	if err != nil {
		return nil, fmt.Errorf("Ebur parser: %v\n %v", err, strings.Join(o.lines, "\n"))
	}

	return data, nil
}

func parseEbur128Summary(list []string, truePeaks bool) (*TEburInfo, error) {
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

	ret := &TEburInfo{
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
