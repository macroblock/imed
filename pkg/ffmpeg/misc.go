package ffmpeg

import (
	"fmt"
	"strings"
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

// ArgsToStrings -
func ArgsToStrings(args []interface{}) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	ret := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg.(type) {
		default:
			ret = append(ret, fmt.Sprintf("%v", arg))
		}
	}
	return ret, nil
}

func getFilter(typ TStreamType, name string, args ...interface{}) (string, error) {
	a, err := ArgsToStrings(args)
	if err != nil {
		return "", err
	}
	switch typ {
	default:
		return "", fmt.Errorf("only video and audio stream types are allowed")
	case streamTypeVideo:
	case streamTypeAudio:
		name = "a" + name
	}
	ret := strings.Join(
		append([]string{name}, a...),
		"=",
	)
	return ret, nil
}
