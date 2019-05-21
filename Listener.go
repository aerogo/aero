package aero

import (
	"net"
	"time"
)

const keepAlivePeriod = 3 * time.Minute

// Listener sets TCP keep-alive timeouts on accepted connections.
type Listener struct {
	*net.TCPListener
}

// Accept accepts incoming connections.
func (listener Listener) Accept() (net.Conn, error) {
	// Accept a new client.
	connection, err := listener.AcceptTCP()

	if err != nil {
		return nil, err
	}

	// Set keep-alive values based on the default values
	// found in the net/http TCP listener implementation.
	err = connection.SetKeepAlive(true)

	if err != nil {
		return nil, err
	}

	err = connection.SetKeepAlivePeriod(keepAlivePeriod)

	if err != nil {
		return nil, err
	}

	// Disable Nagle's algorithm to get lower latency.
	// NoDelay causes TCP packets to be sent immediately
	// when we call Write() instead of waiting to possibly
	// reduce the amount of packets. As Aero is fully aware
	// of the entire response data for most responses and only
	// causes a single Write() call, the delay should be disabled.
	err = connection.SetNoDelay(true)

	if err != nil {
		return nil, err
	}

	return connection, nil
}
