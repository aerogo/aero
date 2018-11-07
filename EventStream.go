package aero

// EventStream includes a channel of events that we can send to
// and a closed channel that we can check for closed connections.
type EventStream struct {
	Events chan *Event
	Closed chan struct{}
}

// NewEventStream creates a new event stream.
func NewEventStream() *EventStream {
	return &EventStream{
		Events: make(chan *Event),
		Closed: make(chan struct{}),
	}
}
