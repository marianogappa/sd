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

func scanToChannel(from io.Reader, to chan string) {
	scanner := bufio.NewScanner(from)
	for scanner.Scan() {
		to <- scanner.Text()
	}
	close(to)
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

func diff(cmd string, timeout time.Duration, i chan string, o chan string, stdout chan string) {
	done := make(chan struct{})

	go scanToChannel(os.Stdin, i)
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

	go func(done chan struct{}) {
		for s := range stdout {
			fmt.Println(s)
		}
		close(done)
	}(done)

	var wg sync.WaitGroup
	wg.Add(len(input))

	for _, v := range input {
		go func(v string, stdout chan string) {
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
		}(v, stdout)
	}

	wg.Wait()
	close(stdout)
	<-done
}

func main() {
	cmd := os.Args[1]

	i := make(chan string)
	o := make(chan string)
	stdout := make(chan string)
	timeout := 15 * time.Second

	diff(cmd, timeout, i, o, stdout)
}
