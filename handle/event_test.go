package handle_test

import (
	"testing"

	"."
	"../client"
)

func TestHandlesEvents(t *testing.T) {
	registryCh := client.NewRegistry()

	payloads := [][]byte{
		[]byte("1|F|12|13\n"),
		[]byte("2|U|12|13\n"),
		[]byte("3|P|12|13\n"),
		[]byte("4|F|13|12\n"),
		[]byte("5|S|12\n"),
		[]byte("6|B\n"),
	}

	expectationsOfClient13 := [][]byte{
		[]byte("1|F|12|13\n"),
		[]byte("3|P|12|13\n"),
		[]byte("5|S|12\n"),
		[]byte("6|B\n"),
	}

	// Registers Client ID 13
	payloadCh13 := make(chan client.Payloader, len(expectationsOfClient13))

	registryCh <- client.RegisterFunc(13, payloadCh13)

	expectationsOfClient12 := [][]byte{
		[]byte("4|F|13|12\n"),
		[]byte("6|B\n"),
	}

	// Registers Client ID 12
	payloadCh12 := make(chan client.Payloader, len(expectationsOfClient13))

	registryCh <- client.RegisterFunc(12, payloadCh12)

	inputCh := make(chan []byte)
	defer close(inputCh)

	handle.Events(inputCh, registryCh)

	for _, p := range payloads {
		inputCh <- p
	}

	for _, inOrderPkt := range expectationsOfClient12 {
		if expected, got := string(inOrderPkt), string((<-payloadCh12).Payload()); expected != got {
			t.Errorf("handle.Events => user %v expected packets in order, should have got %#q, but received %#q", 12, expected, got)
		}
	}

	for _, inOrderPkt := range expectationsOfClient13 {
		if expected, got := string(inOrderPkt), string((<-payloadCh13).Payload()); expected != got {
			t.Errorf("handle.Events => user %v expected packets in order, should have got %#q, but received %#q", 13, expected, got)
		}
	}
}
