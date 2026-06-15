package transport

import (
	"sync"

	"arcan-it.de/arclib/protocol"
)

// Session is the per-connection state established after a successful handshake.
// It owns the Tier-3 AEAD state and optional identity attributes that handlers
// may attach. Send is safe for concurrent use; the underlying writer serializes
// frames so concurrent streaming and request handling cannot interleave bytes.
type Session struct {
	id    uint64
	aead  *protocol.AeadSession
	write func([]byte) error

	mu       sync.RWMutex
	userID   string
	familyID string
	deviceID string
}

func newSession(id uint64, aead *protocol.AeadSession, write func([]byte) error) *Session {
	return &Session{id: id, aead: aead, write: write}
}

// ID returns the session identifier both endpoints derived during the handshake.
func (s *Session) ID() uint64 { return s.id }

// Send seals payload under op into a Tier-3 frame and writes it to the peer.
// A streaming handler calls Send repeatedly (e.g. one MAYA_STREAM frame per
// chunk) and returns no response message from its Handle method.
func (s *Session) Send(op protocol.OpCode, payload []byte) error {
	frame, err := s.aead.Encode(op, payload)
	if err != nil {
		return err
	}
	return s.write(frame)
}

// SetIdentity records authenticated identity attributes for later authorization.
func (s *Session) SetIdentity(userID, familyID, deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userID, s.familyID, s.deviceID = userID, familyID, deviceID
}

// UserID returns the identity attached via SetIdentity (empty if unset).
func (s *Session) UserID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userID
}

// FamilyID returns the identity attached via SetIdentity (empty if unset).
func (s *Session) FamilyID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.familyID
}

// DeviceID returns the identity attached via SetIdentity (empty if unset).
func (s *Session) DeviceID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deviceID
}
