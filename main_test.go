package main

import (
	"io"
	"log"
	"os/exec"
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

type mockUtils struct {
	i io.Reader
}

func (m mockUtils) scanStdinToChannel(i chan string, cancel chan struct{}) {
	scanToChannel(m.i, i, cancel)
}

func TestDiff(t *testing.T) {
	stdout := make(chan string)
	reader := cmdToReader(`echo -e "1\n2\n3\n4"`)

	go diff(`echo -e "1\n2"`, defaultTimeout(), defaultTimeout(), stdout, mockUtils{reader})

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"3", "4"}) != true {
		t.Errorf("result wasn't ['3', '4'], it was %v", lines)
	}
}

func TestDiffWhenInputTimesOut(t *testing.T) {
	stdout := make(chan string)

	reader := cmdToReader(`echo -e "1\n3\n3\n3\n1\n2\n4" && sleep .101 && echo "5"`)

	go diff(`echo -e "1\n2"`, defaultTimeout(), defaultTimeout(), stdout, mockUtils{reader})

	lines := readAndSortBlocking(stdout, 1*time.Second)

	sort.Strings(lines) // order is not deterministic
	if reflect.DeepEqual(lines, []string{"3", "3", "3", "4"}) != true {
		t.Errorf("result wasn't ['3', '3', '3', '4'], it was %v", lines)
	}
}

func TestDiffWhenOutputTimesOut(t *testing.T) {
	stdout := make(chan string)
	reader := cmdToReader(`echo -e "1\n2\n3\n4\n5"`)

	go diff(`echo -e "1\n2" && sleep 1 && echo -e "3\n4"`, defaultTimeout(), defaultTimeout(), stdout, mockUtils{reader})

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"3", "4", "5"}) != true {
		t.Errorf("result wasn't ['3', '4', '5'], it was %v", lines)
	}
}

func TestDiffWhenDelaysAddUpToTimeoutSeparatelyButDoesntTimeout(t *testing.T) {
	stdout := make(chan string)
	reader := cmdToReader(`echo -e "1\n2\n3\n4\n5\n6"`)

	go diff(`echo "1" && sleep .1 && echo "2" && sleep .1 && echo "3" && sleep .1 && echo "4" && sleep .1 && echo "ten"`,
		defaultTimeout(), timeout{firstTime: 200 * time.Millisecond, time: 200 * time.Millisecond}, stdout, mockUtils{reader})

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

func TestEmptyCommand(t *testing.T) {
	stdout := make(chan string)

	reader := cmdToReader(`echo -e "1\n2\n3"`)
	go diff(`echo ""`, defaultTimeout(), defaultTimeout(), stdout, mockUtils{reader})

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{"1", "2", "3"}) != true {
		t.Errorf("result wasn't ['1', '2', '3'], it was %v", lines)
	}
}

func TestEmptyStdin(t *testing.T) {
	stdout := make(chan string)

	go diff(`echo "1\n2\n3"`, defaultTimeout(), defaultTimeout(), stdout, mockUtils{strings.NewReader(``)})

	lines := readAndSortBlocking(stdout, 1*time.Second)

	if reflect.DeepEqual(lines, []string{}) != true {
		t.Errorf("result wasn't [], it was %v", lines)
	}
}

func defaultTimeout() timeout {
	return timeout{firstTime: 100 * time.Millisecond, time: 100 * time.Millisecond}
}

func cmdToReader(cmdString string) io.Reader {
	cmd := exec.Command("bash", "-c", cmdString)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return stdout
}
