package ffmpeg

import (
	"fmt"

	"github.com/malashin/ffinfo"
)

type (
	// TFile -
	TFile struct {
		owner     *TFilterGraph
		name      string
		options   []string
		streams   []*TStream
		videos    []*TStream
		audios    []*TStream
		subtitles []*TStream
		info      *ffinfo.File
	}
)

// NewFile -
func NewFile(owner *TFilterGraph, name string, options ...interface{}) (*TFile, error) {
	opts, err := ArgsToStrings(options)
	if err != nil {
		return nil, err
	}
	o := &TFile{owner: owner, name: name, options: opts}
	return o, nil
}

// Probe -
func (o *TFile) Probe() error {
	if o.info != nil {
		return nil
	}
	info, err := ffinfo.Probe(o.name)
	if err != nil {
		return fmt.Errorf("file: %q %v", o.name, err)
	}
	o.info = info
	o.streams = make([]*TStream, 0, len(o.info.Streams))
	o.videos, o.audios, o.subtitles = nil, nil, nil
	for index, stream := range o.info.Streams {
		codecType := stream.CodecType
		s := &TStream{owner: o, typ: streamTypeUnknown, index: index}
		o.streams = append(o.streams, s)
		switch codecType {
		default:
			return fmt.Errorf("unknown stream #%v codec type (%v)", index, codecType)
		case "video":
			s.typ = streamTypeVideo
			o.videos = append(o.videos, s)
		case "audio":
			s.typ = streamTypeAudio
			o.audios = append(o.audios, s)
		case "subtitle":
			s.typ = streamTypeSubtitle
			o.subtitles = append(o.subtitles, s)
		}
	}
	return nil
}
