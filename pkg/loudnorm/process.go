package loudnorm

import (
	"github.com/malashin/ffinfo"
)

func process(filename string) error {
	finfo, err := ffinfo.Probe(filename)
	if err != nil {
		return err
	}

	for _, stream := range finfo.Streams {
		switch stream.CodecType {
		case "video":
			continue
		case "audio":

		case "subtitle":
		}

	}
	return nil
}
