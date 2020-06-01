package loudnorm

import (
	"fmt"
	"math"
	"strconv"

	"github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/ffmpeg"
	"github.com/macroblock/imed/pkg/misc"
)

func debugPrintf(pattern string, args ...interface{}) {
	if !GlobalDebug {
		return
	}
	fmt.Printf(pattern, args...)
}

func fround(f float64) string {
	precision := 2
	return strconv.FormatFloat(f, 'f', precision, 64)
}

func fdown(f float64) string {
	precision := 2
	if f >= 0.0 {
		str := strconv.FormatFloat(f, 'f', precision+1, 64)
		return str[:len(str)-1]
	}
	str := strconv.FormatFloat(-f, 'f', precision+1, 64)
	return "-" + str[:len(str)-1]
}

func fup(f float64) string {
	precision := 2
	if f >= 0.0 {
		str := strconv.FormatFloat(-f, 'f', precision+1, 64)
		return str[1 : len(str)-1]
	}
	str := strconv.FormatFloat(f, 'f', precision+1, 64)
	return str[:len(str)-1]
}

func froundRatio(f float64) string {
	if f <= 0 {
		return "1:1" //"NaN"
	}
	precision := 2
	return strconv.FormatFloat(1.0/f, 'f', precision, 64) + ":1"
}

func colorizedPrintf(color misc.TTerminalColor, format string, args ...interface{}) {
	printf := ansi.Printf
	c := misc.Color(color)
	r := misc.Color(misc.ColorReset)
	if !misc.IsTerminal() {
		printf = fmt.Printf
		c = ""
		r = ""
	}
	printf(c+format+r, args...)
}

func colorReset() string {
	return misc.Color(misc.ColorReset)
}
func colorizeTo(c misc.TTerminalColor, s string) string {
	if !misc.IsTerminal() {
		return s
	}
	return misc.Color(c) + s + colorReset()
}

func colorizeI(colorize bool, v float64, s string) string {
	if !colorize || !misc.IsTerminal() {
		return s
	}
	c := misc.Color(misc.ColorRed)
	if targetIMin() <= v && v < targetIMax() {
		c = misc.Color(misc.ColorYellow)
	}
	if v == targetI() {
		c = misc.Color(misc.ColorGreen)
	}
	return c + s + colorReset()
}

func colorizeRatio(v float64, s string) string {
	if !misc.IsTerminal() {
		return s
	}
	c := misc.Color(misc.ColorRed)
	v = 1.0 / v
	switch {
	case v == math.NaN():
		c = misc.Color(misc.ColorBlack, misc.ColorBgRed)
	case v <= 1.0:
		c = ""
	case v < 1.4:
		c = misc.Color(misc.ColorGreen)
	case v < 1.9:
		c = misc.Color(misc.ColorYellow)
	}
	return c + s + colorReset()
}

func printStreamParams(stream *TStreamInfo, colorize bool) {

	printf := ansi.Printf

	li := stream.TargetLI
	I2 := math.Min(li.I-li.MP, settings.Loudness.I)
	maxP := math.Inf(-1)
	for _, ch := range stream.astatsInfo.Channels {
		maxP = math.Max(maxP, ch.RMSLevel)
	}
	str := "" //fmt.Sprintf("%v %v: ", stream.astatsInfo.PeakLevel, len(stream.astatsInfo.Channels))
	for _, ch := range stream.astatsInfo.Channels {
		str += strconv.Itoa(int(math.Round(ch.RMSLevel-maxP))) + " "
	}
	printf(" #%2v: %v\n", stream.Index, li.FormatString(colorize))
	// fmt.Printf("    : comp %v chan: %v\n", stream.CompParams, str)
	printf("    :??? %v, [%v], channels: %v\n",
		colorizeI(true, I2, fround(I2)),
		colorizeRatio(stream.CompParams.GetK(), froundRatio(stream.CompParams.GetK())),
		str)
	// fmt.Printf("    : %v, %v\n", fround(stream.CompParams.PreAmp), fround(stream.CompParams.PostAmp))
	printf("    : ST stats clean: %v\n", stream.MiscInfo.toString())
	if stream.MiscInfo.NaNCount > 0 {
		printf("    :      with NaNs: %v\n", colorizeTo(misc.ColorCyan, stream.MiscInfo.toStringWithNaNs()))
	}
}

func initInfo(ebur *ffmpeg.TEburInfo, astats *ffmpeg.TAStatsInfo) (*TLoudnessInfo, *TMiscInfo) {
	li := &TLoudnessInfo{
		I:  ebur.I,
		RA: ebur.LRA,
		TP: ebur.TP,
		TH: ebur.Thresh,
		MP: astats.PeakLevel,
		// CR: -1,
	}
	mi := &TMiscInfo{
		I:          ebur.I,
		MaxST:      ebur.MaxST,
		MinST:      ebur.MinST,
		STSum:      ebur.SumST,
		TotalCount: ebur.CountST,
		NaNCount:   ebur.CountNaN,
		AboveST:    ebur.AboveST,
		BelowST:    ebur.BelowST,
		EqualST:    ebur.EqualST,
	}
	return li, mi
}
