package loudnorm

import (
	"fmt"
	"math"
	"strconv"
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
	maxP := math.Inf(-1)
	minP := math.Inf(+1)
	for _, ch := range stream.astatsInfo.Channels {
		maxP = math.Max(maxP, ch.RMSLevel)
		minP = math.Min(minP, ch.RMSLevel)
	}
	dP := maxP - minP
	str := "" //fmt.Sprintf("%v %v: ", stream.astatsInfo.PeakLevel, len(stream.astatsInfo.Channels))
	for _, ch := range stream.astatsInfo.Channels {
		if maxP == 0.0 {
			str += "NaN "
			continue
		}
		str += strconv.Itoa(int(
			math.Round((ch.RMSLevel-minP)/dP*100),
		)) + " "
	}
	fmt.Printf(" #%2v: %v\n", stream.Index, stream.TargetLI)
	fmt.Printf("    : compression %v\n", stream.CompParams)
	fmt.Printf("    : %v\n", str)
}
