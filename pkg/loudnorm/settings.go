package loudnorm

import (
	"math"

	"github.com/macroblock/imed/pkg/ffmpeg"
)

// -
// var (
// 	GlobalFlagT  = ""
// 	GlobalFlagSS = ""
// )

var settings = TSettings{
	Behavior: tBehavior{
		ScanOnly:    false,
		ForceStereo: false,
	},
	Loudness: tLoudnessSettings{
		I:         -23,
		RA:        math.Inf(+1),
		TP:        math.NaN(), // -1.0,
		MP:        0.0,
		Precision: 0.5,
	},
	Compressor: tCompressorSettings{
		Attack:         0.000, // 0.000,
		Release:        0.050, // 0.010,
		NumTries:       5,
		CorrectionStep: 0.1,
	},
	Edit: tEditSettings{
		ClipPoint:    nil,
		ClipDuration: nil,
	},
}

type (
	// TSettings -
	TSettings struct {
		Behavior   tBehavior
		Loudness   tLoudnessSettings
		Compressor tCompressorSettings
		Edit       tEditSettings
	}

	tBehavior struct {
		ScanOnly    bool
		ForceStereo bool
	}

	tLoudnessSettings struct {
		I         float64
		RA        float64
		TP        float64
		MP        float64
		Precision float64
	}
	tCompressorSettings struct {
		Attack         float64
		Release        float64
		NumTries       int
		CorrectionStep float64
	}
	tEditSettings struct {
		ClipPoint    *ffmpeg.Time
		ClipDuration *ffmpeg.Time
	}
)

// GetSettings -
func GetSettings() TSettings {
	return settings
}

// SetSettings -
func SetSettings(x TSettings) {
	settings = x
}

func (o TSettings) calcDuration(duration float64) (float64, error) {
	if duration <= 0 {
		return -1.0, nil
	}
	if o.Edit.ClipPoint != nil {
		duration -= o.Edit.ClipPoint.Float()
	}
	if o.Edit.ClipDuration != nil {
		duration = math.Min(duration, o.Edit.ClipDuration.Float())
	}
	return duration, nil
}

func (o TSettings) getGlobalFlags() []string {
	ret := []string{}
	if o.Edit.ClipPoint != nil {
		ret = append(ret, "-ss", o.Edit.ClipPoint.String())
	}
	if o.Edit.ClipDuration != nil {
		ret = append(ret, "-t", o.Edit.ClipDuration.String())
	}
	return ret
}
