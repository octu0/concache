package concache

import (
	"time"

	"github.com/octu0/cmap"
)

const (
	defaultSlabSize      int = 1024
	defaultCacheCapacity int = 128
)

type (
	EvictedCb func(string, interface{})
	DeletedCb func(string, interface{})
	UpsertCb  func(exist bool, oldValue interface{}) (newValue interface{})
)

type CacheGetSetDelete interface {
	Set(key string, value interface{}, dur time.Duration)
	SetDefault(key string, value interface{})
	SetNoExpire(key string, value interface{})
	Get(key string) (value interface{}, exist bool)
	Delete(key string) (value interface{}, exist bool)
}

type CacheGetSetUpsertDelete interface {
	CacheGetSetDelete
	Upsert(key string, dur time.Duration, cb UpsertCb)
}

type CacheItemCount interface {
	Count() int
}

type CacheItemDeleteExpire interface {
	DeleteExpired()
}

// compile check
var (
	_ CacheGetSetDelete       = (*Cache)(nil)
	_ CacheGetSetUpsertDelete = (*Cache)(nil)
	_ CacheItemCount          = (*Cache)(nil)
	_ CacheItemDeleteExpire   = (*Cache)(nil)
)

type Cache struct {
	opt     *cacheOpt
	cache   *cmap.CMap
	janitor *janitor
}

func New(funcs ...cacheOptFunc) *Cache {
	opt := newCacheOpt()
	for _, fn := range funcs {
		fn(opt)
	}

	c := &Cache{
		opt:     opt,
		cache:   cmap.New(cmap.WithSlabSize(opt.slabSize), cmap.WithCacheCapacity(opt.cacheCapacity)),
		janitor: newJanitor(opt.cleanupInterval),
	}
	c.initJanitor()
	return c
}

func (c *Cache) initJanitor() {
	if 0 < c.opt.cleanupInterval {
		c.janitor.runBackground(c.DeleteExpired)
		c.janitor.setFinalizer()
	}
}

func (c *Cache) Set(key string, value interface{}, dur time.Duration) {
	ttl := ttlNano(dur)
	c.cache.Set(key, newCacheItem(value, ttl))
}

func (c *Cache) SetDefault(key string, value interface{}) {
	c.Set(key, value, c.opt.defaultExpiration)
}

func (c *Cache) SetNoExpire(key string, value interface{}) {
	c.Set(key, value, 0)
}

func (c *Cache) Upsert(key string, dur time.Duration, cb UpsertCb) {
	ttl := ttlNano(dur)
	c.cache.Upsert(key, func(exists bool, oldValue interface{}) interface{} {
		if exists != true {
			return newCacheItem(cb(false, nil), ttl)
		}

		item := oldValue.(*cacheItem)
		if item.isExpired() {
			return newCacheItem(cb(false, nil), ttl)
		}
		return newCacheItem(cb(true, item.value), ttl)
	})
}

func (c *Cache) UpsertDefault(key string, cb UpsertCb) {
	c.Upsert(key, c.opt.defaultExpiration, cb)
}

func (c *Cache) UpsertNoExpire(key string, cb UpsertCb) {
	c.Upsert(key, 0, cb)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	v, ok := c.cache.Get(key)
	if ok != true {
		return nil, false
	}

	item := v.(*cacheItem)
	if item.isExpired() {
		return nil, false
	}
	return item.value, true
}

func (c *Cache) Delete(key string) (interface{}, bool) {
	v, ok := c.cache.Remove(key)
	if ok != true {
		return nil, false
	}

	item := v.(*cacheItem)
	if c.opt.onDeleted != nil {
		c.opt.onDeleted(key, item.value)
	}
	return item.value, true
}

func (c *Cache) Count() int {
	return c.cache.Len()
}

type deleteExpiredKV struct {
	key   string
	value interface{}
}

func (c *Cache) DeleteExpired() {
	evictedItems := make([]*deleteExpiredKV, 0, c.cache.Len())

	collectEvicted := false
	if c.opt.onEvicted != nil {
		collectEvicted = true
	}

	nowNano := time.Now().UnixNano()
	keys := c.cache.Keys()
	for _, key := range keys {
		v, ok := c.cache.Get(key)
		if ok != true {
			continue
		}
		item := v.(*cacheItem)
		if item.isExpiredFromNano(nowNano) {
			c.cache.Remove(key)
			if collectEvicted {
				evictedItems = append(evictedItems, &deleteExpiredKV{key, item.value})
			}
		}
	}

	if c.opt.onEvicted != nil {
		for _, item := range evictedItems {
			c.opt.onEvicted(item.key, item.value)
		}
	}
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

func newCacheItem(v interface{}, e int64) *cacheItem {
	return &cacheItem{
		value:      v,
		expiration: e,
	}
}

func (c *cacheItem) isExpired() bool {
	return c.isExpiredFromNano(time.Now().UnixNano())
}

func (c *cacheItem) isExpiredFromNano(fromNano int64) bool {
	if 0 == c.expiration {
		return false
	}

	if fromNano < c.expiration {
		return false
	}
	return true
}

func ttlNano(dur time.Duration) int64 {
	if dur < 1 {
		return int64(0)
	}
	return time.Now().Add(dur).UnixNano()
}
