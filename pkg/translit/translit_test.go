package translit

import "testing"

var (
	tableTranslit = []struct {
		inputVal, check string
	}{
		//23456789012345678901234567890
		{inputVal: "абвгдеёжзийклмнопрстуфхцчшщъыьэюя АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ",
			check: "abvgdeezhziyklmnoprstufhcchshshyeyuya_abvgdeezhziyklmnoprstufhcchshshyeyuya"},
		{inputVal: "  qwerty123  абвй  ",
			check: "qwerty123_abvy"},
	}
)

// TestOldFormParseCorrect -
func TestOldFormParseCorrect(t *testing.T) {
	for _, v := range tableTranslit {
		res, err := Do(v.inputVal)
		if err != nil {
			t.Errorf("\n%q\nDo() error: %v", v.inputVal, err)
			continue
		}

		if res != v.check {
			t.Errorf("\nnot equivalent \nin : %q\nres: %q\nchk: %q", v.inputVal, res, v.check)
			continue
		}
	}
}
