package synx

import (
	"sync"
	"time"
)

// call is an in-flight or completed suppressor.Do call
type call struct {
	wg *WaitGroup

	// These fields are written once before the WaitGroup is done
	// and are only read after the WaitGroup is done.
	val interface{}
	err error

	// expiration indicates retention time.
	expiration time.Time

	// These fields are read and written with the suppressor group
	// mutex held before the WaitGroup is done, and are read but
	// not written after the WaitGroup is done.
	chans []chan<- Result
	mu    sync.Mutex
}

// Suppressor represents a class of work and forms a namespace in
// which units of work can be executed with duplicate suppression.
type Suppressor struct {
	m  map[string]*call
	mu sync.Mutex
}

func NewSuppressor() *Suppressor {
	return &Suppressor{
		m: make(map[string]*call),
	}
}

// Result holds the results of Do, so they can be passed
// on a channel.
type Result struct {
	Val interface{}
	Err error
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
// The return a channel that will receive the
// results when they are ready.
//
// The returned channel will not be closed.
func (g *Suppressor) Do(key string, expiration time.Duration, fn func() (interface{}, error)) <-chan Result {
	ch := make(chan Result, 1)
	if c := g.get(key); c != nil {
		c.mu.Lock()
		c.chans = append(c.chans, ch)
		c.mu.Unlock()
		return ch
	}
	c := &call{
		wg:         NewWaitGroup(),
		chans:      []chan<- Result{ch},
		expiration: time.Now().Add(expiration),
	}
	c.wg.Add(1)
	g.put(key, c)

	go g.doCall(c, key, fn)

	return ch
}

// doCall handles the single call for a key.
func (g *Suppressor) doCall(c *call, key string, fn func() (interface{}, error)) {
	defer c.wg.Done()

	c.val, c.err = fn()
	c.mu.Lock()
	for _, ch := range c.chans {
		ch <- Result{c.val, c.err}
	}
	c.chans = c.chans[:0]
	c.mu.Unlock()
}

func (g *Suppressor) get(key string) *call {
	g.mu.Lock()
	defer g.mu.Unlock()

	c, ok := g.m[key]
	if !ok {
		return nil
	}

	if c.expiration.Before(time.Now()) {
		delete(g.m, key)
		return nil
	}

	return c
}

func (g *Suppressor) put(key string, c *call) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.m[key] = c
}
