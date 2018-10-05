package loglevel

import "strings"

// TLevel -
type TLevel int

// General loglevel flags
const (
	TooLow TLevel = -1 + iota
	Panic
	Error
	Warning
	Reset
	Notice
	Info
	Debug
	TooHigh
)

var levelToStr = []string{"PNC", "ERR", "WRN", "RESET", "NTC", "INF", "DBG", "UNKNOWN"}

// Only -
func (o TLevel) Only() TFilter { return 1 << uint(o) }

// Below -
func (o TLevel) Below() TFilter { return o.Only() - 1 }

// Above -
func (o TLevel) Above() TFilter { return TooHigh.Below() &^ o.OrLower() }

// OrLower -
func (o TLevel) OrLower() TFilter { return o.Only()<<1 - 1 }

// OrHigher -
func (o TLevel) OrHigher() TFilter { return TooHigh.Below() &^ o.Below() }

// In -
func (o TLevel) In(f TFilter) bool { return f&o.Only() != 0 }

// NotIn -
func (o TLevel) NotIn(f TFilter) bool { return f&o.Only() == 0 }

// String -
func (o TLevel) String() string {
	if o <= TooLow || o >= TooHigh {
		o = TooHigh
	}
	return levelToStr[o]
}

//============================================================================

// TFilter -
type TFilter uint

// General loglevel filters
const (
	All TFilter = 1<<uint(TooHigh) - 1
)

// Include -
func (o TFilter) Include(f TFilter) TFilter { return o | f }

// Exclude -
func (o TFilter) Exclude(f TFilter) TFilter { return o &^ f }

// Intersect -
func (o TFilter) Intersect(f TFilter) TFilter { return o & f }

// String -
func (o TFilter) String() string {
	sl := []string{}
	for i, x := 0, o&TooHigh.Below(); x != 0; i, x = i+1, x>>1 {
		if x&1 != 0 {
			sl = append(sl, levelToStr[i])
		}
	}
	if len(sl) == 0 {
		return levelToStr[TooHigh]
	}
	return strings.Join(sl, "|")
}
