package orderedtask

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

type Msg struct {
	text   string
	offset uint64
}

func TestStreamingMsgsThroughThePool(test *testing.T) {
	runtime.GOMAXPROCS(2)
	const MsgCnt = 5000
	const PoolSize = 12

	//Create the pool with PoolSize workers
	//  Note: if you lower the pool size, the total runtime gets longer due to the
	//       lack of parallelization.
	//
	pool := NewPool(PoolSize, func(workerlocal map[string]interface{}, t *Task) {
		var buf bytes.Buffer
		if b, ok := workerlocal["buf"]; !ok {
			//worker local is not shared between go routines, so it's a good place to place a store a reusable items like buffers
			buf = bytes.Buffer{}
			workerlocal["buf"] = buf
		} else {
			buf = b.(bytes.Buffer)
		}
		buf.Reset()

		msg := t.Input.(*Msg)
		amt := time.Duration(rand.Intn(5))
		time.Sleep(time.Millisecond * amt) // long running operation
		t.Output = msg
	})

	wg := &sync.WaitGroup{}

	//Consume messages from the pool
	//  Note: they should be in order by the offset
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		expectedoffset := uint64(0)
		for t := range pool.Results() {
			msg := t.Output.(*Msg)
			i++
			if i == MsgCnt-1 {
				return
			} else if msg.offset != expectedoffset {
				test.Fatalf("the offsets weren't in order: got:%d expected:%d", msg.offset, expectedoffset)
			}
			expectedoffset++
		}
	}()

	//Produce messages into the pool
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < MsgCnt; i++ {
			m := &Msg{fmt.Sprintf("{'foo':'%d'}", i), uint64(i)}
			pool.Enqueue(&Task{Index: m.offset, Input: m})
		}
	}()

	wg.Wait()
}

func TestCallerUsesPoolForRPC(test *testing.T) {
	runtime.GOMAXPROCS(2)
	const MsgCnt = 9333
	const PoolSize = 12

	msgchan := make(chan *Task, MsgCnt)

	pool := NewPool(PoolSize, func(workerlocal map[string]interface{}, t *Task) {
		msg := t.Input.(*Msg)
		amt := time.Duration(rand.Intn(9))
		time.Sleep(time.Millisecond * amt)
		t.Output = msg
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < MsgCnt; i++ {
			m := &Msg{fmt.Sprintf("{'foo':'%d'}", i), uint64(i)}
			msgchan <- &Task{Index: m.offset, Input: m}
		}
		close(msgchan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		expectedoffset := uint64(0)

		ticketbox := pool.TicketDispenser()

		for {
			select {
			case <-ticketbox.Tickets():
				t, ok := <-msgchan
				if ok {
					pool.Enqueue(t)
				}
			case t := <-pool.Results():
				ticketbox.ReleaseTicket()

				msg := t.Output.(*Msg)
				i++
				if i == MsgCnt-1 {
					return
				} else if msg.offset != expectedoffset {
					test.Fatalf("out of order: got:%d expected:%d", msg.offset, expectedoffset)
				}
				expectedoffset++
			}
		}
	}()

	wg.Wait()
}

func TestSlowConsumers(test *testing.T) {

	const MsgCnt = 333
	const PoolSize = 2

	msgchan := make(chan *Task, MsgCnt)

	pool := NewPool(PoolSize, func(workerlocal map[string]interface{}, t *Task) {
		msg := t.Input.(*Msg)
		amt := time.Duration(rand.Intn(3))
		time.Sleep(time.Millisecond * amt)
		t.Output = msg
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < MsgCnt; i++ {
			m := &Msg{fmt.Sprintf("{'foo':'%d'}", i), uint64(i)}
			msgchan <- &Task{Index: m.offset, Input: m}
		}
		close(msgchan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		expectedoffset := uint64(0)

		ticketbox := pool.TicketDispenser()

		for {
			select {
			case <-ticketbox.Tickets():
				t, ok := <-msgchan
				if ok {
					pool.Enqueue(t)
				}
			case t := <-pool.Results():
				ticketbox.ReleaseTicket()

				msg := t.Output.(*Msg)
				i++
				if i == MsgCnt-1 {
					return
				} else if msg.offset != expectedoffset {
					test.Fatalf("out of order: got:%d expected:%d", msg.offset, expectedoffset)
				}
				expectedoffset++
				amt := time.Duration(rand.Intn(5))
				time.Sleep(time.Millisecond * amt)
			}
		}
	}()

	wg.Wait()
}

func TestSlowProducers(test *testing.T) {

	const MsgCnt = 633
	const PoolSize = 3

	pool := NewPool(PoolSize, func(workerlocal map[string]interface{}, t *Task) {
		msg := t.Input.(*Msg)
		amt := time.Duration(rand.Intn(10))
		time.Sleep(time.Millisecond * amt)
		t.Output = msg
	})

	msgchan := make(chan *Task, MsgCnt)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < MsgCnt; i++ {
			m := &Msg{fmt.Sprintf("{'foo':'%d'}", i), uint64(i)}
			msgchan <- &Task{Index: m.offset, Input: m}
		}
		close(msgchan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		expectedoffset := uint64(0)

		ticketbox := pool.TicketDispenser()

		for {
			select {
			case <-ticketbox.Tickets():
				t, ok := <-msgchan
				if ok {
					pool.Enqueue(t)
				}
				amt := time.Duration(rand.Intn(15))
				time.Sleep(time.Millisecond * amt)
			case t := <-pool.Results():
				ticketbox.ReleaseTicket()
				msg := t.Output.(*Msg)

				i++
				if i == MsgCnt-1 {
					return
				} else if msg.offset != expectedoffset {
					test.Fatalf("the offsets weren't in order: got:%d expected:%d", msg.offset, expectedoffset)
				}
				expectedoffset++
			}
		}
	}()

	wg.Wait()
}

func TestFastWorkers(test *testing.T) {
	runtime.GOMAXPROCS(6)

	const MsgCnt = 155552
	const PoolSize = 555

	pool := NewPool(PoolSize, func(workerlocal map[string]interface{}, t *Task) {
		msg := t.Input.(*Msg)
		t.Output = msg
	})

	wg := &sync.WaitGroup{}

	//Produce messages into the pool
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < MsgCnt; i++ {
			m := &Msg{fmt.Sprintf("{'foo':'%d'}", i), uint64(i)}
			pool.Enqueue(&Task{Index: m.offset, Input: m})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		expectedoffset := uint64(0)
		for t := range pool.Results() {
			msg := t.Output.(*Msg)

			i++
			if i == MsgCnt-1 {
				return
			} else if msg.offset != expectedoffset {
				test.Fatalf("the offsets weren't in order: got:%d expected:%d", msg.offset, expectedoffset)
			}
			expectedoffset++
		}
	}()

	wg.Wait()
}
