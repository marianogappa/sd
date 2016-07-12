package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
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

func TestReadCmd(t *testing.T) {
	cmdString := `echo -e "one\ntwo"`
	o := make(chan string, 2)

	readCmd(cmdString, o)

	line1 := <-o
	line2 := <-o

	if line1 != "one" {
		t.Error("first line wasn't 'one'")
	}
	if line2 != "two" {
		t.Error("second line wasn't 'two'")
	}
}

type mockUtils struct{}

func (m mockUtils) scanStdinToChannel(i chan string) {}

func TestDiff(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)
	done := make(chan struct{})

	go diff(`echo -e "one\ntwo"`, 100*time.Millisecond, i, stdout, done, mockUtils{})
	i <- "one"
	i <- "two"
	i <- "three"
	i <- "four"

	line1 := <-stdout
	line2 := <-stdout

	if line1 != "four" {
		t.Errorf("first line wasn't 'four', it was %v", line1)
	}
	if line2 != "three" {
		t.Errorf("second line wasn't 'three', it was %v", line2)
	}

	close(done)
}

func TestDiffWhenInputTimesOut(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)
	done := make(chan struct{})

	go diff(`echo -e "one\ntwo"`, 100*time.Millisecond, i, stdout, done, mockUtils{})

	go func() {
		i <- "one"
		i <- "three"
		i <- "three"
		i <- "three"
		i <- "one"
		i <- "two"
		i <- "four"
		time.Sleep(101 * time.Millisecond)
		i <- "five"
	}()

	lines := []string{}
loop:
	for {
		select {
		case line, ok := <-stdout:
			if !ok {
				break loop
			}
			lines = append(lines, line)
		case <-time.After(1 * time.Second):
			break loop
		}
	}

	sort.Strings(lines) // order is not deterministic
	if reflect.DeepEqual(lines, []string{"four", "three", "three", "three"}) != true {
		t.Errorf("result wasn't ['four', 'three', 'three', 'three'], it was %v", lines)
	}

	close(done)
}
