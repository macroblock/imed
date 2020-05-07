package loudnorm

import (
	"fmt"
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
