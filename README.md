# sd - Stream Diff

```
sd [options] 'command'
```

## What does it do?

In principle, this:
```
$ echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'
1
3
5
```

But both STDIN and `command` can be lengthy operations or infinite streams, like:
```
$ for i in {1..1000}; do ( if [[ $(( $RANDOM % 10)) -eq 0 ]] ; then echo "n" ; else echo "y"; fi; sleep .1; ); done | ./sd -h 1 yes
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
mysql schema_1 -Nsr -e "SELECT city FROM users" | sd -h 120 mysql schema_2 -Nsr -e "SELECT city FROM excluded_cities"
```

```
mysql -Nsr -e "SELECT city FROM users" | sd -p 0 -t 10 kafka_consumer --topic excluded_cities > active_cities.txt
```

## Options

**-f --follow**

keeps reading from `STDIN` until `SIGINT` (think tail -f).

**-p --patience %seconds%**

wait for the specified seconds for the first received line. Use `0` for waiting forever.

**-t --timeout %seconds%**

`exit(0)` after specified seconds from last received line. `STDIN` and `command` have independent timeouts. When with `-f`, timeout only applies to `command` (not to `STDIN`).

**-h --hard-timeout %seconds%**

`exit(0)` after the specified seconds (or earlier). Overrides all other options.

## TODO

- output uniques
