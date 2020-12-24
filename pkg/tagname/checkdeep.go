package tagname

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/malashin/ffinfo"
)

func parseSize(str string) (int, int, error) {
	retErr := fmt.Errorf("Invalid size format (want: '\\d+x\\d+' have: %q", str)
	list := strings.Split(str, "x")
	if len(list) != 2 {
		return -1, -1,  retErr
	}
	w, err := strconv.Atoi(list[0])
	if err != nil {
		return -1, -1,  retErr
	}
	h, err := strconv.Atoi(list[1])
	if err != nil {
		return -1, -1,  retErr
	}
	return w, h, nil
}

func checkSize(tn *TTagname, typ string, width, height int) error {
	switch typ {
	case "poster", "poster.gp":
		size, err := tn.GetTag("sizetag")
		if err != nil {
			return err
		}
		if size == "logo" {
			if 900 > width || width > 1500 {
				return fmt.Errorf("Improper size (want width<=1500, have width=%v)", width)
			}
			return nil
		}
		w, h, err := parseSize(size)
		if err != nil {
			return err
		}
		if w != width || h != height {
			return fmt.Errorf("Improper size (want %vx%v, have %vx%v)", width, height, w, h)
		}
	}
	return nil
}

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
	case "poster", "poster.logo", "poster.gp":
		filePath := filepath.Join(tagname.dir, tagname.src)
		file, err := ffinfo.Probe(filePath)
		if err != nil {
			return err
		}
		w, h := 0, 0
		if len(file.Streams)>0 && file.Streams[0].CodecType == "video" {
			w = file.Streams[0].Width
			h = file.Streams[0].Height
		} else {
			return fmt.Errorf("failed to read file canvas size")
		}
		err = checkSize(tagname, typ, w, h)
		if err != nil {
			return err
		}

	case "film", "trailer":
		format, err := tagname.Describe()
		if err != nil {
			return err
		}
		filePath := filepath.Join(tagname.dir, tagname.src)
		file, err := ffinfo.Probe(filePath)
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

		if len(realA) == 1 && realA[0].Language != "---" {
			realA[0].Language = "---"
		}
		if len(format.Audio) == 1 && format.Audio[0].Language != "---" {
			format.Audio[0].Language = "---"
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
			if diff < -0.385 /* */ || 0.385 < diff { // granularity for mpg... it is just an empirical observation
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
		ret += fmt.Sprintf("%v%v", v.Language, v.Channels)
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
