package main

import (
	"strings"
	"testing"
)

func TestScanToChannel(t *testing.T) {
	lines := `one
two`

	from := strings.NewReader(lines)
	to := make(chan string, 2)

	scanToChannel(from, to)

	line1 := <-to
	line2 := <-to

	if line1 != "one" {
		t.Error("first line wasn't 'one'")
	}
	if line2 != "two" {
		t.Error("second line wasn't 'two'")
	}
}
