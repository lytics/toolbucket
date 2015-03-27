#### OrderedTaskPool

The primary use-case for this package would be to insert tasks in order, have them
processed in parallel and have the results consumed in the original order.

Events are emitted ascending index order.   i.e. 1, 2, 3,...,N

###### An example usecase:

You need to read billions of JSON messages out of a Kafka partition, unmarshall the JSON
into a struct and maintain the order of messages.

JSON unmarshalling is very slow and could add days to your processing time, so you want to
unmarshall the data in parallel.  But after you unmarshall the messages, you have to
put them back into the original order they were read from kafka.

Solution use OrderedTaskPool, the Task's index will be the kafka-offset and the Task's input
will be the message bytes.   The Task's result will be the unmarshalled struct.  For you
processor func(),  write code to do the JSON unmarshalling

###### diagram 

```

	        |                            |
	        |                            |
	        |     ---->  worker() --\    |
	        |    /                   \   |
Enqueue()--in----------->  worker() ------out----> Results()
	        |    \                   /   |
	        |     ----> worker() ---/    |
	        |                            |

```
