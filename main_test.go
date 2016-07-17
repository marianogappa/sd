package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestScanToChannel(t *testing.T) {
	stdin := `one
two`

	from := strings.NewReader(stdin)
	to := make(chan string, 2)
	cancel := make(chan struct{})

	scanToChannel(from, to, cancel)

	lines := readAndSortBlocking(to, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"one", "two"}) != true {
		t.Errorf("result wasn't ['one', 'two'], it was %v", lines)
	}
}

func TestReadCmd(t *testing.T) {
	cmdString := `echo -e "one\ntwo"`
	o := make(chan string, 2)
	cancel := make(chan struct{})

	readCmd(cmdString, o, cancel)

	lines := readAndSortBlocking(o, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"one", "two"}) != true {
		t.Errorf("result wasn't ['one', 'two'], it was %v", lines)
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

	lines := readAndSortBlocking(stdout, 1*time.Second)

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

	lines := readAndSortBlocking(stdout, 1*time.Second)

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

	lines := readAndSortBlocking(stdout, 1*time.Second)

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

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"five", "six"}) != true {
		t.Errorf("result wasn't ['five', 'six'], it was %v", lines)
	}
}

func readAndSortBlocking(c chan string, timeout time.Duration) []string {
	lines := []string{}
loop:
	for {
		select {
		case line, ok := <-c:
			if !ok {
				break loop
			}
			lines = append(lines, line)
		case <-time.After(timeout):
			break loop
		}
	}
	sort.Strings(lines) // order is not deterministic

	return lines
}
