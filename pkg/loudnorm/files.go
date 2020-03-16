package loudnorm

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/malashin/ffinfo"
)

type (
	// TFileInfo -
	TFileInfo struct {
		Filename string
		Mode     TMode
		Streams  []*TStreamInfo
	}
	// TStreamInfo -
	TStreamInfo struct {
		Parent      *TFileInfo
		Index       int
		Type        string // "audio", "video", "subtitle"
		ExtName     string
		Codec       string
		Channels    int
		Lang        string
		AudioParams []string
		// Done          bool
		LoudnessInfo  *LoudnessInfo
		W, H          int
		validFormat   bool
		validLoudness bool
		extInputIndex int
	}
)

func (o *TFileInfo) String() string {
	if o == nil {
		return "<nil>"
	}
	ret := fmt.Sprintf("input filename: %v\n", o.Filename)
	ret = fmt.Sprintf("%vmode: %v\n", ret, o.Mode)

	for N, stream := range o.Streams {
		ret = fmt.Sprintf("%vn: %v\n", ret, N)
		ret = fmt.Sprintf("%v  stream index: %v\n", ret, stream.Index)
		ret = fmt.Sprintf("%v  type: %v\n", ret, stream.Type)
		ret = fmt.Sprintf("%v  codec: %v\n", ret, stream.Codec)
		ret = fmt.Sprintf("%v  channels: %v\n", ret, stream.Channels)
		ret = fmt.Sprintf("%v  lang: %v\n", ret, stream.Lang)
		ret = fmt.Sprintf("%v  audio params: %v\n", ret, strings.Join(stream.AudioParams, " "))
		ret = fmt.Sprintf("%v  EBU R128: %v\n", ret, stream.LoudnessInfo)
		ret = fmt.Sprintf("%v  resolution: %vx%v\n", ret, stream.W, stream.H)
	}
	return ret
}

// LoadFile -
func LoadFile(filename string) (*TFileInfo, error) {
	finfo, err := ffinfo.Probe(filename)
	if err != nil {
		return nil, err
	}

	fi := &TFileInfo{Filename: filename}
	for trackN, stream := range finfo.Streams {
		switch stream.CodecType {
		default:
			return nil, fmt.Errorf("unknown stream codec type (%v)", stream.CodecType)
		case "video":
			addVideoStreamInfo(fi, filename, trackN, stream.Width, stream.Height)
		case "subtitle":
			addSubtitleStreamInfo(fi, filename, trackN, stream.CodecName, stream.Tags.Language)
		case "audio":
			addAudioStreamInfo(fi, filename, trackN, stream.CodecName, stream.Channels, stream.Tags.Language)
		}
	}

	err = AttachLoudnessInfo(fi, finfo.Format.Tags.Comment)

	if err != nil {
		return nil, err
	}
	// RecalcParameters(fi)
	return fi, nil
}

func addVideoStreamInfo(fi *TFileInfo, filename string, index int, w, h int) {
	o := &TStreamInfo{
		Parent: fi,
		Index:  index,
		Type:   "video",
		W:      w,
		H:      h,
		// Done:  true,
		validLoudness: true,
	}
	fi.Mode = ModeHD
	if w <= 720 && h <= 576 {
		fi.Mode = ModeSD
	}
	fi.Streams = append(fi.Streams, o)
}

func addAudioStreamInfo(fi *TFileInfo, filename string, index int, codec string, ch int, lang string) {
	o := &TStreamInfo{
		Parent:   fi,
		Index:    index,
		Type:     "audio",
		Codec:    codec,
		Channels: ch,
		Lang:     lang,
	}
	fi.Streams = append(fi.Streams, o)
	// recalcParams(fi, o)
}

func addSubtitleStreamInfo(fi *TFileInfo, filename string, index int, codec string, lang string) {
	o := &TStreamInfo{
		Parent: fi,
		Index:  index,
		Type:   "subtitle",
		Codec:  codec,
		Lang:   lang,
	}
	fi.Streams = append(fi.Streams, o)
}

func generateOutputName(filename string) string {
	path, name := filepath.Split(filename)
	ext := filepath.Ext(name)
	name = strings.TrimSuffix(name, ext)
	return path + name + "-ebur128.mp4"
}

func generateExtFilename(fi *TFileInfo, si *TStreamInfo) string {
	path, name := filepath.Split(fi.Filename)
	ext := filepath.Ext(name)
	name = strings.TrimSuffix(name, ext)
	base := path + name + "-" + strconv.Itoa(si.Index) + "-" + si.Lang //+ "-" + si.Codec
	return base + ".m4a"
}

func generateAudioParams(fi *TFileInfo, si *TStreamInfo) []string {
	switch fi.Mode {
	default:
		panic(fmt.Sprintf("invalid mode (%d) %v", fi.Mode, fi.Mode))
	case ModeHD:
		switch si.Channels {
		default:
			panic(fmt.Sprintf("wrong audio stream parameters: Mode %v, channels: %v", fi.Mode, si.Channels))
		case 2:
			return ac3Params2
		case 6:
			return ac3Params6
		}
	case ModeSD:
		switch si.Channels {
		default:
			panic(fmt.Sprintf("wrong audio stream parameters: Mode %v, channels: %v", fi.Mode, si.Channels))
		case 2:
			return mp2Params2
		case 6:
			return mp2Params6
		}
	case ModeUnknown:
		switch si.Channels {
		default:
			panic(fmt.Sprintf("wrong audio stream parameters: codec %v, channels: %v", si.Codec, si.Channels))
		case 2:
			return alacParams
		case 6:
			return alacParams
		}
	}
}
