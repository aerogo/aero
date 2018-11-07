package aero

// Event represents a single event in an event stream.
type Event struct {
	Name string
	Data interface{}
}
