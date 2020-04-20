package loudnorm

import (
	"fmt"
	"strconv"
	"strings"
)

// PackLoudnessInfoElement -
func PackLoudnessInfoElement(streamNo int, li *TLoudnessInfo) string {
	return fmt.Sprintf("[Stream #:%v]\nL_I  % 6.2f\nL_RA % 6.2f\nL_TP % 6.2f\nL_TH % 6.2f\nL_MP % 6.2f",
		strconv.Itoa(streamNo),
		li.I, li.RA, li.TP, li.TH, li.MP,
	)
}

// PackLoudnessInfo -
func PackLoudnessInfo(fi *TFileInfo) string {
	list := []string{}
	for _, stream := range fi.Streams {
		if stream.LoudnessInfo != nil {
			list = append(list, PackLoudnessInfoElement(stream.Index, stream.LoudnessInfo))
		}
	}
	return strings.Join(list, "\n")
}

// AttachLoudnessInfo -
func AttachLoudnessInfo(fi *TFileInfo, data string) error {
	list := strings.Split(data, "\n")

	dict := map[int]*TLoudnessInfo{}

	for list = skipBlank(list); len(list) != 0; list = skipBlank(list) {
		var (
			streamNo          int
			I, RA, TP, TH, MP float64
			err               error
		)

		list, streamNo, err = parseValI(list, "[Stream #:", "]")
		if err != nil {
			// return err
			break
		}
		list, I, err = parseValF(list, "L_I", "")
		if err != nil {
			return err
		}
		list, RA, err = parseValF(list, "L_RA", "")
		if err != nil {
			return err
		}
		list, TP, err = parseValF(list, "L_TP", "")
		if err != nil {
			return err
		}
		list, TH, err = parseValF(list, "L_TH", "")
		if err != nil {
			return err
		}
		list, MP, err = parseValF(list, "L_MP", "")
		if err != nil {
			return err
		}

		li := &TLoudnessInfo{
			I:  I,
			RA: RA,
			TP: TP,
			TH: TH,
			MP: MP,
		}
		dict[streamNo] = li
	}
	for _, stream := range fi.Streams {
		if li, ok := dict[stream.Index]; ok {
			stream.LoudnessInfo = li
		}
	}

	return nil
}

func skipBlank(list []string) []string {
	for len(list) > 0 && strings.TrimSpace(list[0]) == "" {
		list = list[1:]
	}
	return list
}

func parseVal(list []string, prefix string, trimSuffix string) ([]string, string, error) {
	list = skipBlank(list)
	if len(list) == 0 {
		return nil, "", fmt.Errorf("not enough data")
	}
	s := strings.TrimSpace(list[0])
	if !strings.HasPrefix(s, prefix) {
		return nil, "", fmt.Errorf("does not have prefix %q", prefix)
	}
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, trimSuffix)
	s = strings.TrimSpace(s)
	return list[1:], s, nil
}

func parseValS(list []string, prefix string, trimSuffix string) ([]string, string, error) {
	return parseVal(list, prefix, trimSuffix)
}

func parseValI(list []string, prefix string, trimSuffix string) ([]string, int, error) {
	newList, s, err := parseVal(list, prefix, trimSuffix)
	if err != nil {
		return newList, 0, err
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return newList, 0, err
	}
	return newList, int(val), nil
}

func parseValF(list []string, prefix string, trimSuffix string) ([]string, float64, error) {
	newList, s, err := parseVal(list, prefix, trimSuffix)
	if err != nil {
		return newList, 0.0, err
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return newList, 0.0, err
	}
	return newList, val, nil
}

func parseEbur128Summary(list []string) (*OptionsLight, error) {
	// fmt.Printf("$$$$$$$$\n%q\n", strings.Join(list, "\\n\n"))
	list, _, err := parseValS(list, "Integrated loudness:", "")
	if err != nil {
		return nil, err
	}
	list, I, err := parseValF(list, "I:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, Threshold, err := parseValF(list, "Threshold:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, _, err = parseValS(list, "Loudness range:", "")
	if err != nil {
		return nil, err
	}
	list, LRA, err := parseValF(list, "LRA:", "LU")
	if err != nil {
		return nil, err
	}
	list, Threshold2, err := parseValF(list, "Threshold:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, LRALow, err := parseValF(list, "LRA low:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, LRAHigh, err := parseValF(list, "LRA high:", "LUFS")
	if err != nil {
		return nil, err
	}
	list, _, err = parseValS(list, "True peak:", "")
	if err != nil {
		return nil, err
	}

	list, TP, err := parseValF(list, "Peak:", "dBFS")
	if err != nil {
		return nil, err
	}
	ret := &OptionsLight{
		InputI:       I,
		InputThresh:  Threshold,
		InputLRA:     LRA,
		InputThresh2: Threshold2,
		InputLRALow:  LRALow,
		InputLRAHigh: LRAHigh,
		InputTP:      TP,
	}
	return ret, nil
}
