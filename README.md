# toolbucket
A selection of small Go tool kits for anyone to use.

## tools:

### [orderedtask.Pool](./orderedtask)
The primary use-case for this package would be to insert tasks in order, have them
processed in parallel and have the results consumed in the original order.  

If a task is going to take about 10ms to process and you have 10,000 tasks to perform, then the expected runtime
to perform all the tasks is about `10000 * 10ms == 100 seconds`.  But if you can do them 
parallel with 12 workers, the expect runtime becomes `(10000 * 10ms) / 12 == 8.3 seconds`.

