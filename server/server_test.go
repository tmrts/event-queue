package server_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"."
	"../client"
)

type msgBuffer struct {
	*bytes.Buffer
	io.Closer
}

func TestListensOnProtocol(t *testing.T) {
	expectedMsg := "This is an important message!\n"

	msgBuf := msgBuffer{
		Buffer: bytes.NewBufferString(expectedMsg),
	}

	// listener will block after sending 1 msgBuf
	clientSem := make(chan struct{}, 1)

	listener := func() (client.Interface, error) {
		clientSem <- struct{}{}

		return msgBuf, nil
	}

	msgChan := make(chan string)
	go server.Listen(listener, func(c client.Interface) error {
		buf, err := ioutil.ReadAll(c)
		if err != nil {
			return err
		}

		msgChan <- string(buf)

		return nil
	})

	if msg := <-msgChan; msg != expectedMsg {
		t.Errorf("server.Listen(listener, reader) expected message %#q, got %#q", expectedMsg, msg)
	}
}
