package protocol

import (
	"bytes"
	"testing"
)

func TestHeaderFlagsRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		header Header
	}{
		{
			name: "plaintext basic",
			header: Header{
				Version: 0,
				Tier:    TierPlaintext,
				OpCode:  OpNop,
			},
		},
		{
			name: "encrypted with flags",
			header: Header{
				Version:    1,
				Tier:       TierEncrypted,
				Compressed: true,
				OpCode:     OpSessionInit,
			},
		},
		{
			name: "fragmented",
			header: Header{
				Version:    0,
				Tier:       TierChecksum,
				Fragmented: true,
				OpCode:     OpKeepalive,
			},
		},
		{
			name: "extensions",
			header: Header{
				Version:       0,
				Tier:          TierAuthenticated,
				HasExtensions: true,
				OpCode:        OpError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := tt.header.Flags()

			var decoded Header
			if err := decoded.ParseFlags(flags); err != nil {
				t.Fatalf("ParseFlags: %v", err)
			}

			if decoded.Version != tt.header.Version {
				t.Errorf("Version = %d, want %d", decoded.Version, tt.header.Version)
			}
			if decoded.Tier != tt.header.Tier {
				t.Errorf("Tier = %d, want %d", decoded.Tier, tt.header.Tier)
			}
			if decoded.Compressed != tt.header.Compressed {
				t.Errorf("Compressed = %v, want %v", decoded.Compressed, tt.header.Compressed)
			}
			if decoded.Fragmented != tt.header.Fragmented {
				t.Errorf("Fragmented = %v, want %v", decoded.Fragmented, tt.header.Fragmented)
			}
			if decoded.HasExtensions != tt.header.HasExtensions {
				t.Errorf("HasExtensions = %v, want %v", decoded.HasExtensions, tt.header.HasExtensions)
			}
		})
	}
}

func TestParseFlagsInvalidTier(t *testing.T) {
	// Tier 7 (0b111 << 2 = 0x1C) is invalid
	flags := byte(0x1C) // tier = 7
	var h Header
	if err := h.ParseFlags(flags); err != ErrInvalidTier {
		t.Errorf("ParseFlags(invalid tier) = %v, want ErrInvalidTier", err)
	}
}

func TestHeaderSize(t *testing.T) {
	tests := []struct {
		tier       Tier
		fragmented bool
		expected   int
	}{
		{TierPlaintext, false, 3},
		{TierChecksum, false, 3},
		{TierAuthenticated, false, 7},
		{TierEncrypted, false, 15},
		{TierPFS, false, 63},
		{TierMaxSecurity, false, 71},
		{TierPlaintext, true, 11},
		{TierEncrypted, true, 23},
	}

	for _, tt := range tests {
		h := Header{Tier: tt.tier, Fragmented: tt.fragmented}
		if got := h.HeaderSize(); got != tt.expected {
			t.Errorf("HeaderSize(tier=%d, frag=%v) = %d, want %d",
				tt.tier, tt.fragmented, got, tt.expected)
		}
	}
}

func TestTrailerSize(t *testing.T) {
	tests := []struct {
		tier     Tier
		expected int
	}{
		{TierPlaintext, 0},
		{TierChecksum, 2},
		{TierAuthenticated, 8},
		{TierEncrypted, 16},
		{TierPFS, 16},
		{TierMaxSecurity, 32},
	}

	for _, tt := range tests {
		h := Header{Tier: tt.tier}
		if got := h.TrailerSize(); got != tt.expected {
			t.Errorf("TrailerSize(tier=%d) = %d, want %d", tt.tier, got, tt.expected)
		}
	}
}

func TestHeaderMarshalUnmarshalPlaintext(t *testing.T) {
	h := Header{
		Version: 0,
		Tier:    TierPlaintext,
		OpCode:  OpKeepalive,
	}

	data := h.MarshalHeader()
	if len(data) != 3 {
		t.Fatalf("MarshalHeader len = %d, want 3", len(data))
	}

	var decoded Header
	n, err := decoded.UnmarshalHeader(data)
	if err != nil {
		t.Fatalf("UnmarshalHeader: %v", err)
	}
	if n != 3 {
		t.Errorf("bytes consumed = %d, want 3", n)
	}
	if decoded.OpCode != OpKeepalive {
		t.Errorf("OpCode = 0x%04X, want 0x%04X", decoded.OpCode, OpKeepalive)
	}
}

