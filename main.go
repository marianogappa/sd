package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type iDiffUtils interface {
	scanStdinToChannel(to chan string, cancel chan struct{})
}
type diffUtils struct{}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:

sd [options] 'command'

Examples

  echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'

  while [ 0 ]; do echo $RANDOM; sleep .1; done | sd -h 1 'seq 500'

  mysql schema_1 -Nsr -e "SELECT city FROM users" | sd -h 120 mysql schema_2 -Nsr -e "SELECT city FROM excluded_cities"

  mysql -Nsr -e "SELECT city FROM users" | sd -p 0 -t 10 kafka_consumer --topic excluded_cities > active_cities.txt 

Options

	-f --follow: keeps reading from STDIN until SIGINT (think tail -f).
	-p --patience %seconds%: wait for the specified seconds for the first received line. Use 0 for waiting forever.
	-t --timeout %seconds%: exit(0) after specified seconds from last received line. STDIN and command have independent timeouts. When with -f, timeout only applies to the command (not to STDIN).
	-h --hard-timeout %seconds%: exit(0) after the specified seconds (or earlier). Overrides all other options

`)
	os.Exit(2)
}

func mustResolveOptions() (bool, int, int, int) {
	followHelp := "keeps reading from STDIN until SIGINT (think tail -f)."
	patienceHelp := "wait for the specified seconds for the first received line. Use 0 for waiting forever"
	timeoutHelp := "exit(0) after specified seconds from last received line. STDIN and command have independent timeouts. When with -f, timeout only applies to the command (not to STDIN)."
	hardTimeoutHelp := "exit(0) after the specified seconds (or earlier). Overrides all other options"

	var follow bool
	var patience, timeout, hardTimeout int

	flag.BoolVar(&follow, "follow", false, followHelp)
	flag.BoolVar(&follow, "f", false, followHelp)
	flag.IntVar(&patience, "patience", -1, patienceHelp)
	flag.IntVar(&patience, "p", -1, patienceHelp)
	flag.IntVar(&timeout, "timeout", 10, timeoutHelp)
	flag.IntVar(&timeout, "t", 10, timeoutHelp)
	flag.IntVar(&hardTimeout, "hard-timeout", 0, hardTimeoutHelp)
	flag.IntVar(&hardTimeout, "h", 0, hardTimeoutHelp)

	flag.Usage = usage
	flag.Parse()

	return follow, patience, timeout, hardTimeout
}

func mustResolveTimeouts(follow bool, patience int, timeoutF int, hardTimeout int) (timeout, timeout) {
	var stdinTimeout, cmdTimeout timeout

	if follow {
		stdinTimeout.infinite = true
	}

	if patience == 0 {
		stdinTimeout.firstTimeInfinite = true
		cmdTimeout.firstTimeInfinite = true
	} else if patience == -1 {
		stdinTimeout.firstTime = time.Duration(timeoutF) * time.Second
		cmdTimeout.firstTime = time.Duration(timeoutF) * time.Second
	} else {
		stdinTimeout.firstTime = time.Duration(patience) * time.Second
		cmdTimeout.firstTime = time.Duration(patience) * time.Second
	}

	stdinTimeout.time = time.Duration(timeoutF) * time.Second
	cmdTimeout.time = time.Duration(timeoutF) * time.Second

	if hardTimeout > 0 {
		stdinTimeout.hard = true
		cmdTimeout.hard = true
		stdinTimeout.firstTime = time.Duration(hardTimeout) * time.Second
		cmdTimeout.firstTime = time.Duration(hardTimeout) * time.Second
	}

	return stdinTimeout, cmdTimeout
}

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

func diff(cmd string, stdinTimeout timeout, cmdTimeout timeout, i chan string, stdout chan string, utils iDiffUtils) {
	var diffee []string

	o := make(chan string)
	start := make(chan struct{})
	cancelCmd := make(chan struct{})
	cancelStdin := make(chan struct{})

	stdinFinished := false
	cmdFinished := false

	go utils.scanStdinToChannel(i, cancelStdin)
	go readCmd(cmd, o, cancelCmd)

	var wg sync.WaitGroup

	cmdTimeout.Start()
	stdinTimeout.Start()
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
				stdinTimeout.Reset()
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
				cmdTimeout.Reset()
			}
		case <-*stdinTimeout.c:
			close(cancelStdin)
		case <-*cmdTimeout.c:
			close(cancelCmd)
		}
	}
	wg.Wait()
	close(stdout)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	stdinTimeout, cmdTimeout := mustResolveTimeouts(mustResolveOptions())
	cmd := os.Args[len(os.Args)-1]

	i := make(chan string)
	stdout := make(chan string)
	done := make(chan struct{})

	go printLn(stdout, done)
	diff(cmd, stdinTimeout, cmdTimeout, i, stdout, diffUtils{})
	<-done
}
