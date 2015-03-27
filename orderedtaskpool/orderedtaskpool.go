package orderedtaskpool

import (
	"container/heap"
	"sync"
)

type Task struct {
	Index  uint64
	Input  interface{}
	Output interface{}
}

//
// Finished Task Pool
//
type TaskHeap struct {
	tasks  []*Task
	sortBy func(*Task, *Task) bool
}

var AscendingTaskOrder func(*Task, *Task) bool = func(p1, p2 *Task) bool {
	return (p1.Index < p2.Index)
}
var DecendingTaskOrder func(*Task, *Task) bool = func(p1, p2 *Task) bool {
	return !(p1.Index < p2.Index)
}

func (th *TaskHeap) Len() int           { return len(th.tasks) }
func (th *TaskHeap) Less(i, j int) bool { return th.sortBy(th.tasks[i], th.tasks[j]) }
func (th *TaskHeap) Swap(i, j int)      { th.tasks[i], th.tasks[j] = th.tasks[j], th.tasks[i] }

func (th *TaskHeap) Push(x interface{}) {
	th.tasks = append(th.tasks, x.(*Task))
}

func (th *TaskHeap) Pop() interface{} {
	old := th.tasks
	tail := old[len(old)-1]
	th.tasks = old[0 : len(old)-1]
	return tail
}

func (th *TaskHeap) Next() bool {
	return th.Len() > 0
}

func (th *TaskHeap) Peek() (*Task, bool) {
	if len(th.tasks) == 0 {
		return nil, false
	}
	return th.tasks[0], true
}

func (th *TaskHeap) Dequeue() *Task {
	return heap.Pop(th).(*Task)
}

func (th *TaskHeap) Enqueue(t *Task) {
	heap.Push(th, t)
}

func NewTaskHeap() *TaskHeap {
	taskarray := []*Task{}
	th := &TaskHeap{taskarray, AscendingTaskOrder}
	heap.Init(th)
	return th
}

/*

OrderedTaskPool is the main struct for everything in this package.  Its job is to
coordinate tasks and workers and to ensure emitted events are in index order with
regards to the the task.Index of all enqueued tasks.

*/
type OrderedTaskPool struct {
	//data flow
	finishedtaskheap *TaskHeap
	lowwatermark     *LowWatermark

	//Sync Code
	abort chan bool
	in    chan *Task
	out   chan *Task

	closeonce *sync.Once
	lock      *sync.Mutex
}

func (ms *OrderedTaskPool) Results() <-chan *Task {
	return ms.out
}

func (ms *OrderedTaskPool) Enqueue(m *Task) {
	ms.lock.Lock()
	ms.lowwatermark.Enqueue(m.Index)
	ms.lock.Unlock()

	ms.in <- m
}

func (ms *OrderedTaskPool) enqueueAndDrain(t *Task) {
	//emit tasks if the lowest index is ready.
	ms.lock.Lock()
	defer ms.lock.Unlock()

	ms.finishedtaskheap.Enqueue(t)

	for ok := true; ok; t, ok = ms.finishedtaskheap.Peek() {
		minFinished := t
		minInPool, _ := ms.lowwatermark.Peek()

		if minFinished.Index == minInPool {
			task := ms.finishedtaskheap.Dequeue()
			ms.lowwatermark.Dequeue()

			ms.out <- task
		} else {
			//someone else is still working on the lowest indexed task, let them
			//worry about draining the finished-task's minheap (aka priority queue)
			return
		}
	}
}

func (ms *OrderedTaskPool) Close() {
	ms.closeonce.Do(func() {
		close(ms.abort)
	})
}

func NewOrderedTaskPool(poolsize int, processor func(map[string]interface{}, *Task)) *OrderedTaskPool {
	abort := make(chan bool)
	in := make(chan *Task, 5)
	out := make(chan *Task, 5)
	ms := &OrderedTaskPool{
		finishedtaskheap: NewTaskHeap(),
		lowwatermark:     NewLowWatermark(),
		abort:            abort,
		in:               in,
		out:              out,
		closeonce:        &sync.Once{},
		lock:             &sync.Mutex{},
	}

	//start up worker pool
	for i := 0; i < poolsize; i++ {
		go func() { //start pool worker
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
