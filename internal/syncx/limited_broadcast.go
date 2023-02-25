package syncx

import (
	"context"
	"fmt"
	"sync"
)

type LimitedBroadcast struct {
	notify     chan struct{}
	wait       bool
	maxWaiting int
	waiting    int
	lock       *sync.RWMutex
}

func NewLimitedBroadcast(maxWaiting int) *LimitedBroadcast {
	return &LimitedBroadcast{
		lock:       &sync.RWMutex{},
		maxWaiting: maxWaiting,
		notify:     make(chan struct{}),
	}
}
func (l *LimitedBroadcast) StartWait() {
	l.lock.Lock()
	close(l.notify) // flush any waiters
	l.notify = make(chan struct{})
	l.wait = true
	l.lock.Unlock()
}
func (l *LimitedBroadcast) Signal() {
	l.lock.Lock()
	close(l.notify)
	l.wait = false
	l.lock.Unlock()
}

func (l *LimitedBroadcast) Wait(ctx context.Context) error {
	l.lock.RLock()
	if !l.wait {
		l.lock.RUnlock()
		return nil
	}

	if l.waiting >= l.maxWaiting {
		l.lock.RUnlock()
		return fmt.Errorf("too many requests waiting")
	}
	l.lock.RUnlock()
	l.lock.Lock()
	l.waiting++
	l.lock.Unlock()
	select {
	case <-ctx.Done():
		l.lock.Lock()
		l.waiting--
		l.lock.Unlock()
		return nil
	case <-l.notify:
		return nil
	}
}
