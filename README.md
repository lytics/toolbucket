# toolbucket
A selection of small Go tool kits for anyone to use.

## tools kits:

### [ctxwaitgroup](./ctxwaitgroup)

This is a waitgroup which supports timing/canceling out via a context.

### [orderedtask.Pool](./orderedtask)

Note this code is **Experimental and not production ready**

The primary use-case for this package would be to maintain order 
in a stream of events, but in an intermediate stage have some processing done in parallel.
For example, while processing from a queue one may want to unmarshal in parallel but then 
write the records to a database in order. 


