package concache

import (
	"testing"
	"time"
)

func TestJanitorRun(t *testing.T) {
	c := New(
		WithEvicted(func(key string, value interface{}) {
			if (key == "foo1" || key == "foo2") != true {
				t.Errorf("unexpected key %s", key)
			}
			println("janitor collect " + key)
		}),
		WithCleanupInterval(5*time.Millisecond),
	)
	c.Set("foo1", "bar1", 1*time.Millisecond)
	c.Set("foo2", "bar2", 2*time.Millisecond)

	if _, ok := c.Get("foo1"); ok != true {
		t.Errorf("not expire")
	}
	if _, ok := c.Get("foo2"); ok != true {
		t.Errorf("not expire")
	}

	time.Sleep(10 * time.Millisecond)

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("janitor called")
	}
	if _, ok := c.Get("foo2"); ok {
		t.Errorf("janitor called")
	}
}
