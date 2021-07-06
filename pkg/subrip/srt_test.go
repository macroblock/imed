package subrip

import (
	// "fmt"
	"strings"
	"testing"
	// "unsafe"
)

var data = `0
00:02:17,440 --> 00:02:20,375
Senator, we're making
our final approach into Coruscant.

2
00:02:20,476 --> 00:02:22,501
Very good, Lieutenant.

3
00:02:30,476 --> 00:02:32,501
one line
second one
and the last one

`

func TestSrtCorrect(t *testing.T) {
	srt, err := Parse(strings.NewReader(data))
	if err != nil {
		t.Errorf("Parse error: %v\n", err)
		return
	}
	// for i, v := range srt {
	// t.Errorf("%2v: %v\n", i, v)
	// }

	err = StrictCheck(srt)
	if len(strings.Split(err.Error(), "\n")) != 4 {
		t.Errorf("Check incorrect error: %v\n", err)
	}
	// if err != nil {
	// t.Errorf("%v", err)
	// }

	// t.Errorf("tree:\n%s\n", tree)
	// s := formatStruct(edl)
	// t.Errorf(s)

	// t.Errorf("size of timecode %v", unsafe.Sizeof(Timecode{}))
}
