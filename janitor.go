package concache

import(
  "time"
)

type JanitorDone  chan struct{}

type Janitor struct {
  interval   time.Duration
  done       JanitorDone
}

func newJanitor(interval time.Duration) *Janitor {
  j         := new(Janitor)
  j.interval = interval
  j.done     = make(JanitorDone)
  return j
}
func (j *Janitor) run(c *Cache) {
  tick := time.NewTicker(j.interval)
  defer tick.Stop()
  for {
    select {
    case <-j.done:
      return
    case <-tick.C:
      c.DeleteExpired()
    }
  }
}
func (j *Janitor) Run(c *Cache) {
  go j.run(c)
}
func (j *Janitor) Stop() {
  j.done <-struct{}{}
}
