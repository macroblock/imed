package loudnorm

import (
	"fmt"
)

const peakSafeZone = -0.1

// TCompressParams -
type TCompressParams struct {
	li                     TLoudnessInfo
	PreAmp, PostAmp, Ratio float64
	Correction             float64
}

func newEmptyCompressParams() *TCompressParams {
	return &TCompressParams{Ratio: -1.0, Correction: 1.0}
}

func newCompressParams(li *TLoudnessInfo) *TCompressParams {
	cp := newEmptyCompressParams()
	cp.li = *li
	offs := li.I - targetI()
	if offs >= li.MP {
		cp.PreAmp = -offs
		return cp
	}
	offs = li.MP
	k := targetI() / (li.I - offs)
	cp.PreAmp = -offs
	cp.Ratio = k
	return cp
}

// String -
func (o *TCompressParams) String() string {
	if o == nil {
		return "<nil>"
	}
	ret := fmt.Sprintf("[%s, %s, %s]", fdown(o.PreAmp), froundRatio(o.GetK()), fdown(o.PostAmp))
	return ret
}

// 0.3:1:-30/-30|-20/-5|0/-3:6:0:-90:0.3
func (o *TCompressParams) filterPro() string {
	r := o.GetK() //o.Ratio * o.Correction
	atk := 0.3
	rls := 1.0
	TH0 := o.li.TH - 10 // 10dB seems to be constant value
	TH := o.li.TH
	overhead := 0.0
	CLow := (TH - -overhead)*r + -overhead
	CHigh := -overhead
	limit := -0.1
	ret := fmt.Sprintf("%s:%s:", fround(atk), fround(rls)) +
		fmt.Sprintf("%s/%s|", fdown(TH0), fdown(TH0)) +
		fmt.Sprintf("%s/%s|%s/%s|20/%s:", fdown(TH), fdown(CLow), fdown(CHigh), fdown(CHigh), fdown(CHigh)) +
		// fmt.Sprintf("6:%v:0:%v", -overhead, rls) +
		fmt.Sprintf("6:%s:0:%s", fround(0.0), fround(atk)) +
		fmt.Sprintf(",compand=attacks=%s:points=-80/-80|%s/%s|20/%s", fround(0), fdown(limit), fdown(limit), fdown(limit)) +
		""
	if false {
		return ret
	}
	CHigh = TH - TH*r
	limit = CHigh
	ret = fmt.Sprintf("%s:%s:", fround(atk), fround(rls)) +
		fmt.Sprintf("%s/%s|", fdown(TH), fdown(TH)) +
		fmt.Sprintf("%s/%s|20/%s:", fdown(0.0), fdown(CHigh), fdown(CHigh)) +
		// fmt.Sprintf("6:%v:0:%v", -overhead, rls) +
		fmt.Sprintf("6:%s:%s:%s", fround(0.0), fdown(-90.0), fround(atk)) +
		fmt.Sprintf(",compand=attacks=%s:points=-80/-80|%s/%s|20/%s", fround(0), fdown(limit), fdown(limit), fdown(limit)) +
		// fmt.Sprintf(",compand=attacks=0:decays=0:points=-80/-80|%s/%s|20/%s", fdown(limit), fdown(limit), fdown(limit)) +
		""

	// fmt.Printf("--- CHigh: %v\n", CHigh)

	return ret
}

// BuildFilter -
func (o *TCompressParams) BuildFilter() string {
	if o == nil {
		return "anull"
	}
	if o.Ratio < 0.0 {
		return fmt.Sprintf("volume=%sdB", fdown(o.PreAmp+o.PostAmp))
	}
	// r := o.Ratio * o.Correction
	// ret := fmt.Sprintf("volume=%.4fdB,compand=attacks=%v:decays=%v:"+
	// 	"points=-90/-%.4f|0/0|90/0",
	// 	o.PreAmp,
	// 	settings.Compressor.Attack,
	// 	settings.Compressor.Release,
	// 	90.0*r)
	ret := fmt.Sprintf("volume=%sdB,compand=%v", fdown(o.PreAmp), o.filterPro())
	if o.PostAmp != 0.0 {
		ret += fmt.Sprintf(",volume=%sdB", fdown(o.PostAmp))
	}
	return ret
}

// GetK -
func (o *TCompressParams) GetK() float64 {
	// if o.Ratio < 0.0 {
	// 	return 1.0
	// }
	ret := o.Ratio * o.Correction
	return ret
}
