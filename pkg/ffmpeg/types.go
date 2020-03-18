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

var errHHMMSSMs = fmt.Errorf("hh:mm:ss.ms parser error")

// ParseTime -
func ParseTime(str string) (Time, error) {
	o := Time(0)
	x := strings.Split(str, ":")
	if len(x) != 3 {
		return o, errHHMMSSMs
	}
	y := strings.Split(x[2], ".")
	if len(y) > 2 {
		return o, errHHMMSSMs
	}
	h, err := strconv.Atoi(x[0])
	if err != nil {
		return o, errHHMMSSMs
	}
	m, err := strconv.Atoi(x[1])
	if err != nil {
		return o, errHHMMSSMs
	}
	s, err := strconv.Atoi(y[0])
	if err != nil {
		return o, errHHMMSSMs
	}
	ms := 0
	if len(y) == 2 {
		ms, err = strconv.Atoi(y[1])
		if err != nil {
			return o, errHHMMSSMs
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

// String -
func (o THHMMSSMs) String() string {
	return fmt.Sprintf("%02v:%02v:%02v.%03v", o.HH, o.MM, o.SS, o.Ms)
}

// String -
func (o Time) String() string {
	return o.HHMMSSMs().String()
}
