package ffmpeg

import (
	"fmt"
	"regexp"
	"strings"
)

// TAStatsInfo -
type TAStatsInfo struct {
	BitDepth        string
	NumberOfSamples int
	RMSLevel        float64
	FlatFactor      float64
	PeakLevel       float64
}

// TAStatsParser -
type TAStatsParser struct {
	name  string
	re    *regexp.Regexp
	lines []string
	ret   *TAStatsInfo
}

// NewAStatsParser -
func NewAStatsParser(name string, ret *TAStatsInfo) *TAStatsParser {
	o := &TAStatsParser{name: name, ret: ret}
	o.re = regexp.MustCompile(fmt.Sprintf("\\[%v @ [^ ]+\\] (.*)", name))
	return o
}

// var reAStatsDetect = regexp.MustCompile("\\[Parsed_AStatsdetect_\\d+ @ [^ ]+\\] (.*)")

// Finish -
func (o *TAStatsParser) Finish() error {
	if o.ret == nil {
		o.ret = &TAStatsInfo{}
	}
	data, err := parseAStatsDetect(o.lines)
	if err != nil {
		return fmt.Errorf("AStatsDetect parser: %v\n%v", err, strings.Join(o.lines, "\n"))
	}
	*o.ret = *data
	return nil
}

// Parse -
func (o *TAStatsParser) Parse(line string, eof bool) (bool, error) {
	if val := o.re.FindAllStringSubmatch(line, 1); val != nil {
		o.lines = append(o.lines, val[0][1])
		return true, nil
	}
	return false, nil
}

// GetData -
func (o *TAStatsParser) GetData() *TAStatsInfo {
	return o.ret
}

func parseAStatsDetect(list []string) (*TAStatsInfo, error) {
	// st := map[string]string{}
	vals := NewArgMap()
	prefix := ""
	for _, line := range list {
		switch strings.TrimSpace(line) {
		case "":
			continue
		case "Overall":
			prefix = ":"
			continue
		}
		name, val, err := parseNameVal(line, ":")
		if err != nil {
			return nil, err
		}
		vals.Add(prefix+name, val)
	}
	ret := &TAStatsInfo{}
	// err := error(nil)
	// ret.BitDepth, err = getValS(st, ":Bit depth", "")
	// if err != nil {
	// 	return nil, err
	// }
	// ret.NumberOfSamples, err = getValI(st, ":Number of samples", "")
	// if err != nil {
	// 	return nil, err
	// }
	// ret.RMSLevel, err = getValF(st, ":RMS Level", "")
	// if err != nil {
	// 	return nil, err
	// }
	// ret.FlatFactor, err = getValF(st, ":Flat factor", "")
	// if err != nil {
	// 	return nil, err
	// }
	// ret.PeakLevel, err = getValF(st, ":Peak level dB", "")
	// if err != nil {
	// 	return nil, err
	// }
	vals.GetS(":Bit depth", "", &ret.BitDepth)
	vals.GetI(":Number of samples", "", &ret.NumberOfSamples)
	vals.GetF(":RMS level dB", "", &ret.RMSLevel)
	vals.GetF(":Flat factor", "", &ret.FlatFactor)
	vals.GetF(":Peak level dB", "", &ret.PeakLevel)
	if vals.Error() != nil {
		return nil, vals.Error()
	}
	return ret, nil
}
