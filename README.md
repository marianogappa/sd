# sd - Stream Diff

## What does it do?

In principle, this:
```
$ echo -e "1\n2\n3\n4\n5" | sd 'echo -e "2\n4"'
1
3
5
```

But both STDIN and the string representing a command can be lengthy operations that stream results, like:
```
$ echo "1" && sleep 1 && echo "2" | sd 'sleep 1 && echo "2"'
1
```

## Use case (that motivated this)

```
*mysql_query* | sd *kafka_consumer* | *kafka_producer*
```

## TODO

- separate timeouts for stdin and command
- optional timeout for stdin
- output uniques
- take all parameters from flags
- test for sleeping for less than timeout in between sends, repeatedly
