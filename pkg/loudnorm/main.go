package loudnorm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	targetI    = "-23.0"
	targetLRA  = "20.0"
	targetTP   = "-1.0"
	samplerate = "48k"
)

// Options -
type Options struct {
	InputI            string `json:"input_i"`
	InputTP           string `json:"input_tp"`
	InputLRA          string `json:"input_lra"`
	InputThresh       string `json:"input_thresh"`
	OutputI           string `json:"output_i"`
	OutputTP          string `json:"output_tp"`
	OutputThresh      string `json:"output_thresh"`
	NormalizationType string `json:"normalization_type"`
	TargetOffset      string `json:"target_offset"`
}

// OptionsLight -
type OptionsLight struct {
	InputI      string
	InputThresh string

	InputLRA     string
	InputThresh2 string
	InputLRALow  string
	InputLRAHigh string

	InputTP string
}

// Scan -
func Scan(filePath string, trackN int) (opts *Options, err error) {
	params := []string{
		"-hide_banner",
		"-i", filePath,
		"-map", "0:" + strconv.Itoa(trackN),
		"-filter:a",
		"loudnorm=print_format=json" +
			// ":I=" + targetI +
			// ":LRA=" + targetLRA +
			// ":TP=" + targetTP,
			// ":linear=true" +
			"",
		"-f", "null",
		"NUL",
	}
	c := exec.Command("ffmpeg", params...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err = c.Run()
	if err != nil {
		return nil, errors.New(string(e.Bytes()))
	}
	list := strings.Split(e.String(), "\n")

	if len(list) < 12 {
		fmt.Println(strings.Join(list, "\n"))
		return nil, fmt.Errorf("size of an output info too small")
	}

	found := false
	jsonList := []string{}
	for _, line := range list {
		if strings.HasPrefix(line, "[Parsed_loudnorm_0 @") {
			found = true
			// jsonList = []string{"{"}
			continue
		}
		if !found {
			continue
		}
		jsonList = append(jsonList, line)
	}

	err = json.Unmarshal([]byte(strings.Join(jsonList, "\n")), &opts)
	if err != nil {
		return nil, err
	}

	// fmt.Println(strings.Join(jsonList, "\n"))
	return opts, nil
}

// ScanLight -
func ScanLight(filePath string, trackN int) (opts *OptionsLight, err error) {
	params := []string{
		"-hide_banner",
		"-i", filePath,
		"-map", "0:" + strconv.Itoa(trackN),
		"-filter:a",
		"ebur128" +
			"=peak=true",
		"-f", "null",
		"NUL",
	}
	c := exec.Command("ffmpeg", params...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err = c.Run()
	if err != nil {
		return nil, errors.New(string(e.Bytes()))
	}

	list := strings.Split(e.String(), "\n")

	re := regexp.MustCompile("\\[Parsed_ebur128_0 @ ........\\] Summary:.*")

	found := false
	strList := []string{}
	for _, line := range list {
		if re.MatchString(line) {
			found = true
			continue
		}
		if !found {
			continue
		}
		strList = append(strList, line)
	}

	optsLight, err := parseEbur128Summary(strList)
	if err != nil {
		return nil, err
	}

	// fmt.Println(strings.Join(jsonList, "\n"))
	return optsLight, nil
}

func skipBlank(list []string) []string {
	for len(list) > 0 && strings.TrimSpace(list[0]) == "" {
		list = list[1:]
	}
	return list
}

func parseVal(list []string, prefix string) ([]string, string, error) {
	list = skipBlank(list)
	if len(list) == 0 {
		return nil, "", fmt.Errorf("not enough data")
	}
	s := strings.TrimSpace(list[0])
	if !strings.HasPrefix(s, prefix) {
		return nil, "", fmt.Errorf("does not have prefix %q", prefix)
	}
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, "LU")
	s = strings.TrimSuffix(s, "LUFS")
	s = strings.TrimSuffix(s, "dBFS")
	s = strings.TrimSpace(s)
	return list[1:], s, nil
}

func parseEbur128Summary(list []string) (*OptionsLight, error) {
	list, _, err := parseVal(list, "Integrated loudness:")
	if err != nil {
		return nil, err
	}
	list, I, err := parseVal(list, "I:")
	if err != nil {
		return nil, err
	}
	list, Threshold, err := parseVal(list, "Threshold:")
	if err != nil {
		return nil, err
	}
	list, _, err = parseVal(list, "Loudness range:")
	if err != nil {
		return nil, err
	}
	list, LRA, err := parseVal(list, "LRA:")
	if err != nil {
		return nil, err
	}
	list, Threshold2, err := parseVal(list, "Threshold:")
	if err != nil {
		return nil, err
	}
	list, LRALow, err := parseVal(list, "LRA low:")
	if err != nil {
		return nil, err
	}
	list, LRAHigh, err := parseVal(list, "LRA high:")
	if err != nil {
		return nil, err
	}
	list, _, err = parseVal(list, "True peak:")
	if err != nil {
		return nil, err
	}

	list, TP, err := parseVal(list, "Peak:")
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

// Process -
func Process(filePath string, trackN int, opts *Options) error {
	params := []string{
		"-hide_banner",
		"-i", filePath,
		"-map", "0:" + strconv.Itoa(trackN),
		"-filter:a",
		"loudnorm=print_format=summary" +
			":linear=true" +
			":I=" + targetI +
			":LRA=" + targetLRA +
			":TP=" + targetTP +
			":measured_I=" + opts.InputI +
			":measured_LRA=" + opts.InputLRA +
			":measured_TP=" + opts.InputTP +
			":measured_thresh=" + opts.InputThresh +
			// ":offset=" + opts.TargetOffset,  // it's just difference between internal target_i and i_out
			// "-f", "flac",
			"-codec:a", "flac",
		"-r:a", samplerate,
		// "-ac", "6"
		"-y",
		"test.flac",
	}
	fmt.Println("params: ", params)
	c := exec.Command("ffmpeg", params...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err := c.Run()
	if err != nil {
		fmt.Println("###:", e.String())
		return err
	}
	fmt.Println("###done:", e.String())
	return nil
}
