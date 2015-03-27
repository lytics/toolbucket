package orderedtask

/*
orderedtask.Pool

The primary use-case for this package would be to insert tasks in order, have them
processed in parallel and have the results consumed in the original order.

If a tasks is going to take 10ms and you have 10,000 tasks to perform, then the expected runtime
to perform those tasks one at a time is: `10000 * 10ms == 100 seconds`.  But if you can do them
parallel with 12 workers, the expect runtime becomes `(10000 * 10ms) / 12 == 8.3 seconds`.

Notes:

All Tasks are emitted in ascending index order.   i.e. 1, 2, 3,...,N

Internally were using two MinHeaps, one for the lowest task inserted into the pool and
one for the lowest index of finished tasks.  While the finish tasks index matches the
lowest inserted task, we'll emit the results from the finished tasks.  If the lowest
index for finished tasks is not equal to the lowest inserted task's index, then we wait for
the workers who are processing the lowest indexes to finish before draining.


	        |                            |
	        |                            |
	        |     ---->  worker() --\    |
	        |    /                   \   |
Enqueue()--in----------->  worker() ------out----> Results()
	        |    \                   /   |
	        |     ----> worker() ---/    |
	        |                            |



*/
