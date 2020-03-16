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

type optionsJSON struct {
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

// Options -
type Options struct {
	InputI            float64
	InputTP           float64
	InputLRA          float64
	InputThresh       float64
	OutputI           float64
	OutputTP          float64
	OutputThresh      float64
	NormalizationType string
	TargetOffset      float64
}

// OptionsLight -
type OptionsLight struct {
	InputI      float64
	InputThresh float64

	InputLRA     float64
	InputThresh2 float64
	InputLRALow  float64
	InputLRAHigh float64

	InputTP float64
}

// LoudnessInfo -
type LoudnessInfo struct {
	I  float64 // integrated
	RA float64 // range
	TP float64 // true peaks
	TH float64 // threshold
}

func (o *LoudnessInfo) String() string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf("I: %v, RA: %v, TP: %v, TH: %v", o.I, o.RA, o.TP, o.TH)
}

// SetTargetLI -
func SetTargetLI(li float64) {
	targetI = li
}

// SetTargetLRA -
func SetTargetLRA(lra float64) {
	targetLRA = lra
}

// SetTargetTP -
func SetTargetTP(tp float64) {
	targetTP = tp
}

// Scan -
func Scan(filePath string, trackN int) (*Options, error) {
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
		osNullDevice,
	}
	c := exec.Command("ffmpeg", params...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err := c.Run()
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

	x := &optionsJSON{}
	err = json.Unmarshal([]byte(strings.Join(jsonList, "\n")), &x)
	if err != nil {
		return nil, err
	}

	ret := &Options{
		NormalizationType: x.NormalizationType,
	}
	ret.InputI, err = strconv.ParseFloat(x.InputI, 64)
	if err != nil {
		return nil, err
	}
	ret.InputLRA, err = strconv.ParseFloat(x.InputLRA, 64)
	if err != nil {
		return nil, err
	}
	ret.InputTP, err = strconv.ParseFloat(x.InputTP, 64)
	if err != nil {
		return nil, err
	}
	ret.InputThresh, err = strconv.ParseFloat(x.InputThresh, 64)
	if err != nil {
		return nil, err
	}
	ret.OutputI, err = strconv.ParseFloat(x.OutputI, 64)
	if err != nil {
		return nil, err
	}
	ret.OutputTP, err = strconv.ParseFloat(x.OutputTP, 64)
	if err != nil {
		return nil, err
	}
	ret.OutputThresh, err = strconv.ParseFloat(x.OutputThresh, 64)
	if err != nil {
		return nil, err
	}
	ret.TargetOffset, err = strconv.ParseFloat(x.TargetOffset, 64)
	if err != nil {
		return nil, err
	}

	return ret, nil
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
		osNullDevice,
	}
	if GlobalDebug {
		fmt.Println("### params: ", params)
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

	re := regexp.MustCompile("\\[Parsed_ebur128_0 @ [^ ]+\\] Summary:.*")

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

	fmt.Println("???????")
	optsLight, err := parseEbur128Summary(strList)
	fmt.Println("#######", err)
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

// NormalizeTo -
func NormalizeTo(filePath string, trackN int, fileOut string, audioParams []string, inputI, inputLRA, inputTP, inputThresh float64) (*Options, error) {
	params := []string{
		"-y",
		"-hide_banner",
		"-i", filePath,
		"-map", "0:" + strconv.Itoa(trackN),
		"-filter:a",
		"loudnorm=print_format=json" +
			":linear=true" +
			// ":linear=false" +
			fmt.Sprintf(":I=%.2f:LRA=%.2f:TP=%.2f",
				targetI, targetLRA, targetTP) +
			fmt.Sprintf(":measured_I=%.2f:measured_LRA=%.2f:measured_TP=%.2f:measured_thresh=%.2f",
				inputI, inputLRA, inputTP, inputThresh) +
			// ":offset=" + opts.TargetOffset,  // it's just difference between internal target_i and i_out
			// "-f", "flac",
			"",
	}
	params = append(params, audioParams...)
	params = append(
		params,
		"-ar:a", samplerate,
		fileOut,
	)

	if GlobalDebug {
		fmt.Println("### params: ", params)
	}
	c := exec.Command("ffmpeg", params...)
	var o bytes.Buffer
	var e bytes.Buffer
	c.Stdout = &o
	c.Stderr = &e
	err := c.Run()
	if err != nil {
		fmt.Println("###:", e.String())
		return nil, err
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
			continue
		}
		if found {
			jsonList = append(jsonList, line)
		}
	}

	opts := &Options{}
	err = json.Unmarshal([]byte(strings.Join(jsonList, "\n")), &opts)
	if err != nil {
		return nil, err
	}

	// fmt.Println(strings.Join(jsonList, "\n"))
	return opts, nil
}
