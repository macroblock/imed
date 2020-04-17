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

// -
const (
	NodeUnknown TNodeType = iota
	NodeInput
	NodeOutput
	NodeFilter
)

type (
	// TNodeType -
	TNodeType int

	// TFilterLine -
	TFilterLine struct {
		owner     *TFilterChain
		typ       TNodeType
		key       string
		cachedKey string
		line      string
		out       []*TFilterLine
	}
	// TFilterChain -
	TFilterChain struct {
		TFilterLine

		localBase int
		inputIndex,
		streamIndex,
		localIndex int
	}
)

// NewFilterChain -
func NewFilterChain(input, stream, local int) *TFilterChain {
	ret := &TFilterChain{inputIndex: input, streamIndex: stream, localIndex: local}
	ret.key = ret.GetKey()
	return ret
}

// GetKey -
func (o *TFilterChain) GetKey() string {
	if o.key == "" {
		o.key = fmt.Sprintf("%vx%v", o.inputIndex, o.streamIndex)
		o.key = strings.ReplaceAll(o.key, "-", "m")
	}
	return o.key
}

func newFilterLine(owner *TFilterChain, typ TNodeType) *TFilterLine {
	if owner == nil {
		panic("newFilterLine: owner must be not null")
	}
	ret := &TFilterLine{owner: owner, typ: typ}
	ret.key = ret.getLocalKey()
	return ret
}

func (o *TFilterLine) getLocalKey() string {
	ret := fmt.Sprintf("%v", o.owner.localBase)
	o.owner.localBase++
	return ret
}

// Split -
func (o *TFilterLine) Split() *TFilterLine {
	return newFilterLine(o.owner, NodeFilter)
}
