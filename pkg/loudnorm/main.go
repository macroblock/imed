package loudnorm

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/macroblock/imed/pkg/ffmpeg"
)

// TLoudnessInfo -
type TLoudnessInfo struct {
	I  float64 // integrated
	RA float64 // range
	TP float64 // true peaks
	MP float64 // max peaks
	TH float64 // threshold
	CR float64 // compress ratio

	// Ebur   ffmpeg.TEburInfo
	// Volume ffmpeg.TVolumeInfo
}

func (o *TLoudnessInfo) String() string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf("I: %v, RA: %v, TP: %v, TH: %v, MP: %v, CR: %v",
		o.I, o.RA, o.TP, o.TH, o.MP,
		strconv.FormatFloat(o.CR, 'f', 2, 64))
}

func replaceStatic(pattern string, vals ...string) string {
	for _, val := range vals {
		if strings.Contains(pattern, val) {
			pattern = strings.Replace(pattern, val, fmt.Sprintf("%v", appendBranchStaticValue), -1)
			appendBranchStaticValue++
		}
	}
	return pattern
}

var appendBranchStaticValue = 0

func appendPattern(branches []string, stream *TStreamInfo, comb *ffmpeg.TCombineParser, patterns ...string) []string {
	for _, pattern := range patterns {
		if strings.Contains(pattern, "~vd~") {
			stream.volumeInfo = &ffmpeg.TVolumeInfo{}
			name := ffmpeg.UniqueName("volumedetect")
			comb.Append(ffmpeg.NewVolumeParser(name, stream.volumeInfo))
			pattern = strings.Replace(pattern, "~vd~", name, -1)
		}
		if strings.Contains(pattern, "~astats~") {
			stream.astatsInfo = &ffmpeg.TAStatsInfo{}
			name := ffmpeg.UniqueName("astats")
			comb.Append(ffmpeg.NewAStatsParser(name, stream.astatsInfo))
			pattern = strings.Replace(pattern, "~astats~", name, -1)
		}

		if strings.Contains(pattern, "~ebur~") {
			stream.eburInfo = &ffmpeg.TEburInfo{}
			name := ffmpeg.UniqueName("ebur128")
			params := "=peak=none" // none sample true
			// TruePeaksOn := false
			useTP := !math.IsNaN(targetTP())
			if useTP {
				params = "=peak=true"
			}
			comb.Append(ffmpeg.NewEburParser(name, useTP, stream.eburInfo))
			name += params
			pattern = strings.Replace(pattern, "~ebur~", name, -1)
		}

		if strings.Contains(pattern, "~compressor~") {
			pattern = strings.Replace(pattern, "~compressor~", stream.CompParams.BuildFilter(), -1)
		}

		if strings.Contains(pattern, "~idx~") {
			pattern = strings.Replace(pattern, "~idx~", fmt.Sprintf("%v", stream.Index), -1)
		}

		pattern = replaceStatic(pattern, "~u~", "~u0~", "~u1~", "~u2~", "~u3~")

		branches = append(branches, pattern)
	}

	return branches
}

// Scan -
func Scan(streams []*TStreamInfo) error {
	if len(streams) == 0 {
		return nil
	}
	params := []string{"-hide_banner"}
	params = append(params, getGloblaFlags()...)
	params = append(params, "-i", streams[0].Parent.Filename)

	time := ffmpeg.FloatToTime(streams[0].Parent.Duration)
	combParser := ffmpeg.NewCombineParser(
		ffmpeg.NewAudioProgressParser(time, nil),
	)

	ffmpeg.UniqueReset()
	outputs := []string{}
	filters := []string{}
	for _, stream := range streams {
		filters = appendPattern(filters, stream, combParser, "[0:~idx~]~astats~,~vd~,~ebur~[o~idx~]")
		outputs = appendPattern(outputs, stream, nil, "-map", "[o~idx~]", "-f", "null", os.DevNull)
	}
	params = append(params, "-filter_complex")
	params = append(params, strings.Join(filters, ";"))
	params = append(params, outputs...)

	if GlobalDebug {
		fmt.Println("### params: ", params)
	}

	err := ffmpeg.Run(combParser, params...)
	if err != nil {
		return err
	}

	for i, stream := range streams {
		stream.LoudnessInfo = &TLoudnessInfo{
			I:  stream.eburInfo.I,
			RA: stream.eburInfo.LRA,
			TP: stream.eburInfo.TP,
			TH: stream.eburInfo.Thresh,
			// MP: stream.volumeInfo.MaxVolume,
			MP: stream.astatsInfo.PeakLevel,
			CR: -1.0,
		}
		stream.TargetLI = &TLoudnessInfo{}
		*stream.TargetLI = *stream.LoudnessInfo

		if GlobalDebug {
			fmt.Println("##### stream:", i,
				"\n  ebur >", stream.eburInfo,
				"\n  vol  >", stream.volumeInfo,
				"\n  stat >", stream.astatsInfo)
		}

		comp := &TCompressParams{Ratio: -1.0}
		stream.CompParams = comp

		stream.done = FixLoudness(stream.TargetLI, stream.CompParams)
		if GlobalDebug && stream.done {
			fmt.Println("##### fixed stream:", i,
				"\n  li   >", stream.TargetLI,
				"\n  comp >", stream.CompParams)
		}
	}
	return nil
}

