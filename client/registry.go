package client

import (
	"fmt"
	"os"

	"../log"
)

// Session contains necessary information
// to communicate with users and their followers.
type Session struct {
	isClosed bool

	Chan chan<- Payloader

	Followers UIDSet
}

// IsActive tells whether the given session is activated.
// The meaning of activation is that, it has an open communication
// channel to communicate with the user
func (s *Session) IsActive() bool {
	return s.Chan != nil && !s.isClosed
}

// Send sends a given payload to the owner of the session.
// If the connection has been closed, it returns an error.
func (s *Session) Send(p Payloader) error {
	if !s.IsActive() {
		log.Debug("client.Session: client is inactive")
		return nil
	}

	// Recovers when channel is prematurely closed due to most likely connection errors.
	defer func() {
		if err := recover(); err != nil {
			s.isClosed = true
			log.Debug(fmt.Sprintf("channel.Send panic recovered from %v", err))
		}
	}()

	s.Chan <- p

	return nil
}

// Close closes communication channels of a session.
func (s *Session) Close() error {
	if !s.IsActive() {
		return nil
	}

	// Checks whether the channel is closed or not in a non-blocking manner.
	s.isClosed = true

	close(s.Chan)

	return nil
}

// Registry holds a collection of client and
// their session information
type Registry map[UID]*Session

// tearDown terminates every session registered in the
// current client.Registry.
func (r *Registry) tearDown() error {
	for uid, session := range *r {
		if err := session.Close(); err != nil {
			log.Debug(fmt.Sprintf("session.Close() for user%v, got an error %v", uid, err))
		}
	}

	return nil
}

// RegistryFunc is a function that is used by the client.Registry.
// Given a RegistryFunc to a client.Registry, it executes it on the
// dedicated goroutine of that registry, which is similar to
// Scala akka actor model and provides lock-less safety.
type RegistryFunc func(Registry) error

// RegisterFunc creates a RegistryFunc that registers the given client to the
// Registry when invoked by the Registry itself.
func RegisterFunc(uid UID, payloadCh chan<- Payloader) RegistryFunc {
	return func(clients Registry) error {
		// For preserving the follower list.
		if _, ok := clients[uid]; ok {
			clients[uid] = &Session{
				Chan:      payloadCh,
				Followers: clients[uid].Followers,
			}
		} else {
			clients[uid] = &Session{
				Chan:      payloadCh,
				Followers: make(UIDSet),
			}
		}

		return nil
	}
}

// UnregisterFunc returns a RegistryFunc that
// unregisters a user from the Registry when invoked.
func UnregisterFunc(uid UID) RegistryFunc {
	return func(clients Registry) error {
		delete(clients, uid)

		return nil
	}
}

// NewRegistry creates a new client.Registry and returns
// a RegistryFunc channel for communication purposes.
func NewRegistry() chan<- RegistryFunc {
	funcCh := make(chan RegistryFunc)

	clientRegistry := make(Registry)

	go func(funcCh <-chan RegistryFunc) {
		defer func() {
			clientRegistry.tearDown()

			log.Info("Every notification has been sent.")

			os.Exit(0)
		}()

		for use := range funcCh {
			err := use(clientRegistry)
			if err != nil {
				log.Debug(fmt.Sprintf("error while executing a client.RegistryFunc %#v", err))
				continue
			}
		}

	}(funcCh)

	return funcCh
}
