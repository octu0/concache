package concache

import(
  "time"
  "sync"
)

const(
  DEFAULT_SHARD_COUNT = uint(32)
  offset32            = uint32(2166136261)
  prime32             = uint32(16777619)
)

type CacheItem struct {
  Value      interface{}
  Expiration int64
}
func (item CacheItem) Expired() bool {
  return item.ExpiredFrom(time.Now().UnixNano())
}
func (item CacheItem) ExpiredFrom(fromNano int64) bool {
  if 0 == item.Expiration {
    return false
  }
  if fromNano < item.Expiration {
    return false
  }
  return true
}

type MapShard struct {
  size   uint
  shards []*CacheMap
}

type CacheMap struct {
  id     uint
  items  map[string]CacheItem
  sync.RWMutex
}

func newMapShard(size uint) *MapShard {
  shards := make([]*CacheMap, size)
  for i := 0; i < int(size); i += 1 {
    m        := new(CacheMap)
    m.id      = uint(i)
    m.items   = make(map[string]CacheItem)
    shards[i] = m
  }

  ms       := new(MapShard)
  ms.size   = size
  ms.shards = shards
  return ms
}

func (ms *MapShard) GetShard(key string) *CacheMap {
  idx := uint(fnv32(key)) % ms.size
  return ms.shards[idx]
}
func (ms *MapShard) GetShards() []*CacheMap {
  return ms.shards
}

func fnv32(key string) uint32 {
  h := offset32
  for _, k := range key {
    h = (h ^ uint32(k)) * prime32
  }
  return h
}
