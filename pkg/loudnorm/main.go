package loudnorm

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/macroblock/imed/pkg/ffmpeg"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/types"
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
			parser := ffmpeg.NewEburParser(name, useTP, stream.eburInfo)
			parser.SetOptions(settings.Loudness.STStatTHBelow, settings.Loudness.STStatTHAbove)
			comb.Append(parser)
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

// FixStreamPostAmp -
func FixStreamPostAmp(stream *TStreamInfo) bool {
	ret := false
	if gain, ok := stream.TargetLI.FixAmp(); ok {
		stream.CompParams.GainPostAmp(gain)
		ret = true
	}
	return ret
}

// Scan -
func Scan(streams []*TStreamInfo) error {
	if len(streams) == 0 {
		return nil
	}
	params := []string{"-hide_banner"}
	params = append(params, GetSettings().getGlobalFlags()...)
	params = append(params, "-i", streams[0].Parent.Filename)

	time := types.NewTimecode(0, 0, streams[0].Parent.Duration)
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

	err := ffmpeg.Run(nil, combParser, params...)
	if err != nil {
		return err
	}

	for i, stream := range streams {
		stream.LoudnessInfo, stream.MiscInfo = initInfo(stream.eburInfo, stream.astatsInfo)

		stream.TargetLI = &TLoudnessInfo{}
		*stream.TargetLI = *stream.LoudnessInfo

		if GlobalDebug {
			fmt.Println("##### stream:", i,
				"\n  ebur >", stream.eburInfo,
				"\n  vol  >", stream.volumeInfo,
				"\n  stat >", stream.astatsInfo)
		}

		// fake compress params. Just to print forecast info
		comp := newCompressParams(stream.LoudnessInfo) //&TCompressParams{Ratio: -1.0}
		stream.CompParams = comp
		// print original values
		printStreamParams(stream, true) // colorized LI
		// fmt.Println()

		// compress params to use without compression
		stream.CompParams = newEmptyCompressParams()

		// stream.done = FixLoudnessPostAmp(stream.TargetLI, stream.CompParams)
		stream.done = FixStreamPostAmp(stream)

		if GlobalDebug && stream.done {
			fmt.Println("##### fixed stream:", i,
				"\n  li   >", stream.TargetLI,
				"\n  comp >", stream.CompParams)
		}
	}

	return nil
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
		// compress params to use with compression
		comp := newCompressParams(stream.LoudnessInfo)
		stream.CompParams = comp
	}

	for tries := settings.Compressor.NumTries; tries > 0; tries-- {
		numTries := settings.Compressor.NumTries
		colorizedPrintf(misc.ColorFaint, "Attempt %v/%v:\n", numTries-tries+1, numTries)

		params := []string{"-hide_banner"}
		params = append(params, GetSettings().getGlobalFlags()...)
		params = append(params, "-i", streams[0].Parent.Filename)

		time := types.NewTimecode(0, 0, streams[0].Parent.Duration)
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
			filters = appendPattern(filters, stream, combParser,
				"[0:~idx~]~header~,~compressor~,asplit[~u0~][~u1~];"+
					"[~u0~]~astats~,anullsink;"+
					"[~u1~]~ebur~,anullsink") //[o~idx~]")
			outputs = appendPattern(outputs, stream, nil /*"-map", "[o~idx~]",*/, "-f", "null", os.DevNull)
		}

		if done {
			return nil
		}

		params = append(params, "-filter_complex")
		params = append(params, strings.Join(filters, ";"))
		params = append(params, outputs...)

		if GlobalDebug {
			fmt.Println("### params: ", params)
		}
		err := ffmpeg.Run(nil, combParser, params...)
		if err != nil {
			return err
		}

		done = true
		for i, stream := range streams {
			stream.TargetLI, stream.MiscInfo = initInfo(stream.eburInfo, stream.astatsInfo)
			// stream.TargetLI.CR = stream.CompParams.GetK()

			if GlobalDebug {
				fmt.Println("##### stream:", i,
					"\n  ebur >", stream.eburInfo,
					"\n  vol  >", stream.volumeInfo,
					// "\n  K    >", stream.CompParams.GetK(),
					// "\n  CR   >", 1/stream.CompParams.GetK(), ": 1")
					"\n  stats>", stream.astatsInfo,
					"\n  comp >", stream.CompParams)
			}

			if !stream.TargetLI.CanFix() {
				stream.CompParams.Correction -= settings.Compressor.CorrectionStep
			}
			printStreamParams(stream, false) // LI stats without color

			// fmt.Printf("--- i - mp: %v\n", stream.TargetLI.I-stream.TargetLI.MP)
			// stream.done = FixLoudnessPostAmp(stream.TargetLI, stream.CompParams)
			stream.done = FixStreamPostAmp(stream)

			if GlobalDebug && stream.done {
				fmt.Println("##### fixed stream:", i,
					"\n  li   >", stream.TargetLI,
					"\n  comp >", stream.CompParams)
			}
			done = done && stream.done
		}
		fmt.Println()

		if done {
			return nil
		}
	}
	return fmt.Errorf("not enough tries")
}
