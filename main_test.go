package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestScanToChannel(t *testing.T) {
	stdin := `1
2`

	from := strings.NewReader(stdin)
	to := make(chan string, 2)
	cancel := make(chan struct{})

	scanToChannel(from, to, cancel)

	lines := readAndSortBlocking(to, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"1", "2"}) != true {
		t.Errorf("result wasn't ['1', '2'], it was %v", lines)
	}
}

func TestReadCmd(t *testing.T) {
	cmdString := `echo -e "1\n2"`
	o := make(chan string, 2)
	cancel := make(chan struct{})

	readCmd(cmdString, o, cancel)

	lines := readAndSortBlocking(o, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"1", "2"}) != true {
		t.Errorf("result wasn't ['1', '2'], it was %v", lines)
	}
}

type mockUtils struct{}

func (m mockUtils) scanStdinToChannel(i chan string) {}

func TestDiff(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo -e "1\n2"`, 100*time.Millisecond, i, stdout, mockUtils{})
	i <- "1"
	i <- "2"
	i <- "3"
	i <- "4"

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"3", "4"}) != true {
		t.Errorf("lines weren't ['3', '4'], it was %v", lines)
	}
}

func TestDiffWhenInputTimesOut(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo -e "1\n2"`, 100*time.Millisecond, i, stdout, mockUtils{})

	go func() {
		i <- "1"
		i <- "3"
		i <- "3"
		i <- "3"
		i <- "1"
		i <- "2"
		i <- "4"
		time.Sleep(101 * time.Millisecond)
		i <- "5"
	}()

	lines := readAndSortBlocking(stdout, 1*time.Second)

	sort.Strings(lines) // order is not deterministic
	if reflect.DeepEqual(lines, []string{"3", "3", "3", "4"}) != true {
		t.Errorf("result wasn't ['3', '3', '3', '4'], it was %v", lines)
	}
}

func TestDiffWhenOutputTimesOut(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo -e "1\n2" && sleep 1 && echo -e "3\n4"`, 100*time.Millisecond, i, stdout, mockUtils{})

	i <- "1"
	i <- "2"
	i <- "3"
	i <- "4"
	i <- "5"

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"3", "4", "5"}) != true {
		t.Errorf("result wasn't ['3', '4', '5'], it was %v", lines)
	}
}

func TestDiffWhenDelaysAddUpToTimeoutSeparatelyButDoesntTimeout(t *testing.T) {
	i := make(chan string)
	stdout := make(chan string)

	go diff(`echo "1" && sleep .1 && echo "2" && sleep .1 && echo "3" && sleep .1 && echo "4" && sleep .1 && echo "ten"`,
		200*time.Millisecond, i, stdout, mockUtils{})

	i <- "1"
	i <- "2"
	i <- "5"
	i <- "3"
	i <- "4"
	i <- "6"

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"5", "6"}) != true {
		t.Errorf("result wasn't ['5', '6'], it was %v", lines)
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
