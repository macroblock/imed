package ffmpeg

import (
	"fmt"
	"regexp"
	"time"
)

type tAudioProgressParser struct {
	callback IAudioProgress
}

var reAudioProgress = regexp.MustCompile("size=.+ time=(\\d{2}:\\d{2}:\\d{2}.\\d+) bitrate=.+ speed=.+")

// Parse -
func (o *tAudioProgressParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {
	// if o == nil || o.callback == nil {
	// 	return false, false, fmt.Errorf("audio progress parser: either receiver or callback is nil")
	// }
	if val := reAudioProgress.FindAllStringSubmatch(line, 1); val != nil {
		t, err := ParseTime(val[0][1])
		if err != nil {
			return true, eof, err
		}
		err = o.callback.Callback(t)
		return true, eof, err
	}
	return false, eof, nil
}

type tDefaultAudioProgressCallback struct {
	lastTime time.Time
	total    Time
}

func (o *tDefaultAudioProgressCallback) Callback(t Time) error {
	ct := time.Now()
	zeroTime := time.Time{}
	if o.lastTime == zeroTime {
		o.lastTime = ct
	}
	fmt.Printf("%v / %v, delta: %v\n", t, o.total, time.Since(o.lastTime))
	return nil
}

// NewAudioProgressParser -
func NewAudioProgressParser(totalLen Time, callback IAudioProgress) IParser {
	// fmt.Printf("@@@@: %q\n", line)
	if callback == nil {
		callback = &tDefaultAudioProgressCallback{total: totalLen}
	}
	return &tAudioProgressParser{
		callback: callback,
	}
}
