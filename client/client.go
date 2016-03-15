// Package client contains interfaces and structs
// for clients and registries that record the connected clients.
package client

import "io"

// Interface contains bare minimum methods a client connection must support.
type Interface io.ReadWriteCloser

// Payloader contains Payload method that returns byte slice.
type Payloader interface {
	Payload() []byte
}
