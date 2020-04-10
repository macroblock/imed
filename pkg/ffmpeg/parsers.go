package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

// IAudioProgress -
type IAudioProgress interface {
	Callback(Time) error
}

/////////////////////////////////////////////////////////////////////////////////////

type tCombineParser struct {
	parsers []IParser
}

// NewCombineParser -
func NewCombineParser(parsers ...IParser) IParser {
	return &tCombineParser{parsers: parsers}
}

// Parse -
func (o *tCombineParser) Parse(line string, eof bool) (accepted bool, finished bool, err error) {
	accepted = false
	finished = true
	err = error(nil)
	for i := range o.parsers {
		if o.parsers[i] == nil {
			continue
		}
		if accepted {
			if o.parsers[i] != nil {
				finished = false
			}
			continue
		}
		acc, fin, e := o.parsers[i].Parse(line, eof)
		accepted = acc
		if e != nil {
			return false, true, e
		}
		if fin {
			o.parsers[i] = nil
		} else {
			finished = false
		}
	}
	// return accepted, finished, nil
	return accepted, false, nil
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
