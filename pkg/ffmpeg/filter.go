package ffmpeg

import (
	"fmt"
	"strings"
)

type (
	// TFilterSpan -
	TFilterSpan struct {
		owner      *TFilterGraph
		outputName string
		filters    []string
		outs       []*TFilterSpan
	}

	// TFilterInput -
	TFilterInput struct {
		TFilterSpan
		inputStreams []*TStream
	}

	tctx struct {
		inputIndex  int
		splitIndex  int
		outIndex    int
		files       []*TFile
		filemap     map[string]int
		inputs      []string
		filterSpans []string
		outputs     []string
	}

	// TFilterGraph -
	TFilterGraph struct {
		inputs  []*TFilterInput
		outputs []*TFilterSpan
		ctx     *tctx
	}
)

// NewFilterGraph -
func NewFilterGraph() *TFilterGraph {
	ret := &TFilterGraph{}
	return ret
}

func newFilterSpan(owner *TFilterGraph) *TFilterSpan {
	return &TFilterSpan{owner: owner}
}

// NewInput -
func (o *TFilterGraph) NewInput(streams []*TStream) *TFilterSpan {
	ret := &TFilterInput{
		TFilterSpan:  *newFilterSpan(o),
		inputStreams: streams,
	}
	o.inputs = append(o.inputs, ret)
	return &ret.TFilterSpan
}

// Probe -
func (o *TFilterGraph) Probe() error {
	for _, i := range o.inputs {
		for _, s := range i.inputStreams {
			err := s.owner.Probe()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ResetGraph -
func (o *TFilterGraph) ResetGraph() {
	o.inputs = nil
	o.outputs = nil
}

// Build -
func (o *TFilterGraph) Build() ([]string, error) {
	ctx := &tctx{filemap: map[string]int{}}
	o.ctx = ctx
	defer func() { o.ctx = nil }()

	for _, input := range o.inputs {
		err := input.build()
		if err != nil {
			return nil, err
		}
	}
	ret := ctx.inputs
	ret = append(ret, ctx.filterSpans...)
	ret = append(ret, ctx.outputs...)
	return ret, nil
}

func (o *TFilterSpan) build(inLink string, typ TStreamType) error {
	if o.outputName == "" {
		return o.buildSpan(inLink, typ)
	}
	return o.buildOutput(inLink, typ)
}

func (o *TFilterSpan) buildOutput(inLink string, typ TStreamType) error {
	// ctx := o.owner.ctx
	// span := inLink
	return nil
}

func (o *TFilterSpan) buildSpan(inLink string, typ TStreamType) error {
	ctx := o.owner.ctx
	span := inLink
	filters := strings.Join(o.filters, ":")
	if filters == "" {
		null, err := getFilter(typ, "null")
		if err != nil {
			return err
		}
		filters = null
	}
	span += filters
	split, err := getFilter(typ, "split", len(o.outs))
	if err != nil {
		return err
	}
	span += split
	for _, split := range o.outs {
		link := fmt.Sprintf("[s%v]", ctx.splitIndex)
		span += link
		ctx.splitIndex++
		err := split.build(link, typ)
		if err != nil {
			return err
		}
	}
	ctx.filterSpans = append(ctx.filterSpans, span)
	return nil
}

func (o *TFilterInput) build() error {
	ctx := o.owner.ctx
	inLink := ""
	typ := streamTypeUnknown
	for _, stream := range o.inputStreams {
		if typ == streamTypeUnknown {
			typ = stream.typ
		}
		if typ != stream.typ {
			return fmt.Errorf("different stream types in one input")
		}
		file := stream.owner
		index, ok := ctx.filemap[file.name]
		if !ok {
			index = ctx.inputIndex
			ctx.inputIndex++

			ctx.filemap[file.name] = index
			ctx.files = append(ctx.files, file)
		}
		inLink += fmt.Sprintf("[%v:%v]", index, stream.index)
	}
	if typ == streamTypeUnknown {
		return fmt.Errorf("unknown stream type in the input")
	}
	return o.TFilterSpan.build(inLink, typ)
}

///////////////////////////////////////////////////////////////////////

// Split -
func (o *TFilterSpan) Split() *TFilterSpan {
	out := newFilterSpan(o.owner)
	o.outs = append(o.outs, out)
	return out
}

// Filter -
func (o *TFilterSpan) Filter(filter string) *TFilterSpan {
	o.filters = append(o.filters, filter)
	return o
}

// Output -
func (o *TFilterSpan) Output(filename string, options ...interface{}) {
	opts, err := ArgsToStrings(options)
	if err != nil {
		panic("here")
	}
	out := newFilterSpan(o.owner)
	out.outputName = filename
	out.filters = opts
	o.outs = append(o.outs, out)
	o.owner.outputs = append(o.owner.outputs, out)
}
