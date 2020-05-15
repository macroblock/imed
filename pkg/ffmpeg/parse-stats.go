package ffmpeg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TChannelInfo -
type TChannelInfo struct {
	BitDepth string
	// NumberOfSamples int
	RMSLevel   float64 // dB
	FlatFactor float64
	PeakLevel  float64 // dB
}

// TAStatsInfo -
type TAStatsInfo struct {
	Channels []TChannelInfo
	TChannelInfo
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
	chnum := 0
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
		switch name {
		case "":
			continue
		case "Channel":
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			prefix = strconv.Itoa(v-1) + ":"
			continue
		}
		// fmt.Printf("prefix: %q %q\n", prefix, name)
		vals.Add(prefix+name, val)
	}
	ret := &TAStatsInfo{}
	ret.Channels = make([]TChannelInfo, chnum)
	for i := -1; i < len(ret.Channels); i++ {
		prefix := ":"
		p := &ret.TChannelInfo
		if i >= 0 {
			p = &ret.Channels[i]
			prefix = strconv.Itoa(i) + ":"
		}
		vals.GetS(prefix+"Bit depth", "", &p.BitDepth)
		// vals.GetI(prefix+"Number of samples", "", &p.NumberOfSamples)
		vals.GetF(prefix+"RMS level dB", "", &p.RMSLevel)
		vals.GetF(prefix+"Flat factor", "", &p.FlatFactor)
		vals.GetF(prefix+"Peak level dB", "", &p.PeakLevel)
		if vals.Error() != nil {
			return nil, vals.Error()
		}
	}
	return ret, nil
}
