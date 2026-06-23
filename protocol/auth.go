package protocol

import (
	"errors"
	"fmt"

	"arcan-it.de/arclib/crypto"
	"arcan-it.de/arclib/msgpack"
)

// Identity authentication over an established (anonymous) session
// (SIGMA-style). The hybrid KEX in handshake.go gives confidentiality and
// forward secrecy but not peer identity; this layer adds mutual authentication
// by having each endpoint sign the handshake transcript with its long-term
// hybrid identity key (crypto.IdentityKey) and verify the peer's signature,
// then run a caller-supplied authorization check (e.g. an allowlist or pin).
//
// This is an arcmail extension to the MyClerk handshake (the base spec leaves
// peer authentication to a higher layer), kept transport-agnostic: callers
// provide send/recv that frame and encrypt the exchanged bytes over the session
// codec, so the signed identities are exchanged confidentially (identity
// protection). The responder authenticates first.
//
// Distinct context labels per role bind each signature to a direction so a
// signature observed in one role cannot be reflected as the other.

// FrameCodec is the minimal interface for a per-connection frame codec; both
// AeadSession (Tier 3) and Tier5Session (Tier 5) implement it, so a transport
// can carry either after the handshake.
type FrameCodec interface {
	Encode(op OpCode, plaintext []byte) ([]byte, error)
	Decode(frame []byte) (OpCode, []byte, error)
}

var (
	_ FrameCodec = (*AeadSession)(nil)
	_ FrameCodec = (*Tier5Session)(nil)
)

// Authentication errors.
var (
	ErrAuthFraming = errors.New("protocol: malformed identity-auth message")
	ErrAuthVerify  = errors.New("protocol: peer identity signature invalid")
)

var (
	authInitiatorContext = []byte("arcmail-auth-initiator-v1")
	authResponderContext = []byte("arcmail-auth-responder-v1")
)

// authMessage carries an endpoint's hybrid identity public key and its
// signature over (role-context || transcript).
type authMessage struct {
	Pub []byte `msgpack:"pub"`
	Sig []byte `msgpack:"sig"`
}

func signedTranscript(context []byte, transcript [32]byte) []byte {
	msg := make([]byte, 0, len(context)+len(transcript))
	msg = append(msg, context...)
	msg = append(msg, transcript[:]...)
	return msg
}

func marshalAuth(id *crypto.IdentityKey, context []byte, transcript [32]byte) ([]byte, error) {
	sig, err := id.Sign(signedTranscript(context, transcript))
	if err != nil {
		return nil, fmt.Errorf("protocol: sign transcript: %w", err)
	}
	raw, err := msgpack.Marshal(authMessage{Pub: id.PublicKey(), Sig: sig})
	if err != nil {
		return nil, fmt.Errorf("protocol: encode auth message: %w", err)
	}
	return raw, nil
}

// verifyPeerAuth unmarshals raw, verifies the signature under peerContext, and
// returns the peer's verified public key.
func verifyPeerAuth(raw, peerContext []byte, transcript [32]byte) ([]byte, error) {
	var m authMessage
	if err := msgpack.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthFraming, err)
	}
	if !crypto.IdentityVerify(m.Pub, signedTranscript(peerContext, transcript), m.Sig) {
		return nil, ErrAuthVerify
	}
	return m.Pub, nil
}

// AuthenticateInitiator runs the initiator side of the mutual identity exchange.
// It first verifies the responder (and runs authorize on its key, e.g. a pin),
// then presents its own identity. send/recv exchange raw message bytes over the
// already-encrypted session; authorize decides whether the peer key is trusted.
func AuthenticateInitiator(id *crypto.IdentityKey, transcript [32]byte,
	send func([]byte) error, recv func() ([]byte, error), authorize func(peerPub []byte) error) error {

	raw, err := recv()
	if err != nil {
		return fmt.Errorf("protocol: recv responder auth: %w", err)
	}
	peerPub, err := verifyPeerAuth(raw, authResponderContext, transcript)
	if err != nil {
		return err
	}
	if err := authorize(peerPub); err != nil {
		return err
	}

	mine, err := marshalAuth(id, authInitiatorContext, transcript)
	if err != nil {
		return err
	}
	if err := send(mine); err != nil {
		return fmt.Errorf("protocol: send initiator auth: %w", err)
	}
	return nil
}

// AuthenticateResponder runs the responder side: it presents its identity first,
// then verifies the initiator and runs authorize on the initiator's key (e.g. an
// allowlist).
func AuthenticateResponder(id *crypto.IdentityKey, transcript [32]byte,
	send func([]byte) error, recv func() ([]byte, error), authorize func(peerPub []byte) error) error {

	mine, err := marshalAuth(id, authResponderContext, transcript)
	if err != nil {
		return err
	}
	if err := send(mine); err != nil {
		return fmt.Errorf("protocol: send responder auth: %w", err)
	}

	raw, err := recv()
	if err != nil {
		return fmt.Errorf("protocol: recv initiator auth: %w", err)
	}
	peerPub, err := verifyPeerAuth(raw, authInitiatorContext, transcript)
	if err != nil {
		return err
	}
	return authorize(peerPub)
}
