package edl

import (
	"fmt"
	"strings"
	"testing"

	// "unsafe"
)

var data =
`TITLE: f_files_s06
FCM: NON-DROP FRAME

001  BL       V     C        00:00:00:00 00:02:23:00 00:00:00:00 00:02:23:00

002  AX       V     C        00:02:23:00 00:45:27:00 00:02:23:00 00:45:27:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e01_HD__Proxy__.mp4

003  AX       A     C        00:02:23:00 00:45:27:00 00:02:23:00 00:45:27:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e01_AUDIOENG51.m4a

004  AX       A2    C        00:02:23:00 00:45:27:00 00:02:23:00 00:45:27:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e01_AUDIORUS20.m4a

005  BL       V     C        00:00:00:00 00:16:56:00 00:45:27:00 01:02:23:00

006  AX       V     C        00:02:23:00 00:44:06:00 01:02:23:00 01:44:06:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e02_HD__Proxy__.mp4

007  AX       A     C        00:02:23:00 00:44:06:00 01:02:23:00 01:44:06:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e02_AUDIOENG51.m4a

008  AX       A2    C        00:02:23:00 00:44:06:00 01:02:23:00 01:44:06:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e02_AUDIORUS20.m4a

009  BL       V     C        00:00:00:00 00:18:17:00 01:44:06:00 02:02:23:00

010  AX       V     C        00:02:23:01 00:44:56:00 02:02:23:00 02:44:56:00
* FROM CLIP NAME: You_Me_and_the_Apocalypse_s01e03_HD__Proxy__.mp4
`
func formatStruct(v interface{}) string {
	const pad = "####"
	s := fmt.Sprintf("%+v", v)
	s = pad + s[1:len(s)-1]
	// s = strings.Replace(s, ":{", "$$$$"+pad, -1)

	s = strings.Replace(s, "} ", "\n"+pad, -1)
	s = strings.Replace(s, "}", "", -1)
	s = strings.Replace(s, "{", "\n"+pad+pad, -1)
	s = strings.Replace(s, " ", "\n"+pad+pad, -1)
	s = strings.Replace(s, "#", " ", -1)
	x := strings.Split(s, "\n")
	for i := range x {
		x[i] = strings.Replace(x[i], ":", ": ", 1)
	}
	s = strings.Join(x, "\n")
	return s
}

func TestEdlCorrect(t *testing.T) {
	edl, tree, err := Parse(data)
	if err != nil {
		t.Errorf("Parse error: %v\n", err)
		return
	}
	list := edl.Split()
	if len(list) != 3 {
		t.Errorf("Split error: len %v != 3\n", len(list))
		return
	}
	for i, edl := range list {
		t.Errorf("== %v == ---------------------------------------------\n", i)
		t.Errorf("%v\n", formatStruct(edl))
	}

	_ = tree

	// t.Errorf("tree:\n%s\n", tree)
	// s := formatStruct(edl)
	// t.Errorf(s)

	// t.Errorf("size of timecode %v", unsafe.Sizeof(Timecode{}))
}

