package ffmpeg

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/macroblock/imed/pkg/misc"
)

type tAudioProgressParser struct {
	callback IAudioProgress
}

var reAudioProgress = regexp.MustCompile("size=.+ time=(\\d{2}:\\d{2}:\\d{2}.\\d+) bitrate=.+ speed=.+")

// Finish -
func (o *tAudioProgressParser) Finish() error {
	err := o.callback.Callback(-1)

	// clearLine := strings.Repeat(" ", 78)
	// fmt.Println(clearLine + "\r----+")
	// fmt.Println()

	return err
}

// Parse -
func (o *tAudioProgressParser) Parse(line string, eof bool) (accepted bool, err error) {
	// if o == nil || o.callback == nil {
	// 	return false, false, fmt.Errorf("audio progress parser: either receiver or callback is nil")
	// }
	if val := reAudioProgress.FindAllStringSubmatch(line, 1); val != nil {
		// t, err := ParseTime(val[0][1])
		t, err := ParseTimecode(val[0][1])
		if err != nil {
			return true, err
		}
		err = o.callback.Callback(t)
		return true, err
	}
	return false, nil
}

type tDefaultAudioProgressCallback struct {
	lastTime   time.Time
	total      Timecode
	maxInfoLen int
}

func (o *tDefaultAudioProgressCallback) Callback(t Timecode) error {
	if t < 0.0 {
		t = o.total
		clearLine := strings.Repeat(" ", o.maxInfoLen)
		// fmt.Printf(clearLine+"\relapsed: %v", time.Since(o.lastTime))
		fmt.Printf(clearLine + "\r")
		return nil
	}
	ct := time.Now()
	zeroTime := time.Time{}
	if o.lastTime == zeroTime {
		o.lastTime = ct
	}
	percents := "N/A"
	if o.total > 0 {
		percents = strconv.Itoa(int(math.Round(100*t.InSeconds()/o.total.InSeconds()))) + "%"
	}
	info := fmt.Sprintf(" %v %v/%v, elapsed: %v            \r",
		percents, t, o.total, time.Since(o.lastTime))
	o.maxInfoLen = misc.MaxInt(len(info), o.maxInfoLen)
	fmt.Print(info)

	return nil
}

// NewAudioProgressParser -
func NewAudioProgressParser(totalLen Timecode, callback IAudioProgress) IParser {
	// fmt.Printf("@@@@: %q\n", line)
	if callback == nil {
		if totalLen.InSeconds() < 1 { // !!!HACK!!!
			totalLen = NewTimecode(0, 0, -totalLen.InSeconds())
		}
		callback = &tDefaultAudioProgressCallback{total: totalLen}
	}
	return &tAudioProgressParser{
		callback: callback,
	}
}
