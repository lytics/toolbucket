package orderedtask

type TicketDispenser struct {
	freetickets chan bool
}

func NewTicketDispenser(size int) *TicketDispenser {
	freetickets := make(chan bool, size)

	//fillup the ticketbox with tickets
	for i := 0; i < size; i++ {
		freetickets <- true
	}
	return &TicketDispenser{freetickets}
}

func (t *TicketDispenser) Tickets() <-chan bool {
	return t.freetickets
}

func (t *TicketDispenser) ReleaseTicket() {
	t.freetickets <- true
}
