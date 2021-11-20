package event

// Event represents a single event in an event stream.
type Event struct {
	Name string
	Data interface{}
}

// New creates a new event
func New(name string, data interface{}) *Event {
	return &Event{
		Name: name,
		Data: data,
	}
}
