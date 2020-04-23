package loudnorm

import (
	"fmt"
	"math"
	"strconv"
)

// var (
// 	targetI     = -23.0
// 	targetLRA   = 20.0
// 	targetTP    = -1.0
// 	targetUseTP = true
// 	samplerate  = "48k"
// )

func targetI() float64 {
	return settings.Loudness.I
}

func targetIMin() float64 {
	return settings.Loudness.I - settings.Loudness.Precision
}

func targetIMax() float64 {
	return settings.Loudness.I + settings.Loudness.Precision
}

func targetLRA() float64 {
	return settings.Loudness.RA
}

func targetTP() float64 {
	return settings.Loudness.TP
}

func targetMP() float64 {
	return settings.Loudness.MP
}

const loudnessDeltaLI = 0.5

// ValidLoudness -
func ValidLoudness(li *TLoudnessInfo) bool {
	if li == nil {
		return false
	}
	if targetIMax() <= li.I || li.I <= targetIMin() {
		return false
	}
	if !math.IsNaN(targetTP()) && li.TP > targetTP() {
		return false
	}
	if !math.IsNaN(targetTP()) && li.RA > targetLRA() {
		return false
	}
	return true
}

// SuitableLoudness -
func SuitableLoudness(li *TLoudnessInfo) bool {
	// defer fmt.Printf("@@@@@@@@@ !!!!! %+v", li)
	if li == nil {
		return false
	}
	if li.I <= targetIMin() {
		// fmt.Println("IMin")
		return false
	}
	if !math.IsNaN(targetTP()) && li.TP > targetTP() {
		// fmt.Println("TP")
		return false
	}
	if !math.IsNaN(targetTP()) && li.RA > targetLRA() {
		// fmt.Println("LRA")
		return false
	}
	return true
}

func ftostr(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func feq(a, b float64, okNaN bool) bool {
	// in order to "-0" != "0"
	if okNaN {
		// fmt.Printf("0@@@@@ ok\n")
		return true
	}
	// fmt.Printf("1@@@@@ %v, %v\n", a, b)
	if a == b {
		return true
	}
	// fmt.Printf("2@@@@@ %v, %v\n", a, b)
	return ftostr(a) == ftostr(b)
}

func isNaN(a, b float64) bool {
	return math.IsNaN(a) || math.IsNaN(b)
}

// LoudnessIsEqual -
func LoudnessIsEqual(a, b *TLoudnessInfo) bool {
	if GlobalDebug {
		fmt.Printf("@@a: %v\n", a)
		fmt.Printf("@@b: %v\n", b)
	}
	if feq(a.I, b.I, false) &&
		feq(a.RA, b.RA, isNaN(a.RA, b.RA)) &&
		feq(a.TP, b.TP, isNaN(a.TP, b.TP)) &&
		feq(a.MP, b.MP, false) &&
		feq(a.TH, b.TH, false) &&
		feq(a.CR, b.CR, false) &&
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
	postAmp := targetI() - li.I
	if postAmp > 0.0 {
		postAmp = 0.0
	}
	compParams.PostAmp = postAmp
	li.I += postAmp
	li.TP += postAmp
	li.TH += postAmp
	li.MP += postAmp
	// fmt.Printf("@@@@@ Post Amp: %v\n", postAmp)
	// stream.done = true
	// fmt.Println("##### stream:", i,
	// 	"\n  li      >", li,
	// 	"\n  postAmp >", stream.CompParams.PostAmp)
	return true
}
