package loudnorm

import (
	"fmt"
	"strconv"
	"strings"
)

// PackLoudnessInfoElement -
func PackLoudnessInfoElement(streamNo int, li *LoudnessInfo) string {
	return fmt.Sprintf("[Stream #:%v]\nL_I  %v\nL_RA %v\nL_TP %v\nL_TH %v",
		strconv.Itoa(streamNo),
		alignStr(4, li.I),
		alignStr(4, li.RA),
		alignStr(4, li.TP),
		alignStr(4, li.TH),
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

	dict := map[int]*LoudnessInfo{}

	for list = skipBlank(list); len(list) != 0; list = skipBlank(list) {
		var (
			streamNo, I, RA, TP, TH string
			err                     error
		)

		list, streamNo, err = parseVal(list, "[Stream #:", "]")
		if err != nil {
			return err
		}
		n, err := strconv.Atoi(streamNo)
		if err != nil {
			return err
		}
		list, I, err = parseVal(list, "L_I", "")
		if err != nil {
			return err
		}
		list, RA, err = parseVal(list, "L_RA", "")
		if err != nil {
			return err
		}
		list, TP, err = parseVal(list, "L_TP", "")
		if err != nil {
			return err
		}
		list, TH, err = parseVal(list, "L_TH", "")
		if err != nil {
			return err
		}
		li := &LoudnessInfo{
			I:  I,
			RA: RA,
			TP: TP,
			TH: TH,
		}
		dict[n] = li
	}
	for _, stream := range fi.Streams {
		if li, ok := dict[stream.Index]; ok {
			stream.LoudnessInfo = li
		}
	}

	return nil
}
