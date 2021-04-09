package main

import (
	"bufio"
	"fmt"
	"io"
	// "log"
	"os"
	"path/filepath"
	// "regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/rtimg/pkg"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	inputPath = "input.txt"
	outputPath = "output.txt"

	useClipboard = false
)

func findProvider(key *rtimg.TKey) string {
	ret := ""
	for i := 0; ; i++ {
		seg, ok := key.Segment(i)
		if !ok {
			return ""
		}
		if seg == "posters" {
			return ret
		}
		ret = seg
	}
}

func doJob(files []string) ([]string, error) {
	var ret []string
	var errors []string

	appendError := func(name string, err error) {
		if err != nil {
			errors = append(errors, fmt.Sprintf("%v:\n    %v",name, err))
		}
	}

	for _, filePath := range files {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}
		filePath, err := filepath.Abs(filePath)
		if err != nil {
			appendError(filePath, err)
			continue
		}

		tn, err := tagname.NewFromFilename(filePath, false)
		if err != nil {
			tn = nil
		}
		key, err := rtimg.FindKey(filePath, tn)
		if err != nil {
			appendError(filePath, err)
			continue
		}

		name := key.Name()
		data := key.Data()
		if data == nil {
			appendError(filePath, fmt.Errorf("unreachable: something wrong with a <key>"))
			continue
		}
		po := findProvider(key)
		if po == "" {
			appendError(filePath, fmt.Errorf("cannot detect segment 'pravoobladatel'"))
			continue
		}

		jobType := "### Error ###"
		switch data.Type {
		default:
			appendError(filePath, fmt.Errorf("unsupported type %q", data.Type))
			continue
		case "gp":
			ext := filepath.Ext(filePath)
			switch ext {
			default:
				appendError(filePath, fmt.Errorf("unsupported extension %q for type %q", ext, data.Type))
				continue
			case ".jpg":
			case ".png":
				jobType = "Постер"
			case ".psd":
				jobType = "Постер (исходник)"
			} // switch ext
		} // switch data.Type

		s := name + "\t" + filepath.Base(filePath) + "\t" + jobType + "\t" + "\t" + po + "\n"
		ret = append(ret, s)
	}
	// fmt.Println(ret)
	if len(errors) > 0 {
		return nil, fmt.Errorf("%v", strings.Join(errors, "\n"))
	}
	return ret, nil
}

func subMain() error {
	for _, arg := range os.Args[1:] {
		switch arg {
		default: return fmt.Errorf("unsupported flag %q\n  Use -c to use clipboard mode", arg)
		case "-c": useClipboard = true
		}
	}

	files, err := readLines(inputPath)
	if err != nil {
		return err
	}
	output, err := doJob(files)
	if err != nil {
		return err
	}
	writeStringArrayTo(outputPath, output, 0775)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := subMain()
	if err != nil {
		fmt.Println(err)

		if useClipboard {
			fmt.Println("Press any key to continue...")
			err := waitForAnyKey()
			if err != nil {
				fmt.Println(err)
			}
		}
		os.Exit(-1)
	}
}

func readLines(path string) ([]string, error) {
	var reader io.Reader

	if useClipboard {
		if clipboard.Unsupported {
			return nil, fmt.Errorf("%s", "clipboard unsupported for this OS")
		}
		text, err := clipboard.ReadAll()
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(text)
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeStringArrayTo(filename string, strArray []string, perm os.FileMode) error {
	if useClipboard {
		clipboard.WriteAll(strings.Join(strArray, ""))
		return nil
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		// log.Panic(err)
		return err
	}
	defer f.Close()
	for _, v := range strArray {
		if _, err = f.WriteString(v); err != nil {
			return err
		}
	}
	return nil
}

// waitForAnyKey await for any key press to continue.
func waitForAnyKey() error {
	fd := int(os.Stdin.Fd())
	if !terminal.IsTerminal(fd) {
		return fmt.Errorf("it's not a terminal descriptor")
	}
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("cannot set raw mode")
	}
	defer terminal.Restore(fd, state)

	b := [1]byte{}
	os.Stdin.Read(b[:])
	return nil
}
