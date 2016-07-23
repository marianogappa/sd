# sd - Stream Diff
Diffs two streams, timing them out if necessary.

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

In principle, this:
```
$ echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'
1
3
5
```

But both `STDIN` and `COMMAND` can be lengthy operations or infinite streams, like:
```
$ for i in {1..100}; do ( if [[ $(($RANDOM%10)) -eq 0 ]]; then echo "n"; fi; ); done | sd -h 1 yes
n
n
n
...
```

## Examples

- Echo the range of numbers 1 to 5, filtering out even numbers.
```
echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'
```
- Echo random numbers only if they're not in the 1-500 range. Stop after 1 second has passed.
```
while :; do echo $RANDOM; sleep .1; done | sd -h 1 'seq 500'
```
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

## I mostly diff files that do terminate; why `sd` over `grep -Fxvf`?
Because `sd` is orders of magnitude faster than `grep -Fxvf` (depending on hardware and `GOMAXPROCS`). `sd` violently parallelises work using goroutines (at the expense of CPU usage), while remaining idle while waiting for stream IO.

```
$ seq 10000 > a
$ seq 5001 10000 > b
$ time bash -c "grep -Fxvf a b > c"

real    0m14.519s
user    0m14.475s
sys     0m0.031s

$ wc -l c
    5000 c
    
$ time bash -c "cat a | sd 'cat b' > c"

real    0m0.107s
user    0m0.445s
sys     0m0.040s

$ wc -l c
    5000 c
    
$ time bash -c "seq 10000 | sd 'seq 5001 10000' > c"

real    0m0.105s
user    0m0.443s
sys     0m0.039s

$ wc -l c
    5000 c
```
