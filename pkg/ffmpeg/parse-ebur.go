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
	name        string
	re          *regexp.Regexp
	active      bool
	truePeaks   bool
	accepted    bool
	finished    bool
	linesToRead int
	lines       []string
	ret         *TEburInfo
}

// NewEburParser -
func NewEburParser(name string, truePeaks bool, ret *TEburInfo) *TEburParser {
	linesToRead := 10
	if truePeaks {
		linesToRead = 13
	}
	o := &TEburParser{name: name, truePeaks: truePeaks, linesToRead: linesToRead, ret: ret}
	o.re = regexp.MustCompile(fmt.Sprintf("\\[%v @ [^ ]+\\] Summary:.*", name))
	return o
}

// var reEbur128 = regexp.MustCompile("\\[Parsed_ebur128_\\d+ @ [^ ]+\\] Summary:.*")

// Finish -
func (o *TEburParser) Finish() error {
	// if !o.finished {
	// 	return nil, fmt.Errorf("Ebur parser: uncompleted\n %v", strings.Join(o.lines, "\n"))
	// }

	if o.ret == nil {
		o.ret = &TEburInfo{}
	}
	data, err := parseEbur128Summary(o.lines, o.truePeaks)
	if err != nil {
		return fmt.Errorf("Ebur parser: %v\n %v", err, strings.Join(o.lines, "\n"))
	}
	*o.ret = *data

	return nil
}

// Parse -
func (o *TEburParser) Parse(line string, eof bool) (bool, error) {

	if !o.accepted && !o.active {
		if o.re.MatchString(line) {
			o.accepted = true
			o.active = true
		}
		return o.accepted, nil
	}
	if line != "" && line[0] != ' ' {
		o.accepted = false
		o.active = false
		return false, nil
	}

	o.lines = append(o.lines, line)

	return true, nil
}

// GetData -
func (o *TEburParser) GetData() *TEburInfo {
	// if !o.finished {
	// 	return nil, fmt.Errorf("Ebur parser: uncompleted\n %v", strings.Join(o.lines, "\n"))
	// }

	// data, err := parseEbur128Summary(o.lines, o.truePeaks)
	// if err != nil {
	// 	return nil, fmt.Errorf("Ebur parser: %v\n %v", err, strings.Join(o.lines, "\n"))
	// }

	// return data, nil
	return o.ret
}

func parseEbur128Summary(list []string, truePeaks bool) (*TEburInfo, error) {
	st := map[string]string{}
	prefix := ""

	for _, line := range list {
		name, val, err := parseNameVal(line, ":")
		if err != nil {
			return nil, err
		}
		switch name {
		case "":
			continue
		case "Integrated loudness":
			prefix = "IL"
		case "Loudness range":
			prefix = "LR"
		case "True peak":
			prefix = "TP"
		case "Sample peak":
			prefix = "SP"
		}
		st[prefix+"."+name] = val
	}
	ret := &TEburInfo{}
	err := error(nil)
	// fmt.Printf("$$$$$$$$\n%q\n", strings.Join(list, "\\n\n"))
	// list, _, err := parseValS(list, "Integrated loudness:", "")
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Printf("$$$$$$$$\n%q\n", st)
	ret.I, err = getValF(st, "IL.I", "LUFS")
	if err != nil {
		return nil, err
	}
	ret.Thresh, err = getValF(st, "IL.Threshold", "LUFS")
	if err != nil {
		return nil, err
	}
	// _, err = getValS( "Loudness range", "")
	// if err != nil {
	// 	return nil, err
	// }
	ret.LRA, err = getValF(st, "LR.LRA", "LU")
	if err != nil {
		return nil, err
	}
	ret.Thresh2, err = getValF(st, "LR.Threshold", "LUFS")
	if err != nil {
		return nil, err
	}
	ret.LRALow, err = getValF(st, "LR.LRA low", "LUFS")
	if err != nil {
		return nil, err
	}
	ret.LRAHigh, err = getValF(st, "LR.LRA high", "LUFS")
	if err != nil {
		return nil, err
	}

	ret.TP = math.NaN()
	if truePeaks {
		ret.TP, err = getValF(st, "TP.Peak", "dBFS")
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}
