package tagname

import (
	"fmt"

	"github.com/malashin/ffinfo"
)

var langMap = map[string]string{"rus": "r", "eng": "e"}

func convLang(lang string) string {
	ok := false
	if lang, ok = langMap[lang]; !ok {
		lang = "-"
	}
	return lang
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
	case "film", "trailer":
		format, err := tagname.Describe()
		if err != nil {
			return err
		}
		file, err := ffinfo.Probe(tagname.src)
		if err != nil {
			return err
		}

		realA := "a"
		realS := "s"
		for _, s := range file.Streams {
			switch s.CodecType {
			default:
				return fmt.Errorf(fmtCheckError("unsupported codec type", s.CodecType, "", tagname.src))
			case "video":
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
				if format.Sar != sar {
					return fmt.Errorf(fmtCheckError("SAR", format.Sar, sar, tagname.src))
				}
			case "audio":
				lang := convLang(s.Tags.Language)
				realA += fmt.Sprintf("%v%v", lang, s.Channels)
			case "subtitle":
				lang := convLang(s.Tags.Language)
				realS += fmt.Sprintf("%v", lang)
			}
		}

		if len(realA) == 3 && (realA[1] == '-' || realA[1] == 'e') {
			realA = fmt.Sprintf("%vr%v", string(realA[0]), string(realA[2]))
		}
		if format.Audio != realA {
			return fmt.Errorf(fmtCheckError("audio", format.Audio, realA, tagname.src))
		}
		if format.Subtitle != realS {
			return fmt.Errorf(fmtCheckError("subtitle", format.Subtitle, realS, tagname.src))
		}
	}
	return nil
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
