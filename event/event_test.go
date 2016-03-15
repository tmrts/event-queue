package event_test

import (
	"fmt"
	"testing"

	"."
	"../client"
)

const (
	// Two's complement
	UINT64_MAX = ^uint64(0)
)

func TestParsesUIDs(t *testing.T) {
	tests := []struct {
		n   client.UID
		err error
	}{
		{n: 0},
		{n: 1519210928371298},
		{n: client.UID(UINT64_MAX)},
	}

	for _, testCase := range tests {
		buf := []byte(fmt.Sprint(testCase.n))

		n, err := client.ParseUID(buf)
		if err != testCase.err {
			t.Errorf("event.ParseUID([]byte(%v)) expected error %v, got %v", testCase.n, testCase.err, err)
		}

		if n != testCase.n {
			t.Errorf("event.ParseUID([]byte(%v)) got a different number %v", testCase.n, n)
		}
	}
}

func TestParsesIncorrectlyFormattedPackets(t *testing.T) {
	packets := []string{
		"",
		"\n",
		"F\n",
		"S\n",
		"B\n",
		"11\\|B\n",
		"11|P|12\n",
		"S|11|12\n",
		"11|PM|12\n",
		"11\n",
		"11|B",
		"11|U|21|11",
		"11|F|U|11\n",
	}

	for _, msg := range packets {
		payload := []byte(msg)

		_, err := event.Parse(payload)
		if expectedErr := event.IncorrectFormatError; expectedErr != err {
			t.Errorf("event.Parse(%#q) expected %q error, got %v", msg, expectedErr, err)
		}
	}
}

func TestParsesPayloads(t *testing.T) {
	packets := []struct {
		Payload  []byte
		Sequence uint64
		Action   event.Action
		UIDs     []client.UID
		Message  string
	}{
		{
			Sequence: 11,
			Action:   event.FollowAction,
			Message:  "11|F|12|12\n",
			UIDs:     []client.UID{12, 12},
		},
		{
			Sequence: 11,
			Action:   event.UnfollowAction,
			Message:  "11|U|12|12\n",
			UIDs:     []client.UID{12, 12},
		},
		{
			Sequence: 11,
			Action:   event.PrivateMessageAction,
			Message:  "11|P|12|12\n",
			UIDs:     []client.UID{12, 12},
		},
		{
			Sequence: 11,
			Action:   event.StatusUpdateAction,
			Message:  "11|S|12\n",
			UIDs:     []client.UID{12},
		},
		{
			Sequence: 11,
			Action:   event.BroadcastAction,
			Message:  "11|B\n",
			UIDs:     []client.UID{},
		},
	}

	for _, pkt := range packets {
		pkt.Payload = []byte(pkt.Message)

		msg, err := event.Parse(pkt.Payload)
		if err != nil {
			t.Errorf("event.Parse(%#q) -> got error %v", pkt.Message, err)
			continue
		}

		if pkt.Sequence != msg.Sequence() {
			t.Errorf("event.Parse(%#q) expected id %v, got %v", pkt.Message, pkt.Sequence, msg.Sequence())
		}

		if pkt.Action != msg.Action() {
			t.Errorf("event.Parse(%#q) expected action %v, got %v", pkt.Message, pkt.Action, msg.Action())
		}

		if expected, got := fmt.Sprint(pkt.UIDs), fmt.Sprint(msg.UIDs()); expected != got {
			t.Errorf("event.Parse(%#q) expected action %v, got %v", pkt.Message, expected, got)
		}
	}
}
