package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

var input, output []string

type iDiffUtils interface {
	scanStdinToChannel(to chan string)
}
type diffUtils struct{}

func scanToChannel(from io.Reader, to chan string) {
	scanner := bufio.NewScanner(from)
	for scanner.Scan() {
		to <- scanner.Text()
	}
	close(to)
}

func (d diffUtils) scanStdinToChannel(i chan string) {
	scanToChannel(os.Stdin, i)
}

func readCmd(cmdString string, o chan string) {
	cmd := exec.Command("bash", "-c", cmdString)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanToChannel(stdout, o)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func diffLine(v string, stdout chan string, wg *sync.WaitGroup) {
	found := false
	for _, w := range output {
		if v == w {
			found = true
		}
	}
	if !found {
		stdout <- v
	}
	wg.Done()
}

func printLn(stdout chan string, done chan struct{}) {
	for s := range stdout {
		fmt.Println(s)
	}
	close(done)
}

func diff(cmd string, timeout time.Duration, i chan string, stdout chan string, done chan struct{}, utils iDiffUtils) {
	o := make(chan string)

	go utils.scanStdinToChannel(i)
	go readCmd(cmd, o)

	inTimer := time.NewTimer(timeout)
in:
	for {
		select {
		case s, ok := <-i:
			if !ok {
				break in
			}
			input = append(input, s)
			inTimer.Reset(timeout)
		case <-inTimer.C:
			break in
		}
	}

	outTimer := time.NewTimer(timeout)
out:
	for {
		select {
		case s, ok := <-o:
			if !ok {
				break out
			}
			output = append(output, s)
			outTimer.Reset(timeout)
		case <-outTimer.C:
			break out
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(input))

	for _, v := range input {
		go diffLine(v, stdout, &wg)
	}

	wg.Wait()
	close(stdout)
	<-done
}

func main() {
	cmd := os.Args[1]

	i := make(chan string)
	stdout := make(chan string)
	done := make(chan struct{})
	timeout := 15 * time.Second

	go printLn(stdout, done)
	diff(cmd, timeout, i, stdout, done, diffUtils{})
}
