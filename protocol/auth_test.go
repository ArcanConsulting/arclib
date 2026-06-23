package protocol

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"arcan-it.de/arclib/crypto"
)

// pipeAuth wires an initiator and responder over two in-memory channels and runs
// the mutual identity exchange, returning both sides' errors.
func pipeAuth(t *testing.T,
	clientID, serverID *crypto.IdentityKey,
	clientTranscript, serverTranscript [32]byte,
	clientAuthorize, serverAuthorize func([]byte) error,
) (initErr, respErr error) {
	t.Helper()
	i2r := make(chan []byte, 1)
	r2i := make(chan []byte, 1)
	recvOrClosed := func(ch chan []byte) ([]byte, error) {
		b, ok := <-ch
		if !ok {
			return nil, errors.New("peer closed connection")
		}
		return b, nil
	}

	ic := make(chan error, 1)
	rc := make(chan error, 1)
	go func() {
		// Closing i2r on return models the conn closing when a side aborts,
		// so the peer's recv unblocks instead of hanging.
		err := AuthenticateInitiator(clientID, clientTranscript,
			func(b []byte) error { i2r <- b; return nil },
			func() ([]byte, error) { return recvOrClosed(r2i) },
			clientAuthorize)
		close(i2r)
		ic <- err
	}()
	go func() {
		err := AuthenticateResponder(serverID, serverTranscript,
			func(b []byte) error { r2i <- b; return nil },
			func() ([]byte, error) { return recvOrClosed(i2r) },
			serverAuthorize)
		close(r2i)
		rc <- err
	}()
	select {
	case initErr = <-ic:
	case <-time.After(3 * time.Second):
		t.Fatal("initiator auth timed out")
	}
	select {
	case respErr = <-rc:
	case <-time.After(3 * time.Second):
		t.Fatal("responder auth timed out")
	}
	return initErr, respErr
}

func newIDs(t *testing.T) (client, server *crypto.IdentityKey) {
	t.Helper()
	c, err := crypto.GenerateIdentityKey()
	if err != nil {
		t.Fatalf("client id: %v", err)
	}
	s, err := crypto.GenerateIdentityKey()
	if err != nil {
		t.Fatalf("server id: %v", err)
	}
	return c, s
}

func pinTo(expected []byte) func([]byte) error {
	return func(got []byte) error {
		if !bytes.Equal(got, expected) {
			return errors.New("untrusted peer identity")
		}
		return nil
	}
}

func TestMutualAuthSuccess(t *testing.T) {
	client, server := newIDs(t)
	tr := [32]byte{1, 2, 3}
	ie, re := pipeAuth(t, client, server, tr, tr,
		pinTo(server.PublicKey()), pinTo(client.PublicKey()))
	if ie != nil || re != nil {
		t.Fatalf("auth failed: initiator=%v responder=%v", ie, re)
	}
}

func TestMutualAuthInitiatorRejectsUnknownServer(t *testing.T) {
	client, server := newIDs(t)
	_, impostor := newIDs(t)
	tr := [32]byte{9}
	ie, _ := pipeAuth(t, client, server, tr, tr,
		pinTo(impostor.PublicKey()), // client expects a different server key
		pinTo(client.PublicKey()))
	if ie == nil {
		t.Fatal("initiator accepted an unpinned server identity")
	}
}

func TestMutualAuthResponderRejectsUnknownClient(t *testing.T) {
	client, server := newIDs(t)
	other, _ := newIDs(t)
	tr := [32]byte{7}
	_, re := pipeAuth(t, client, server, tr, tr,
		pinTo(server.PublicKey()),
		pinTo(other.PublicKey())) // server allowlist does not contain this client
	if re == nil {
		t.Fatal("responder accepted a client not on the allowlist")
	}
}

func TestMutualAuthTranscriptMismatch(t *testing.T) {
	client, server := newIDs(t)
	// A MITM that completed a separate KEX on each leg yields different
	// transcripts; the signatures then fail to verify.
	ie, re := pipeAuth(t, client, server, [32]byte{1}, [32]byte{2},
		pinTo(server.PublicKey()), pinTo(client.PublicKey()))
	if ie == nil && re == nil {
		t.Fatal("auth succeeded despite transcript mismatch")
	}
	if ie != nil && !errors.Is(ie, ErrAuthVerify) {
		t.Errorf("initiator error = %v, want ErrAuthVerify", ie)
	}
}
