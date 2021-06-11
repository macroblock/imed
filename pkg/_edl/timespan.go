package edl

import (
	"fmt"
)

type Timespan struct {
	In Timecode
	Out Timecode
}

func NewTimespan(in, out Timecode) Timespan {
	return Timespan{In: in, Out: out}
}

func (o Timespan) String() string {
	return fmt.Sprintf("(%v-%v)", o.In, o.Out)
}

func GapMs(a, b Timespan) int {
	aMax := a.Max().ms
	bMin := b.Min().ms
	if aMax <= bMin {
		return bMin - aMax
	}
	aMin := a.Min().ms
	bMax := b.Max().ms
	if aMin >= bMax {
		return bMax - aMin
	}
	return 0
}

func Unite(a, b Timespan) Timespan {
	min := Min(a.Min(), b.Min())
	max := Max(a.Max(), b.Max())
	return NewTimespan(min, max)
}

func (o Timespan) Min() Timecode {
	return Min(o.In, o.Out)
}

func (o Timespan) Max() Timecode {
	return Max(o.In, o.Out)
}

func (o Timespan) Duration() Timecode {
	return NewTimecodeFromMs(o.Out.ms - o.In.ms)
}

func (o Timespan) Extends(ts Timespan) bool {
	if o.Min() == ts.Max() {
		return true
	}
	return false
}

func (o Timespan) ExtendedBy(ts Timespan) bool {
	if o.Max() == ts.Min() {
		return true
	}
	return false
}

func (o Timespan) Unite(ts Timespan) Timespan {
	return Unite(o, ts)
}
