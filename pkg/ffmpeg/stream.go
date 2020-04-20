package ffmpeg

const (
	streamTypeUnknown = TStreamType(iota)
	streamTypeVideo
	streamTypeAudio
	streamTypeSubtitle
)

type (
	// TStreamType -
	TStreamType = int

	// TStream -
	TStream struct {
		owner *TFile
		typ   TStreamType // "audio", "video", "subtitle"
		index int
	}
)
