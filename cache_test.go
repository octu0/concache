package concache

import (
	"testing"
	"time"
)

func TestSetGetDeleteBasic(t *testing.T) {
	c := New()
	c.Set("foo", "bar", 0)
	val, ok := c.Get("foo")
	if ok != true {
		t.Errorf("foo exists")
	}

	v, ok := val.(string)
	if ok != true {
		t.Errorf("value is string")
	}
	if v != "bar" {
		t.Errorf("value is 'bar'")
	}

	if _, ok := c.Get("helloworld"); ok {
		t.Errorf("not set key")
	}

	oldval, ok := c.Delete("foo")
	if ok != true {
		t.Errorf("foo exists")
	}
	if oldval.(string) != "bar" {
		t.Errorf("oldval is 'bal'")
	}

	if _, ok := c.Delete("foo"); ok {
		t.Errorf("already deleted")
	}
}
func TestSetGetWithExpiration(t *testing.T) {
	check := func(c *Cache, key string, value string, willExpire bool) {
		val, ok := c.Get(key)
		if ok != true {
			t.Errorf("not expire")
		}
		if val.(string) != value {
			t.Errorf("%s value is %s", key, value)
		}
		time.Sleep(5 * time.Millisecond)
		if _, ok := c.Get(key); ok != true {
			t.Errorf("%s not expire", key)
		}
		time.Sleep(5 * time.Millisecond)
		if willExpire {
			v, ok := c.Get(key)
			if ok {
				t.Errorf("%s should expire", key)
			}
			if v != nil {
				t.Errorf("%s should expire value is nil", key)
			}
		} else {
			v, ok := c.Get(key)
			if ok != true {
				t.Errorf("%s shoud alive", key)
			}
			if v != value {
				t.Errorf("%s shoud alive value = %s", key, value)
			}
		}
	}
	c := New()
	c.Set("foo", "bar", 10*time.Millisecond)
	check(c, "foo", "bar", true)

	c.SetNoExpire("foo1", "bar1")
	check(c, "foo1", "bar1", false)

	c.SetDefault("foo_d", "bar_d")
	check(c, "foo_d", "bar_d", false) // default expiration 0

	c1 := New(
		DefaultExpiration(10 * time.Millisecond),
	)
	c1.SetDefault("foo2", "bar2")
	check(c1, "foo2", "bar2", true)

	c1.SetNoExpire("foo3", "bar3")
	check(c1, "foo3", "bar3", false)

	c1.Set("foo_0", "bar_0", 0)
	check(c1, "foo_0", "bar_0", false)

	c2 := New(
		DefaultExpiration(100 * time.Millisecond),
	)
	c2.Set("foo4", "bar4", 10*time.Millisecond)
	check(c2, "foo4", "bar4", true)
	c2.SetDefault("foo5", "bar5")
	check(c2, "foo5", "bar5", false)
}

func TestUpsertBasic(t *testing.T) {
	c := New()
	c.Upsert("foo1", 5*time.Millisecond, func(exist bool, oldValue interface{}) interface{} {
		if exist {
			t.Errorf("currently not exist key")
		}
		if oldValue != nil {
			t.Errorf("not exist value")
		}
		return "hello world"
	})
	val, ok := c.Get("foo1")
	if ok != true {
		t.Errorf("Upserted key")
	}
	if val.(string) != "hello world" {
		t.Errorf("upsert value")
	}

	time.Sleep(10 * time.Millisecond)

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("expired key")
	}

	c.Upsert("foo1", 5*time.Millisecond, func(exist bool, oldValue interface{}) interface{} {
		if exist {
			t.Errorf("expired key")
		}
		if oldValue != nil {
			t.Errorf("expired value")
		}
		return "hello2 world2"
	})

	val, ok = c.Get("foo1")
	if ok != true {
		t.Errorf("upsert key")
	}
	if val.(string) != "hello2 world2" {
		t.Errorf("upsert value")
	}

	c.Upsert("foo1", 15*time.Millisecond, func(exist bool, oldValue interface{}) interface{} {
		if exist != true {
			t.Errorf("not expired key")
		}
		if oldValue.(string) != "hello2 world2" {
			t.Errorf("previous value 'hello2 world2'")
		}
		return "hello3 world3"
	})

	val, ok = c.Get("foo1")
	if ok != true {
		t.Errorf("upsert key")
	}
	if val.(string) != "hello3 world3" {
		t.Errorf("upsert value")
	}

	time.Sleep(10 * time.Millisecond)

	if _, ok := c.Get("foo1"); ok != true {
		t.Errorf("update expire time: previous 10ms current 12ms")
	}

	time.Sleep(15 * time.Millisecond)

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("expired key")
	}
}

