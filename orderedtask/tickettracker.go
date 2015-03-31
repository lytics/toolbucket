package orderedtask

type TicketBox struct {
	freetickets chan bool
}

func NewTicketBox(size int) *TicketBox {
	freetickets := make(chan bool, size+1)

	//fillup the ticketbox with tickets
	for i := 0; i < size; i++ {
		freetickets <- true
	}
	return &TicketBox{freetickets}
}

func (t *TicketBox) Tickets() <-chan bool {
	return t.freetickets
}

func (t *TicketBox) ReturnTicket() {
	t.freetickets <- true
}
