package subrip

import (
	"fmt"
	"strconv"
	"strings"
)

var t35Min = NewTimecode(0, 35, 0) // 35 minutes

type Record struct {
	ID int // used in error messages only
	In Timecode
	Out Timecode
	Text string
}

type Options int
const (
	OptMultipleErrors Options = iota
	OptCheckNegativeTimecode
	OptCheckChunkId
	OptCheck35MinGap
)

func MildOptions() Options {
	return NewOptions()
}

func StrictOptions() Options {
	return ^MildOptions()
}

func NewOptions(args ...Options) Options {
	o := Options(0)
	for _, v := range args {
		o = o |	(1<<v)
	}
	return o
}

func(o Options) On(v Options) bool {
	return o & (1<<v) != 0
}

func(o Options) Off(v Options) bool {
	return o & (1<<v) == 0
}

func MildCheck(srt []Record) error {
	return CheckOpt(srt, MildOptions())
}

func StrictCheck(srt []Record) error {
	return CheckOpt(srt, StrictOptions())
}

func CheckOpt(srt []Record, opt Options) error {
	var (
		errs []string
		tc Timecode
		prevID int
		prevOut Timecode
	)
	errMsg := func(format string, args ...interface{}) bool {
		err := fmt.Errorf(format, args...)
		if opt.Off(OptMultipleErrors) {
			errs = append(errs, err.Error())
			return true
		}
		errs = append(errs, "    "+err.Error())
		return false
	}

	first := true
	for _, v := range srt {
		if opt.On(OptCheckNegativeTimecode) {
			if v.In < 0 || v.Out < 0 {
				if errMsg("chunk:%v: negative timecode", v.ID) {
					return fmt.Errorf(strings.Join(errs, "\n"))
				}
			}
		}
		if v.In > v.Out {
			if errMsg("chunk:%v: <IN> > <OUT>", v.ID) {
				return fmt.Errorf(strings.Join(errs, "\n"))
			}
		}
		if first {
			tc = v.Out
			if v.ID != 1 {
				if errMsg("chunk:%v: starting chunk id != 1", v.ID) {
					return fmt.Errorf(strings.Join(errs, "\n"))
				}
			}
			prevID = v.ID
			prevOut = v.Out
			first = false
			continue
		}
		if tc > v.In {
			if errMsg("chunk:%v: overlaps with previous chunk", v.ID) {
				return fmt.Errorf(strings.Join(errs, "\n"))
			}

		}
		if opt.On(OptCheckChunkId) {
			if v.ID - prevID != 1 {
				if errMsg("chunk:%v: nonhomogeneus chunk id", v.ID) {
					return fmt.Errorf(strings.Join(errs, "\n"))
				}
			}
		}
		if opt.On(OptCheck35MinGap) {
			if v.Out - v.In > t35Min ||
				v.In - prevOut > t35Min {
				if errMsg("chunk:%v: delta time > 35 min", v.ID) {
					return fmt.Errorf(strings.Join(errs, "\n"))
				}
			}
		}
		prevID = v.ID
		prevOut = v.Out
	}
	if len(errs) == 0 {
		return nil
	}
	for i := range errs {
		errs[i] = "  "+errs[i]
	}
	return fmt.Errorf("Error(s):\n%v", strings.Join(errs, "\n"))
}

func TimeToStr(t Timecode) string {
	return strings.Replace(t.String(), ".", ",", 1)
}

func ToString(srt []Record) string {
	ret := ""
	for _, v := range srt {
		ret += strconv.Itoa(v.ID) + "\n"
		ret += fmt.Sprintf("%v --> %v\n", TimeToStr(v.In), TimeToStr(v.Out))
		ret += v.Text + "\n\n"
	}
	return ret
}

func genFillerTCs(in, out Timecode) []Timecode {
	var ret []Timecode
	times := int((out.InSeconds() - in.InSeconds()) / t35Min.InSeconds())
	tc := (out - in) / 2 + in
	for i := 0; i < times; i++ {
		ret = append(ret, tc)
		tc += t35Min
	}
	return ret
}

func Fix(srt []Record) []Record {
	var (
		ret []Record
		prevOut Timecode
	)
	first := true
	id := 1
	for _, v := range srt {
		if first {
			v.ID = id
			ret = append(ret, v)
			prevOut = v.Out
			first = false
			id++
			continue
		}
		if v.In - prevOut > t35Min {
			timecodes := genFillerTCs(prevOut, v.In)
			for _, tc := range timecodes {
				r := Record{ ID: id, In: tc, Out: tc, Text: "<i>" }
				ret = append(ret, r)
				id++
			}
		}
		if v.Out - v.In > t35Min {
			timecodes := genFillerTCs(v.Out, v.In)
			prev := v.In
			for _, tc := range timecodes {
				r := Record{ ID: id, In: prev, Out: tc, Text: v.Text }
				ret = append(ret, r)
				id++
				prev = tc
			}
			v.In = prev
		}
		v.ID = id
		ret = append(ret, v)
		prevOut = v.Out
		id++
	}
	return ret
}

func transformTC(tc Timecode, zp Timecode, scale float64, move Timecode) Timecode {
	if scale == 1.0 {
		return tc - zp + move
	}
	// return ffmpeg.FloatToTime((tc - zp).Float()*scale) + move
	return NewTimecode(0, 0, (tc - zp).InSeconds()*scale) + move
}

func Transform(srt *[]Record, zp Timecode, scale float64, move Timecode) {
	for i, v := range *srt {
		(*srt)[i].In = transformTC(v.In, zp, scale, move)
		(*srt)[i].Out = transformTC(v.Out, zp, scale, move)
	}
}