func TestUpsertConcurrent(t *testing.T) {
	type rendez struct {
		done chan struct{}
	}
	run := func(c *Cache, ch1 chan rendez, ch2 chan rendez, key string) {
		go func() {
			for {
				select {
				case r := <-ch1:
					t := 10 * time.Millisecond
					c.Upsert(key, t, func(exist bool, oldValue interface{}) interface{} {
						if exist {
							old := oldValue.([]string)
							return append(old, "hello2")
						}
						return []string{"hello1"}
					})
					r.done <- struct{}{}
					return
				}
			}
		}()
		go func() {
			for {
				select {
				case r := <-ch2:
					t := 10 * time.Millisecond
					c.Upsert(key, t, func(exist bool, oldValue interface{}) interface{} {
						if exist {
							old := oldValue.([]string)
							return append(old, "world2")
						}
						return []string{"world1"}
					})
					r.done <- struct{}{}
					return
				}
			}
		}()
	}

	ch1 := make(chan rendez)
	ch2 := make(chan rendez)
	c1 := New()
	run(c1, ch1, ch2, "foobar")

	done1 := make(chan struct{})
	done2 := make(chan struct{})
	ch1 <- rendez{done1}
	<-done1
	ch2 <- rendez{done2}
	<-done2

	val1, ok1 := c1.Get("foobar")
	if ok1 != true {
		t.Errorf("upsert key")
	}
	if v1, ok := val1.([]string); ok != true {
		t.Errorf("string slice")
	} else {
		if len(v1) != 2 {
			t.Errorf("2 string")
		}
		if (v1[0] == "hello1" && v1[1] == "world2") != true {
			t.Errorf("first upsert not exist / second upsert exist")
		}
	}

	ch3 := make(chan rendez)
	ch4 := make(chan rendez)
	c2 := New()
	run(c2, ch3, ch4, "quux")

	done3 := make(chan struct{})
	done4 := make(chan struct{})

	ch4 <- rendez{done4}
	<-done4
	ch3 <- rendez{done3}
	<-done3

	val2, ok2 := c2.Get("quux")
	if ok2 != true {
		t.Errorf("upsert key")
	}
	if v2, ok := val2.([]string); ok != true {
		t.Errorf("string slice")
	} else {
		if len(v2) != 2 {
			t.Errorf("2 string")
		}
		if (v2[0] == "world1" && v2[1] == "hello2") != true {
			t.Errorf("expect: drifted")
		}
	}

	ch5 := make(chan rendez)
	ch6 := make(chan rendez)
	c3 := New()
	run(c3, ch5, ch6, "qwerty")

	done5 := make(chan struct{})
	done6 := make(chan struct{})

	c3.Set("qwerty", []string{"hoge"}, 10*time.Millisecond)

	ch5 <- rendez{done5}
	<-done5

	if val3, ok := c3.Get("qwerty"); ok != true {
		t.Errorf("first upsert")
	} else {
		v3 := val3.([]string)
		if len(v3) != 2 {
			t.Errorf("2 string")
		}
		if (v3[0] == "hoge" && v3[1] == "hello2") != true {
			t.Errorf("prepend 'hoge'")
		}
	}

	ch6 <- rendez{done6}
	<-done6

	if val3, ok := c3.Get("qwerty"); ok != true {
		t.Errorf("second upsert")
	} else {
		v3 := val3.([]string)
		if len(v3) != 3 {
			t.Errorf("3 string")
		}
		if (v3[0] == "hoge" && v3[1] == "hello2" && v3[2] == "world2") != true {
			t.Errorf("prepend 'hoge'")
		}
	}
}

func TestDeleteExpired(t *testing.T) {
	type tuple struct {
		key   string
		value string
	}
	expects := []tuple{
		tuple{"foo1", "bar1"},
		tuple{"foo2", "bar2"},
		tuple{"foo3", "bar3"},
	}
	c := New(
		Evicted(func(key string, value interface{}) {
			found := false
			for _, e := range expects {
				if e.key == key {
					found = true

					if value.(string) != e.value {
						t.Errorf("expect value: %s actual: %s", e.value, value.(string))
					}
				}
			}
			if found != true {
				t.Errorf("expect key not found: %s", key)
			}
		}),
	)
	c.Set("foo1", "bar1", 5*time.Millisecond)
	c.Set("foo2", "bar2", 6*time.Millisecond)
	c.Set("foo3", "bar3", 7*time.Millisecond)
	c.Set("foo4", "bar4", 20*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("expire key (soft delete)")
	}
	if _, ok := c.Get("foo2"); ok {
		t.Errorf("expire key (soft delete)")
	}
	if _, ok := c.Get("foo3"); ok {
		t.Errorf("expire key (soft delete)")
	}

	c.DeleteExpired()

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("expire key (hard delete)")
	}
	if _, ok := c.Get("foo2"); ok {
		t.Errorf("expire key (hard delete)")
	}
	if _, ok := c.Get("foo3"); ok {
		t.Errorf("expire key (hard delete)")
	}
}
