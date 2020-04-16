package loudnorm

var (
	targetI    = -23.0
	targetLRA  = 20.0
	targetTP   = -1.0
	samplerate = "48k"
)

// ValidLoudness -
func ValidLoudness(li *TLoudnessInfo) bool {
	if li == nil {
		return false
	}
	if targetI+0.5 > li.I && li.I > targetI-0.5 && // should it be +/-1.0 ?
		//tLRA+1.0 >= RA && tTP >= TP
		true {
		return true
	}
	// fmt.Printf("####### invalid %v\n", tI)
	return false
}

// SuitableLoudness -
func SuitableLoudness(li *TLoudnessInfo) bool {
	if li == nil {
		return false
	}
	if li.I > targetI-0.5 && // should it be +/-1.0 ?
		//tLRA+1.0 >= RA && tTP >= TP
		true {
		return true
	}
	// fmt.Printf("####### invalid %v\n", tI)
	return false
}
