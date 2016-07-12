package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

var input, output []string

func readStdin(i chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		ok := scanner.Scan()
		if !ok {
			close(i)
			break
		}
		i <- scanner.Text()
	}
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

	scanner := bufio.NewScanner(stdout)
	for {
		ok := scanner.Scan()

		if !ok {
			close(o)
			break
		}
		o <- scanner.Text()
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	cmd := os.Args[1]

	i := make(chan string)
	o := make(chan string)
	stdout := make(chan string)
	done := make(chan struct{})
	timeout := 15 * time.Second

	go readStdin(i)
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
