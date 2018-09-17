package ctxwaitgroup

import (
	"context"
	"fmt"
	"sync"
)

var ErrClosed = fmt.Errorf("waitgroup closed")
var ErrNegWg = fmt.Errorf("neg waitgroup")

type WaitGroup struct {
	cond   *sync.Cond
	closed bool

	waiters int
}

func NewWaitGroup() *WaitGroup {
	cond := sync.NewCond(&sync.Mutex{})
	return &WaitGroup{
		cond:   cond,
		closed: false,
	}
}

func (wg *WaitGroup) Inc() bool {
	wg.cond.L.Lock()
	defer wg.cond.L.Unlock()

	if wg.closed == true {
		return false
	}
	wg.waiters++
	return true
}

func (wg *WaitGroup) Dec() error {
	wg.cond.L.Lock()
	defer wg.cond.L.Unlock()
	wg.waiters--
	if wg.waiters < 0 {
		wg.cond.Signal()
		return ErrNegWg
	}
	wg.cond.Signal()
	return nil
}

func (wg *WaitGroup) Wait(ctx context.Context) error {
	exited := make(chan struct{})
	defer close(exited)

	wg.cond.L.Lock()
	defer wg.cond.L.Unlock()

	go func() {
		select {
		case <-ctx.Done():
		case <-exited:
			return
		}
		wg.cond.L.Lock()
		defer wg.cond.L.Unlock()
		wg.cond.Signal() // after the context is Done, we'll signal the Wait loop so it can wake up and check the ctx
	}()

	wg.closed = true
	for {
		if wg.waiters == 0 {
			return nil
		} else if wg.waiters < 0 {
			return ErrNegWg
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			//release the lock and wait until signaled.  On awake we'll require the lock.
			// After wait requires the lock we have to recheck the wait condition
			// (calling wg.waiters).
			wg.cond.Wait()
		}
	}
}
