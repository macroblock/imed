package ffmpeg

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

type tAudioProgressParser struct {
	callback IAudioProgress
}

var reAudioProgress = regexp.MustCompile("size=.+ time=(\\d{2}:\\d{2}:\\d{2}.\\d+) bitrate=.+ speed=.+")

// Finish -
func (o *tAudioProgressParser) Finish() error {
	err := o.callback.Callback(-1)
	fmt.Println()
	return err
}

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
	if t < 0 {
		t = o.total
	}
	ct := time.Now()
	zeroTime := time.Time{}
	if o.lastTime == zeroTime {
		o.lastTime = ct
	}
	percents := "N/A"
	if o.total > 0 {
		percents = strconv.Itoa(int(math.Round(100*t.Float()/o.total.Float()))) + "%"
	}
	fmt.Printf(" %v %v / %v, elapsed: %v            \r",
		percents, t, o.total, time.Since(o.lastTime))
	return nil
}

// NewAudioProgressParser -
func NewAudioProgressParser(totalLen Time, callback IAudioProgress) IParser {
	// fmt.Printf("@@@@: %q\n", line)
	if callback == nil {
		if totalLen < 1000 { // !!!HACK!!!
			totalLen = -1000
		}
		callback = &tDefaultAudioProgressCallback{total: totalLen}
	}
	return &tAudioProgressParser{
		callback: callback,
	}
}
