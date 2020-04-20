package loudnorm

import (
	"fmt"
	"os"
	"strconv"

	"github.com/macroblock/imed/pkg/ffmpeg"
)

type optionsJSON struct {
	InputI            string `json:"input_i"`
	InputTP           string `json:"input_tp"`
	InputLRA          string `json:"input_lra"`
	InputThresh       string `json:"input_thresh"`
	OutputI           string `json:"output_i"`
	OutputTP          string `json:"output_tp"`
	OutputThresh      string `json:"output_thresh"`
	NormalizationType string `json:"normalization_type"`
	TargetOffset      string `json:"target_offset"`
}

// // Options -
// type Options struct {
// 	InputI            float64
// 	InputTP           float64
// 	InputLRA          float64
// 	InputThresh       float64
// 	OutputI           float64
// 	OutputTP          float64
// 	OutputThresh      float64
// 	NormalizationType string
// 	TargetOffset      float64
// }

// OptionsLight -
type OptionsLight struct {
	InputI      float64
	InputThresh float64

	InputLRA     float64
	InputThresh2 float64
	InputLRALow  float64
	InputLRAHigh float64

	InputTP float64
}

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
	return fmt.Sprintf("I: %v, RA: %v, TP: %v, TH: %v, MP: %v", o.I, o.RA, o.TP, o.TH, o.MP)
}

// SetTargetLI -
func SetTargetLI(li float64) {
	targetI = li
}

// SetTargetLRA -
func SetTargetLRA(lra float64) {
	targetLRA = lra
}

// SetTargetTP -
func SetTargetTP(tp float64) {
	targetTP = tp
}

// Scan -
func Scan(streams []*TStreamInfo) error {
	if len(streams) == 0 {
		return nil
	}
	params := []string{
		"-hide_banner",
		"-i", streams[0].Parent.Filename,
		"-filter_complex",
	}
	outputs := []string{}

	time, err := ffmpeg.ParseTime("11:22:33.44")
	if err != nil {
		return err
	}
	combParser := ffmpeg.NewCombineParser(
		ffmpeg.NewAudioProgressParser(time, nil),
	)

	ffmpeg.UniqueReset()
	filter := ""
	n := 0
	for _, stream := range streams {
		vdFilter := ffmpeg.UniqueName("volumedetect")
		eburFilter := ffmpeg.UniqueName("ebur128")
		output := fmt.Sprintf("[o%v]", n)
		filter += fmt.Sprintf("[0:%v]%v,%v=peak=true%v", stream.Index, vdFilter, eburFilter, output)
		n++
		outputs = append(outputs, "-map", output, "-f", "null", os.DevNull)

		stream.volumeInfo = &ffmpeg.TVolumeInfo{}
		stream.eburInfo = &ffmpeg.TEburInfo{}
		combParser.Append(
			ffmpeg.NewVolumeParser(vdFilter, stream.volumeInfo),
			ffmpeg.NewEburParser(eburFilter, true, stream.eburInfo),
		)
	}
	params = append(params, filter)
	params = append(params, outputs...)

	if GlobalDebug {
		fmt.Println("### params: ", params)
	}

	err = ffmpeg.Run(combParser, params...)
	if err != nil {
		return err
	}

	for i, stream := range streams {
		stream.LoudnessInfo = &TLoudnessInfo{
			I:  stream.eburInfo.I,
			RA: stream.eburInfo.LRA,
			TP: stream.eburInfo.TP,
			TH: stream.eburInfo.Thresh,
			MP: stream.volumeInfo.MaxVolume,
			CR: -1.0,
		}
		fmt.Println("##### stream:", i,
			"\n  ebur >", stream.eburInfo,
			"\n  vol  >", stream.volumeInfo)
	}
	return nil
}

// TCompressParams -
type TCompressParams struct {
	PreAmp, PostAmp, Ratio float64
	Correction             float64
}

// BuildFilter -
func (o TCompressParams) BuildFilter() string {
	r := o.Ratio * o.Correction
	if r < 0.0 {
		return fmt.Sprintf("volume=%.4fdB", o.PreAmp+o.PostAmp)
	}
	ret := fmt.Sprintf("volume=%.4fdB,compand=0:0.01:-90/-%.4f|0/0", o.PreAmp, 90.0*r)
	if o.PostAmp != 0.0 {
		ret = fmt.Sprintf("%v,volume=%.4fdB", ret, o.PostAmp)
	}
	return ret
}

