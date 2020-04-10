package ffmpeg

import (
	"fmt"
	"sync"
)

var mtx sync.Mutex
var id int

// UniqueReset -
func UniqueReset() {
	mtx.Lock()
	defer mtx.Unlock()
	id = 0
}

// UniqueName -
func UniqueName(name string) string {
	mtx.Lock()
	defer mtx.Unlock()
	ret := fmt.Sprintf("%v@%v", name, id)
	id++
	return ret
}
