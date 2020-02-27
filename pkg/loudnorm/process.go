package loudnorm

import (
	"bytes"
	"errors"
	"os/exec"
	"strconv"

	"github.com/malashin/ffinfo"
)

func callFFMPEG(args ...string) error {
	c := exec.Command("ffmpeg", args...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err := c.Run()
	if err != nil {
		return errors.New(string(e.Bytes()))
	}
	return err
}

func process(filename string) error {
	finfo, err := ffinfo.Probe(filename)
	if err != nil {
		return err
	}

	for trackN, stream := range finfo.Streams {
		switch stream.CodecType {
		case "video":
			continue
		case "audio":
			_ = stream.Tags.Language
			callFFMPEG(
				"-hide_banner",
				"-i", filename,
				"-map", "0:"+strconv.Itoa(trackN),
				"-f", "null",
				"NUL",
			)

		case "subtitle":
		}

	}
	return nil
}