func TestHeaderMarshalUnmarshalAuthenticated(t *testing.T) {
	h := Header{
		Version:  0,
		Tier:     TierAuthenticated,
		OpCode:   OpSessionInit,
		Sequence: 42,
	}

	data := h.MarshalHeader()

	var decoded Header
	_, err := decoded.UnmarshalHeader(data)
	if err != nil {
		t.Fatalf("UnmarshalHeader: %v", err)
	}
	if decoded.Sequence != 42 {
		t.Errorf("Sequence = %d, want 42", decoded.Sequence)
	}
	if decoded.Tier != TierAuthenticated {
		t.Errorf("Tier = %d, want %d", decoded.Tier, TierAuthenticated)
	}
}

func TestHeaderMarshalUnmarshalEncrypted(t *testing.T) {
	h := Header{
		Version:   0,
		Tier:      TierEncrypted,
		OpCode:    OpSessionAck,
		Sequence:  100,
		SessionID: 0xDEADBEEFCAFEBABE,
	}

	data := h.MarshalHeader()

	var decoded Header
	_, err := decoded.UnmarshalHeader(data)
	if err != nil {
		t.Fatalf("UnmarshalHeader: %v", err)
	}
	if decoded.Sequence != 100 {
		t.Errorf("Sequence = %d, want 100", decoded.Sequence)
	}
	if decoded.SessionID != 0xDEADBEEFCAFEBABE {
		t.Errorf("SessionID = 0x%X, want 0xDEADBEEFCAFEBABE", decoded.SessionID)
	}
}

func TestHeaderMarshalUnmarshalPFS(t *testing.T) {
	nonce := [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	ecdhKey := [32]byte{0xAA, 0xBB, 0xCC}
	ecdhKey[31] = 0xFF

	h := Header{
		Version:    0,
		Tier:       TierPFS,
		OpCode:     OpKeyExchangeInit,
		Sequence:   999,
		SessionID:  0x1234567890ABCDEF,
		Nonce:      nonce,
		KeyID:      7,
		ECDHPublic: ecdhKey,
	}

	data := h.MarshalHeader()

	var decoded Header
	_, err := decoded.UnmarshalHeader(data)
	if err != nil {
		t.Fatalf("UnmarshalHeader: %v", err)
	}
	if decoded.Sequence != 999 {
		t.Errorf("Sequence = %d, want 999", decoded.Sequence)
	}
	if decoded.SessionID != 0x1234567890ABCDEF {
		t.Errorf("SessionID = 0x%X, want 0x1234567890ABCDEF", decoded.SessionID)
	}
	if decoded.Nonce != nonce {
		t.Errorf("Nonce = %v, want %v", decoded.Nonce, nonce)
	}
	if decoded.KeyID != 7 {
		t.Errorf("KeyID = %d, want 7", decoded.KeyID)
	}
	if decoded.ECDHPublic != ecdhKey {
		t.Errorf("ECDHPublic mismatch")
	}
}

func TestHeaderMarshalUnmarshalMaxSecurity(t *testing.T) {
	h := Header{
		Version:   0,
		Tier:      TierMaxSecurity,
		OpCode:    OpSessionRotate,
		Sequence:  12345,
		SessionID: 0xAAAABBBBCCCCDDDD,
		Timestamp: 1700000000000,
		Nonce:     [12]byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120},
		KeyID:     255,
	}

	data := h.MarshalHeader()

	var decoded Header
	_, err := decoded.UnmarshalHeader(data)
	if err != nil {
		t.Fatalf("UnmarshalHeader: %v", err)
	}
	if decoded.Timestamp != 1700000000000 {
		t.Errorf("Timestamp = %d, want 1700000000000", decoded.Timestamp)
	}
	if decoded.Sequence != 12345 {
		t.Errorf("Sequence = %d, want 12345", decoded.Sequence)
	}
	if decoded.KeyID != 255 {
		t.Errorf("KeyID = %d, want 255", decoded.KeyID)
	}
}

func TestHeaderMarshalUnmarshalFragmented(t *testing.T) {
	h := Header{
		Version:    0,
		Tier:       TierChecksum,
		Fragmented: true,
		OpCode:     OpNop,
		FragmentInfo: &FragmentInfo{
			MessageID:      0x12345678,
			FragmentIndex:  3,
			TotalFragments: 10,
		},
	}

	data := h.MarshalHeader()

	var decoded Header
	_, err := decoded.UnmarshalHeader(data)
	if err != nil {
		t.Fatalf("UnmarshalHeader: %v", err)
	}
	if !decoded.Fragmented {
		t.Error("Fragmented should be true")
	}
	if decoded.FragmentInfo == nil {
		t.Fatal("FragmentInfo should not be nil")
	}
	if decoded.FragmentInfo.MessageID != 0x12345678 {
		t.Errorf("MessageID = 0x%X, want 0x12345678", decoded.FragmentInfo.MessageID)
	}
	if decoded.FragmentInfo.FragmentIndex != 3 {
		t.Errorf("FragmentIndex = %d, want 3", decoded.FragmentInfo.FragmentIndex)
	}
	if decoded.FragmentInfo.TotalFragments != 10 {
		t.Errorf("TotalFragments = %d, want 10", decoded.FragmentInfo.TotalFragments)
	}
}

