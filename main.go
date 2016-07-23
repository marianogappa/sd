package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type iDiffUtils interface {
	scanStdinToChannel(to chan string, cancel chan struct{})
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

func (d diffUtils) scanStdinToChannel(i chan string, cancel chan struct{}) {
	scanToChannel(os.Stdin, i, cancel)
}

func readCmd(cmdString string, o chan string, cancel chan struct{}) {
	var stderr bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", cmdString)
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	go scanToChannel(stdout, o, cancel)
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

func processStdin(stdinCh chan string, diffee *[]string, stdinTimeout timeout, cancelStdin chan struct{}, start chan struct{}, stdout chan string, wg *sync.WaitGroup) {
	stdinTimeout.Start()
	var innerWg sync.WaitGroup

loop:
	for {
		select {
		case s, ok := <-stdinCh:
			if !ok {
				break loop
			}
			innerWg.Add(1)
			go diffLine(s, stdout, diffee, start, &innerWg)
			stdinTimeout.Reset()
		case <-*stdinTimeout.c:
			close(cancelStdin)
			break loop
		}
	}

	innerWg.Wait()
	wg.Done()
}

func processCmd(cmdCh chan string, diffee *[]string, cmdTimeout timeout, cancelCmd chan struct{}, start chan struct{}, wg *sync.WaitGroup) {
	cmdTimeout.Start()
	for {
		select {
		case s, ok := <-cmdCh:
			if !ok {
				close(start)
				wg.Done()
				return
			}
			*diffee = append(*diffee, s)
			cmdTimeout.Reset()
		case <-*cmdTimeout.c:
			close(cancelCmd)
		}
	}
}

func diff(cmd string, stdinTimeout timeout, cmdTimeout timeout, stdout chan string, utils iDiffUtils) {
	var diffee []string

	stdinCh := make(chan string)
	cmdCh := make(chan string)
	start := make(chan struct{})
	cancelCmd := make(chan struct{})
	cancelStdin := make(chan struct{})

	go utils.scanStdinToChannel(stdinCh, cancelStdin)
	go readCmd(cmd, cmdCh, cancelCmd)

	var wg sync.WaitGroup
	wg.Add(2)

	go processStdin(stdinCh, &diffee, stdinTimeout, cancelStdin, start, stdout, &wg)
	go processCmd(cmdCh, &diffee, cmdTimeout, cancelCmd, start, &wg)

	wg.Wait()
	close(stdout)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	stdinTimeout, cmdTimeout := mustResolveTimeouts(mustResolveOptions())
	cmd := os.Args[len(os.Args)-1]

	stdout := make(chan string)
	done := make(chan struct{})

	go printLn(stdout, done)
	diff(cmd, stdinTimeout, cmdTimeout, stdout, diffUtils{})
	<-done
}
