package event

// Stream includes a channel of events that we can send to
// and a closed channel that we can check for closed connections.
type Stream struct {
	Events chan *Event
	Closed chan struct{}
}

// NewStream creates a new event stream.
func NewStream() *Stream {
	return &Stream{
		Events: make(chan *Event),
		Closed: make(chan struct{}),
	}
}