// TCompressParams -
type TCompressParams struct {
	li                     TLoudnessInfo
	PreAmp, PostAmp, Ratio float64
	Correction             float64
}

func newCompressParams() *TCompressParams {
	return &TCompressParams{Ratio: -1.0}
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
	overhead := 3.0
	CLow := (TH - -overhead)*r + -overhead
	CHigh := -overhead
	ret := fmt.Sprintf("%v:%v:", atk, rls) +
		fmt.Sprintf("%v/%v|", TH0, TH0) +
		fmt.Sprintf("%v/%v|%v/%v|100/%v:", TH, CLow, CHigh, CHigh, CHigh) +
		// fmt.Sprintf("6:%v:0:%v", -overhead, rls) +
		fmt.Sprintf("6:%v:0:%v", 0, rls) +
		// fmt.Sprintf(",alimiter=attack=%v:release=%v:level_in=%vdB:level_out=%vdB:level=true", atk, rls, -overhead/2, -overhead/2)+
		// fmt.Sprintf(",alimiter=level_in=%vdB:level_out=%vdB:level=false", -1.0, -1.5) +
		// fmt.Sprintf(",alimiter=level_in=%v:level_out=%v:level=false", 1.0, 1.0) +
		fmt.Sprintf(",alimiter=attack=%v:release=%v:level_in=%v:level_out=%v:level=true", 50, 100, 0.95, 1.0) + // try atk:7 rls:100
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
	r := o.Ratio * o.Correction
	ret := fmt.Sprintf("volume=%.4fdB,compand=attacks=%v:decays=%v:"+
		"points=-90/-%.4f|0/0|90/0",
		o.PreAmp,
		settings.Compressor.Attack,
		settings.Compressor.Release,
		90.0*r)
	ret = fmt.Sprintf("volume=%.4fdB,compand=%v", o.PreAmp, o.filterPro())
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

func calcCompressParams(li *TLoudnessInfo) *TCompressParams {

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

// RenderParameters -
func RenderParameters(streams []*TStreamInfo) error {
	if len(streams) == 0 {
		return nil
	}

	for _, stream := range streams {
		if stream.LoudnessInfo == nil {
			return fmt.Errorf("stream %v:%v has no loudness info", stream.Parent.Filename, stream.Index)
		}
		comp := calcCompressParams(stream.LoudnessInfo)
		stream.CompParams = comp
	}

	for tries := 5; tries > 0; tries-- {

		params := []string{"-hide_banner"}
		params = append(params, getGloblaFlags()...)
		params = append(params, "-i", streams[0].Parent.Filename)

		time := ffmpeg.FloatToTime(streams[0].Parent.Duration)
		combParser := ffmpeg.NewCombineParser(
			ffmpeg.NewAudioProgressParser(time, nil),
		)

		done := true
		// first := true
		ffmpeg.UniqueReset()
		outputs := []string{}
		filters := []string{}
		for _, stream := range streams {
			if stream.done {
				continue
			}
			done = false
			// if !first {
			stream.CompParams.Correction -= settings.Compressor.CorrectionStep
			// first = false
			// }
			filters = appendPattern(filters, stream, combParser, "[0:~idx~]~compressor~,~astats~,~vd~,~ebur~[o~idx~]")
			outputs = appendPattern(outputs, stream, nil, "-map", "[o~idx~]", "-f", "null", os.DevNull)

			fmt.Printf("        #%v: %v\n          : %v\n", stream.Index, stream.TargetLI, stream.CompParams)
		}
		if done {
			// fmt.Println("--- All ok. continue ---")
			return nil
		}
		params = append(params, "-filter_complex")
		params = append(params, strings.Join(filters, ";"))
		params = append(params, outputs...)

		if GlobalDebug {
			fmt.Println("### params: ", params)
		}
		err := ffmpeg.Run(combParser, params...)
		if err != nil {
			return err
		}

		done = true
		for i, stream := range streams {
			stream.TargetLI = &TLoudnessInfo{
				I:  stream.eburInfo.I,
				RA: stream.eburInfo.LRA,
				TP: stream.eburInfo.TP,
				TH: stream.eburInfo.Thresh,
				// MP: stream.volumeInfo.MaxVolume,
				MP: stream.astatsInfo.PeakLevel,
				CR: stream.CompParams.GetK(),
			}
			if GlobalDebug {
				fmt.Println("##### stream:", i,
					"\n  ebur >", stream.eburInfo,
					"\n  vol  >", stream.volumeInfo,
					// "\n  K    >", stream.CompParams.GetK(),
					// "\n  CR   >", 1/stream.CompParams.GetK(), ": 1")
					"\n  stats>", stream.astatsInfo,
					"\n  comp >", stream.CompParams)
			}

			stream.done = FixLoudness(stream.TargetLI, stream.CompParams)
			if GlobalDebug && stream.done {
				fmt.Println("##### fixed stream:", i,
					"\n  li   >", stream.TargetLI,
					"\n  comp >", stream.CompParams)
			}
			done = done && stream.done
		}
		if done {
			return nil
		}
	}
	return fmt.Errorf("not enough tries")
}
