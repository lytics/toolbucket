package orderedtask

import "container/heap"

//
// Uint64Heap is a min heap used by LowerWatermark to track the lowwatermark of incoming tasks ids.
//
type Uint64Heap struct {
	arr []uint64
}

func (h *Uint64Heap) Len() int           { return len(h.arr) }
func (h *Uint64Heap) Less(i, j int) bool { return h.arr[i] < h.arr[j] }
func (h *Uint64Heap) Swap(i, j int)      { h.arr[i], h.arr[j] = h.arr[j], h.arr[i] }

func (h *Uint64Heap) Push(x interface{}) {
	h.arr = append(h.arr, x.(uint64))
}

func (h *Uint64Heap) Pop() interface{} {
	old := h.arr
	n := len(old)
	x := old[n-1]
	h.arr = old[0 : n-1]
	return x
}

func (h *Uint64Heap) Peek() (uint64, bool) {
	if h.Len() == 0 {
		return ^uint64(0), false
	}
	return h.arr[0], true
}

func (h *Uint64Heap) Enqueue(mark uint64) {
	heap.Push(h, mark)
}

func (h *Uint64Heap) Dequeue() uint64 {
	return heap.Pop(h).(uint64)
}

func NewUint64Heap() *Uint64Heap {
	minheap := &Uint64Heap{[]uint64{}}
	heap.Init(minheap)
	return minheap
}

//
// LowWatermark keeps track of the low watermark of incoming tasks ids.  So
// that we only release tasks if they are in order.
//
type LowWatermark struct {
	minheap *Uint64Heap
}

func (lh LowWatermark) Peek() (uint64, bool) {
	return lh.minheap.Peek()
}

func (lh LowWatermark) Enqueue(mark uint64) {
	lh.minheap.Enqueue(mark)
}

func (lh LowWatermark) Dequeue() uint64 {
	return lh.minheap.Dequeue()
}

func NewLowWatermark() *LowWatermark {
	lh := &LowWatermark{
		minheap: NewUint64Heap(),
	}
	return lh
}
