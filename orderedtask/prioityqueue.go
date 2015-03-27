package orderedtask

import "container/heap"

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