// GetK -
func (o TCompressParams) GetK() float64 {
	ret := o.Ratio * o.Correction
	if ret < 0.0 {
		return 1.0
	}
	return ret
}

func calcCompressParams(li *TLoudnessInfo) *TCompressParams {
	diffLU := targetI - li.I
	if diffLU <= 0.0 {
		return &TCompressParams{PreAmp: diffLU, PostAmp: 0.0, Ratio: -1.0, Correction: 1.0}
	}
	exceededVal := li.MP + diffLU
	if exceededVal <= 0.0 {
		return &TCompressParams{PreAmp: diffLU, PostAmp: 0.0, Ratio: -1.0, Correction: 1.0}
	}
	offs := -li.MP
	k := targetI / (li.I + offs)
	return &TCompressParams{PreAmp: offs, PostAmp: 0.0, Ratio: k, Correction: 1.0}
}

// RenderParameters -
func RenderParameters(streams []*TStreamInfo) error {
	const compressCorrectionStep = 0.05
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
		params := []string{
			"-hide_banner",
			"-i", streams[0].Parent.Filename,
			"-filter_complex",
		}
		outputs := []string{}

		time, err := ffmpeg.ParseTime("11:22:33.44")
		if err != nil {
			return err
		}
		combParser := ffmpeg.NewCombineParser(
			ffmpeg.NewAudioProgressParser(time, nil),
		)
		ffmpeg.UniqueReset()
		filter := ""
		n := 0
		for _, stream := range streams {
			if stream.done {
				continue
			}
			stream.CompParams.Correction -= compressCorrectionStep
			vdFilter := ffmpeg.UniqueName("volumedetect")
			eburFilter := ffmpeg.UniqueName("ebur128")
			output := fmt.Sprintf("[o%v]", n)
			filter += fmt.Sprintf("[0:%v]%v,%v,%v=peak=true%v",
				stream.Index, stream.CompParams.BuildFilter(), vdFilter, eburFilter, output)
			n++
			outputs = append(outputs, "-map", output, "-f", "null", os.DevNull)

			stream.volumeInfo = &ffmpeg.TVolumeInfo{}
			stream.eburInfo = &ffmpeg.TEburInfo{}
			combParser.Append(
				ffmpeg.NewVolumeParser(vdFilter, stream.volumeInfo),
				ffmpeg.NewEburParser(eburFilter, true, stream.eburInfo),
			)
		}
		params = append(params, filter)
		params = append(params, outputs...)

		if GlobalDebug {
			fmt.Println("### params: ", params)
		}
		err = ffmpeg.Run(combParser, params...)
		if err != nil {
			return err
		}

		done := true
		for i, stream := range streams {
			stream.TargetLI = &TLoudnessInfo{
				I:  stream.eburInfo.I,
				RA: stream.eburInfo.LRA,
				TP: stream.eburInfo.TP,
				TH: stream.eburInfo.Thresh,
				MP: stream.volumeInfo.MaxVolume,
				CR: stream.CompParams.GetK(),
			}
			fmt.Println("##### stream:", i,
				"\n  ebur >", stream.eburInfo,
				"\n  vol  >", stream.volumeInfo,
				"\n  K    >", stream.CompParams.GetK(),
				"\n  CR   >", 1/stream.CompParams.GetK(), ": 1")
			if SuitableLoudness(stream.TargetLI) {
				postAmp := targetI - stream.TargetLI.I
				if postAmp > 0.0 {
					postAmp = 0.0
				}
				stream.CompParams.PostAmp = postAmp
				stream.TargetLI.I += postAmp
				stream.TargetLI.TP += postAmp
				stream.TargetLI.TH += postAmp
				stream.TargetLI.MP += postAmp
				stream.done = true
				fmt.Println("##### stream:", i,
					"\n  li      >", stream.TargetLI,
					"\n  postAmp >", stream.CompParams.PostAmp)
			}
			done = done && stream.done
		}
		if done {
			return nil
		}
	}
	return fmt.Errorf("not enough tries")
}

