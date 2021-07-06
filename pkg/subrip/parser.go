package subrip

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseError(ln int, line string, msg string) error {
	return fmt.Errorf("line:%v: expected %v, got %q", ln, msg, line)
}

func Parse(rd io.Reader) ([]Record, error) {
	const (
		stReady = iota
		stTime
		stTextStart
		stText
	)
	// skip BOM
	br := bufio.NewReader(rd)
	{
		r, _, err := br.ReadRune()
		if err != nil {
			return nil, err
		}
		if r != '\ufeff' {
			br.UnreadRune()
		}
	}

	scanner := bufio.NewScanner(br)
	state := stReady
	ln := 0
	r := Record{}
	ret := []Record{}
	for scanner.Scan() {
		line := scanner.Text()
		ln++
		switch state {
		default:
			panic("unreachable")
		case stReady:
			if len(line) == 0 {
				continue
			}
			// chunk number
			err := error(nil)
			r.ID, err = strconv.Atoi(line)
			if err != nil {
				return nil, parseError(ln, line, "chunk number")
			}
			state = stTime
		case stTime:
			t := strings.Split(line, "-->")
			if len(t) != 2 {
				return nil, parseError(ln, line, "hh:mm:ss,ms --> hh:mm:ss,ms")
			}
			err := error(nil)
			r.In, err = ParseTimecode(strings.TrimSpace(t[0]))
			if err != nil {
				return nil, parseError(ln, t[0], "hh:mm:ss,ms")
			}
			r.Out, err = ParseTimecode(strings.TrimSpace(t[1]))
			if err != nil {
				return nil, parseError(ln, t[1], "hh:mm:ss,ms")
			}
			state = stTextStart
		case stTextStart:
			if len(line) == 0 {
				return nil, parseError(ln, line, "text")
			}
			r.Text = line
			state = stText
		case stText:
			if len(line) != 0 {
				r.Text += "\n" + line
				continue
			}
			ret = append(ret, r)
			state = stReady
		}
	} // for scanner.Scan()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if state != stReady {
		return nil, fmt.Errorf("unexpected <EOF>")
	}
	return ret, nil
}
