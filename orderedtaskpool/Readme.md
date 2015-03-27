#### OrderedTaskPool

The primary use-case for this package would be to insert tasks in order, have them
processed in parallel and have the results consumed in the original order.

Events are emitted ascending index order.   i.e. 1, 2, 3,...,N

#### An example usecase:

You need to read billions of JSON messages out of a Kafka partition, unmarshall the JSON
into a struct in parallel and maintain the order of messages after they're unmarshalled.

JSON unmarshalling is very slow and could add days to your processing time, so you want to
unmarshall the data in parallel.  But after you unmarshall the messages, you have to
put them back into the original order they were read from kafka.

Solution use OrderedTaskPool, the Task's index will be the kafka-offset and the Task's input
will be the message bytes.   The Task's result will be the unmarshalled struct.  For you
processor func(),  write code to do the JSON unmarshalling

#### diagram 

```
	        |                            |
	        |     ---->  worker() --\    |
	        |    /                   \   |
Enqueue()--in----------->  worker() ------out----> Results()
	        |    \                   /   |
	        |     ----> worker() ---/    |
	        |                            |
```


#### code example:

structs

```go
	type Visitor struct {
		Name      string
		ClickLink string
		VisitTime int64
	}
	type Msg struct {
		Offset uint64
		Body   []bytes
	}

```
Example using of the pull

```go
	const PoolSize = 16

	//Create the pool with PoolSize workers
	pool := NewOrderedTaskPool(PoolSize, func(workerlocal map[string]interface{}, t *Task) {
		//For some serialization frameworks you get better performance if you can reuse buffers in a thread-safe manor.
		// workerlocal is used for storing state for each go routine worker.
		/*
			var buf bytes.Buffer
			if b, ok := workerlocal["buf"]; !ok {
				//worker local is not shared between go routines, so it's a good place to place a store a reusable items like buffers
				buf = bytes.Buffer{}
				workerlocal["buf"] = buf
			} else {
				buf = b.(bytes.Buffer)
			}
			buf.Reset()
		*/

		var visit Visitor
		msg := t.Input.(*Msg)
		err := json.Unmarshal(msg.Body, &vistor)
		if err != nil {
			fmt.Println("error:", err)
		}

		t.Output = visit
	})
	defer pool.Close()

	wg := &sync.WaitGroup{}

	//Produce messages from kafka, into the pool to be unmarshalled
	// and then Consume messages from the pool as structs
	wg.Add(1)
	go func() {
		defer wg.Done()

		kafkaconsumer := createKafkaConsumer() //for example this could be a https://github.com/Shopify/sarama consumer, reading messages from kafka8.

		for {
			select {
			case event := <-kafkaconsumer.Events():
				if event.Err != nil {
					log.Printf("error: consumer: %v", event.Err)
					continue
				}
				//Take the events in from kafka and pass them off to the pool to be unmarshal'ed
				pool.Enqueue(&Task{Index: event.Offset, Input: &Msg{event.Offset, event.Message}})
			case res := <-pool.Results():
				vis := t.Output.(*Visitor)
				fmt.Printf("visitor: off:%v text:%v \n", vis.Name, vis.ClickLink, vis.VisitTime)
				//process the unmarshalled Visitor struct
			}
		}
	}()
```






