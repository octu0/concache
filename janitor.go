package concache

import (
	"context"
	"runtime"
	"time"
)

type janitorFunc func()

type janitor struct {
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
}

func newJanitor(interval time.Duration) *janitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &janitor{
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
	}
}

func (j *janitor) stop() {
	j.cancel()
}

func (j *janitor) run(fn janitorFunc) {
	tick := time.NewTicker(j.interval)
	defer tick.Stop()

	for {
		select {
		case <-j.ctx.Done():
			return
		case <-tick.C:
			fn()
		}
	}
}

func (j *janitor) runBackground(fn janitorFunc) {
	go j.run(fn)
}

func (j *janitor) setFinalizer() {
	runtime.SetFinalizer(j, stopJanitor)
}

func stopJanitor(j *janitor) {
	runtime.SetFinalizer(j, nil) // clear finalizer
	j.stop()
}
