package protocol

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestFragmenter_SmallMessage(t *testing.T) {
	f := NewFragmenter()

	payload := make([]byte, 100)
	msg := NewMessage(TierPlaintext, OpCode(0x01), payload)

	fragments, err := f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fragments))
	}

	if fragments[0] != msg {
		t.Fatal("expected original message to be returned as-is")
	}

	if fragments[0].Header.Fragmented {
		t.Fatal("small message should not be marked as fragmented")
	}
}

func TestFragmenter_LargeMessage(t *testing.T) {
	f := NewFragmenter()

	// Create a payload slightly larger than 2 fragments
	payloadSize := MaxFragmentSize*2 + 100
	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	msg := NewMessage(TierAuthenticated, OpCode(0x42), payload)
	msg.Header.Sequence = 999

	fragments, err := f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFragments := 3
	if len(fragments) != expectedFragments {
		t.Fatalf("expected %d fragments, got %d", expectedFragments, len(fragments))
	}

	// Verify each fragment preserves header fields
	for i, frag := range fragments {
		if !frag.Header.Fragmented {
			t.Errorf("fragment %d: expected Fragmented=true", i)
		}
		if frag.Header.Tier != TierAuthenticated {
			t.Errorf("fragment %d: expected TierAuthenticated, got %v", i, frag.Header.Tier)
		}
		if frag.Header.OpCode != OpCode(0x42) {
			t.Errorf("fragment %d: expected OpCode 0x42, got 0x%04X", i, frag.Header.OpCode)
		}
		if frag.Header.Sequence != 999 {
			t.Errorf("fragment %d: expected Sequence=999, got %d", i, frag.Header.Sequence)
		}
		if frag.Header.FragmentInfo == nil {
			t.Fatalf("fragment %d: FragmentInfo is nil", i)
		}
		if frag.Header.FragmentInfo.FragmentIndex != uint16(i) {
			t.Errorf("fragment %d: expected index %d, got %d", i, i, frag.Header.FragmentInfo.FragmentIndex)
		}
		if frag.Header.FragmentInfo.TotalFragments != uint16(expectedFragments) {
			t.Errorf("fragment %d: expected total %d, got %d", i, expectedFragments, frag.Header.FragmentInfo.TotalFragments)
		}
	}

	// Verify combined payload matches original
	var combined []byte
	for _, frag := range fragments {
		combined = append(combined, frag.Payload...)
	}
	if !bytes.Equal(combined, payload) {
		t.Fatal("combined fragment payloads do not match original")
	}
}

func TestFragmentAssembler_InOrder(t *testing.T) {
	f := NewFragmenter()
	fa := NewFragmentAssembler()

	payloadSize := MaxFragmentSize*3 + 500
	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i % 251)
	}

	msg := NewMessage(TierEncrypted, OpCode(0x10), payload)
	msg.Header.SessionID = 12345

	fragments, err := f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Add fragments in order
	for i, frag := range fragments {
		result, err := fa.AddFragment(frag)
		if i < len(fragments)-1 {
			if !errors.Is(err, ErrFragmentPending) {
				t.Fatalf("fragment %d: expected ErrFragmentPending, got %v", i, err)
			}
			if result != nil {
				t.Fatalf("fragment %d: expected nil result", i)
			}
		} else {
			if err != nil {
				t.Fatalf("last fragment: unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("last fragment: expected completed message")
			}
			if !bytes.Equal(result.Payload, payload) {
				t.Fatal("reassembled payload does not match original")
			}
			if result.Header.Fragmented {
				t.Fatal("reassembled message should not be marked fragmented")
			}
			if result.Header.FragmentInfo != nil {
				t.Fatal("reassembled message should have nil FragmentInfo")
			}
			if result.Header.SessionID != 12345 {
				t.Errorf("expected SessionID=12345, got %d", result.Header.SessionID)
			}
			if result.Header.OpCode != OpCode(0x10) {
				t.Errorf("expected OpCode=0x10, got 0x%04X", result.Header.OpCode)
			}
		}
	}
}

func TestFragmentAssembler_OutOfOrder(t *testing.T) {
	f := NewFragmenter()
	fa := NewFragmentAssembler()

	payloadSize := MaxFragmentSize*4 + 1
	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i % 199)
	}

	msg := NewMessage(TierPlaintext, OpCode(0x20), payload)

	fragments, err := f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fragments) != 5 {
		t.Fatalf("expected 5 fragments, got %d", len(fragments))
	}

	// Add in reverse order: 4, 2, 0, 3, 1
	order := []int{4, 2, 0, 3, 1}
	for i, idx := range order {
		result, err := fa.AddFragment(fragments[idx])
		if i < len(order)-1 {
			if !errors.Is(err, ErrFragmentPending) {
				t.Fatalf("step %d (frag %d): expected ErrFragmentPending, got %v", i, idx, err)
			}
			if result != nil {
				t.Fatalf("step %d: expected nil result", i)
			}
		} else {
			if err != nil {
				t.Fatalf("last step: unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected completed message on last fragment")
			}
			if !bytes.Equal(result.Payload, payload) {
				t.Fatal("reassembled payload does not match original")
			}
		}
	}
}

