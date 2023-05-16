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
	return o.FormatString(false)
}

// FormatString -
func (o *TLoudnessInfo) FormatString(colorize bool) string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf("I: %s,  RA: %s,  TP: %s,  TH: %s,  MP: %s",
		colorizeI(colorize, o.I, fround(o.I)), fround(o.RA), fround(o.TP), fround(o.TH), fround(o.MP))
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

func targetLimit() float64 {
	ret := 0.0
	if !math.IsNaN(targetMP()) {
		ret = targetMP()
	}
	if !math.IsNaN(targetTP()) {
		ret = math.Min(ret, targetTP())
	}
	return ret
}

const loudnessDeltaLI = 0.5

// IsValid -
func (o *TLoudnessInfo) IsValid() bool {
	if o == nil {
		return false
	}
	if o.I <= targetIMin() || targetIMax() <= o.I {
		return false
	}
	if !math.IsNaN(targetTP()) && o.TP > targetTP() {
		return false
	}
	if !math.IsNaN(targetTP()) && o.RA > targetLRA() {
		return false
	}
	return true
}

// IsSuitable -
func (o *TLoudnessInfo) IsSuitable() bool {
	if GlobalDebug {
		fmt.Printf("## suitable?:\n"+
			"    I %0.3v,  tI %0.3v\n"+
			"   MP %0.3v, tMP %0.3v\n"+
			"   TP %0.3v, tTP %0.3v\n"+
			"   RA %0.3v, tRA %0.3v\n",
			o.I, targetIMin(),
			o.MP, targetMP(),
			o.TP, targetTP(),
			o.RA, targetLRA())
	}
	if o == nil {
		return false
	}
	if o.I <= targetIMin() {
		//fmt.Println("## IMin")
		return false
	}
	if !math.IsNaN(targetMP()) && o.MP > targetMP() {
		//fmt.Println("## MP")
		return false
	}
	if !math.IsNaN(targetTP()) && o.TP > targetTP() {
		//fmt.Println("## TP")
		return false
	}
	if !math.IsNaN(targetLRA()) && o.RA > targetLRA() {
		//fmt.Println("## LRA")
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

// // IsEqual -
// func (o *TLoudnessInfo) IsEqual(v *TLoudnessInfo) bool {
// 	if GlobalDebug {
// 		fmt.Printf("@@a: %v\n", o)
// 		fmt.Printf("@@b: %v\n", v)
// 	}
// 	if feq(o.I, v.I, false) &&
// 		feq(o.RA, v.RA, isNaN(o.RA, v.RA)) &&
// 		feq(o.TP, v.TP, isNaN(o.TP, v.TP)) &&
// 		feq(o.MP, v.MP, false) &&
// 		feq(o.TH, v.TH, false) &&
// 		// feq(a.CR, b.CR, false) &&
// 		true {
// 		return true
// 	}
// 	return false
// }

// Normalize -
func (o *TLoudnessInfo) Normalize() {
	// o.I -= o.MP
	// o.TP -= o.MP
	// o.TH -= o.MP
	// o.MP -= o.MP
	o.Amp(o.Headroom())
}

func (o *TLoudnessInfo) Headroom() float64 {
	ret := -o.MP
	if !math.IsNaN(o.TP) {
		ret = math.Min(ret, -o.TP)
	}
	ret += targetLimit()
	return ret
}

// Amp -
func (o *TLoudnessInfo) Amp(postAmp float64) {
	o.I += postAmp
	o.TP += postAmp
	o.TH += postAmp
	o.MP += postAmp
}

// CanFix -
func (o *TLoudnessInfo) CanFix() bool {
	l := &TLoudnessInfo{}
	*l = *o
	l.Normalize()
	if l.IsSuitable() {
		//fmt.Println("## suitable")
		return true
	}
	//fmt.Println("## not suitable")
	return false
}

func (o *TLoudnessInfo) calcPostAmp() (float64, bool) {
	if !o.CanFix() {
		return 0.0, false
	}
	postAmp := targetI() - o.I
	postAmp = math.Min(postAmp, o.Headroom()) //-o.MP)
	return postAmp, true
}

// FixAmp -
func (o *TLoudnessInfo) FixAmp() (float64, bool) {
	postAmp, ok := o.calcPostAmp()
	if !ok {
		return 0.0, false
	}
	o.Amp(postAmp)
	return postAmp, true
}

// // FixLoudnessPostAmp -
// func FixLoudnessPostAmp(li *TLoudnessInfo, compParams *TCompressParams) bool {
// 	postAmp, ok := li.calcPostAmp()
// 	if !ok {
// 		return false
// 	}
// 	compParams.PostAmp += postAmp
// 	li.Amp(postAmp)
// 	// fmt.Println("##### stream:", i,
// 	// 	"\n  li      >", li,
// 	// 	"\n  postAmp >", stream.CompParams.PostAmp)
// 	return true
// }
