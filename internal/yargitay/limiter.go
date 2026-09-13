// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"context"
	"sync"
	"time"
)

type waiter struct{ ready chan struct{} }
type limiter struct {
	mu       sync.Mutex
	interval time.Duration
	maxQueue int
	queue    []*waiter
	next     time.Time
}

var processQueue = &limiter{interval: 3 * time.Second, maxQueue: 100}

func sharedLimiter(c Config) *limiter {
	processQueue.mu.Lock()
	defer processQueue.mu.Unlock()
	processQueue.interval = max(processQueue.interval, c.Interval)
	processQueue.maxQueue = min(processQueue.maxQueue, c.MaxQueue)
	return processQueue
}
func (l *limiter) deferFor(d time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if until := time.Now().Add(d); until.After(l.next) {
		l.next = until
	}
}
func (l *limiter) acquire(ctx context.Context) (func(), error) {
	w := &waiter{ready: make(chan struct{})}
	l.mu.Lock()
	if len(l.queue) >= l.maxQueue {
		l.mu.Unlock()
		return nil, ErrCapacity
	}
	l.queue = append(l.queue, w)
	if len(l.queue) == 1 {
		close(w.ready)
	}
	l.mu.Unlock()
	var once sync.Once
	release := func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			for i, q := range l.queue {
				if q == w {
					l.queue = append(l.queue[:i], l.queue[i+1:]...)
					if i == 0 && len(l.queue) > 0 {
						close(l.queue[0].ready)
					}
					return
				}
			}
		})
	}
	select {
	case <-ctx.Done():
		release()
		return nil, ErrUnavailable
	case <-w.ready:
	}
	for {
		if ctx.Err() != nil {
			release()
			return nil, ErrUnavailable
		}
		l.mu.Lock()
		delay := time.Until(l.next)
		if delay <= 0 {
			l.next = time.Now().Add(l.interval)
			l.mu.Unlock()
			return release, nil
		}
		l.mu.Unlock()
		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			t.Stop()
			release()
			return nil, ErrUnavailable
		case <-t.C:
		}
	}
}
