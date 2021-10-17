package synx

import (
	"runtime"
	"sync/atomic"
	"time"
)

const (
	unlocked = int32(0)
	locked   = -1
)

// Spinlock is a spinlock-based implementation of lock.
type Spinlock struct {
	state int32
}

// TryLock performs a non-blocking attempt to acquire the locker and
// returns true if successful.
func (l *Spinlock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&l.state, unlocked, locked)
}

// Lock waits until the locker with be released (if it is not) and then acquire it.
func (l *Spinlock) Lock() {
	wait(l.TryLock)
}

// Unlock relases the locker.
func (l *Spinlock) Unlock() {
	if !atomic.CompareAndSwapInt32(&l.state, locked, unlocked) {
		panic(`Unlock()-ing non-locked locker`)
	}
}

func wait(slowFn func() bool) {
	for i := 0; !slowFn(); {
		if i < 2 {
			time.Sleep(50 * time.Microsecond)
			i++
			continue
		}

		// NOTE: if after trying a short timeout,
		// it was not possible to take true,
		// then release the scheduler resources.
		runtime.Gosched()
	}
}
