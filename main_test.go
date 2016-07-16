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
	cancel := make(chan struct{})

	scanToChannel(from, to, cancel)

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
	cancel := make(chan struct{})

	readCmd(cmdString, o, cancel)

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

	go diff(`echo -e "one\ntwo"`, 100*time.Millisecond, i, stdout, mockUtils{})
	i <- "one"
	i <- "two"
	i <- "three"
	i <- "four"

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

	if reflect.DeepEqual(lines, []string{"four", "three"}) != true {
		t.Errorf("lines weren't ['four', 'three'], it was %v", lines)
	}
}

func TestDiffWhenInputTimesOut(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo -e "one\ntwo"`, 100*time.Millisecond, i, stdout, mockUtils{})

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
}

func TestDiffWhenOutputTimesOut(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo -e "one\ntwo" && sleep 1 && echo -e "three\nfour"`, 100*time.Millisecond, i, stdout, mockUtils{})

	i <- "one"
	i <- "two"
	i <- "three"
	i <- "four"
	i <- "five"

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
	if reflect.DeepEqual(lines, []string{"five", "four", "three"}) != true {
		t.Errorf("result wasn't ['five', 'four', 'three'], it was %v", lines)
	}
}

func TestDiffWhenDelaysAddUpToTimeoutSeparatelyButDoesntTimeout(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo "one" && sleep .1 && echo "two" && sleep .1 && echo "three" && sleep .1 && echo "four" && sleep .1 && echo "ten"`,
		200*time.Millisecond, i, stdout, mockUtils{})

	i <- "one"
	i <- "two"
	i <- "five"
	i <- "three"
	i <- "four"
	i <- "six"

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

	if reflect.DeepEqual(lines, []string{"five", "six"}) != true {
		t.Errorf("result wasn't ['five', 'six'], it was %v", lines)
	}
}
