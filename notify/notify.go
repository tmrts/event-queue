// Package notify contains functions that operate on a client.Registry to send user notifications.
package notify

import (
	"fmt"

	"../client"
	"../event"
	"../log"
)

// Func transforms an event.Packet to a function that
// takes a client.Registry and modifies it.
type Func func(pkt event.Packet) client.RegistryFunc

var notifyMap = map[event.Action]Func{
	event.FollowAction:         Follow,
	event.UnfollowAction:       Unfollow,
	event.PrivateMessageAction: PrivateMessage,
	event.StatusUpdateAction:   StatusUpdate,
	event.BroadcastAction:      Broadcast,
}

// New returns a notification closure for the
// given event.Packet.
func FuncFor(pkt event.Packet) client.RegistryFunc {
	return notifyMap[pkt.Action()](pkt)
}

// Follow returns a closure that adds a follower
// to the target user's list and sends a notification to that user when invoked.
// If the target user hasn't connected to the client.Registry yet,
// it registers that user into the client.Registry, but as inactive
// so that a list of followers can be kept
func Follow(pkt event.Packet) client.RegistryFunc {
	return func(clients client.Registry) error {
		UIDs := pkt.UIDs()

		// TODO: should the event be ignored if from == to? probably for others as well.
		from, to := UIDs[0], UIDs[1]

		if _, ok := clients[to]; !ok {
			// Initializes a record for the unconnected client.
			clients[to] = &client.Session{
				Followers: make(client.UIDSet),
			}
		}

		targetClient := clients[to]
		targetClient.Followers.Add(from)

		if err := targetClient.Send(pkt); err != nil {
			log.Debug(fmt.Sprintf("notify.Follow: for client %v, got error %#q", to, err))
			client.UnregisterFunc(to)(clients)
			return err
		}

		return nil
	}
}

// Unfollow returns a closure that removes a follower
// from the target user's list when invokes.
// It doesn't send any notifications.
func Unfollow(pkt event.Packet) client.RegistryFunc {
	return func(clients client.Registry) error {
		UIDs := pkt.UIDs()
		from, to := UIDs[0], UIDs[1]

		if _, ok := clients[to]; !ok {
			return fmt.Errorf("for packet numbered %v client %#q is not connected", pkt.Sequence(), to)
		}

		targetClient := clients[to]

		if targetClient.Followers.Contains(from) {
			delete(targetClient.Followers, from)
		}

		return nil
	}
}

// StatusUpdate returns a closure that notifies every
// follower of the target user when invoked.
func StatusUpdate(pkt event.Packet) client.RegistryFunc {
	return func(clients client.Registry) error {
		from := pkt.UIDs()[0]

		if _, ok := clients[from]; !ok {
			return fmt.Errorf("for packet numbered %v client %v is not connected", pkt.Sequence(), from)
		}

		targetClient := clients[from]

		for uid := range targetClient.Followers {
			follower, ok := clients[uid]
			if !ok {
				// Client is no longer present, delete from followers
				delete(targetClient.Followers, uid)
				continue
			}

			if !follower.IsActive() {
				continue
			}

			if err := follower.Send(pkt); err != nil {
				log.Debug(fmt.Sprintf("notify.StatusUpdate: for client %v, got error %#q", uid, err))
				delete(targetClient.Followers, uid)

				client.UnregisterFunc(uid)(clients)
			}
		}

		return nil
	}
}

// PrivateMessage returns a closure that sends a private message
// notification to the target user when invoked.
func PrivateMessage(pkt event.Packet) client.RegistryFunc {
	return func(clients client.Registry) error {
		to := pkt.UIDs()[1]

		if _, ok := clients[to]; !ok {
			return fmt.Errorf("for packet numbered %v client %v is not connected", pkt.Sequence(), to)
		}

		targetClient := clients[to]

		if err := targetClient.Send(pkt); err != nil {
			log.Debug(fmt.Sprintf("notify.PrivateMessage: for client %v, got error %#q", to, err))
			client.UnregisterFunc(to)(clients)
			return err
		}

		return nil
	}
}

// Broadcast returns a closure that notifies every client
// registered in the client.Registry when invoked.
func Broadcast(pkt event.Packet) client.RegistryFunc {
	return func(clients client.Registry) error {
		for uid, c := range clients {
			if err := c.Send(pkt); err != nil {
				log.Debug(fmt.Sprintf("notify.Broadcast: for client %v, got error %#q", uid, err))
				client.UnregisterFunc(uid)(clients)
			}
		}

		return nil
	}
}
