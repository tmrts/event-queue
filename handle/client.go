package handle

import (
	"fmt"

	"../client"
	"../log"
)

// Client manages communications to/from a client.Interface.
// Returns a channel that signals when a connection lifetime has ended.
func Client(conn client.Interface, payloadCh <-chan client.Payloader) {
	go func(payloadCh <-chan client.Payloader) {
		defer conn.Close()

		for pkt := range payloadCh {
			_, err := conn.Write(pkt.Payload())
			if err != nil {
				log.Debug(fmt.Sprintf("handle.Client: while forwarding packets to a client, got error %#q", err))
				return
			}
		}

	}(payloadCh)
}
