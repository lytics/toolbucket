package ctxwaitgroup

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func TestWgEmpty(t *testing.T) {
	wg := NewWaitGroup()
	e := wg.Wait(context.Background())
	assert.Tf(t, e == nil, "")
}

func TestWgNegErr(t *testing.T) {
	wg := NewWaitGroup()
	e := wg.Inc()
	assert.Tf(t, e == nil, "got: %v", e)
	e = wg.Dec()
	assert.Tf(t, e == nil, "got: %v", e)
	e = wg.Dec()
	assert.Tf(t, e == ErrNegWg, "got: %v", e)
}

func TestWgOne(t *testing.T) {
	wg := NewWaitGroup()
	e := wg.Inc()
	assert.Tf(t, e == nil, "")

	st := time.Now()
	time.AfterFunc(200*time.Millisecond, func() { wg.Dec() })
	e = wg.Wait(context.Background())
	assert.Tf(t, e == nil, "")
	assert.Tf(t, time.Since(st) >= (200*time.Millisecond), "")
}

func TestWgTen(t *testing.T) {
	wg := NewWaitGroup()
	for i := 0; i < 10; i++ {
		e := wg.Inc()
		assert.Tf(t, e == nil, "")
	}

	decCnt := int64(0)

	st := time.Now()
	time.AfterFunc(200*time.Millisecond, func() {
		for i := 0; i < 10; i++ {
			e := wg.Dec()
			assert.Tf(t, e == nil, "")
			atomic.AddInt64(&decCnt, 1)
		}
	})

	e := wg.Wait(context.Background())
	assert.Tf(t, e == nil, "")
	assert.Tf(t, time.Since(st) >= (200*time.Millisecond), "")
	c := atomic.LoadInt64(&decCnt)
	assert.Tf(t, c == 10, " got: %v", c)
}

func TestWgIncWhileWaiting(t *testing.T) {
	wg := NewWaitGroup()
	for i := 0; i < 10; i++ {
		e := wg.Inc()
		assert.Tf(t, e == nil, "")
	}

	decCnt := int64(0)

	st := time.Now()
	time.AfterFunc(200*time.Millisecond, func() {
		for i := 0; i < 9; i++ {
			e := wg.Dec()
			assert.Tf(t, e == nil, "")
			atomic.AddInt64(&decCnt, 1)
		}
	})
	time.AfterFunc(300*time.Millisecond, func() {
		e := wg.Inc()
		assert.Tf(t, e == ErrClosed, "got: %v", e)

		e = wg.Dec()
		assert.Tf(t, e == nil, "")
		atomic.AddInt64(&decCnt, 1)
	})
	e := wg.Wait(context.Background())
	assert.Tf(t, e == nil, "")
	assert.Tf(t, time.Since(st) >= (200*time.Millisecond), "")
	c := atomic.LoadInt64(&decCnt)
	assert.Tf(t, c == 10, " got: %v", c)
}

func TestWgCtxTimeout(t *testing.T) {
	exited := make(chan struct{})
	defer close(exited)

	wg := NewWaitGroup()
	e := wg.Inc()
	assert.Tf(t, e == nil, "")

	st := time.Now()

	ctx, can := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer can()
	time.AfterFunc(800*time.Millisecond, func() {
		select {
		case <-exited:
			return
		default:
			t.Fatalf("testcase timed out")
		}
	})

	e = wg.Wait(ctx)
	assert.Tf(t, e == context.DeadlineExceeded, "err: %v", e)
	assert.Tf(t, time.Since(st) >= (200*time.Millisecond), "got: %v", time.Since(st))
}
