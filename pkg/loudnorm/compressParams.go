package loudnorm

import (
	"fmt"
	"strconv"
)

// TCompressParams -
type TCompressParams struct {
	li                     TLoudnessInfo
	PreAmp, PostAmp, Ratio float64
	Correction             float64
}

func newEmptyCompressParams() *TCompressParams {
	return &TCompressParams{Ratio: -1.0}
}

func newCompressParams(li *TLoudnessInfo) *TCompressParams {
	diffLU := targetI() - li.I
	if diffLU <= 0.0 {
		return &TCompressParams{li: *li, PreAmp: diffLU, PostAmp: 0.0, Ratio: -1.0, Correction: 1.0}
	}
	exceededVal := li.MP + diffLU
	if exceededVal <= 0.0 {
		return &TCompressParams{li: *li, PreAmp: diffLU, PostAmp: 0.0, Ratio: -1.0, Correction: 1.0}
	}
	offs := -li.MP
	k := targetI() / (li.I + offs)
	return &TCompressParams{li: *li, PreAmp: offs, PostAmp: 0.0, Ratio: k, Correction: 1.0}
}

// String -
func (o *TCompressParams) String() string {
	if o == nil {
		return "<nil>"
	}
	ret := ""
	ret += "[" + strconv.FormatFloat(o.PreAmp, 'f', 2, 64) + ","
	ret += " " + strconv.FormatFloat(1/o.GetK(), 'f', 2, 64) + ":1,"
	ret += " " + strconv.FormatFloat(o.PostAmp, 'f', 2, 64) + ""
	ret += "]"
	return ret
}

// 0.3:1:-30/-30|-20/-5|0/-3:6:0:-90:0.3
func (o *TCompressParams) filterPro() string {
	r := o.Ratio * o.Correction
	atk := 0.3
	rls := 1.0
	TH0 := o.li.TH - 10 // 10dB seems to be constant value
	TH := o.li.TH
	overhead := 0.0
	CLow := (TH - -overhead)*r + -overhead
	CHigh := -overhead
	limit := -0.1
	ret := fmt.Sprintf("%v:%v:", atk, rls) +
		fmt.Sprintf("%v/%v|", TH0, TH0) +
		fmt.Sprintf("%v/%v|%v/%v|20/%v:", TH, CLow, CHigh, CHigh, CHigh) +
		// fmt.Sprintf("6:%v:0:%v", -overhead, rls) +
		fmt.Sprintf("6:%v:0:%v", 0, rls) +
		// fmt.Sprintf(",alimiter=attack=%v:release=%v:level_in=%vdB:level_out=%vdB:level=true", atk, rls, -overhead/2, -overhead/2)+
		// fmt.Sprintf(",alimiter=level_in=%vdB:level_out=%vdB:level=false", -1.0, -1.5) +
		// fmt.Sprintf(",alimiter=level_in=%v:level_out=%v:level=false", 1.0, 1.0) +
		// fmt.Sprintf(",alimiter=attack=%v:release=%v:level_in=%v:level_out=%v:level=true", 50, 100, 0.95, 1.0) + // try atk:7 rls:100
		fmt.Sprintf(",compand=attacks=%v:points=-80/-80|%v/%v|20/%v", 0, limit, limit, limit) +
		""

	return ret
}

// BuildFilter -
func (o *TCompressParams) BuildFilter() string {
	if o == nil {
		return "anull"
	}
	if o.Ratio < 0.0 {
		return fmt.Sprintf("volume=%.4fdB", o.PreAmp+o.PostAmp)
	}
	// r := o.Ratio * o.Correction
	// ret := fmt.Sprintf("volume=%.4fdB,compand=attacks=%v:decays=%v:"+
	// 	"points=-90/-%.4f|0/0|90/0",
	// 	o.PreAmp,
	// 	settings.Compressor.Attack,
	// 	settings.Compressor.Release,
	// 	90.0*r)
	ret := fmt.Sprintf("volume=%.4fdB,compand=%v", o.PreAmp, o.filterPro())
	if o.PostAmp != 0.0 {
		ret += fmt.Sprintf(",volume=%.4fdB", o.PostAmp)
	}
	return ret
}

// GetK -
func (o *TCompressParams) GetK() float64 {
	if o.Ratio < 0.0 {
		return 1.0
	}
	ret := o.Ratio * o.Correction
	return ret
}
