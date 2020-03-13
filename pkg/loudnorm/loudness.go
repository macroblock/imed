package loudnorm

import (
	"strconv"
)

var (
	targetI    = "-23.0"
	targetLRA  = "20.0"
	targetTP   = "-1.0"
	samplerate = "48k"
)

// ValidLoudness -
func ValidLoudness(li *LoudnessInfo) bool {
	if li == nil {
		return false
	}
	tI, err := strconv.ParseFloat(targetI, 64)
	if err != nil {
		panic(err)
	}
	tLRA, err := strconv.ParseFloat(targetLRA, 64)
	if err != nil {
		panic(err)
	}
	tTP, err := strconv.ParseFloat(targetTP, 64)
	if err != nil {
		panic(err)
	}
	I, err := strconv.ParseFloat(li.I, 64)
	if err != nil {
		panic(err)
	}
	RA, err := strconv.ParseFloat(li.RA, 64)
	if err != nil {
		panic(err)
	}
	TP, err := strconv.ParseFloat(li.TP, 64)
	if err != nil {
		panic(err)
	}
	if tI+0.5 > I && I > tI-0.5 && // should it be +/-1.0 ?
		tLRA+1.0 >= RA && tTP >= TP {
		return true
	}
	// fmt.Printf("####### invalid %v\n", tI)
	return false
}
