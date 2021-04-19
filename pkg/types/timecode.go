package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/macroblock/imed/pkg/misc"
)

type (
	Timecode float64

	HHMMSSMs struct {
		HH, MM, SS, Ms int
	}
)

var (
	globalMutex = sync.Mutex{}
	globalFps float64 = 25.0

	errHHMMSSMsTemplate = "%v parse error: %v"
)

var errval = Timecode(math.NaN())

// String -
func (o HHMMSSMs) String() string {
	if o.HH >= 0 && o.MM >= 0 && o.SS >= 0 && o.Ms >= 0 {
		return fmt.Sprintf("%02v:%02v:%02v.%03v", o.HH, o.MM, o.SS, o.Ms)
	}
	abs := misc.AbsInt
	return fmt.Sprintf("-%02v:%02v:%02v.%03v", abs(o.HH), abs(o.MM), abs(o.SS), abs(o.Ms))
}

func errMsg(val string, err error) error {
	return fmt.Errorf(errHHMMSSMsTemplate, val, err)
}

func SetFps(fps float64) {
	if fps <= 0.0 {
		panic("SetFps(): timecode fps cannot be <= 0")
	}
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalFps = fps
}

func GetFps() float64 {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	ret := globalFps
	return ret
}

// NewTimecode -
func NewTimecode(h, m, s float64) Timecode {
	return Timecode((h*60+m)*60 + s)
}

// ParseTimecode -
func ParseTimecode(t string) (Timecode, error) {
	ret, err := ParseHHMMSS(t)
	if err != nil {
		e := error(nil)
		ret, e = ParseHHMMSSFr(t)
		if e != nil {
			return Timecode(0), err
		}
	}
	return ret, nil
}

// ParseHHMMSS - hh:mm:ss format where ss can be float64
func ParseHHMMSS(t string) (Timecode, error) {
	o := Timecode(0)
	// fmt.Printf("here %v", t)
	x := strings.Split(t, ":")
	i := len(x)
	if i > 3 {
		return o, errMsg(t, fmt.Errorf("too many colons"))
	}
	h, m, s := 0.0, 0.0, 0.0
	signbit := false
	err := error(nil)
	i--
	s, signbit, err = parseHelper(x, i, true, signbit, err)
	i--
	m, signbit, err = parseHelper(x, i, false, signbit, err)
	i--
	h, signbit, err = parseHelper(x, i, false, signbit, err)
	if err != nil {
		return errval, err
	}
	ret := (h*60+m)*60 + s
	if signbit {
		return Timecode(-ret), nil
	}
	return Timecode(ret), nil
}

// ParseHHMMSSFr - hh:mm:ss:fr format
func ParseHHMMSSFr(t string) (Timecode, error) {
	fps := GetFps()
	errval := Timecode(math.NaN())
	x := strings.Split(t, ":")
	i := len(x)
	if i > 4 {
		return errval, errMsg(t, fmt.Errorf("too many colons"))
	}
	h, m, s, fr := 0.0, 0.0, 0.0, 0.0
	signbit := false
	err := error(nil)
	if i == 4 {
		i--
		fr, signbit, err = parseHelper(x, i, false, signbit, err)
	}
	i--
	s, signbit, err = parseHelper(x, i, false, signbit, err)
	i--
	m, signbit, err = parseHelper(x, i, false, signbit, err)
	i--
	h, signbit, err = parseHelper(x, i, false, signbit, err)
	if err != nil {
		return errval, err
	}
	ret := (h*60+m)*60 + s + fr/fps
	if signbit {
		return Timecode(-ret), nil
	}
	return Timecode(ret), nil
}

// HHMMSSMs -
func (o Timecode) HHMMSSMs() HHMMSSMs {
	v := int(o*1000)
	ms := v % 1000
	v = v / 1000
	s := v % 60
	v = v / 60
	m := v % 60
	h := v / 60
	return HHMMSSMs{HH: h, MM: m, SS: s, Ms: ms}
}

// String -
func (o Timecode) String() string {
	return o.HHMMSSMs().String()
}

// InSeconds -
func (o Timecode) InSeconds() float64 {
	return float64(o)
}

func parseHelper(s []string, index int, isFloat, signbit bool, err error) (float64, bool, error) {
	if index < 0 || err != nil {
		return 0.0, signbit, err
	}
	NaN := math.NaN()
	v := 0.0
	if isFloat {
		v, err = strconv.ParseFloat(s[index], 64)
	} else {
		x, e := strconv.Atoi(s[index])
		v = float64(x)
		err = e
	}
	if err != nil {
		return NaN, false, errMsg(strings.Join(s, ":"), err)
	}
	sb := math.Signbit(v)

	// if it is the firstmost subvalue
	if index == 0 {
		return math.Abs(v), sb, nil
	}
	if sb {
		return NaN, false, errMsg(strings.Join(s, ":"), fmt.Errorf("only first subvalue can have a sign"))
	}
	return v, signbit, nil
}

