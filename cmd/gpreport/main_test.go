package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const pathSep = string(filepath.Separator)

var dataIn = []string{
	"xxx/yyy/PO/posters/project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"yyy/PO/posters/project_b/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"/PO/posters/project_c/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"PO/posters/project_d/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"posters/project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"posters/for_service/для сервиса/600x600.jpg",
	"posters/season/3 сезон/600x600.jpg",
	"posters/season2/42 сезон/600x600.jpg",
	"posters/season3/42 сезон/для сервиса/600x600.jpg",
	"posters/season2/для сервиса/42 сезон/600x600.jpg",
	// "/posters/project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	// "project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
}

var dataOut = []string{
	"project_a",
	"project_b",
	"project_c",
	"project_d",
	"project_a",
	"for_service",
	"season",
	"season2",
	"season3",
	"для сервиса",
}

func TestJob(t *testing.T) {
	for i := range dataIn {
		dataIn[i] = strings.Replace(dataIn[i], "/", pathSep, -1)
	}
	res, err := doJob(dataIn)
	if err != nil {
		t.Error(err)
		return
	}

	if len(res) != len(dataOut) {
		t.Errorf("lengths aren't equal (%v != %v)", len(res), len(dataOut))
		return
	}
	for i, line := range res {
		name := strings.Split(line, "\t")[0]
		if name != dataOut[i] {
			t.Errorf("data line %v: %q != %q)", i, name, dataOut[i])
		}
	}
}
