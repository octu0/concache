package concache

import(
  "time"
  "runtime"
)

type Cache struct {
  shard              *MapShard
  janitor            *Janitor
  defaultExpiration  time.Duration
  cleanupInterval    time.Duration
  evictedCb          EvictedCb
  deletedCb          DeletedCb
}

func New(funcs ...OptionsFunc) *Cache {
  opts := new(Options)
  for _, fn := range funcs {
    fn(opts)
  }

  c                  := new(Cache)
  c.shard             = newMapShard(opts.ShardSize)
  c.defaultExpiration = opts.DefaultExpiration
  c.cleanupInterval   = opts.CleanupInterval
  c.evictedCb         = opts.OnEvicted
  c.deletedCb         = opts.OnDeleted

  if 0 < c.cleanupInterval {
    janitor  := newJanitor(c.cleanupInterval)
    c.janitor = janitor
    janitor.Run(c)
    runtime.SetFinalizer(c, stopJanitor)
  }
  return c
}
func stopJanitor(c *Cache) {
  c.janitor.Stop()
}

func (c *Cache) Set(key string, value interface{}, dur time.Duration) {
  ttl   := createTTL(dur)
  shard := c.shard.GetShard(key)
  shard.Lock()
  shard.items[key] = CacheItem{ Value: value, Expiration: ttl }
  shard.Unlock()
}
func (c *Cache) SetDefault(key string, value interface{}) {
  c.Set(key, value, c.defaultExpiration)
}
func (c *Cache) SetNoExpire(key string, value interface{}) {
  c.Set(key, value, 0)
}

type UpsertCb func(exist bool, oldValue interface{}) (newValue interface{})
func (c *Cache) Upsert(key string, dur time.Duration, cb UpsertCb) {
  var newValue interface{}
  ttl   := createTTL(dur)
  shard := c.shard.GetShard(key)
  shard.Lock()
  item, ok := shard.items[key]
  if ok {
    if item.Expired() {
      newValue = cb(false, nil)
    } else {
      newValue = cb(true, item.Value)
    }
  } else {
    newValue = cb(false, nil)
  }
  shard.items[key] = CacheItem{ Value: newValue, Expiration: ttl }
  shard.Unlock()
}
func (c *Cache) UpsertDefault(key string, cb UpsertCb) {
  c.Upsert(key, c.defaultExpiration, cb)
}
func (c *Cache) UpsertNoExpire(key string, cb UpsertCb) {
  c.Upsert(key, 0, cb)
}

func (c *Cache) Get(key string) (interface{}, bool) {
  shard := c.shard.GetShard(key)
  shard.RLock()
  item, ok := shard.items[key]
  if ok != true {
    shard.RUnlock()
    return nil, false
  }
  if item.Expired() {
    shard.RUnlock()
    return nil, false
  }
  shard.RUnlock()
  return item.Value, true
}

func (c *Cache) Delete(key string) (interface{}, bool) {
  shard := c.shard.GetShard(key)
  shard.Lock()
  item, ok := shard.items[key]
  if ok != true {
    shard.Unlock()
    return nil, false
  }
  delete(shard.items, key)
  c.onDeleted(key, item.Value)
  shard.Unlock()
  return item.Value, true
}

func (c *Cache) Count() int {
  count := 0
  for _, shard := range c.shard.GetShards() {
    shard.RLock()
    count += len(shard.items)
    shard.RUnlock()
  }
  return count
}

type kv struct {
  key   string
  value interface{}
}
func (c *Cache) DeleteExpired() {
  var evictedItems []kv
  now            := time.Now().UnixNano()
  collectEvicted := false
  if c.evictedCb != nil {
    collectEvicted = true
    evictedItems   = make([]kv, 0)
  }

  for _, shard := range c.shard.GetShards() {
    shard.Lock()
    for key, item := range shard.items {
      if item.ExpiredFrom(now) != true {
        delete(shard.items, key)
        if collectEvicted {
          evictedItems = append(evictedItems, kv{key, item.Value})
        }
      }
    }
    shard.Unlock()
  }
  for _, item := range evictedItems {
    c.onExpired(item.key, item.value)
  }
}
func (c *Cache) onDeleted(key string, value interface{}) {
  if c.deletedCb != nil {
    c.deletedCb(key, value)
  }
}
func (c *Cache) onExpired(key string, value interface{}) {
  if c.evictedCb != nil {
    c.evictedCb(key, value)
  }
}

func createTTL(dur time.Duration) int64 {
  ttl := int64(0)
  if 0 < dur {
    ttl = time.Now().Add(dur).UnixNano()
  }
  return ttl
}
