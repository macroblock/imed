package loudnorm

import (
	"fmt"
	"math"
	"strconv"
)

// TLoudnessInfo -
type TLoudnessInfo struct {
	I  float64 // integrated
	RA float64 // range
	TP float64 // true peaks
	MP float64 // max peaks
	TH float64 // threshold
	// CR float64 // compress ratio
}

// TMiscInfo -
type TMiscInfo struct {
	I          float64
	MaxST      float64 // max short term
	MinST      float64 // max short term
	STSum      float64
	TotalCount int
	NaNCount   int
	AboveST    int
	BelowST    int
	EqualST    int
}

func (o *TLoudnessInfo) String() string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf("I: %s,  RA: %s,  TP: %s,  TH: %s,  MP: %s",
		colorizeI(o.I, fround(o.I)), fround(o.RA), fround(o.TP), fround(o.TH), fround(o.MP))
}

func (o *TMiscInfo) toString() string {
	if o == nil {
		return "<nil>"
	}
	count := o.TotalCount - o.NaNCount
	sum := o.STSum
	min := o.MinST - o.I
	max := o.MaxST - o.I
	k := sum/float64(count) - o.I

	totalTime := o.AboveST + o.EqualST + o.BelowST
	below := int(math.Round(100 * float64(o.BelowST) / float64(totalTime)))
	equal := int(math.Round(100 * float64(o.EqualST) / float64(totalTime)))
	above := int(math.Round(100 * float64(o.AboveST) / float64(totalTime)))
	return fmt.Sprintf("(%s, %s) k: %s, time%%%% %v<%v<%v %%%%",
		fround(min), fround(max), fround(k), below, equal, above)
}

func (o *TMiscInfo) toStringWithNaNs() string {
	const NaNSilece = -120 //TODO!!!: something with the ugly constant
	if o == nil {
		return "<nil>"
	}
	count := o.TotalCount
	sum := o.STSum + float64(o.NaNCount*NaNSilece)
	NaNPercentage := 100 * float64(o.NaNCount) / float64(o.TotalCount)
	min := o.MinST - o.I
	if o.NaNCount > 0 {
		min = math.Min(o.MinST, NaNSilece) - o.I
	}
	max := o.MaxST - o.I
	k := sum/float64(count) - o.I

	totalTime := o.AboveST + o.EqualST + o.BelowST + o.NaNCount
	below := int(math.Round(100 * float64(o.BelowST+o.NaNCount) / float64(totalTime)))
	equal := int(math.Round(100 * float64(o.EqualST) / float64(totalTime)))
	above := int(math.Round(100 * float64(o.AboveST) / float64(totalTime)))
	return fmt.Sprintf("(%s, %s) k: %s, time%%%% %v<%v<%v %%%%, NaNs: %v%%",
		fround(min), fround(max), fround(k), below, equal, above, fround(NaNPercentage))
}

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
		// feq(a.CR, b.CR, false) &&
		true {
		return true
	}
	return false
}

func normLi(li *TLoudnessInfo) {
	li.I -= li.MP
	li.TP -= li.MP
	li.TH -= li.MP
	li.MP -= li.MP
}

// CanFixLoudness -
func CanFixLoudness(li *TLoudnessInfo) bool {
	l := *li
	normLi(&l)
	if SuitableLoudness(&l) {
		// fmt.Println("--- not suitable")
		return true
	}
	return false
}

// FixLoudness -
func FixLoudness(li *TLoudnessInfo, compParams *TCompressParams) bool {
	// l := *li
	// normLi(&l)
	// if !SuitableLoudness(&l) {
	// 	// fmt.Println("--- not suitable")
	// 	return false
	// }
	if !CanFixLoudness(li) {
		return false
	}
	postAmp := targetI() - li.I
	postAmp = math.Min(postAmp, -li.MP)
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
