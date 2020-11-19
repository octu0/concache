package concache

import (
	"testing"
	"time"
)

func TestHashingfnv32(t *testing.T) {
	if fnv32("foo") != fnv32("foo") {
		t.Errorf("same hash key")
	}
	if fnv32("foo") == fnv32("bar") {
		t.Errorf("not same hash key")
	}
	println(fnv32("foo"))
	println(fnv32("foo"))
	println(fnv32("bar"))
}

func TestGetShardBasic(t *testing.T) {
	ms := newMapShard(1)
	s1 := ms.GetShard("hello")
	s1.items["hello"] = CacheItem{"mark1", 0}
	s2 := ms.GetShard("world")
	s2.items["world"] = CacheItem{"mark2", 0}

	s1 = ms.GetShard("hello")
	if _, ok := s1.items["hello"]; ok != true {
		t.Errorf("should be exist")
	}
	s2 = ms.GetShard("world")
	if _, ok := s2.items["world"]; ok != true {
		t.Errorf("should be exist")
	}

	s1item := s1.items["hello"]
	s2item := s2.items["world"]
	if s1item.Value == s2item.Value {
		t.Errorf("different key")
	}
	if len(s1.items) != len(s2.items) {
		t.Errorf("shard size is 1 (s1 and s2 same shard)")
	}
}
func TestGetShard(t *testing.T) {
	ms1 := newMapShard(1)
	if ms1.GetShard("foo").id != ms1.GetShard("foo").id {
		t.Errorf("same key same shard")
	}
	if ms1.GetShard("foo").id != ms1.GetShard("bar").id {
		t.Errorf("same shard")
	}
	ms2 := newMapShard(2)
	if ms2.GetShard("foo").id != ms2.GetShard("foo").id {
		t.Errorf("same key same shard")
	}
	if ms2.GetShard("foo").id == ms2.GetShard("bar").id {
		t.Errorf("maybe not same shard")
	}
	ms128 := newMapShard(128)
	if ms128.GetShard("foo").id != ms128.GetShard("foo").id {
		t.Errorf("same key same shard")
	}
	if ms128.GetShard("foo").id == ms128.GetShard("bar").id {
		t.Errorf("maybe not same shard")
	}
}

func TestGetShards(t *testing.T) {
	ms1 := newMapShard(1)
	if len(ms1.GetShards()) != 1 {
		t.Errorf("shard size = 1")
	}
	ms2 := newMapShard(16)
	if len(ms2.GetShards()) != 16 {
		t.Errorf("shard size = 16")
	}
	ms3 := newMapShard(128)
	if len(ms3.GetShards()) != 128 {
		t.Errorf("shard size = 128")
	}
}

func TestShardLockSingleShardSameKey(t *testing.T) {
	// GetShard same key locking
	// expect: same shard lock
	KEY := "foo"

	ms1 := newMapShard(1)
	shard1 := ms1.GetShard(KEY)
	println("shard1 lock[1]")
	shard1.Lock()
	shard1.items["bar"] = CacheItem{"mark1", 0}

	ch1 := make(chan struct{})
	ch2 := make(chan bool)
	go func() {
		for {
			select {
			case <-time.After(10 * time.Millisecond):
				ch2 <- true
				return
			case <-ch1:
				ch2 <- false
				return
			}
		}
	}()

	go func() {
		s1 := ms1.GetShard(KEY)
		println("shard1 lock[2]")
		s1.Lock()
		shard1.items["bar"] = CacheItem{"mark2", 0}
		s1.Unlock()
		println("shard1 unlock[2]")
		ch1 <- struct{}{}
	}()

	timeout := <-ch2
	if timeout != true {
		t.Errorf("shard1 locked by main")
	}
	shard1.Unlock()
	println("shard1 unlock[1]")
}
func TestShardLockSingleShardDiffererntKey(t *testing.T) {
	// GetShard different key locking
	// expect: same shard lock
	KEY1 := "foo1"
	KEY2 := "foo2"

	ms1 := newMapShard(1)
	shard1 := ms1.GetShard(KEY1)
	println("shard1 lock[1]")
	shard1.Lock()
	shard1.items["bar"] = CacheItem{"mark1", 0}

	ch1 := make(chan struct{})
	ch2 := make(chan bool)
	go func() {
		for {
			select {
			case <-time.After(10 * time.Millisecond):
				ch2 <- true
				return
			case <-ch1:
				ch2 <- false
				return
			}
		}
	}()
	go func() {
		s1 := ms1.GetShard(KEY2)
		println("shard1 lock[2]")
		s1.Lock()
		shard1.items["bar"] = CacheItem{"mark2", 0}
		s1.Unlock()
		println("shard1 unlock[2]")
		ch1 <- struct{}{}
	}()

	timeout := <-ch2
	if timeout != true {
		t.Errorf("shard1 locked by main")
	}
	shard1.Unlock()
	println("shard1 unlock[1]")
}

func TestShardLockMultiShardDiffererntKey(t *testing.T) {
	// GetShard different key locking
	// expect: shard by shard lock
	KEY1 := "foo"
	KEY2 := "bar"

	ms2 := newMapShard(2)
	shard1 := ms2.GetShard(KEY1)
	println("shard1 lock[1]")
	shard1.Lock()
	shard1.items["bar"] = CacheItem{"mark1", 0}

	ch1 := make(chan struct{})
	ch2 := make(chan bool)
	go func() {
		for {
			select {
			case <-time.After(10 * time.Millisecond):
				ch2 <- true
				return
			case <-ch1:
				ch2 <- false
				return
			}
		}
	}()
	go func() {
		s1 := ms2.GetShard(KEY2)
		println("shard1 lock[2]")
		s1.Lock()
		shard1.items["bar"] = CacheItem{"mark2", 0}
		s1.Unlock()
		println("shard1 unlock[2]")
		ch1 <- struct{}{}
	}()

	timeout := <-ch2
	if timeout {
		t.Errorf("different shard lock")
	}
	shard1.Unlock()
	println("shard1 unlock[1]")
}
