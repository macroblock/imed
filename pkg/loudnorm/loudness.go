package loudnorm

import (
	"strconv"
)

var (
	targetI     = -23.0
	targetLRA   = 20.0
	targetTP    = -1.0
	targetUseTP = true
	samplerate  = "48k"
)

const loudnessDeltaLI = 0.5

// ValidLoudness -
func ValidLoudness(li *TLoudnessInfo) bool {
	if li == nil {
		return false
	}
	if targetI+loudnessDeltaLI <= li.I || li.I <= targetI-loudnessDeltaLI {
		return false
	}
	if li.TP > targetTP {
		return false
	}
	if li.RA > targetLRA {
		return false
	}
	return true
}

// SuitableLoudness -
func SuitableLoudness(li *TLoudnessInfo) bool {
	if li == nil {
		return false
	}
	if li.I <= targetI-loudnessDeltaLI {
		return false
	}
	if li.TP > targetTP {
		return false
	}
	if li.RA > targetLRA {
		return false
	}
	return true
}

func ftostr(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func feq(a, b float64) bool {
	// in order to "-0" != "0"
	if a == b {
		return true
	}
	return ftostr(a) == ftostr(b)
}

// LoudnessIsEqual -
func LoudnessIsEqual(a, b *TLoudnessInfo) bool {
	// fmt.Printf("@@a: %v\n", a)
	// fmt.Printf("@@b: %v\n", a)
	if feq(a.I, b.I) &&
		feq(a.RA, b.RA) &&
		feq(a.TP, b.TP) &&
		feq(a.MP, b.MP) &&
		feq(a.TH, b.TH) &&
		feq(a.CR, b.CR) &&
		true {
		return true
	}
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
