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
		ScanOnly: false,
	},
	Loudness: tLoudnessSettings{
		I:         -23,
		RA:        math.Inf(+1),
		TP:        -1.0,
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
		ScanOnly bool
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

func calcDuration(duration float64) (float64, error) {
	if duration <= 0 {
		return -1.0, nil
	}
	if settings.Edit.ClipPoint != nil {
		duration -= settings.Edit.ClipPoint.Float()
	}
	if settings.Edit.ClipDuration != nil {
		duration -= settings.Edit.ClipDuration.Float()
	}
	// if GlobalFlagSS != "" {
	// 	val, err := ffmpeg.ParseTime(GlobalFlagSS)
	// 	if err != nil {
	// 		return duration, err
	// 	}
	// 	duration -= val.Float()
	// }
	// if GlobalFlagT != "" {
	// 	val, err := ffmpeg.ParseTime(GlobalFlagT)
	// 	if err != nil {
	// 		return duration, err
	// 	}
	// 	duration -= val.Float()
	// }
	return duration, nil
}

func getGloblaFlags() []string {
	ret := []string{}
	if settings.Edit.ClipPoint != nil {
		ret = append(ret, "-ss", settings.Edit.ClipPoint.String())
	}
	if settings.Edit.ClipDuration != nil {
		ret = append(ret, "-t", settings.Edit.ClipDuration.String())
	}
	// if GlobalFlagSS != "" {
	// 	ret = append(ret, "-ss", GlobalFlagSS)
	// }
	// if GlobalFlagT != "" {
	// 	ret = append(ret, "-t", GlobalFlagT)
	// }
	return ret
}
