# protohackers

Solutions to https://protohackers.com

Done:

- [x] Smoke Test
- [x] Prime Time
- [x] Means to an End
- [x] Budget Chat
- [x] Unusual Database Program

## Deploying

> TODO

```shell
make deploy
```

## Testing

```shell
go test -v ./...
```

## Notes

### 0

- Took me a hot minute to figure out how to do this

### 1

- JSON JSON JSON
- Learned about prime numbers

### 2

- Learned about Big Endian and sending numbers over the network
- Spent forever figuring out why 200k test case was constantly failing. Turns out the way I calculated the mean was causing an int32 overflow.
- Experimented with [B-Tree](https://github.com/tidwall/btree) library which was a lot of fun.

### 3

- The most fun challenge yet
- Learned about multi-threaded chat servers
- Parsing
- Avoiding deadlocks

### 4

- Oh dang UDP is way different than TCP

### 5

- Regex parsing
    - Go's lack of forward/negative lookahead
- TCP forwarding
- Goroutines for reading/writing from the client -> server, server -> upstream
