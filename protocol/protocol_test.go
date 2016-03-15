package protocol_test

import (
	"bufio"
	"io"
	"net"
	"testing"

	"."
)

func TestAcceptsTCPConnections(t *testing.T) {
	addr := ":9229"
	accept := protocol.TCP(addr)

	pingMsg, pongMsg := "Hello, World!\n", "!dlroW, olleH\n"
	msgChan := make(chan string, 1)

	go func() {
		conn, err := accept()
		if err != nil {
			t.Fatalf("protocol.TCP(%#q) got error %v", addr, err)
		}

		defer conn.Close()

		rdr := bufio.NewReader(conn)
		buf, err := rdr.ReadBytes('\n')
		if err != nil {
			t.Fatalf("bufio.ReadBytes(conn) got error %v", err)
		}

		msgChan <- string(buf)

		if _, err := io.WriteString(conn, pongMsg); err != nil {
			t.Fatalf("bufio.Write(conn, %#q) got error %v", pongMsg, err)
		}
	}()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("net.Dial(tcp, %#q) got error %v", addr, err)
	}

	defer conn.Close()

	if _, err := io.WriteString(conn, pingMsg); err != nil {
		t.Errorf("io.WriteString(conn, %#q) got error %v", pingMsg, err)
	}

	if expected, got := pingMsg, <-msgChan; got != expected {
		t.Errorf("io.WriteString(conn, %#q) client received %#q instead of %#q", expected, got)
	}

	rdr := bufio.NewReader(conn)
	buf, err := rdr.ReadBytes('\n')
	if err != nil {
		t.Errorf("bufio.ReadBytes(conn) got error %v", err)
	}

	if expected, got := pongMsg, string(buf); expected != got {
		t.Errorf("bufio.ReadBytes(conn) expected %#q, got %#q", expected, got)
	}
}
