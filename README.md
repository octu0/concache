# concache

`concache` is in-memory key:value cache. `concache` provides thread-safe `map[string]interface{}` with expiration times.
`concache` a high-performance solution to this by sharding the map with minimal time spent waiting for locks.

## Installation

```
$ go get github.com/octu0/concache
```

## Usage

Import the package

```
import(
  "time"
  "github.com/octu0/concache"
)

func main(){
  cache := concache.New(
    concache.DefaultShardSize(),
    concache.DefaultExpiration(5 * time.Second),
    concache.CleanupInterval(10 * time.Minute),
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
  if _, ok := cache.Get("world"); ok != true {
    println("world expired")
  }
}
```
