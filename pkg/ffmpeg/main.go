package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
)

// IParser -
type IParser interface {
	Parse(line string, eof bool) (accepted bool, finished bool, err error)
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
	c := exec.Command("ffmpeg", args...)
	// var o bytes.Buffer
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
	c.Stdout = nil //&o
	errPipe, err := c.StderrPipe()
	if err != nil {
		return err
	}
	err = c.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(errPipe)
	scanner.Split(scanLines)
	ok := scanner.Scan()
	for ok {
		line := scanner.Text()
		// fmt.Printf("@@@@: %q\n", line)
		ok = scanner.Scan()
		_, finished, err := parser.Parse(line, !ok) // !ok ==> EOF
		if err != nil {
			return err
		}
		if finished {
			break
		}
	}
	err = c.Wait()
	if err != nil {
		return err
	}
	return nil
}
