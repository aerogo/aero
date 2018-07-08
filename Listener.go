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

	connection.SetKeepAlive(true)
	connection.SetKeepAlivePeriod(keepAlivePeriod)
	connection.SetNoDelay(true)

	return connection, nil
}
