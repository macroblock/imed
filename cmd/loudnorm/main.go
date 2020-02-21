package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/translit"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()

	flagFiles     []string
	flagClipboard bool
	flagU         bool
	flagD         string
)

func doProcess(path string) {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	dir, name := filepath.Split(path)
	ext := ""

	file, err := os.Open(path)
	retif.Error(err, "cannot open file: ", path)

	stat, err := file.Stat()
	retif.Error(err, "cannot get filestat: ", path)

	err = file.Close()
	retif.Error(err, "cannot close file: ", path)

	if !stat.IsDir() {
		ext = filepath.Ext(path)
	}
	name = strings.TrimSuffix(name, ext)
	name, _ = translit.Do(name)
	name = upper(name)
	err = os.Rename(path, dir+name+ext)
	retif.Error(err, "cannot rename file")

	log.Notice("result: " + dir + name + ext)
}

func upper(str string) string {
	if !flagU {
		return str
	}
	s := ""
	for i, r := range str {
		if i == 0 {
			r = unicode.ToUpper(r)
		}
		s += string(r)
	}
	return s
}

// LoudOptions -
type LoudOptions struct {
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

const (
	targetI    = "-23.0"
	targetLRA  = "10.0"
	targetTP   = "-1.0"
	samplerate = "48k"
)

// LoudInfo -
func LoudInfo(filePath string) (opts *LoudOptions, err error) {
	params := []string{
		"-hide_banner",
		"-i", filePath,
		"-filter:a",
		"loudnorm=print_format=json" +
			":I=" + targetI +
			":LRA=" + targetLRA +
			":TP=" + targetTP,
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
		return nil, fmt.Errorf("size of output info too small")
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

// LoudProcess -
func LoudProcess(filePath string, opts *LoudOptions) error {
	params := []string{
		"-hide_banner",
		"-i", filePath,
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
			":offset=" + opts.TargetOffset,
		// "-f", "flac",
		"-codec:a", "flac",
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

func mainFunc() error {
	if len(flagFiles) == 0 && !flagClipboard {
		return cli.ErrorNotEnoughArguments()
	}

	opts, err := LoudInfo(flagFiles[0])
	if err != nil {
		return err
	}

	fmt.Println("opts: ", opts)

	err = LoudProcess(flagFiles[0], opts)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// setup log
	newLogger := misc.NewSimpleLogger
	if misc.IsTerminal() {
		newLogger = misc.NewAnsiLogger
	}
	log.Add(
		newLogger(loglevel.Warning.OrLower(), ""),
		newLogger(loglevel.Info.Only().Include(loglevel.Notice.Only()), "~x\n"),
	)

	defer func() {
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// command line interface
	cmdLine := cli.New("!PROG! the program that translit text in filenames or|and clipboard.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information about that flag."),
		cli.Flag("-h -help      : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-c -clipboard : transtlit clipboard data.", &flagClipboard),
		cli.Flag("-u            : upper case first letter.", &flagU),
		cli.Flag("-d -delimiter : delimiter to separate multiple files. CR by default.", &flagD),
		cli.Flag(": files to be processed", &flagFiles),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
