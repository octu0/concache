package concache

import(
  "time"
)

type EvictedCb   func(string, interface{})
type DeletedCb   func(string, interface{})

type OptionsFunc func(opts *Options)

type Options struct {
  ShardSize         uint
  DefaultExpiration time.Duration
  CleanupInterval   time.Duration
  OnEvicted         EvictedCb
  OnDeleted         DeletedCb
}

func DefaultShardSize() OptionsFunc {
  return func(opts *Options){
    opts.ShardSize = DEFAULT_SHARD_COUNT
  }
}
func ShardSize(size uint) OptionsFunc {
  return func(opts *Options){
    opts.ShardSize = size
  }
}
func DefaultExpiration(dur time.Duration) OptionsFunc {
  return func(opts *Options) {
    opts.DefaultExpiration = dur
  }
}
func CleanupInterval(intval time.Duration) OptionsFunc {
  return func(opts *Options) {
    opts.CleanupInterval = intval
  }
}
func Evicted(cb EvictedCb) OptionsFunc {
  return func(opts *Options) {
    opts.OnEvicted = cb
  }
}
func Deleted(cb DeletedCb) OptionsFunc {
  return func(opts *Options) {
    opts.OnDeleted = cb
  }
}
