package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	// Time -
	Time int

	// THHMMSSMs -
	THHMMSSMs struct {
		HH, MM, SS, Ms int
	}
)

var errHHMMSSMsTemplate = "%v parse error: %v"

func errHHMMSSMs(val string, err error) error {
	return fmt.Errorf(errHHMMSSMsTemplate, val, err)
}

func errHHMMSSFr(val string, err error) error {
	return fmt.Errorf(errHHMMSSMsTemplate, val, err)
}

// FloatToTime -
func FloatToTime(f float64) Time {
	return Time(f * 1000)
}

// ParseHHMMSSFr - hh:mm:ss:fr format
func ParseHHMMSSFr(t string, msPerFrame int) (Time, error) {
	o := Time(0)
	x := strings.Split(t, ":")
	if len(x) > 4 {
		return o, errHHMMSSFr(t, fmt.Errorf("too many colons"))
	}
	h, m, s, fr := 0, 0, 0, 0
	done := false
	for _, str := range x {
		if done {
			return o, errHHMMSSFr(t, fmt.Errorf("unreachable"))
		}
		val, err := strconv.Atoi(str)
		if err != nil {
			return o, errHHMMSSFr(t, err)
		}
		h = m
		m = s
		s = fr
		fr = val
	}
	o = Time(((h*60+m)*60+s)*1000 + fr*msPerFrame)
	return o, nil
}

// ParseTime - hh:mm:ss.ms format
func ParseTime(t string) (Time, error) {
	o := Time(0)
	// fmt.Printf("here %v", t)
	x := strings.Split(t, ":")
	if len(x) > 3 {
		return o, errHHMMSSMs(t, fmt.Errorf("too many colons"))
	}
	h, m, s, ms := 0, 0, 0, 0
	done := false
	for _, str := range x {
		if done {
			return o, errHHMMSSMs(t, fmt.Errorf("unreachable"))
		}
		switch {
		default:
			val, err := strconv.Atoi(str)
			if err != nil {
				return o, errHHMMSSMs(t, err)
			}
			h = m
			m = s
			s = ms
			ms = val
		case strings.Contains(str, "."):
			done = true
			y := strings.Split(str, ".")
			if len(y) != 2 {
				return o, errHHMMSSMs(t, fmt.Errorf("too many dots"))
			}
			a, err := strconv.Atoi(y[0])
			if err != nil {
				return o, errHHMMSSMs(t, err)
			}
			b, err := strconv.Atoi(y[1])
			if err != nil {
				return o, errHHMMSSMs(t, err)
			}
			h = s
			m = ms
			s = a
			ms = b
		}
	}
	o = Time(((h*60+m)*60+s)*1000 + ms)
	return o, nil
}

// HHMMSSMs -
func (o Time) HHMMSSMs() THHMMSSMs {
	v := int(o)
	ms := v % 1000
	v = v / 1000
	s := v % 60
	v = v / 60
	m := v % 60
	h := v / 60
	return THHMMSSMs{HH: h, MM: m, SS: s, Ms: ms}
}

// Float -
func (o Time) Float() float64 {
	return float64(o) / 1000.0
}

// String -
func (o THHMMSSMs) String() string {
	return fmt.Sprintf("%02v:%02v:%02v.%03v", o.HH, o.MM, o.SS, o.Ms)
}

// String -
func (o Time) String() string {
	return o.HHMMSSMs().String()
}
