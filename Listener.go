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
	connection, err := listener.AcceptTCP()

	if err != nil {
		return nil, err
	}

	err = connection.SetKeepAlive(true)

	if err != nil {
		return nil, err
	}

	err = connection.SetKeepAlivePeriod(keepAlivePeriod)

	if err != nil {
		return nil, err
	}

	err = connection.SetNoDelay(true)

	if err != nil {
		return nil, err
	}

	return connection, nil
}