func TestFragmentAssembler_DuplicateFragment(t *testing.T) {
	f := NewFragmenter()
	fa := NewFragmentAssembler()

	payloadSize := MaxFragmentSize + 1
	payload := make([]byte, payloadSize)
	msg := NewMessage(TierPlaintext, OpCode(0x01), payload)

	fragments, err := f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Add first fragment
	_, err = fa.AddFragment(fragments[0])
	if !errors.Is(err, ErrFragmentPending) {
		t.Fatalf("expected ErrFragmentPending, got %v", err)
	}

	// Add same fragment again
	_, err = fa.AddFragment(fragments[0])
	if !errors.Is(err, ErrDuplicateFragment) {
		t.Fatalf("expected ErrDuplicateFragment, got %v", err)
	}
}

func TestFragmentAssembler_Cleanup(t *testing.T) {
	fa := NewFragmentAssembler()

	// Manually insert a stale buffer
	fa.mu.Lock()
	fa.pending[1] = &fragmentBuffer{
		fragments: make([]*Message, 4),
		total:     4,
		received:  1,
		createdAt: time.Now().Add(-2 * time.Minute),
	}
	fa.pending[2] = &fragmentBuffer{
		fragments: make([]*Message, 2),
		total:     2,
		received:  1,
		createdAt: time.Now(),
	}
	fa.mu.Unlock()

	removed := fa.Cleanup(time.Minute)
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}

	fa.mu.Lock()
	if _, exists := fa.pending[1]; exists {
		t.Fatal("stale buffer should have been removed")
	}
	if _, exists := fa.pending[2]; !exists {
		t.Fatal("fresh buffer should still exist")
	}
	fa.mu.Unlock()
}

func TestFragmentAssembler_FragmentOutOfRange(t *testing.T) {
	fa := NewFragmentAssembler()

	msg := &Message{
		Header: Header{
			Fragmented: true,
			OpCode:     OpCode(0x01),
			FragmentInfo: &FragmentInfo{
				MessageID:      42,
				FragmentIndex:  5,
				TotalFragments: 3,
			},
		},
		Payload: []byte("data"),
	}

	_, err := fa.AddFragment(msg)
	if !errors.Is(err, ErrFragmentOutOfRange) {
		t.Fatalf("expected ErrFragmentOutOfRange, got %v", err)
	}
}

func TestFragmenter_TooManyFragments(t *testing.T) {
	f := NewFragmenter()

	// MaxFragments * MaxFragmentSize + 1 would require more than MaxFragments
	payloadSize := MaxFragments*MaxFragmentSize + 1
	payload := make([]byte, payloadSize)
	msg := NewMessage(TierPlaintext, OpCode(0x01), payload)

	_, err := f.Fragment(msg)
	if !errors.Is(err, ErrTooManyFragments) {
		t.Fatalf("expected ErrTooManyFragments, got %v", err)
	}
}

func TestFragmenter_ExactBoundary(t *testing.T) {
	f := NewFragmenter()

	// Exactly MaxFragmentSize should NOT fragment
	payload := make([]byte, MaxFragmentSize)
	msg := NewMessage(TierPlaintext, OpCode(0x01), payload)

	fragments, err := f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fragments) != 1 {
		t.Fatalf("expected 1 fragment for exact boundary, got %d", len(fragments))
	}

	// MaxFragmentSize + 1 SHOULD fragment
	payload = make([]byte, MaxFragmentSize+1)
	msg = NewMessage(TierPlaintext, OpCode(0x01), payload)

	fragments, err = f.Fragment(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fragments) != 2 {
		t.Fatalf("expected 2 fragments, got %d", len(fragments))
	}
}

func TestFragmentAssembler_MismatchedTotal(t *testing.T) {
	fa := NewFragmentAssembler()

	// First fragment says total=4
	msg1 := &Message{
		Header: Header{
			Fragmented: true,
			OpCode:     OpCode(0x01),
			FragmentInfo: &FragmentInfo{
				MessageID:      99,
				FragmentIndex:  0,
				TotalFragments: 4,
			},
		},
		Payload: []byte("first"),
	}

	_, err := fa.AddFragment(msg1)
	if !errors.Is(err, ErrFragmentPending) {
		t.Fatalf("expected ErrFragmentPending, got %v", err)
	}

	// Second fragment says total=3 (mismatch)
	msg2 := &Message{
		Header: Header{
			Fragmented: true,
			OpCode:     OpCode(0x01),
			FragmentInfo: &FragmentInfo{
				MessageID:      99,
				FragmentIndex:  1,
				TotalFragments: 3,
			},
		},
		Payload: []byte("second"),
	}

	_, err = fa.AddFragment(msg2)
	if !errors.Is(err, ErrMismatchedTotal) {
		t.Fatalf("expected ErrMismatchedTotal, got %v", err)
	}
}
