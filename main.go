package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"./client"
	"./handle"
	"./log"
	"./protocol"
	"./server"
)

var (
	registryChan = client.NewRegistry()
)

var (
	ClientListenerPort = os.Getenv("clientListenerPort")
	EventListenerPort  = os.Getenv("eventListenerPort")
)

func init() {
	if EventListenerPort != "" {
		EventListenerPort = ":" + EventListenerPort
	} else {
		EventListenerPort = ":9090"
	}

	if ClientListenerPort != "" {
		ClientListenerPort = ":" + ClientListenerPort
	} else {
		ClientListenerPort = ":9099"
	}
}

// Handles event sources and supports multiple event sources at the same time.
func handleEventSourceConnections(conn client.Interface) error {
	rdr := bufio.NewReader(conn)

	payloadCh := make(chan []byte)
	defer close(payloadCh)

	handle.Events(payloadCh, registryChan)

	for {
		payload, err := rdr.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Error(fmt.Sprintf("server.Listen: event source thrown read error %#q", err))
			}

			break
		}

		// No more to read close the loop
		if len(payload) == 0 {
			break
		}

		payloadCh <- payload
	}

	return nil
}

// Handles new event consumer connections.
func handleClientConnections(conn client.Interface) error {
	rdr := bufio.NewReader(conn)

	// Trims the line-feed at the end
	buf, _, err := rdr.ReadLine()
	if err != nil {
		return err
	}

	uid, err := client.ParseUID(buf)
	if err != nil {
		return err
	}

	payloadCh := make(chan client.Payloader)

	handle.Client(conn, payloadCh)

	// Sends a closure that registers client to the client registry.
	registryChan <- client.RegisterFunc(client.UID(uid), payloadCh)

	return nil
}

func main() {
	go func() {
		log.Info("Starting the event source handler...")
		err := server.Listen(protocol.TCP(EventListenerPort), handleEventSourceConnections)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Info("Starting the client handler...")
	err := server.Listen(protocol.TCP(ClientListenerPort), handleClientConnections)
	if err != nil {
		log.Fatal(err)
	}
}
