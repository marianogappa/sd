# sd - Stream Diff

```
sd [OPTIONS] 'COMMAND'
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
$ for i in {1..100}; do ( if [[ $(($RANDOM%10)) -eq 0 ]]; then echo "n"; fi; ); done | ./sd -h 1 yes
n
n
n
...
```

## Examples

```
echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'
```

```
while :; do echo $RANDOM; sleep .1; done | sd -h 1 'seq 500'
```

```
mysql db1 -Nsre "SELECT city FROM user" | sd -h 120 mysql db2 -Nsre "SELECT city FROM excluded_city"
```

```
mysql -Nsre "SELECT city FROM user" | sd -p 0 -t 10 kafka_consumer --topic excluded_city > city.txt
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

## TODO

- output uniques
