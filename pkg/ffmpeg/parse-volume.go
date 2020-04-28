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
	name  string
	re    *regexp.Regexp
	lines []string
	ret   *TVolumeInfo
}

// NewVolumeParser -
func NewVolumeParser(name string, ret *TVolumeInfo) *TVolumeParser {
	o := &TVolumeParser{name: name, ret: ret}
	o.re = regexp.MustCompile(fmt.Sprintf("\\[%v @ [^ ]+\\] (.*)", name))
	return o
}

// var reVolumeDetect = regexp.MustCompile("\\[Parsed_volumedetect_\\d+ @ [^ ]+\\] (.*)")

// Finish -
func (o *TVolumeParser) Finish() error {
	if o.ret == nil {
		o.ret = &TVolumeInfo{}
	}
	data, err := parseVolumeDetect(o.lines)
	if err != nil {
		return fmt.Errorf("VolumeDetect parser: %v\n%v", err, strings.Join(o.lines, "\n"))
	}
	*o.ret = *data
	return nil
}

// Parse -
func (o *TVolumeParser) Parse(line string, eof bool) (bool, error) {
	if val := o.re.FindAllStringSubmatch(line, 1); val != nil {
		o.lines = append(o.lines, val[0][1])
		return true, nil
	}
	return false, nil
}

// GetData -
func (o *TVolumeParser) GetData() *TVolumeInfo {
	return o.ret
}

func parseVolumeDetect(list []string) (*TVolumeInfo, error) {
	vals := NewArgMap()
	for _, line := range list {
		name, val, err := parseNameVal(line, ":")
		if err != nil {
			return nil, err
		}
		if name != "" {
			vals.Add(name, val)
		}
	}
	ret := &TVolumeInfo{}
	vals.GetI("n_samples", "", &ret.NSamples)
	vals.GetF("mean_volume", "dB", &ret.MeanVolume)
	vals.GetF("max_volume", "dB", &ret.MaxVolume)
	if vals.Error() != nil {
		return nil, vals.Error()
	}
	return ret, nil
}