// Normalize -
func Normalize(filePath string, trackN int, li *TLoudnessInfo) (*TCompressParams, error) {
	comp := calcCompressParams(li)

	// s := fmt.Sprintf("%v.5", k)
	// compStr := fmt.Sprintf("compand=0:0.01:-90/%v|0/0", s)

	for tries := 5; tries > 0; tries-- {
		comp.Correction -= 0.1
		filter := comp.BuildFilter()

		fmt.Println("filter: ", filter)
		params := []string{
			"-hide_banner",
			"-i", filePath,
			"-map", "0:" + strconv.Itoa(trackN),
			"-filter:a",
			"" +
				filter +
				",volumedetect" +
				",ebur128" +
				"=peak=true" +
				"",
			"-f", "null",
			osNullDevice,
		}
		if GlobalDebug {
			fmt.Println("### params: ", params)
		}
		time, err := ffmpeg.ParseTime("11:22:33.44")
		if err != nil {
			return nil, err
		}
		eburParser := ffmpeg.NewEburParser("replace me", true, nil)
		volumeParser := ffmpeg.NewVolumeParser("replace me", nil)
		err = ffmpeg.Run(
			ffmpeg.NewCombineParser(
				ffmpeg.NewAudioProgressParser(time, nil),
				volumeParser,
				eburParser,
			),
			params...,
		)
		if err != nil {
			return nil, err
		}

		eburInfo := eburParser.GetData()
		// if err != nil {
		// 	return nil, err
		// }
		// volumeInfo, err := volumeParser.GetData()
		// if err != nil {
		// 	return nil, err
		// }

		fmt.Printf("ebur: %v\n", eburInfo)
		fmt.Printf("ti-0.5: %v\nti+0.5: %v\n", targetI-0.5, targetI+0.5)

		if targetI-0.5 < eburInfo.I && eburInfo.I < targetI+0.5 {
			// return &TLoudnessInfo{
			// 	I:  eburInfo.I,
			// 	RA: eburInfo.LRA,
			// 	TP: eburInfo.TP,
			// 	TH: eburInfo.Thresh,
			// 	MP: volumeInfo.MaxVolume,
			// 	CR: k * correction,
			// }, nil
			comp.PostAmp = targetI - eburInfo.I
			if comp.PostAmp > 0.0 {
				comp.PostAmp = 0.0
			}
			return comp, nil
		}
	}
	return nil, fmt.Errorf("max tries with no result")
}

// NormalizeTo -
// func NormalizeTo(filePath string, trackN int, fileOut string, audioParams []string, inputI, inputLRA, inputTP, inputThresh float64) (*Options, error) {
// 	params := []string{
// 		"-y",
// 		"-hide_banner",
// 		"-i", filePath,
// 		"-map", "0:" + strconv.Itoa(trackN),
// 		"-filter:a",
// 		"loudnorm=print_format=json" +
// 			":linear=true" +
// 			// ":linear=false" +
// 			fmt.Sprintf(":I=% 6.2f:LRA=% 6.2f:TP=% 6.2f",
// 				targetI, targetLRA, targetTP) +
// 			fmt.Sprintf(":measured_I=% 6.2f:measured_LRA=% 6.2f:measured_TP=% 6.2f:measured_thresh=% 6.2f",
// 				inputI, inputLRA, inputTP, inputThresh) +
// 			// ":offset=" + opts.TargetOffset,  // it's just difference between internal target_i and i_out
// 			// "-f", "flac",
// 			"",
// 	}
// 	params = append(params, audioParams...)
// 	params = append(
// 		params,
// 		"-ar:a", samplerate,
// 		fileOut,
// 	)

// 	if GlobalDebug {
// 		fmt.Println("### params: ", params)
// 	}
// 	c := exec.Command("ffmpeg", params...)
// 	var o bytes.Buffer
// 	var e bytes.Buffer
// 	c.Stdout = &o
// 	c.Stderr = &e
// 	err := c.Run()
// 	if err != nil {
// 		fmt.Println("###:", e.String())
// 		return nil, err
// 	}

// 	list := strings.Split(e.String(), "\n")

// 	if len(list) < 12 {
// 		fmt.Println(strings.Join(list, "\n"))
// 		return nil, fmt.Errorf("size of an output info too small")
// 	}

// 	found := false
// 	jsonList := []string{}
// 	for _, line := range list {
// 		if strings.HasPrefix(line, "[Parsed_loudnorm_0 @") {
// 			found = true
// 			continue
// 		}
// 		if found {
// 			jsonList = append(jsonList, line)
// 		}
// 	}

// 	opts := &Options{}
// 	err = json.Unmarshal([]byte(strings.Join(jsonList, "\n")), &opts)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// fmt.Println(strings.Join(jsonList, "\n"))
// 	return opts, nil
// }
