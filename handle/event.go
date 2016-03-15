// Package handle contains methods that utilize
// goroutines to handle clients and events
package handle

import (
	"fmt"

	"../client"
	"../event"
	"../log"
	"../notify"
)

const (
	startingIndex = 1
)

// Events funnels out-of-order packets and sends them in a sorted fashiong
// as client.RegistryFunc closures.
func Events(payloadCh <-chan []byte, registryCh chan<- client.RegistryFunc) {
	packets := make(map[uint64]client.RegistryFunc)

	go func(payloadCh <-chan []byte, registryCh chan<- client.RegistryFunc) {
		var index uint64 = startingIndex

		defer close(registryCh)

		for payload := range payloadCh {
			pkt, err := event.Parse(payload)
			if err != nil {
				// TODO(tmrts): might try to read the packet sequence no and skip that packet
				//              to make sure the flow continues.
				log.Debug(fmt.Sprintf("event.Parse(%#q) got error %#q", string(payload), err))
				continue
			}

			seq := pkt.Sequence()
			// Ignores packets with same sequence numbers or
			// lower than current index numbers.
			if _, ok := packets[seq]; !ok && seq >= index {
				packets[seq] = notify.FuncFor(pkt)
			}

			for {
				pkt, ok := packets[index]
				if !ok {
					break
				}

				registryCh <- pkt

				// Evicts used event packets
				// NOTE: Bulk delete might increase performance
				delete(packets, index)

				index++
			}
		}

		// Send the remaning events
		for {
			pkt, ok := packets[index]
			if !ok {
				break
			}

			registryCh <- pkt
			index++
		}
	}(payloadCh, registryCh)
}
