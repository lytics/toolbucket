#### orderedtask.Pool 

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

###### diagram 

```
	        |                            |
	        |     ---->  worker() --\    |
	        |    /                   \   |
Enqueue()--in----------->  worker() ------out----> Results()
	        |    \                   /   |
	        |     ----> worker() ---/    |
	        |                            |
```



#### An example usecase:

You need to read billions of JSON messages out of a Kafka partition, unmarshall the JSON
into a struct and maintain the order of messages.

JSON unmarshalling is very slow at this scale and could add days to your processing time, so you want
to unmarshall the data in parallel.  But sometimes you need those unmarshall messages, in the
original order they arrived in.

The orderedtask.Pool can do this.  The Task's index in this case would be the kafka offset and the Task's input
would be the message.Body() `[]bytes`.   The Task's result would be the unmarshalled struct.  For your
processor func(),  write code to do the JSON unmarshalling into a struct.


##### Code Examples:
import line: `import "github.com/lytics/toolbucket/orderedtask"`

Working examples of both examples can be found in `orderedtaskpool_test.go`.  

###### Producer / Consumer example 

```go
	type Visit struct {
		Name      string
		ClickLink string
		VisitTime int64
	}
	type Msg struct {
		Offset uint64
		Body   []bytes
	}

	const PoolSize = 16

	func ProcessMessages() {
		//Create the pool with PoolSize workers
		pool := orderedtask.NewPool(PoolSize, func(workerlocal map[string]interface{}, t *orderedtask.Task) {
			// workerlocal is used for storing go routine local state that isn't shared between workers.
			//   i.e. if you need to reuse a buffer between calls to the function.
			/*
				var buf bytes.Buffer
				// Checking if "buf" is created.
				// we only want to create the buffer once on the first call to this worker!
				if b, ok := workerlocal["buf"]; !ok {
					buf = bytes.Buffer{}
					workerlocal["buf"] = buf
				} else {
					buf = b.(bytes.Buffer)
				}
				buf.Reset()
			*/

			var v Visit
			msg := t.Input.(*Msg) //Task.Input is anything you want processed in the pull.
			err := json.Unmarshal(msg.Body, &v)
			if err != nil {
				fmt.Println("error:", err)
			}

			t.Output = v //Task.Output is were you write the results from the task.
		})
		defer pool.Close() //Closing the pool shuts down all the workers.

		//here we consume messages from kafka, insert them into the pool to be unmarshalled
		//and then consume the messages from the pool as structs

		go func() {
			kafkaconsumer := createKafkaConsumer() //for example this could be a https://github.com/Shopify/sarama consumer, reading messages from kafka8.
			defer kafkaconsumer.Close()
			for {
				select {
				case event := <-kafkaconsumer.Events():
					if event.Err != nil {
						log.Printf("error: consumer: %v", event.Err)
						continue
					}
					//Take the events in from kafka and pass them off to the pool to be unmarshal'ed
					pool.Enqueue(&orderedtask.Task{Index: event.Offset, Input: &Msg{event.Offset, event.Message}})
				}
			}
		}()

		for {
			select {
			case res := <-pool.Results():
				//The results will arrive here in the same order they were Enqueued.  
				vis := t.Output.(*Visit)
				fmt.Printf("Visit: name:%v click:%v ts:%v \n", vis.Name, vis.ClickLink, vis.VisitTime)
				//process the unmarshalled Visit struct
			}
		}
	}
```








