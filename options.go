package concache

import (
	"time"
)

type cacheOptFunc func(opt *cacheOpt)

type cacheOpt struct {
	slabSize          int
	cacheCapacity     int
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	onEvicted         EvictedCb
	onDeleted         DeletedCb
}

func newCacheOpt() *cacheOpt {
	return &cacheOpt{
		slabSize:      defaultSlabSize,
		cacheCapacity: defaultCacheCapacity,
	}
}

func WithDefaultSlabSize() cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.slabSize = defaultSlabSize
	}
}

func WithSlabSize(size int) cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.slabSize = size
	}
}

func WithDefaultCacheCapacity() cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.cacheCapacity = defaultCacheCapacity
	}
}

func WithCacheCapacity(size int) cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.cacheCapacity = size
	}
}

func WithDefaultExpiration(dur time.Duration) cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.defaultExpiration = dur
	}
}

func WithCleanupInterval(intval time.Duration) cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.cleanupInterval = intval
	}
}

func WithEvicted(cb EvictedCb) cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.onEvicted = cb
	}
}

func WithDeleted(cb DeletedCb) cacheOptFunc {
	return func(opt *cacheOpt) {
		opt.onDeleted = cb
	}
}
