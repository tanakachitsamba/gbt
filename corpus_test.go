package main

import "testing"

func TestGetCountOfCharsASCII(t *testing.T) {
	doc := "hello world"
	want := 11

	if got := getCountOfChars(doc); got != want {
		t.Errorf("getCountOfChars(%q) = %d, want %d", doc, got, want)
	}

	if got := countChars(doc); got != want {
		t.Errorf("countChars(%q) = %d, want %d", doc, got, want)
	}
}

func TestGetCountOfCharsMultibyte(t *testing.T) {
	doc := "こんにちは世界"
	want := 7

	if got := getCountOfChars(doc); got != want {
		t.Errorf("getCountOfChars(%q) = %d, want %d", doc, got, want)
	}

	if got := countChars(doc); got != want {
		t.Errorf("countChars(%q) = %d, want %d", doc, got, want)
	}
}
