// Package protocol contains network protocol
// interfaces and functions that can be used
// with the server package to serve clients
package protocol

import (
	"fmt"
	"net"
	"time"

	"../client"
	"../log"
)

var (
	TCP_TIMEOUT          time.Duration = 5 * time.Second
	TCP_KEEPALIVE_PERIOD time.Duration = 10 * time.Second
)

// Handler is function that mutates a client.Interface and returns an error if there is any.
type Handler func(client.Interface) error

// Listener returns a connection that implements client.Interface and any errors it encounters.
type Listener func() (client.Interface, error)

// TCP creates a listener on the given address that accepts TCP connections.
// Throws a panic if binding is unsuccessful.
func TCP(addr string) Listener {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(fmt.Sprintf("while binding to address %#q, got error %#q", addr, err))
		panic(err)
	}

	// Sets options for TCP socket connections
	// TODO(tmrts); Optimize heartbeat pings
	setOptions := func(conn net.Conn) error {
		tcp := conn.(*net.TCPConn)

		if err := tcp.SetKeepAlive(true); err != nil {
			return err
		}

		if err := tcp.SetKeepAlivePeriod(TCP_KEEPALIVE_PERIOD); err != nil {
			return err
		}

		return nil
	}

	return func() (client.Interface, error) {
		conn, err := listener.Accept()
		if err != nil {
			return nil, err
		}

		if err := setOptions(conn); err != nil {
			return nil, err
		}

		return conn, nil
	}
}
