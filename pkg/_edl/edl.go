package edl

import (
	// "fmt"
)

type Edl struct {
	convErr bool
	Title string
	Clips []*Clip
}

func NewEdl(title string) *Edl {
	return &Edl{ Title: title }
}

func (o *Edl) Split() []*Edl {
	index := 0
	nextNotBL := func() *Clip {
		for index < len(o.Clips) {
			clip := o.Clips[index]
			index++
			if clip.Type != "BL" {
				return clip
			}
		}
		index = -1
		return nil
	}
	startNextEdl := func(clip *Clip) (*Edl, Timespan) {
		edl := NewEdl(o.Title)
		edl.Clips = append(edl.Clips, clip)
		return edl, clip.Timespan
	}
	clip := nextNotBL()
	if clip == nil {
		o.Clips = nil
		return []*Edl{ o }
	}
	var ret []*Edl

	edl, edlSpan := startNextEdl(clip)
	for clip = nextNotBL(); clip != nil; clip = nextNotBL() {
		if GapMs(edlSpan, clip.Timespan) != 0 {
			ret = append(ret, edl)
			edl, edlSpan = startNextEdl(clip)
			continue
		}
		edl.Clips = append(edl.Clips, clip)
		edlSpan = edlSpan.Unite(clip.Timespan)
	}
	if len(edl.Clips) != 0 {
		ret = append(ret, edl)
	}
	return ret
}

