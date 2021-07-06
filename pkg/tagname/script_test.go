package tagname

import (
	"testing"
)

// TestCorrect -
func TestCorrect(t *testing.T) {
	src := `
fmt := import("fmt")
imed := import("imed")

process := func(path) {
	// fmt.println("!!!", filename)
	if imed.fileext("test.xxx") != ".xxx" {
		return error("must be '.xxx'")
	}

	tn := imed.tagname("sd_9999_xxx__film", false)
	err := tn.err()
	if is_error(err) {
		return err
	}
	num := tn.len()
	tn.add_tag("mtag", "m"+num)
	return [tn, tn] 
}

filename = process(filename)
`
	s, err := NewScript(src)
	if err != nil {
		t.Errorf("NewScript() error: %v", err)
		return
	}
	ret, err := s.Run("test")
	if err != nil {
		t.Errorf("Run() error: %v", err)
		return
	}
	if len(ret) != 2 {
		t.Errorf("Run() error: len(ret) != 2 (%v)", len(ret))
		return
	}
	if ret[0] == nil {
		t.Error("Run() error: ret[0] == nil")
		return
	}

	if v, _ := ret[0].ConvertTo("old"); v != "xxx_9999__sd_ar2_m6" {
		t.Errorf("error: invalid return value %v", v)
		return
	}
}
