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
