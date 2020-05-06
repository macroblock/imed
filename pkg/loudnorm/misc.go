package loudnorm

import "fmt"

func debugPrintf(pattern string, args ...interface{}) {
	if !GlobalDebug {
		return
	}
	fmt.Printf(pattern, args...)
}
