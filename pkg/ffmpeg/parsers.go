package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

// IAudioProgress -
type IAudioProgress interface {
	Callback(Timecode) error
}

/////////////////////////////////////////////////////////////////////////////////////

// TCombineParser -
type TCombineParser struct {
	parsers []IParser
}

// NewCombineParser -
func NewCombineParser(parsers ...IParser) *TCombineParser {
	return &TCombineParser{parsers: parsers}
}

// Append -
func (o *TCombineParser) Append(parsers ...IParser) {
	o.parsers = append(o.parsers, parsers...)
}

// Finish -
func (o *TCombineParser) Finish() error {
	for _, parser := range o.parsers {
		if parser == nil {
			continue
		}
		err := parser.Finish()
		if err != nil {
			return err
		}
	}
	return nil
}

// Parse -
func (o *TCombineParser) Parse(line string, eof bool) (bool, error) {
	for i := range o.parsers {
		accepted, err := o.parsers[i].Parse(line, eof)
		if err != nil {
			return false, err
		}
		if accepted {
			return true, nil
		}
	}
	return false, nil
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
