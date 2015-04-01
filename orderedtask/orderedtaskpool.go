package orderedtask

// import github.com/lytics/toolbucket/orderedtask

import (
	"sync"
	"time"
)

type Task struct {
	Index  uint64
	Input  interface{}
	Output interface{}
}

/*

Pool is the main struct for everything in this package.  Its job is to
coordinate tasks and workers and to ensure emitted events are in index order with
regards to the the task.Index of all enqueued tasks.

*/
type Pool struct {
	//data flow
	finishedtaskheap *TaskHeap
	lowwatermark     *LowWatermark
	semaphore        *TicketDispenser

	//Sync Code
	abort chan bool
	in    chan *Task
	out   chan *Task

	closeonce *sync.Once
	lock      *sync.Mutex
}

func (ms *Pool) Results() <-chan *Task {
	return ms.out
}

func (ms *Pool) Enqueue(m *Task) {
	ms.lock.Lock()
	ms.lowwatermark.Enqueue(m.Index)
	ms.lock.Unlock()

	ms.in <- m
}

func (ms *Pool) enqueueAndDrain(t *Task) {
	//emit tasks if the lowest index is ready.
	ms.lock.Lock()
	defer ms.lock.Unlock()

	ms.finishedtaskheap.Enqueue(t)

	for t, ok := ms.finishedtaskheap.Peek(); ok; t, ok = ms.finishedtaskheap.Peek() {
		minFinished := t
		minInPool, _ := ms.lowwatermark.Peek()

		if minFinished.Index == minInPool {
			//we have to release the lock incase the out chan is full
			// to allow others to be able to enqueue
			if len(ms.out) < cap(ms.out) {
				task := ms.finishedtaskheap.Dequeue()
				ms.lowwatermark.Dequeue()
				ms.out <- task
			} else {
				time.Sleep(time.Microsecond * 100)
			}
		} else {
			//someone else is still working on the lowest indexed task, let them
			//worry about draining the finished-task's minheap (aka priority queue)
			return
		}
	}
}

func (ms *Pool) Close() {
	ms.closeonce.Do(func() {
		close(ms.abort)
	})
}

func (ms *Pool) AquireTicket() <-chan bool {
	return ms.semaphore.Tickets()
}

func (ms *Pool) ReleaseTicket() {
	ms.semaphore.ReleaseTicket()
}

func NewPool(poolsize int, processor func(map[string]interface{}, *Task)) *Pool {
	abort := make(chan bool)
	in := make(chan *Task, poolsize+1)
	out := make(chan *Task, poolsize+1)

	ms := &Pool{
		finishedtaskheap: NewTaskHeap(),
		lowwatermark:     NewLowWatermark(),
		abort:            abort,
		in:               in,
		out:              out,
		closeonce:        &sync.Once{},
		lock:             &sync.Mutex{},
		//counting semaphore to protect deadlocking when the pool is used in an RPC manner
		//  i.e. read/write from the same go func.
		//  See: TestPoolRPC
		semaphore: NewTicketDispenser(cap(in) - 1),
	}

	//start up worker pool
	for i := 0; i < poolsize; i++ {
		go func() {
			var workerlocal map[string]interface{} = make(map[string]interface{}, 1)
			for {
				select {
				case t := <-ms.in:
					processor(workerlocal, t)
					ms.enqueueAndDrain(t)
				case <-ms.abort:
					return
				}
			}
		}()
	}

	return ms
}
