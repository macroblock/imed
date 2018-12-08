package tagname

import "strconv"

// HMSMSToSeconds - converts timecode (H:M:S.MS) to seconds float64 (S.MS).
func HMSMSToSeconds(hhmmssms string) float64 {
	var hh, mm, ss, ms float64
	var buffer string
	length := len(hhmmssms)
	timecode := []string{}

	for i := length - 1; i >= 0; i-- {
		if hhmmssms[i] == '.' {
			buffer = "." + buffer
			ms, _ = strconv.ParseFloat(buffer, 64)
			buffer = ""
		} else if hhmmssms[i] == ':' {
			timecode = append(timecode, buffer)
			buffer = ""
		} else if i == 0 {
			if buffer != "" {
				timecode = append(timecode, string(hhmmssms[i])+buffer)
			} else {
				timecode = append(timecode, string(hhmmssms[i]))
			}
		} else {
			buffer = string(hhmmssms[i]) + buffer
		}
	}

	length = len(timecode)

	if length == 1 {
		ss, _ = strconv.ParseFloat(timecode[0], 64)
	} else if length == 2 {
		ss, _ = strconv.ParseFloat(timecode[0], 64)
		mm, _ = strconv.ParseFloat(timecode[1], 64)
	} else if length == 3 {
		ss, _ = strconv.ParseFloat(timecode[0], 64)
		mm, _ = strconv.ParseFloat(timecode[1], 64)
		hh, _ = strconv.ParseFloat(timecode[2], 64)
	}

	return hh*3600 + mm*60 + ss + ms
}
