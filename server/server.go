// Package server contains methods to serve clients
// over a communications channel.
package server

import (
	"fmt"

	"../log"

	"../client"
	"../protocol"
)

// Listen uses the given protocol.Listener to accept connections and protocol.Handler
// to handle those connections.
func Listen(accept protocol.Listener, handle protocol.Handler) error {
	// TODO(tmrts): Make protocol.Listener a variadic argument for allowing
	//              a user to listen on multiple ports/protocols with few lines of code.
	// TODO(tmrts); server.Listen(protocol.TCP(addr1), protocol.TCP(addr2), handleFunc, upstreamPort)
	//              can be used as a simple reverse proxy as well
	for {
		c, err := accept()
		if err != nil {
			return fmt.Errorf("server.Listen: error while accepting a connection %#q", err)
		}

		go func(conn client.Interface) {
			err := handle(conn)
			if err != nil {
				log.Error(fmt.Sprintf("server.Listen: error while handling a connection %#q", err))
				return
			}
		}(c)
	}

	return nil
}
