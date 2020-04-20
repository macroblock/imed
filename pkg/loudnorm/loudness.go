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

// FixLoudness -
func FixLoudness(li *TLoudnessInfo, compParams *TCompressParams) bool {
	if !SuitableLoudness(li) {
		return false
	}
	postAmp := targetI - li.I
	if postAmp > 0.0 {
		postAmp = 0.0
	}
	compParams.PostAmp = postAmp
	li.I += postAmp
	li.TP += postAmp
	li.TH += postAmp
	li.MP += postAmp
	// stream.done = true
	// fmt.Println("##### stream:", i,
	// 	"\n  li      >", li,
	// 	"\n  postAmp >", stream.CompParams.PostAmp)
	return true
}
