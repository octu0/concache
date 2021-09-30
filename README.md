# `concache`

[![MIT License](https://img.shields.io/github/license/octu0/concache)](https://github.com/octu0/concache/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/octu0/cocache?status.svg)](https://godoc.org/github.com/octu0/concache)
[![Go Report Card](https://goreportcard.com/badge/github.com/octu0/concache)](https://goreportcard.com/report/github.com/octu0/concache)
[![Releases](https://img.shields.io/github/v/release/octu0/concache)](https://github.com/octu0/concache/releases)

`concache` is in-memory key:value cache. `concache` provides thread-safe `map[string]interface{}` with expiration times.
`concache` a high-performance solution to this by sharding the map with minimal time spent waiting for locks.

## Documentation

https://godoc.org/github.com/octu0/concache

## Installation

```
$ go get github.com/octu0/concache
```

## Example

```go
import(
  "time"
  "github.com/octu0/concache"
)

func main(){
  cache := concache.New(
    concache.WithDefaultExpiration(5 * time.Second),
    concache.WithCleanupInterval(10 * time.Minute),
  )

  cache.Set("hello", "123", 1 * time.Second)
  cache.SetDefault("world", "456")
  if value, ok := cache.Get("hello"); ok {
    println("hello " + value.(string))
  }
  if value, ok := cache.Get("world"); ok {
    println("world " + value.(string))
  }

  time.Sleep(1 * time.Second)

  if _, ok := cache.Get("hello"); ok != true {
    println("hello expired")
  }
  if _, ok := cache.Get("world"); ok {
    println("world not expired")
  }
}
```

## License

MIT, see LICENSE file for details.
