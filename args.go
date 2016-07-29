package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:

sd [options] 'command'

Examples

  echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'

  while [ 0 ]; do echo $RANDOM; sleep .1; done | sd -h 1 'seq 500'

  mysql schema_1 -Nsr -e "SELECT city FROM users" | sd -h 120 mysql schema_2 -Nsr -e "SELECT city FROM excluded_cities"

  mysql -Nsr -e "SELECT city FROM users" | sd -p 0 -t 10 kafka_consumer --topic excluded_cities > active_cities.txt 

Options

	-f --follow: keeps reading from STDIN until SIGINT or its end.
	-i --infinite: keeps reading from COMMAND until it ends rather than timing it out. Note that if the stream doesn't end, sd just blocks forever and does nothing.
	-p --patience %seconds%: wait for the specified seconds for the first received line. Use 0 for waiting forever.
	-t --timeout %seconds%: exit(0) after specified seconds from last received line. STDIN and command have independent timeouts. When with -f, timeout only applies to the command (not to STDIN).
	-h --hard-timeout %seconds%: exit(0) after the specified seconds (or earlier). Overrides all other options.

`)
	os.Exit(2)
}

func mustResolveOptions(args []string) (bool, bool, int, int, int) {
	followHelp := "keeps reading from STDIN until SIGINT or its end."
	infiniteHelp := "keeps reading from COMMAND until it ends rather than timing it out. Note that if the stream doesn't end, sd just blocks forever and does nothing."
	patienceHelp := "wait for the specified seconds for the first received line. Use 0 for waiting forever."
	timeoutHelp := "exit(0) after specified seconds from last received line. STDIN and command have independent timeouts. When with -f, timeout only applies to the command (not to STDIN)."
	hardTimeoutHelp := "exit(0) after the specified seconds (or earlier). Overrides all other options."

	var follow, infinite bool
	var patience, timeout, hardTimeout int

	flag.BoolVar(&follow, "follow", false, followHelp)
	flag.BoolVar(&follow, "f", false, followHelp)
	flag.BoolVar(&infinite, "infinite", false, infiniteHelp)
	flag.BoolVar(&infinite, "i", false, infiniteHelp)
	flag.IntVar(&patience, "patience", -1, patienceHelp)
	flag.IntVar(&patience, "p", -1, patienceHelp)
	flag.IntVar(&timeout, "timeout", 10, timeoutHelp)
	flag.IntVar(&timeout, "t", 10, timeoutHelp)
	flag.IntVar(&hardTimeout, "hard-timeout", 0, hardTimeoutHelp)
	flag.IntVar(&hardTimeout, "h", 0, hardTimeoutHelp)

	flag.Usage = usage
	flag.CommandLine.Parse(args)

	return follow, infinite, patience, timeout, hardTimeout
}

func mustResolveTimeouts(follow bool, infinite bool, patience int, timeoutF int, hardTimeout int) (timeout, timeout) {
	var stdinTimeout, cmdTimeout timeout

	if follow {
		stdinTimeout.infinite = true
	}

	if infinite {
		cmdTimeout.infinite = true
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
