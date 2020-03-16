package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/macroblock/imed/pkg/loudnorm"
)

func testLoudness(I float64, path string) {
	t := time.Now()
	opts, err := loudnorm.Scan(path, 0)
	dt := time.Since(t)

	if err != nil {
		fmt.Printf("### error: %v\n", err)
	}
	res := "#FAILED"
	if opts.InputI <= I+0.1 && opts.InputI >= I-0.1 {
		res = " PASSED"
	}
	fmt.Printf("%v: (%2.3f, LRA: %v, Thresh: %v, TP: %v, Offs: %v) %v %q\n",
		res, opts.InputI, opts.InputLRA, opts.InputThresh, opts.InputTP, opts.TargetOffset, dt, filepath.Base(path))
}

func testLoudness2(I float64, path string) {
	t := time.Now()
	opts, err := loudnorm.ScanLight(path, 0)
	dt := time.Since(t)

	if err != nil {
		fmt.Printf("### error: %v\n", err)
	}
	res := "#FAILED"
	if opts.InputI <= I+0.1 && opts.InputI >= I-0.1 {
		res = " PASSED"
	}
	// fmt.Printf("%v: (%2.3f) %v2.3 %q\n", res, val, dt, filepath.Base(path))
	fmt.Printf("%v: (%2.3f, LRA: %v, Thresh: %v %v, TP: %v) %v %q\n",
		res, opts.InputI, opts.InputLRA, opts.InputThresh, opts.InputThresh2, opts.InputTP, dt, filepath.Base(path))
}

func main() {
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-1-16bit.wav")
	testLoudness(-33.0, "../../../test/#test_sound/seq-3341-2-16bit.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-3-16bit-v02.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-4-16bit-v02.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-5-16bit-v02.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-6-5channels-16bit.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-6-6channels-WAVEEX-16bit.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-7_seq-3342-5-24bit.wav")
	testLoudness(-23.0, "../../../test/#test_sound/seq-3341-2011-8_seq-3342-6-24bit-v02.wav")

	fmt.Println()

	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-1-16bit.wav")
	testLoudness2(-33.0, "../../../test/#test_sound/seq-3341-2-16bit.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-3-16bit-v02.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-4-16bit-v02.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-5-16bit-v02.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-6-5channels-16bit.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-6-6channels-WAVEEX-16bit.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-7_seq-3342-5-24bit.wav")
	testLoudness2(-23.0, "../../../test/#test_sound/seq-3341-2011-8_seq-3342-6-24bit-v02.wav")
}
