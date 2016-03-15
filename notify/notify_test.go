package notify_test

import (
	"strconv"
	"testing"

	"../client"
	"../event"
	"../notify"
)

var UIDs = []client.UID{
	15, 2, 92, 71, 87,
}

func populateClientRegistry() (chan<- client.RegistryFunc, chan client.Payloader) {
	registryCh := client.NewRegistry()
	payloadCh := make(chan client.Payloader, len(UIDs))

	for _, uid := range UIDs {
		registryCh <- client.RegisterFunc(uid, payloadCh)
	}

	return registryCh, payloadCh
}

func TestIgnoresNotificationIfUserIsntThere(t *testing.T) {
	registryCh, _ := populateClientRegistry()
	defer close(registryCh)

	payload := []byte("123123|F|15|200\n")
	pkt, err := event.Parse(payload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
	}

	registryCh <- notify.FuncFor(pkt)
}
func TestBroadcastsNotification(t *testing.T) {
	registryCh, payloadCh := populateClientRegistry()
	defer close(registryCh)
	defer close(payloadCh)

	payload := []byte("123123|B\n")
	pkt, err := event.Parse(payload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
	}

	registryCh <- notify.FuncFor(pkt)

	for _, uid := range UIDs {
		if got := (<-payloadCh).Payload(); string(got) != string(payload) {
			t.Errorf("notify.Broadcast => expected every user to receive %#q, but client%v received %#q", string(payload), uid, string(got))
		}
	}
}

func userInRegistry(uid client.UID, registryChan chan<- client.RegistryFunc) bool {
	boolCh := make(chan bool)

	registryChan <- func(r client.Registry) error {
		_, ok := r[uid]

		boolCh <- ok

		return nil
	}

	return <-boolCh
}

func userInFollowers(from client.UID, to client.UID, registryChan chan<- client.RegistryFunc) bool {
	boolCh := make(chan bool)

	registryChan <- func(r client.Registry) error {
		boolCh <- r[to].Followers.Contains(from)

		return nil
	}

	return <-boolCh
}

func TestFollowsTarget(t *testing.T) {
	registryCh, payloadCh := populateClientRegistry()
	defer close(registryCh)
	defer close(payloadCh)

	from, to := client.UID(15), client.UID(2)
	payload := []byte("123123|F|15|2\n")
	pkt, err := event.Parse(payload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
	}

	registryCh <- notify.FuncFor(pkt)

	if got := (<-payloadCh).Payload(); string(got) != string(payload) {
		t.Errorf("notify.Follow => expected user %v to receive %#q, but it received %#q", to, string(payload), string(got))
	}

	if !userInRegistry(to, registryCh) {
		t.Errorf("notify.Follow => expected user %v to be created, but it wasn't", from)
	}
	if !userInFollowers(from, to, registryCh) {
		t.Errorf("notify.Follow => expected user %v to be a follower of user %v, but it isn't", from, to)
	}
}

func TestFollowsInactiveTargets(t *testing.T) {
	registryCh, _ := populateClientRegistry()
	defer close(registryCh)

	from, to := client.UID(15), client.UID(200)
	payload := []byte("123123|F|15|200\n")
	pkt, err := event.Parse(payload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
	}

	registryCh <- notify.FuncFor(pkt)

	if !userInFollowers(from, to, registryCh) {
		t.Errorf("notify.Follow => expected user %v to be a follower of user %v, but it isn't", from, to)
	}
}

func TestUnfollowsTargetWithoutNotification(t *testing.T) {
	registryCh, payloadCh := populateClientRegistry()
	defer close(registryCh)
	defer close(payloadCh)

	from, to := client.UID(15), client.UID(2)
	payload := []byte("123123|F|15|2\n")
	pkt, err := event.Parse(payload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
	}

	registryCh <- notify.FuncFor(pkt)

	if !userInFollowers(from, to, registryCh) {
		t.Errorf("notify.Follow => expected user %v to be a follower of user %v, but it isn't", from, to)
	}

	unfollowPayload := []byte("123123|U|15|2\n")
	unfollowPkt, err := event.Parse(unfollowPayload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(unfollowPayload), err)
	}

	registryCh <- notify.FuncFor(unfollowPkt)

	if userInFollowers(from, to, registryCh) {
		t.Errorf("notify.Unfollow => expected user %v to stop being a follower of user %v, but it didn't", from, to)
	}

	// Extract the first notification and check if there is any activity
	// in the payload channel after the Follow notification
	<-payloadCh
	select {
	case notification := <-payloadCh:
		t.Errorf("notify.Unfollow => user %v shouldn't have received payload, but it received %#q", to, string(notification.Payload()))
	default:
	}
}

func TestSendsAPrivateMessageToTarget(t *testing.T) {
	registryCh, payloadCh := populateClientRegistry()
	defer close(registryCh)
	defer close(payloadCh)

	_, to := client.UID(15), client.UID(2)
	payload := []byte("123123|P|15|2\n")
	pkt, err := event.Parse(payload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
	}

	registryCh <- notify.FuncFor(pkt)

	if got := (<-payloadCh).Payload(); string(got) != string(payload) {
		t.Errorf("notify.PrivateMessage => expected user %v to receive %#q, but it received %#q", to, string(payload), string(got))
	}
}

func TestSendsStatusUpdateToAllFollowers(t *testing.T) {
	registryCh, payloadCh := populateClientRegistry()
	defer close(registryCh)
	defer close(payloadCh)

	target, followers := UIDs[0], UIDs[1:]
	for _, uid := range followers {
		payload := []byte("123123|F|" + strconv.FormatUint(uint64(uid), 10) + "|" + strconv.FormatUint(uint64(target), 10) + "\n")
		pkt, err := event.Parse(payload)
		if err != nil {
			t.Fatalf("parse.Event(%#q) got error %v", string(payload), err)
		}

		registryCh <- notify.FuncFor(pkt)
		// Empty the payload channel
		<-payloadCh
	}

	statusPayload := []byte("123123|S|" + strconv.FormatUint(uint64(target), 10) + "\n")
	statusPkt, err := event.Parse(statusPayload)
	if err != nil {
		t.Fatalf("parse.Event(%#q) got error %v", string(statusPayload), err)
	}

	registryCh <- notify.FuncFor(statusPkt)

	for _, uid := range UIDs[:0] {
		if got := (<-payloadCh).Payload(); string(got) != string(statusPayload) {
			t.Errorf("notify.StatusUpdate => expected every follower of %v to receive %#q, but client%v received %#q", target, string(statusPayload), uid, string(got))
		}
	}

	// Shouldn't have received more than len(UIDs)-1 notifications
	select {
	case notification := <-payloadCh:
		t.Errorf("notify.StatusUpdate => target user %v shouldn't have received a payload, but it received %#q", target, string(notification.Payload()))
	default:
	}
}
