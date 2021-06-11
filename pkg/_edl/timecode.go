package edl

import (
	"fmt"
	// "strconv"
)

const (
	Milliseconds = 1
	Seconds = 1000*Milliseconds
	Minutes = 60*Seconds
	Hours = 60*Minutes
)

type timecodeDisplayMode = int
const (
	TimecodeDisplayMs timecodeDisplayMode = iota
	TimecodeDisplayFr
	TimecodeDisplayFrFps
	TimecodeDisplayFrRest
	TimecodeDisplayFrRestFps
)

var globTimecodeDisplayMode = TimecodeDisplayMs

func SetTimecodeDispayMode(mode timecodeDisplayMode) {
	globTimecodeDisplayMode = mode
}

func TimecodeDispayMode() timecodeDisplayMode {
	return globTimecodeDisplayMode
}

func SetFPS(fps int) {
	globFps = fps
}

func FPS(fps int) int {
	return globFps
}

type Timecode struct {
	ms int
}

func Min(a, b Timecode) Timecode {
	if a.ms < b.ms {
		return a
	}
	return b
}

func Max(a, b Timecode) Timecode {
	if a.ms > b.ms {
		return a
	}
	return b
}

func div(num, den int) (int, int) {
	return num / den, num % den
}

func hhmmssms2ms(fps, hh, mm, ss, ms int) int {
	return ((hh*60+mm)*60+ss)*1000 + ms
}

func hhmmssfr2ms(fps, hh, mm, ss, fr int) (int, int) {
	x, rest := div(1000, fps)
	return hhmmssms2ms(fps, hh, mm, ss, x*fr), rest
}

func ms2hhmmssms(ms int) (int, int, int, int) {
	var hh, mm, ss, rest int
	rest, ms = div(ms, 1000)
	rest, ss = div(rest, 60)
	hh, mm = div(rest, 60)
	return hh, mm, ss, ms
}

func ms2hhmmssfr(fps int, ms int) (int, int, int, int, int) {
	var hh, mm, ss, fr, rest int
	hh, mm, ss, ms = ms2hhmmssms(ms)
	fr, rest = div(ms*fps, 1000)
	return hh, mm, ss, fr, int(rest)
}

func NewTimecode() Timecode {
	return Timecode{}
}

func NewTimecodeFromMs(ms int) Timecode {
	return Timecode{ ms: ms }
}

func NewTimecodeFromHHMMSSMs(hh, mm, ss, ms int) Timecode {
	return Timecode{ ms: hhmmssms2ms(globFps, hh, mm, ss, ms) }
}

func NewTimecodeFromHHMMSSFr(hh, mm, ss, fr int) (Timecode, int) {
	ms, rest := hhmmssfr2ms(globFps, hh, mm, ss, fr)
	return Timecode{ ms: ms }, rest
}

func (o Timecode) ToMs() int { return o.ms }
func (o Timecode) ToHHMMSSMs() (int, int, int, int) { return ms2hhmmssms(o.ms) }
func (o Timecode) ToHHMMSSFr(fps int) (int, int, int, int, int) { return ms2hhmmssfr(fps, o.ms) }

func (o Timecode) HH() (int) {
	return o.ms/(60*60*1000)
}

func (o Timecode) MM() (int) {
	return (o.ms%(60*60*1000))/(60*100)
}

func (o Timecode) SS() (int) {
	return (o.ms%(60*1000))/1000
}

func (o Timecode) Ms() (int) {
	return o.ms%1000
}

// Fr - returns frames and remainder. Check the last one please
func (o Timecode) Fr(fps int) (int, int) {
	return div(o.Ms() * fps, 1000)
}

func(o Timecode) String() string {
	switch globTimecodeDisplayMode {
	default:
		panic(fmt.Sprintf("invalid timecode display mode (%v)", globTimecodeDisplayMode))
	case TimecodeDisplayMs:
		hh, mm, ss, ms := o.ToHHMMSSMs()
		return fmt.Sprintf("%02v:%02v:%02v.%03v", hh, mm, ss, ms)
	case TimecodeDisplayFr:
		hh, mm, ss, fr, _ := o.ToHHMMSSFr(globFps)
		return fmt.Sprintf("%02v:%02v:%02v:%02v", hh, mm, ss, fr)
	case TimecodeDisplayFrFps:
		fps := globFps
		hh, mm, ss, fr, _ := o.ToHHMMSSFr(fps)
		return fmt.Sprintf("%02v:%02v:%02v:%02v@%v", hh, mm, ss, fr, fps)
	case TimecodeDisplayFrRest:
		hh, mm, ss, fr, rest := o.ToHHMMSSFr(globFps)
		return fmt.Sprintf("%02v:%02v:%02v:%02v+%03v", hh, mm, ss, fr, rest)
	case TimecodeDisplayFrRestFps:
		fps := globFps
		hh, mm, ss, fr, rest := o.ToHHMMSSFr(fps)
		return fmt.Sprintf("%02v:%02v:%02v:%02v+%03v@%v", hh, mm, ss, fr, rest, fps)
	}
}

