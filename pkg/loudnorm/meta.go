package loudnorm

import (
	"fmt"
	"strconv"
	"strings"
)

// PackLoudnessInfoElement -
func PackLoudnessInfoElement(streamNo int, li *TLoudnessInfo) string {
	return fmt.Sprintf("[Stream #:%v]\nL_I : % 6.2f\nL_RA: % 6.2f\nL_TP: %v\nL_TH: % 6.2f\nL_MP: % 6.2f\nL_CR: % 6.2f",
		strconv.Itoa(streamNo),
		li.I, li.RA, li.TP, li.TH, li.MP, li.CR,
	)
}

// PackTargetLoudnessInfo -
func PackTargetLoudnessInfo(fi *TFileInfo) string {
	list := []string{}
	for _, stream := range fi.Streams {
		if stream.TargetLI != nil {
			list = append(list, PackLoudnessInfoElement(stream.Index, stream.TargetLI))
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
			I, RA, TP, MP, TH float64
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
			MP: MP,
			TH: TH,
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
