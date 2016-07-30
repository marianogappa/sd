package main

import (
	"flag"
	"fmt"
	"log"
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
}

type options struct {
	follow      bool
	infinite    bool
	patience    int
	timeoutF    int
	hardTimeout int
}

func defineOptions(fs *flag.FlagSet) *options {
	followHelp := "keeps reading from STDIN until SIGINT or its end."
	infiniteHelp := "keeps reading from COMMAND until it ends rather than timing it out. Note that if the stream doesn't end, sd just blocks forever and does nothing."
	patienceHelp := "wait for the specified seconds for the first received line. Use 0 for waiting forever."
	timeoutHelp := "exit(0) after specified seconds from last received line. STDIN and command have independent timeouts. When with -f, timeout only applies to the command (not to STDIN)."
	hardTimeoutHelp := "exit(0) after the specified seconds (or earlier). Overrides all other options."

	var o options
	setDefaultOptions(&o)

	fs.BoolVar(&o.follow, "follow", o.follow, followHelp)
	fs.BoolVar(&o.follow, "f", o.follow, followHelp)
	fs.BoolVar(&o.infinite, "infinite", o.infinite, infiniteHelp)
	fs.BoolVar(&o.infinite, "i", o.infinite, infiniteHelp)
	fs.IntVar(&o.patience, "patience", o.patience, patienceHelp)
	fs.IntVar(&o.patience, "p", o.patience, patienceHelp)
	fs.IntVar(&o.timeoutF, "timeout", o.timeoutF, timeoutHelp)
	fs.IntVar(&o.timeoutF, "t", o.timeoutF, timeoutHelp)
	fs.IntVar(&o.hardTimeout, "hard-timeout", o.hardTimeout, hardTimeoutHelp)
	fs.IntVar(&o.hardTimeout, "h", o.hardTimeout, hardTimeoutHelp)

	fs.Usage = usage

	return &o
}

func setDefaultOptions(o *options) {
	o.follow = false
	o.infinite = false
	o.patience = -1
	o.timeoutF = 10
	o.hardTimeout = 0
}

func resolveOptions(args []string) (*options, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	o := defineOptions(fs)
	err := fs.Parse(args)
	if err != nil {
		return o, err
	}

	return o, nil
}

func mustResolveOptions(args []string) *options {
	o, err := resolveOptions(args)
	if err != nil {
		log.Fatal(err)
	}

	return o
}

func resolveTimeouts(options *options) (timeout, timeout) {
	var stdinTimeout, cmdTimeout timeout

	if options.follow {
		stdinTimeout.infinite = true
	}

	if options.infinite {
		cmdTimeout.infinite = true
	}

	if options.patience == 0 {
		stdinTimeout.firstTimeInfinite = true
		cmdTimeout.firstTimeInfinite = true
	} else if options.patience == -1 {
		stdinTimeout.firstTime = time.Duration(options.timeoutF) * time.Second
		cmdTimeout.firstTime = time.Duration(options.timeoutF) * time.Second
	} else {
		stdinTimeout.firstTime = time.Duration(options.patience) * time.Second
		cmdTimeout.firstTime = time.Duration(options.patience) * time.Second
	}

	stdinTimeout.time = time.Duration(options.timeoutF) * time.Second
	cmdTimeout.time = time.Duration(options.timeoutF) * time.Second

	if options.hardTimeout > 0 {
		stdinTimeout.hard = true
		cmdTimeout.hard = true
		stdinTimeout.firstTime = time.Duration(options.hardTimeout) * time.Second
		cmdTimeout.firstTime = time.Duration(options.hardTimeout) * time.Second
	}

	return stdinTimeout, cmdTimeout
}
