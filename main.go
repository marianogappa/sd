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

type iDiffUtils interface {
	scanStdinToChannel(to chan string)
}
type diffUtils struct{}

func scanToChannel(from io.Reader, to chan string, cancel chan struct{}) {
	scanner := bufio.NewScanner(from)
	intermediate := make(chan string)

	go func() {
		for scanner.Scan() {
			intermediate <- scanner.Text()
		}
		close(intermediate)
	}()

loop:
	for {
		select {
		case s, ok := <-intermediate:
			if !ok {
				break loop
			}
			to <- s
		case <-cancel:
			break loop
		}
	}
	close(to)
}

func (d diffUtils) scanStdinToChannel(i chan string) {
	cancel := make(chan struct{})
	scanToChannel(os.Stdin, i, cancel)
}

func readCmd(cmdString string, o chan string, cancel chan struct{}) {
	cmd := exec.Command("bash", "-c", cmdString)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	go scanToChannel(stdout, o, cancel)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func diffLine(v string, stdout chan string, diffee *[]string, start chan struct{}, wg *sync.WaitGroup) {
loop:
	for {
		select {
		case <-start:
			break loop
		}
	} // wait until diffee finishes loading

	found := false
	for _, w := range *diffee {
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

func diff(cmd string, timeout time.Duration, i chan string, stdout chan string, utils iDiffUtils) {
	var diffee []string

	o := make(chan string)
	start := make(chan struct{})
	cancel := make(chan struct{})

	stdinFinished := false
	cmdFinished := false

	go utils.scanStdinToChannel(i)
	go readCmd(cmd, o, cancel)

	var wg sync.WaitGroup

	outTimer := time.NewTimer(timeout)
	inTimer := time.NewTimer(timeout)
loop:
	for {
		select {
		case s, ok := <-i:
			if !stdinFinished {
				if !ok {
					stdinFinished = true
					if cmdFinished {
						break loop
					}
					continue
				}
				wg.Add(1)
				go diffLine(s, stdout, &diffee, start, &wg)
				inTimer.Reset(timeout)
			}
		case s, ok := <-o:
			if !cmdFinished {
				if !ok {
					cmdFinished = true
					close(start)
					if stdinFinished {
						break loop
					}
					continue
				}
				diffee = append(diffee, s)
				outTimer.Reset(timeout)
			}
		case <-inTimer.C:
			if cmdFinished {
				break loop
			}
		case <-outTimer.C:
			close(cancel)
		}
	}
	wg.Wait()
	close(stdout)
}

func main() {
	cmd := os.Args[1]

	i := make(chan string)
	stdout := make(chan string)
	done := make(chan struct{})
	timeout := 2 * time.Second

	go printLn(stdout, done)
	diff(cmd, timeout, i, stdout, diffUtils{})
	<-done
}
