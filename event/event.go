// Package event contains utilities for parsing and verifying event protocol packets
package event

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"../client"
)

// Action denotes which type of event the packet is
type Action int

const (
	BroadcastAction Action = 1 << iota
	FollowAction
	UnfollowAction
	PrivateMessageAction
	StatusUpdateAction
)

var (
	IncorrectFormatError = errors.New("payload is formatted incorrectly")

	eventPattern = regexp.MustCompile(`^\d+\|(([FUP]\|\d+\|\d+)|B|(S\|\d+))\n$`)
)

var actions = map[string]Action{
	"B": BroadcastAction,
	"F": FollowAction,
	"U": UnfollowAction,
	"P": PrivateMessageAction,
	"S": StatusUpdateAction,
}

// Parse parses an event packet residing in a byte slice.
func Parse(buf []byte) (Packet, error) {
	msg := string(buf)

	if !eventPattern.MatchString(msg) {
		return nil, IncorrectFormatError
	}

	tokens := strings.Split(strings.TrimSuffix(msg, "\n"), "|")

	seq, err := strconv.ParseUint(tokens[0], 10, 64)
	if err != nil {
		return nil, err
	}

	var ts []client.UID
	for _, token := range tokens[2:] {
		n, err := client.ParseUID([]byte(token))
		if err != nil {
			return nil, err
		}

		ts = append(ts, n)
	}

	return packet{
		seq:    seq,
		action: actions[tokens[1]],
		buffer: buf,
		uIDs:   ts,
	}, nil
}

// Packet contains behavior of an event.Packet.
type Packet interface {
	String() string
	Payload() []byte
	Action() Action
	Sequence() uint64
	UIDs() []client.UID
}

type packet struct {
	seq    uint64
	action Action
	buffer []byte
	uIDs   []client.UID
}

// Sequence returns the sequence number of the packet.
func (m packet) Sequence() uint64 {
	return m.seq
}

// Sequence returns the string form of the payload.
func (m packet) String() string {
	return string(m.Payload())
}

// Sequence returns payload in a byte slice.
func (m packet) Payload() []byte {
	return m.buffer
}

// UIDs returns the user IDs mentioned in the event.Packet.
func (m packet) UIDs() []client.UID {
	return m.uIDs
}

// Action denotes the type of event.Packet.
func (m packet) Action() Action {
	return m.action
}
