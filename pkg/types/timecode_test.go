package types

import (
	// "fmt"
	// "strings"
	"testing"

	// "unsafe"
)

func hhmmssms(hh, mm, ss, ms int) Timecode {
	return Timecode((float64(hh)*60+float64(mm))*60+float64(ss) + float64(ms)/1000)
}

type correct struct {
	in string
	out Timecode
}

type incorrect struct {
	in string
	out string
}

var typesParseHHMMSSCorrect = []correct {
	correct{ "01:02:03.456", NewTimecode(1, 2, 3.456) },
	correct{ "+01:02:03.456", NewTimecode(1, 2, 3.456) },
	correct{ "1", NewTimecode(0, 0, 1.0) },
	correct{ "01.234", NewTimecode(0, 0, 1.234) },
	correct{ "01:02", NewTimecode(0, 1, 2.0) },
	correct{ "01:02:03", NewTimecode(1, 2, 3.0) },
	correct{ "+00:01:02.345", NewTimecode(0, 1, 2.345) },
	correct{ "-00:01:02.345", NewTimecode(0, -1, -2.345) },
	correct{ "-01:02:03.456", NewTimecode(-1, -2, -3.456) },
	correct{ "-1", NewTimecode(0, 0, -1.0) },
	correct{ "-01.234", NewTimecode(0, 0, -1.234) },
	correct{ "-01:02", NewTimecode(0, -1, -2.0) },
	correct{ "-01:02:03", NewTimecode(-1, -2, -3.0) },
	correct{ "08:09:010", NewTimecode(8, 9, 10.0) },
	correct{ "00:33:20", NewTimecode(0, 0, 2000) },
	correct{ "00:15:00", NewTimecode(0, 0, 900) },
}

var typesParseHHMMSSIncorrect = []incorrect {
	incorrect{ "01:err:03.456", "01:err:03.456 parse error: strconv.Atoi: parsing \"err\": invalid syntax" },
	incorrect{ "--01:02:03.456", "--01:02:03.456 parse error: strconv.Atoi: parsing \"--01\": invalid syntax" },
	incorrect{ "++01:02:03.456", "++01:02:03.456 parse error: strconv.Atoi: parsing \"++01\": invalid syntax" },
	incorrect{ "01:-02:03.456", "01:-02:03.456 parse error: only first subvalue can have a sign" },
	incorrect{ "01:+02:03.456", "01:+02:03.456 parse error: only first subvalue can have a sign" },
	incorrect{ "00:+00:01.234", "00:+00:01.234 parse error: only first subvalue can have a sign" },
	incorrect{ "01:02:03:4.567", "01:02:03:4.567 parse error: too many colons" },
}

var typesParseHHMMSSFrCorrect = []correct {
	correct{ "01:02:03:04", NewTimecode(1, 2, 3.160) },
	correct{ "+01:02:03:04", NewTimecode(1, 2, 3.160) },
	correct{ "1", NewTimecode(0, 0, 1.0) },
	correct{ "01:02", NewTimecode(0, 1, 2.0) },
	correct{ "01:02:03", NewTimecode(1, 2, 3.0) },
	correct{ "-00:01:02:10", NewTimecode(0, -1, -2.400) },
	correct{ "-01:02:03:10", NewTimecode(-1, -2, -3.400) },
	correct{ "-1", NewTimecode(0, 0, -1.0) },
	correct{ "-01:02", NewTimecode(0, -1, -2.0) },
	correct{ "-01:02:03", NewTimecode(-1, -2, -3.0) },
	correct{ "08:09:010", NewTimecode(8, 9, 10.0) },
}

var typesParseHHMMSSFrIncorrect = []incorrect {
	incorrect{ "01:err:03:04", "01:err:03:04 parse error: strconv.Atoi: parsing \"err\": invalid syntax" },
	incorrect{ "--01:02:03:04", "--01:02:03:04 parse error: strconv.Atoi: parsing \"--01\": invalid syntax" },
	incorrect{ "++01:02:03:04", "++01:02:03:04 parse error: strconv.Atoi: parsing \"++01\": invalid syntax" },
	incorrect{ "01:-02:03:04", "01:-02:03:04 parse error: only first subvalue can have a sign" },
	incorrect{ "01:+02:03:04", "01:+02:03:04 parse error: only first subvalue can have a sign" },
	incorrect{ "00:+00:01:04", "00:+00:01:04 parse error: only first subvalue can have a sign" },
	incorrect{ "01:02:03:04:05", "01:02:03:04:05 parse error: too many colons" },
}

var typesParseTimecodeCorrect = []correct {
	correct{ "01:02:03.456", NewTimecode(1, 2, 3.456) },
	correct{ "01:02:03:04", NewTimecode(1, 2, 3.160) },
}

func testCorrect(t *testing.T, title string, fn func(string)(Timecode, error), data []correct) {
	l := len(data)
	for i, v := range data {
		out, err := fn(v.in)
		if err != nil {
			t.Errorf("%v[#%v/%v] error: %v\n", title, i, l, err)
			continue
		}
		if v.out != out {
			t.Errorf("%v[#%v/%v] expected %v, got %v\n", title, i, l, v.out, out)
		}
		// if v.in != out.String() {
			// t.Errorf("%v[#%v/%v] tostring expected %v, got %v\n", title, i, l, v.in, out.String())
		// }
	}
}

func testIncorrect(t *testing.T, title string, fn func(string)(Timecode, error), data []incorrect) {
	l := len(data)
	for i, v := range data {
		_, err := fn(v.in)
		if err == nil {
			t.Errorf("%v[#%v/%v] do not have an error\n", title, i, l)
			continue
		}
		if v.out != err.Error() {
			t.Errorf("%v[#%v/%v] expected error %q, got %q\n", title, i, l, v.out, err)
		}
	}
}

func TestMsCorrect(t *testing.T) {
	testCorrect(t, "ParseHHMMSS", ParseHHMMSS, typesParseHHMMSSCorrect )
}

func TestMsIncorrect(t *testing.T) {
	testIncorrect(t, "ParseHHMMSS", ParseHHMMSS, typesParseHHMMSSIncorrect)
}

func TestFrCorrect(t *testing.T) {
	testCorrect(t, "ParseHHMMSSFr", ParseHHMMSSFr, typesParseHHMMSSFrCorrect)
}

func TestFrIncorrect(t *testing.T) {
	testIncorrect(t, "ParseHHMMSSFr", ParseHHMMSSFr, typesParseHHMMSSFrIncorrect)
}

func TestTypeCorrect(t *testing.T) {
	testCorrect(t, "Parse", ParseTimecode, typesParseTimecodeCorrect)
}
