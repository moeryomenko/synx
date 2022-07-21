package synx

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Suppressor represents a class of work and forms a namespace in
// which units of work can be executed with duplicate suppression.
type Suppressor struct {
	group      singleflight.Group
	timers     map[string]time.Time
	mu         sync.Mutex
	exparition time.Duration
}

func NewSuppressor(expiration time.Duration) *Suppressor {
	return &Suppressor{
		group:      singleflight.Group{},
		timers:     make(map[string]time.Time),
		exparition: expiration,
	}
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
// The return a channel that will receive the
// results when they are ready.
//
// The returned channel will not be closed.
func (s *Suppressor) Do(key string, fn func() (any, error)) <-chan singleflight.Result {
	s.mu.Lock()
	now := time.Now()
	timer, ok := s.timers[key]
	if !ok {
		s.timers[key] = now
		timer = now
	}
	if now.After(timer.Add(s.exparition)) {
		s.group.Forget(key)
		delete(s.timers, key)
	}
	s.mu.Unlock()

	return s.group.DoChan(key, fn)
}
