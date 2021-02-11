package main

import (
    "testing"
)

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
	_, err := doJob(data)
	if err != nil {
		t.Error(err)
	}
}
