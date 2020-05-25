package loudnorm

import (
	"fmt"
	"math"
	"strconv"

	"github.com/macroblock/imed/pkg/ffmpeg"
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
		return "NaN"
	}
	precision := 2
	return strconv.FormatFloat(1.0/f, 'f', precision, 64) + ":1"
}

func printStreamParams(stream *TStreamInfo) {
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
	fmt.Printf(" #%2v: %v\n", stream.Index, li)
	// fmt.Printf("    : comp %v chan: %v\n", stream.CompParams, str)
	fmt.Printf("    :??? %v, [%v], channels: %v\n", fround(I2), froundRatio(stream.CompParams.GetK()), str)
	fmt.Printf("    : %v\n", stream.MiscInfo.String())
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
		I:       ebur.I,
		MaxST:   ebur.MaxST,
		MinST:   ebur.MinST,
		STSum:   ebur.SumST,
		STCount: ebur.CountST,
	}
	return li, mi
}
