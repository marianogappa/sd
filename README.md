# sd - Stream Diff
Diffs two streams of line-separated strings, timing them out if necessary.

[![Build Status](https://img.shields.io/travis/MarianoGappa/sd.svg)](https://travis-ci.org/MarianoGappa/sd)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/MarianoGappa/sd/master/LICENSE)

```
sd [OPTIONS] 'COMMAND'
```
## Options

**-f --follow**

keeps reading from `STDIN` until `SIGINT` (think tail -f).

**-p --patience %seconds%**

wait for the specified seconds for the first received line. Use `0` for waiting forever.

**-t --timeout %seconds%**

`exit(0)` after specified seconds from last received line. `STDIN` and `COMMAND` have independent timeouts. When with `-f`, timeout only applies to `COMMAND` (not to `STDIN`).

**-h --hard-timeout %seconds%**

`exit(0)` after the specified seconds (or earlier). Overrides all other options.

## Installing

Find the latest binaries for your OS in the [Releases](https://github.com/MarianoGappa/sd/releases/) section.

Or via Go:
```
go install github.com/MarianoGappa/sd
```

## What does it do?

`sd` is very similar to [comm](https://en.wikipedia.org/wiki/Comm), but with two important differences:

1. `comm` diffs files. `sd` can diff files, but it's meant for diffing streams; especially infinite ones. (To be fair, `comm - file2` reads from STDIN which could be an infinite stream, but file2 has to end)
2. `comm` requires the files to be sorted. `sd` doesn't; it compares each line in one stream to all lines in the other.

A closer approximation to `sd` is `grep -Fxvf` (which doesn't require sorting), again with two important differences:

1. Again, `grep -Fxvf` can't be used for streams that don't finish.
2. `grep -Fxvf` is optimised for regexes, which makes it orders of magnitude slower than `sd`.

A basic idea of what `sd` does is:
```
$ echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'
1
3
5
```

But both `STDIN` and `COMMAND` can be lengthy operations or infinite streams.

## Use case examples

- Query user cities and excluded cities at the same time. Output only non-excluded cities. If it takes longer than 120 seconds, stop.
```
mysql db1 -Nsre "SELECT city FROM user" | sd -h 120 mysql db2 -Nsre "SELECT city FROM excluded_city"
```
- Query user cities and consume excluded cities from a Kafka consumer at the same time, and output only non-excluded cities to output.txt. Allow infinite time to pass for the first streamed element (as query could take a long time to run), but then timeout if 10 seconds pass since the last streamed message.
```
mysql -Nsre "SELECT city FROM user" | sd -p 0 -t 10 kafka_consumer --topic excluded_city > city.txt
```

## Details

- Note that `sd` does not guarantee order of output, nor uniqueness. If you need those, just `| sort | uniq`.
- Because `sd` compares every line of `STDIN` against all lines in the second stream, and although it's very fast, execution time will increase linearly as one stream grows and quadratically as both streams grow. 1M^2 comparisons might become impractical.
