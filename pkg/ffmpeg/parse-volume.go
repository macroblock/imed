package ffmpeg

import (
	"fmt"
	"regexp"
	"strings"
)

// TVolumeInfo -
type TVolumeInfo struct {
	NSamples   int
	MeanVolume float64
	MaxVolume  float64
	Histogram  []int
}

// TVolumeParser -
type TVolumeParser struct {
	lines []string
}

// NewVolumeParser -
func NewVolumeParser() *TVolumeParser {
	o := &TVolumeParser{}
	return o
}

var reVolumeDetect = regexp.MustCompile("\\[Parsed_volumedetect_\\d+ @ [^ ]+\\] (.*)")

// Parse -
func (o *TVolumeParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {

	if val := reVolumeDetect.FindAllStringSubmatch(line, 1); val != nil {
		accepted = true
		o.lines = append(o.lines, val[0][1])
		return accepted, eof, nil
	}

	return accepted, eof, nil
}

// GetData -
func (o *TVolumeParser) GetData() (*TVolumeInfo, error) {
	// if !o.finished {
	// 	return nil, fmt.Errorf("Ebur parser: incomplete data\n %v", strings.Join(o.lines, "\n"))
	// }

	data, err := parseVolumeDetect(o.lines)
	if err != nil {
		return nil, fmt.Errorf("VolumeDetect parser: %v\n %v", err, strings.Join(o.lines, "\n"))
	}

	return data, nil
}

func parseVolumeDetect(list []string) (*TVolumeInfo, error) {

	list, nsamples, err := parseValI(list, "n_samples:", "")
	if err != nil {
		return nil, err
	}
	list, meanVolume, err := parseValF(list, "mean_volume:", "dB")
	if err != nil {
		return nil, err
	}
	list, maxVolume, err := parseValF(list, "max_volume:", "dB")
	if err != nil {
		return nil, err
	}
	ret := &TVolumeInfo{
		NSamples:   nsamples,
		MeanVolume: meanVolume,
		MaxVolume:  maxVolume,
	}
	return ret, nil
}
