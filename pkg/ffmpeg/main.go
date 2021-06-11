package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// -
const (
	Stdout int = iota
	Stderr
)

// IParser -
type IParser interface {
	Parse(line string, eof bool) (accepted bool, err error)
	Finish() error
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {

	for i := 0; i < len(data); i++ {
		switch {
		case data[i] == '\n':
			return i + 1, data[:i], nil
		case data[i] == '\r':
			if i == len(data)-1 {
				if atEOF {
					// \r, EOF
					return i + 1, data[:i], nil
				}
				// \r, EOBuffer -> need more data
				return 0, nil, nil
			}
			if data[i+1] == '\n' {
				// \r, \n
				return i + 2, data[:i], nil
			}
			// \r, !\n
			return i + 1, data[:i], nil
		}
	}

	if atEOF {
		// + 1 brings to not stuck on empty buffer
		return len(data) + 1, data, nil
	}
	// need more data
	return 0, nil, nil
}


// Run -
func Run(parser IParser, args ...string) error {
	return RunPipe(Stderr, parser, args...)
}

// RunStderr -
func RunStderr(parser IParser, args ...string) error {
	return RunPipe(Stderr, parser, args...)
}

// RunStdout -
func RunStdout(parser IParser, args ...string) error {
	return RunPipe(Stdout, parser, args...)
}

// RunPipe -
func RunPipe(pipeSelector int, parser IParser, args ...string) error {
	c := exec.Command("ffmpeg", args...)
	if parser == nil {
		var errBuf bytes.Buffer
		c.Stdout = nil // &o
		c.Stderr = &errBuf
		err := c.Run()
		if err != nil {
			return errors.New(errBuf.String())
		}
		return nil
	}

	err := error(nil)
	pipe := io.ReadCloser(nil)

	switch pipeSelector {
	default: err = fmt.Errorf("unsupported pipe selector %v", pipeSelector)
	case Stderr:
		c.Stdout = nil
		pipe, err = c.StderrPipe()
	case Stdout:
		c.Stderr = nil
		pipe, err = c.StdoutPipe()
	}
	if err != nil {
		return err
	}

	err = c.Start()
	if err != nil {
		return err
	}

	buffer := []string{}
	scanner := bufio.NewScanner(pipe)
	scanner.Split(scanLines)
	ok := scanner.Scan()
	for ok {
		line := scanner.Text()
		// fmt.Printf("@@@@: %q\n", line)
		ok = scanner.Scan()
		accepted, err := parser.Parse(line, !ok) // !ok ==> EOF
		if err != nil {
			return err
		}
		if !accepted {
			buffer = append(buffer, line)
		}
	}
	err = c.Wait()
	if err != nil {
		return err
	}
	err = parser.Finish()
	if err != nil {
		err = fmt.Errorf("somewhere below:\n%v\n\nError: %v", strings.Join(buffer, "\n"), err)
		return err
	}
	return nil
}
