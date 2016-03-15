package handle_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"../client"
	"../handle"
)

type payload string

func (p payload) Payload() []byte {
	return []byte(p)
}

type buffer struct {
	bytes.Buffer
	io.Closer
}

func TestSendsMessageToClientEmitter(t *testing.T) {
	msg := "12|B|12|2323\n"

	recorder, payloadCh := new(buffer), make(chan client.Payloader)
	defer close(payloadCh)

	handle.Client(recorder, payloadCh)

	payloadCh <- payload(msg)

	if expected, got := msg, recorder.String(); got != expected {
		t.Errorf("handle.Client expected msg %#q, got %#q", expected, got)
	}
}

type mockBuffer struct {
	io.Reader
	isClosed chan bool
}

func (buf *mockBuffer) Write([]byte) (int, error) {
	return -1, errors.New("write operation not possible")
}

func (buf *mockBuffer) Close() error {
	buf.isClosed <- true

	return nil
}

func TestChecksReleaseOfResources(t *testing.T) {
	msg := "12|B|12|2323\n"

	c := &mockBuffer{
		isClosed: make(chan bool, 1),
	}

	payloadCh := make(chan client.Payloader)
	defer close(payloadCh)

	handle.Client(c, payloadCh)

	payloadCh <- payload(msg)

	if isClosed := <-c.isClosed; !isClosed {
		t.Error("handle.Client should have closed `payloadCh` channel after client.Interface has been closed")
	}
}
