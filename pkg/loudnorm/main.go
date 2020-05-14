package loudnorm

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/macroblock/imed/pkg/ffmpeg"
)

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

		if strings.Contains(pattern, "~header~") {
			if settings.Behavior.ForceStereo && stream.Channels > 2 {
				filter := "pan=stereo|FL<1.0*FL+0.707*FC+0.5*BL|FR<1.0*FR+0.707*FC+0.5*BR"
				pattern = strings.Replace(pattern, "~header~", filter+",~header~", -1)
			}
			pattern = strings.Replace(pattern, "~header~", "anull", -1)
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
	params = append(params, GetSettings().getGlobalFlags()...)
	params = append(params, "-i", streams[0].Parent.Filename)

	time := ffmpeg.FloatToTime(streams[0].Parent.Duration)
	combParser := ffmpeg.NewCombineParser(
		ffmpeg.NewAudioProgressParser(time, nil),
	)

	ffmpeg.UniqueReset()
	outputs := []string{}
	filters := []string{}
	for _, stream := range streams {
		filters = appendPattern(filters, stream, combParser, "[0:~idx~]~header~,~astats~,~ebur~,anullsink") //[o~idx~]")
		outputs = appendPattern(outputs, stream, nil /*"-map", "[o~idx~]",*/, "-f", "null", os.DevNull)
	}
	params = append(params, "-filter_complex")
	params = append(params, strings.Join(filters, ";"))
	params = append(params, outputs...)

	debugPrintf("### params: %v\n", params)

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

	for _, stream := range streams {
		printStreamParams(stream)
	}
	return nil
}

func printStreamParams(stream *TStreamInfo) {
	fmt.Printf("        #%v: %v\n", stream.Index, stream.TargetLI)
	fmt.Printf("          : compression %v\n", stream.CompParams)
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
		comp := newCompressParams(stream.LoudnessInfo)
		stream.CompParams = comp
	}

	first := true
	for tries := 5; tries > 0; tries-- {

		params := []string{"-hide_banner"}
		params = append(params, GetSettings().getGlobalFlags()...)
		params = append(params, "-i", streams[0].Parent.Filename)

		time := ffmpeg.FloatToTime(streams[0].Parent.Duration)
		combParser := ffmpeg.NewCombineParser(
			ffmpeg.NewAudioProgressParser(time, nil),
		)

		done := true
		ffmpeg.UniqueReset()
		outputs := []string{}
		filters := []string{}
		for _, stream := range streams {
			if stream.done {
				continue
			}
			done = false
			if !first {
				stream.CompParams.Correction -= settings.Compressor.CorrectionStep
			}
			first = false
			filters = appendPattern(filters, stream, combParser,
				"[0:~idx~]~header~,~compressor~,asplit[~u0~][~u1~];"+
					"[~u0~]~astats~,anullsink;"+
					"[~u1~]~ebur~,anullsink") //[o~idx~]")
			outputs = appendPattern(outputs, stream, nil /*"-map", "[o~idx~]",*/, "-f", "null", os.DevNull)
		}

		if done || settings.Behavior.ScanOnly {
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

			// fmt.Printf("--- i - mp: %v\n", stream.TargetLI.I-stream.TargetLI.MP)
			stream.done = FixLoudness(stream.TargetLI, stream.CompParams)
			if GlobalDebug && stream.done {
				fmt.Println("##### fixed stream:", i,
					"\n  li   >", stream.TargetLI,
					"\n  comp >", stream.CompParams)
			}
			done = done && stream.done
		}

		for _, stream := range streams {
			printStreamParams(stream)
		}

		if done {
			return nil
		}
	}
	return fmt.Errorf("not enough tries")
}
