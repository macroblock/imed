package tagname

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/malashin/ffinfo"
)

func checkDeep(tagname *TTagname) error {
	typ, err := tagname.GetType()
	if err != nil {
		return err
	}
	switch typ {
	default:
		if err == nil {
			err = fmt.Errorf(fmtCheckError("unsupported type", typ, "", tagname.src))
		}
		return nil
	case "film", "trailer":
		format, err := tagname.Describe()
		if err != nil {
			return err
		}
		file, err := ffinfo.Probe(tagname.src)
		if err != nil {
			return err
		}

		type tduration struct {
			idx int
			dur float64
		}
		duration := []tduration{}
		videoDur := tduration{}
		realA := []TAudio{}
		realS := []string{}
		for index, s := range file.Streams {
			// dur, _ := file.StreamDuration(index)
			// fmt.Printf("#%v: %v\n", index, dur)
			switch s.CodecType {
			default:
				return fmt.Errorf(fmtCheckError("unsupported codec type", s.CodecType, "", tagname.src))
			case "video":
				dur, err := file.StreamDuration(index)
				if dur < 0 {
					return fmt.Errorf("stream #%v of file %q: %v", index, tagname.src, err)
				}
				log.Warningf(err, tagname.src)
				videoDur = tduration{idx: index, dur: dur}

				if index != 0 {
					return fmt.Errorf(fmtCheckError("index of the video stream", "0", strconv.Itoa(index), tagname.src))
				}
				realRes := TResolution{s.Width, s.Height}
				if format.resolution != realRes {
					return fmt.Errorf(fmtCheckError("resolution", format.resolution.String(), realRes.String(), tagname.src))
				}
				sar := s.SampleAspectRatio
				// fix ffmpeg SAR
				switch sar {
				case "", "0:1":
					sar = "1:1"
				}
				if format.Sar != "" && format.Sar != sar {
					return fmt.Errorf(fmtCheckError("SAR", format.Sar, sar, tagname.src))
				}
			case "audio":
				dur, err := file.StreamDuration(index)
				if dur < 0 {
					return fmt.Errorf("get stream duration of stream #%v of file %q: %v", index, tagname.src, err)
				}
				log.Warningf(err, tagname.src)
				duration = append(duration, tduration{idx: index, dur: dur})

				lang := s.Tags.Language
				if lang == "" || len(lang) != 3 {
					lang = "---"
				}
				realA = append(realA, TAudio{lang, s.Channels})
			case "subtitle":
				lang := s.Tags.Language
				if lang == "" || len(lang) != 3 {
					lang = "---"
				}
				realS = append(realS, lang)
			}
		}

		if len(realA) == 1 && realA[0].language != "---" {
			realA[0].language = "---"
		}
		if len(format.Audio) == 1 && format.Audio[0].language != "---" {
			format.Audio[0].language = "---"
		}
		a1 := audioToStr(format.Audio)
		a2 := audioToStr(realA)
		if a1 != a2 {
			return fmt.Errorf(fmtCheckError("audio", a1, a2, tagname.src))
		}
		s1 := strings.Join(format.Subtitle, " ")
		s2 := strings.Join(realS, " ")
		if s1 != s2 {
			return fmt.Errorf(fmtCheckError("subtitle", s1, s2, tagname.src))
		}

		ok := true
		for _, v := range duration {
			diff := videoDur.dur - v.dur
			if diff < -0.1281 || 0.0221 < diff { // granularity for mpg... it is just an empirical observation
				ok = false
			}
		}
		if !ok {
			errStr := "different stream duration:"
			for _, v := range duration {
				diff := videoDur.dur - v.dur
				errStr += fmt.Sprintf("\n            #%v: %.4f seconds (%+.4f)", v.idx, v.dur, diff)
			}
			return fmt.Errorf(errStr)
		}
	} // switch typ
	return nil
}

func audioToStr(a []TAudio) string {
	ret := ""
	for _, v := range a {
		ret += fmt.Sprintf("%v%v", v.language, v.channels)
	}
	return ret
}

func fmtCheckError(title, a, b, filename string) string {
	switch {
	case a == "" && b == "":
		return fmt.Sprintf("%v\n            %v", filename, title)
	case a != "" && b != "":
		return fmt.Sprintf("%v\n            %v: %v != %v", filename, title, a, b)
	}
	if a == "" {
		a = b
	}
	return fmt.Sprintf("%v\n            %v: %v", filename, title, a)
}
