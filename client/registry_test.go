package client_test

import (
	"testing"

	"."
)

func TestRegistersClientsToRegistry(t *testing.T) {
	registryCh := client.NewRegistry()
	defer close(registryCh)

	var emptyPayloadCh chan client.Payloader

	UIDs := []client.UID{
		15, 2, 92, 71, 87,
	}

	for _, uid := range UIDs {
		registryCh <- client.RegisterFunc(uid, emptyPayloadCh)
	}

	userCh := make(chan client.UIDSet)

	registryCh <- func(r client.Registry) error {
		UIDs := make(client.UIDSet)

		for id := range r {
			UIDs.Add(id)
		}

		userCh <- UIDs

		return nil
	}

	results := <-userCh
	for _, uid := range UIDs {
		if _, ok := results[uid]; !ok {
			t.Errorf("client.RegisterFunc => expected %v to be registered, but it wasn't", uid)
		}
	}
}

func TestUnregistersClientsToRegistry(t *testing.T) {
	uid := client.UID(92)

	registryCh := client.NewRegistry()

	registryCh <- func(r client.Registry) error {
		r[uid] = nil

		return nil
	}

	registryCh <- client.UnregisterFunc(uid)

	boolCh := make(chan bool)
	registryCh <- func(r client.Registry) error {
		_, ok := r[uid]

		boolCh <- ok

		return nil
	}

	if ok := <-boolCh; ok {
		t.Errorf("client.UnregisterFunc => expected %v to be unregistered, but it wasn't", uid)
	}
}
