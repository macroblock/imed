package main 

import (
	"path/filepath"
	"strings"
	"testing"
)

const pathSep = string(filepath.Separator)

var (
	tableCorrect = []struct {
		input string
		out string
		err error
	}{
		//23456789012345678901234567890
		{
			input: "xxx__logo_600x600.jpg",
			out: "g_haslogo_600x600.jpg",
			err: nil,
		},
		{
			input: "path/to/file/xxx__logo_1800x1000.jpg",
			out: "path/to/file/g_hastitle_logo_1800x1000.jpg",
			err: nil,
		},
		{
			input: "xxx__background_1000x1500.jpg",
			out: "g_iconic_background_1000x1500.jpg",
			err: nil,
		},
		{
			input: "xxx__poster_800x600.jpg",
			out: "g_iconic_poster_800x600.jpg",
			err: nil,
		},
	}
	// tableIncorrect = []string{
		//23456789012345678901234567890
		// "The_name_s01_zzz_2018__hd_q0w0",
		// "sd_2018_Sobibor__12_q0w2_trailer.mpg",
	// }
)

type ttag struct {
	typ, val string
}

// TestCorrect -
func TestCorrect(t *testing.T) {
	for i := range tableCorrect {
		v := &tableCorrect[i]
		v.out = strings.Replace(v.out, "/", pathSep, -1)
	}
	for _, v := range tableCorrect {
		out, err := doProcess(v.input, false)
		if err != v.err {
			t.Errorf("\n%q\nerror: %v", v.input, err)
			continue
		}
		if out != v.out {
			t.Errorf("\n%q\nerror: %v != %v", v.input, out, v.out)
			continue
		}
	}
}

// // TestIncorrect -
// func TestIncorrect(t *testing.T) {
	// for _, v := range tableIncorrect {
		// tn, err := tagname.NewFromFilename(v, false)
		// if err != nil {
			// t.Errorf("\n%q\nNewFromFilename() error:\n%v", v, err)
			// continue
		// }
		// sizeLimit, err := rtimg.CheckImage(tn, false)
		// _ = sizeLimit
		// if err == nil {
			// t.Errorf("\n%q\nhas no error", v)
			// continue
		// }
	// }
// }

