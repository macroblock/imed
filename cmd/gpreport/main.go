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

	"golang.org/x/crypto/ssh/terminal"
)

var (
	inputPath = "input.txt"
	outputPath = "output.txt"

	useClipboard = false
)

func doJob(files []string) ([]string, error) {
	var ret []string

	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		fmt.Println(file)

		dir := filepath.Dir(file)
		file := filepath.Base(file)
		po := ""
		if x := strings.Split(dir, string(os.PathSeparator)); len(x) > 1 {
			po = x[1]
		}

		tn, err := tagname.NewFromFilename(file, false)
		if err != nil {
			return nil, err
		}
		typ, err := tn.GetTag("type")
		if typ == "poster" {
			return nil, fmt.Errorf("type must be %q, not %q", file, "poster.gp", typ)
		}

		jobType := "### Error ###"

		switch typ {
		default: return nil, fmt.Errorf("unsupported type %q", typ)
		case "poster.gp":
			ext, _ := tn.GetTag("ext")
			switch ext {
			default:
				return nil, fmt.Errorf("unsupported extension %q for type %q", ext, typ)
			case ".jpg":
				jobType = "Постер"
			case ".psd":
				jobType = "Постер (исходник)"
			} // switch ext
		} // switch typ

		s := file + "\t" + jobType + "\t" + "\t" + po + "\n"
		// fmt.Print(s)
		ret = append(ret, s)
	}
	return ret, nil
}

func subMain() error {
	for _, arg := range os.Args[1:] {
		switch arg {
		default: return fmt.Errorf("unsupported flag %q\n  Use -c to use clipboard mode")
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
