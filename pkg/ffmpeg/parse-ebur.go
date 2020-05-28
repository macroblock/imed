package ffmpeg

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// TEburInfo -
type TEburInfo struct {
	// Timeline []TTimelineElement

	I      float64
	Thresh float64

	LRA     float64
	Thresh2 float64
	LRALow  float64
	LRAHigh float64

	TP float64

	MaxST    float64
	MinST    float64
	SumST    float64
	CountST  int
	CountNaN int
	AboveST  int
	BelowST  int
	EqualST  int
}

// TEburParser -
type TEburParser struct {
	name        string
	re          *regexp.Regexp
	reTime      *regexp.Regexp
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
	o.reTime = regexp.MustCompile(fmt.Sprintf("\\[%v @ [^ ]+\\] (t:.*)", name))
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

	if val := o.reTime.FindAllStringSubmatch(line, 1); val != nil {
		o.lines = append(o.lines, val[0][1])
		return true, nil
	}

	if !o.accepted && !o.active {
		if o.re.MatchString(line) {
			o.accepted = true
			o.active = true
			return true, nil
		}
		return false, nil
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
	return o.ret
}

func parseEbur128Summary(list []string, truePeaks bool) (*TEburInfo, error) {
	vals := NewArgMap()
	prefix := ""

	maxLST := math.Inf(-1)
	minLST := math.Inf(+1)
	countLST := 0
	countNaN := 0
	sumLST := 0.0
	arrayST := []float64{}

	started := false
	for _, line := range list {
		// fmt.Printf("line: %q\n", line)
		if strings.HasPrefix(line, "t: ") {
			i := strings.Index(line, "S:")
			s := strings.TrimSpace(line[i+2:])
			i = strings.Index(s, " ")
			s = s[:i]
			val, err := valToF(s, "")
			if err != nil {
				return nil, err
			}
			if !started && (val < -120.0 || math.IsNaN(val)) {
				continue
			}
			started = true
			countLST++

			arrayST = append(arrayST, val)

			if math.IsNaN(val) {
				countNaN++
				continue
			}
			sumLST += val
			maxLST = math.Max(val, maxLST)
			minLST = math.Min(val, minLST)
			// fmt.Println("min max", minLST, maxLST)
			continue
		}

		// fmt.Printf("line: %q\n", line)
		name, val, err := parseNameVal(line, ":")
		if err != nil {
			return nil, err
		}
		switch name {
		case "":
			continue
		case "Integrated loudness":
			prefix = "IL."
		case "Loudness range":
			prefix = "LR."
		case "True peak":
			prefix = "TP."
		case "Sample peak":
			prefix = "SP."
		}
		vals.Add(prefix+name, val)
	}

	ret := &TEburInfo{}
	vals.GetF("IL.I", "LUFS", &ret.I)
	vals.GetF("IL.Threshold", "LUFS", &ret.Thresh)
	vals.GetF("LR.LRA", "LU", &ret.LRA)
	vals.GetF("LR.Threshold", "LUFS", &ret.Thresh2)
	vals.GetF("LR.LRA low", "LUFS", &ret.LRALow)
	vals.GetF("LR.LRA high", "LUFS", &ret.LRAHigh)
	ret.TP = math.NaN()
	if truePeaks {
		vals.GetF("TP.Peak", "dBFS", &ret.TP)
	}
	if vals.Error() != nil {
		return nil, vals.Error()
	}

	// if NaNs more than one percent
	// if float64(countNaN)/float64(countLST) > 0.001 {
	// 	fmt.Printf("\n#internal <ebur parser>: too much NaNs (%v/%v > 0.001)\n\n", countNaN, countLST)
	// 	// minLST = math.NaN()
	// 	// maxLST = math.NaN()
	// }
	stAbove, stEqual, stBelow := 0, 0, 0
	stNaN := 0
	for _, st := range arrayST {
		switch {
		case math.IsNaN(st):
			stNaN++
		case st < ret.I-1.0:
			stBelow++
		case st >= ret.I+1.0:
			stAbove++
		default:
			stEqual++
		}
	}
	ret.MaxST = maxLST
	ret.MinST = minLST
	ret.SumST = sumLST
	ret.CountST = countLST
	ret.CountNaN = countNaN // !!!
	ret.AboveST = stAbove
	ret.BelowST = stBelow
	ret.EqualST = stEqual
	ret.CountNaN = stNaN // !!!

	if ret == nil {
		return nil, fmt.Errorf("eburInfo == nil")
	}
	return ret, nil
}