func TestUnmarshalHeaderTooShort(t *testing.T) {
	var h Header
	_, err := h.UnmarshalHeader([]byte{0x00, 0x01})
	if err != ErrMessageTooShort {
		t.Errorf("UnmarshalHeader(2 bytes) = %v, want ErrMessageTooShort", err)
	}
}

func TestUnmarshalHeaderTierTruncated(t *testing.T) {
	// Authenticated tier requires 7 bytes total, give it only 5
	flags := byte(TierAuthenticated << FlagTierShift)
	data := []byte{flags, 0x00, 0x03, 0x00, 0x00}

	var h Header
	_, err := h.UnmarshalHeader(data)
	if err != ErrMessageTooShort {
		t.Errorf("UnmarshalHeader(truncated) = %v, want ErrMessageTooShort", err)
	}
}

func TestMessageMarshalUnmarshal(t *testing.T) {
	payload := []byte("hello, protocol")
	msg := NewMessage(TierPlaintext, OpKeepalive, payload)

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Message
	if err := decoded.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload = %q, want %q", decoded.Payload, payload)
	}
	if decoded.Header.OpCode != OpKeepalive {
		t.Errorf("OpCode = 0x%04X, want 0x%04X", decoded.Header.OpCode, OpKeepalive)
	}
}

func TestMessageMarshalUnmarshalWithTrailer(t *testing.T) {
	payload := []byte("authenticated data")
	msg := &Message{
		Header: Header{
			Version:  0,
			Tier:     TierAuthenticated,
			OpCode:   OpSessionInit,
			Sequence: 1,
		},
		Payload: payload,
		Trailer: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Message
	if err := decoded.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload = %q, want %q", decoded.Payload, payload)
	}
	if !bytes.Equal(decoded.Trailer, msg.Trailer) {
		t.Errorf("Trailer = %v, want %v", decoded.Trailer, msg.Trailer)
	}
}

func TestMessageMarshalPayloadTooLarge(t *testing.T) {
	msg := NewMessage(TierPlaintext, OpNop, make([]byte, MaxPayloadSize+1))
	_, err := msg.Marshal()
	if err != ErrPayloadTooLarge {
		t.Errorf("Marshal(too large) = %v, want ErrPayloadTooLarge", err)
	}
}

func TestMessageWriteTo(t *testing.T) {
	msg := NewMessage(TierPlaintext, OpNop, []byte("test"))

	var buf bytes.Buffer
	n, err := msg.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	expected := int64(msg.Header.HeaderSize() + len(msg.Payload))
	if n != expected {
		t.Errorf("WriteTo returned %d bytes, want %d", n, expected)
	}
}

func TestMessageReadWithLength(t *testing.T) {
	payload := []byte("read test")
	msg := NewMessage(TierPlaintext, OpKeepalive, payload)
	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	reader := bytes.NewReader(data)
	var decoded Message
	if err := decoded.ReadWithLength(reader, len(data)); err != nil {
		t.Fatalf("ReadWithLength: %v", err)
	}

	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload = %q, want %q", decoded.Payload, payload)
	}
}

func TestMessageReadWithLengthTooLarge(t *testing.T) {
	var m Message
	err := m.ReadWithLength(bytes.NewReader(nil), MaxMessageSize+1)
	if err != ErrPayloadTooLarge {
		t.Errorf("ReadWithLength(too large) = %v, want ErrPayloadTooLarge", err)
	}
}

func TestMessageString(t *testing.T) {
	msg := NewMessage(TierEncrypted, OpSessionInit, []byte("x"))
	s := msg.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage(TierPFS, OpKeyExchangeInit, []byte{0x01})
	if msg.Header.Tier != TierPFS {
		t.Errorf("Tier = %d, want %d", msg.Header.Tier, TierPFS)
	}
	if msg.Header.OpCode != OpKeyExchangeInit {
		t.Errorf("OpCode = 0x%04X, want 0x%04X", msg.Header.OpCode, OpKeyExchangeInit)
	}
	if msg.Header.Version != 0 {
		t.Errorf("Version = %d, want 0", msg.Header.Version)
	}
}
