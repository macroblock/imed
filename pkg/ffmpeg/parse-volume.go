package ffmpeg

import (
	"fmt"
	"regexp"
	"strconv"
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
func (o *TVolumeParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {

	// if val := reVolumeDetect.FindAllStringSubmatch(line, 1); val != nil {
	if val := o.re.FindAllStringSubmatch(line, 1); val != nil {
		accepted = true
		o.lines = append(o.lines, val[0][1])
		return accepted, eof, nil
	}

	return accepted, eof, nil
}

// GetData -
func (o *TVolumeParser) GetData() *TVolumeInfo {
	// if !o.finished {
	// 	return nil, fmt.Errorf("Ebur parser: incomplete data\n %v", strings.Join(o.lines, "\n"))
	// }

	// data, err := parseVolumeDetect(o.lines)
	// if err != nil {
	// 	return nil, fmt.Errorf("VolumeDetect parser: %v\n%v", err, strings.Join(o.lines, "\n"))
	// }

	// return data, nil
	return o.ret
}

func parseNameVal(line, delim string) (string, string, error) {
	s := strings.TrimSpace(line)
	if s == "" {
		return "", "", nil
	}
	x := strings.Split(s, delim)
	if len(x) > 2 {
		return "", "", fmt.Errorf("more than one delimeter in line %q", line)
	}
	if len(x) < 2 {
		return "", "", fmt.Errorf("no delimeter in line %q", line)
	}
	return strings.TrimSpace(x[0]), strings.TrimSpace(x[1]), nil
}

func valToS(val, trimSuffix string) (string, error) {
	s := strings.TrimSpace(val)
	if !strings.HasSuffix(val, trimSuffix) {
		return "", fmt.Errorf("%q does not have suffix %q", val, trimSuffix)
	}
	s = strings.TrimSuffix(s, trimSuffix)
	s = strings.TrimSpace(s)
	return s, nil
}

func valToI(val, trimSuffix string) (int, error) {
	s, err := valToS(val, trimSuffix)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(ret), nil
}

func valToF(val, trimSuffix string) (float64, error) {
	s, err := valToS(val, trimSuffix)
	if err != nil {
		return 0.0, err
	}
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return ret, nil
}

func getValS(st map[string]string, name, suffix string) (string, error) {
	if s, ok := st[name]; ok {
		return valToS(s, suffix)
	}
	return "", fmt.Errorf("%v not found", name)
}

func getValI(st map[string]string, name, suffix string) (int, error) {
	if s, ok := st[name]; ok {
		return valToI(s, suffix)
	}
	return 0, fmt.Errorf("%v not found", name)
}

func getValF(st map[string]string, name, suffix string) (float64, error) {
	if s, ok := st[name]; ok {
		return valToF(s, suffix)
	}
	return 0.0, fmt.Errorf("%v not found", name)
}

func parseVolumeDetect(list []string) (*TVolumeInfo, error) {
	st := map[string]string{}

	for _, line := range list {
		name, val, err := parseNameVal(line, ":")
		if err != nil {
			return nil, err
		}
		if name != "" {
			st[name] = val
		}
	}
	ret := &TVolumeInfo{}
	err := error(nil)
	ret.NSamples, err = getValI(st, "n_samples", "")
	if err != nil {
		return nil, err
	}
	ret.MeanVolume, err = getValF(st, "mean_volume", "dB")
	if err != nil {
		return nil, err
	}
	ret.MaxVolume, err = getValF(st, "max_volume", "dB")
	if err != nil {
		return nil, err
	}
	return ret, nil
}
