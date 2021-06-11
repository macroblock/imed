package main

import (
    "path/filepath"
    "strings"
    "testing"
)

const pathSep = string(filepath.Separator)

var data = []string {
	"xxx/yyy/PO/posters/project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"yyy/PO/posters/project_b/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"/PO/posters/project_c/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"PO/posters/project_d/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	"posters/project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	// "/posters/project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
	// "project_a/google_apple_feed/jpg/g_iconic_poster_600x800.jpg",
}

func TestJob(t *testing.T) {
	for i := range data {
		data[i] = strings.Replace(data[i], "/", pathSep, -1)
	}
	_, err := doJob(data)
	if err != nil {
		t.Error(err)
	}
}
