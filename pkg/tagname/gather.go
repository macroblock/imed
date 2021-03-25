package tagname

import (
	"fmt"
	"strings"

	"github.com/malashin/ffinfo"
)

var errorFFInfoIsNil = fmt.Errorf("ffinfo is nil")

// GatherExtension -
func (o *TTagname) GatherExtension() (string, error) {
	return gatherWrapper(o, GatherExtension)
}

// GatherSizeTag -
func (o *TTagname) GatherSizeTag() (string, error) {
	return gatherWrapper(o, GatherSizeTag)
}

// GatherATag -
func (o *TTagname) GatherATag() (string, error) {
	return gatherWrapper(o, GatherATag)
}

// GatherSTag -
func (o *TTagname) GatherSTag() (string, error) {
	return gatherWrapper(o, GatherSTag)
}

// GatherSDHD -
func (o *TTagname) GatherSDHD() (string, error) {
	return gatherWrapper(o, GatherSDHD)
}

func gatherWrapper(o *TTagname, fn func(*ffinfo.File)(string, error)) (string, error) {
	info, err := o.FFInfo()
	if err != nil {
		return "", err
	}
	return fn(info)
}

// GatherExtension -
func GatherExtension(info *ffinfo.File) (string, error) {
	if info == nil {
		return "", errorFFInfoIsNil
	}
	if len(info.Streams)<1 {
		return "", fmt.Errorf("len(info.Streams)<1")
	}
	codecName := strings.ToLower(info.Streams[0].CodecName)
	switch codecName {
	default:
		codecName = "." + codecName
	case "mjpeg":
		codecName = ".jpg"
	}
	return codecName, nil
}

// GatherSizeTag -
func GatherSizeTag(info *ffinfo.File) (string, error) {
	if info == nil {
		return "", errorFFInfoIsNil
	}
	if len(info.Streams)<1 {
		return "", fmt.Errorf("len(info.Streams)<1")
	}
	size := fmt.Sprintf("%vx%v", info.Streams[0].Width, info.Streams[0].Height)

	return size, nil
}

// GatherATag -
func GatherATag(info *ffinfo.File) (string, error) {
	if info == nil {
		return "", errorFFInfoIsNil
	}
	if len(info.Streams)<1 {
		return "", fmt.Errorf("len(info.Streams)<1")
	}

	var audio []TAudio
	for _, s := range info.Streams {
		if s.CodecType != "audio" {
			continue
		}
		lang := s.Tags.Language
		if len(lang) != 3 {
			lang = "---"
		}
		audio = append(audio, TAudio{lang, s.Channels})
	}

	if len(audio) == 0 {
		return "", nil
	}
	if len(audio) == 1 && audio[0].Language != "---" {
		audio[0].Language = "rus"
	}

	ret := "a"
	for i, v := range audio {
		lang := v.Language
		switch lang {
		case "---": return "", fmt.Errorf("audio stream #%v has unsupported language tag", i)
		case "rus": lang = "r"
		case "eng": lang = "e"
		}
		ret += fmt.Sprintf("%v%v", lang, v.Channels)
	}
	return ret, nil
}

// GatherSTag -
func GatherSTag(info *ffinfo.File) (string, error) {
	if info == nil {
		return "", errorFFInfoIsNil
	}
	if len(info.Streams)<1 {
		return "", fmt.Errorf("len(info.Streams)<1")
	}

	var subs []string
	for index, s := range info.Streams {
		if s.CodecType != "subtitle" {
			continue
		}
		lang := s.Tags.Language
		if len(lang) != 3 {
			return "", fmt.Errorf("stream #%v (subtitle) has unsupported language tag %q", index, lang)
		}
		subs = append(subs, lang)
	}
	if len(subs) == 0 {
		return "", nil
	}

	ret := "s"
	for _, v := range subs {
		switch v {
		case "rus": v = "r"
		case "eng": v = "e"
		}
		ret += v
	}
	return ret, nil
}

// GatherSDHD -
func GatherSDHD(info *ffinfo.File) (string, error) {
	if info == nil {
		return "", errorFFInfoIsNil
	}
	if len(info.Streams)<1 {
		return "", fmt.Errorf("len(info.Streams)<1")
	}

	sdhd := ""
	for index, s := range info.Streams {
		if s.CodecType != "video" {
			continue
		}
		size := fmt.Sprintf("%vx%v",s.Width, s.Height)
		v := ""
		switch size {
		default:
			return "", fmt.Errorf("stream #%v (video) has unsupported resolution %v", index, size)
		case "720x576": v = "sd"
		case "1920x1080": v = "hd"
                case "3840x2160": v = "4k"
		}
		if sdhd != "" && sdhd != v {
			return "", fmt.Errorf("video streams have unequal resolution")
		}
		sdhd = v
	}
	return sdhd, nil
}
