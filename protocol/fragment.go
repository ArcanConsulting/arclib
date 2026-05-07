package protocol

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Fragment reassembly constants.
const (
	// MaxFragments is the maximum number of fragments per message.
	MaxFragments = 256

	// FragmentTimeout is the default timeout for fragment reassembly.
	FragmentTimeout = 30 * time.Second
)

// Fragment errors.
var (
	ErrFragmentPending    = errors.New("fragment pending, message not yet complete")
	ErrDuplicateFragment  = errors.New("duplicate fragment received")
	ErrTooManyFragments   = errors.New("too many fragments")
	ErrFragmentOutOfRange = errors.New("fragment index out of range")
	ErrMismatchedTotal    = errors.New("fragment total mismatch")
)

// Fragmenter splits large messages into fragments.
type Fragmenter struct {
	nextMessageID atomic.Uint32
}

// NewFragmenter creates a new Fragmenter.
func NewFragmenter() *Fragmenter {
	return &Fragmenter{}
}

// nextID returns the next unique message ID using atomic increment.
func (f *Fragmenter) nextID() uint32 {
	return f.nextMessageID.Add(1)
}

// Fragment splits a message into fragments if it exceeds MaxFragmentSize.
// If the payload fits in a single fragment, the original message is returned as-is.
func (f *Fragmenter) Fragment(msg *Message) ([]*Message, error) {
	if len(msg.Payload) <= MaxFragmentSize {
		return []*Message{msg}, nil
	}

	payload := msg.Payload
	numFragments := (len(payload) + MaxFragmentSize - 1) / MaxFragmentSize
	if numFragments > MaxFragments {
		return nil, ErrTooManyFragments
	}

	messageID := f.nextID()
	fragments := make([]*Message, numFragments)
	totalFrags := uint16(numFragments) //nolint:gosec // bounds checked above (max 256)

	for i := range numFragments {
		start := i * MaxFragmentSize
		end := start + MaxFragmentSize
		if end > len(payload) {
			end = len(payload)
		}

		fragIdx := uint16(i) //nolint:gosec // bounds checked above (max 256)

		frag := &Message{
			Header: Header{
				Version:       msg.Header.Version,
				Tier:          msg.Header.Tier,
				Compressed:    msg.Header.Compressed,
				Fragmented:    true,
				HasExtensions: msg.Header.HasExtensions,
				OpCode:        msg.Header.OpCode,
				Sequence:      msg.Header.Sequence,
				SessionID:     msg.Header.SessionID,
				Timestamp:     msg.Header.Timestamp,
				Nonce:         msg.Header.Nonce,
				KeyID:         msg.Header.KeyID,
				ECDHPublic:    msg.Header.ECDHPublic,
				FragmentInfo: &FragmentInfo{
					MessageID:      messageID,
					FragmentIndex:  fragIdx,
					TotalFragments: totalFrags,
				},
			},
			Payload: make([]byte, end-start),
		}
		copy(frag.Payload, payload[start:end])
		fragments[i] = frag
	}

	return fragments, nil
}

// fragmentBuffer holds fragments for a single message being reassembled.
type fragmentBuffer struct {
	fragments []*Message
	received  int
	total     uint16
	header    Header
	createdAt time.Time
}

// FragmentAssembler reassembles fragmented messages.
type FragmentAssembler struct {
	mu      sync.Mutex
	pending map[uint32]*fragmentBuffer
}

// NewFragmentAssembler creates a new FragmentAssembler.
func NewFragmentAssembler() *FragmentAssembler {
	return &FragmentAssembler{
		pending: make(map[uint32]*fragmentBuffer),
	}
}

// AddFragment adds a fragment to the assembler. Returns the completed message
// when all fragments have been received, or nil with ErrFragmentPending if more
// fragments are needed. Returns an error for invalid fragments.
func (fa *FragmentAssembler) AddFragment(msg *Message) (*Message, error) {
	if !msg.Header.Fragmented || msg.Header.FragmentInfo == nil {
		return msg, nil
	}

	info := msg.Header.FragmentInfo

	if info.TotalFragments == 0 || info.TotalFragments > MaxFragments {
		return nil, ErrTooManyFragments
	}

	if info.FragmentIndex >= info.TotalFragments {
		return nil, ErrFragmentOutOfRange
	}

	fa.mu.Lock()
	defer fa.mu.Unlock()

	buf, exists := fa.pending[info.MessageID]
	if !exists {
		buf = &fragmentBuffer{
			fragments: make([]*Message, info.TotalFragments),
			total:     info.TotalFragments,
			header:    msg.Header,
			createdAt: time.Now(),
		}
		fa.pending[info.MessageID] = buf
	}

	// Verify total matches
	if info.TotalFragments != buf.total {
		return nil, ErrMismatchedTotal
	}

	// Detect duplicates
	if buf.fragments[info.FragmentIndex] != nil {
		return nil, ErrDuplicateFragment
	}

	buf.fragments[info.FragmentIndex] = msg
	buf.received++

	if buf.received < int(buf.total) {
		return nil, ErrFragmentPending
	}

	// All fragments received — reassemble
	totalSize := 0
	for _, frag := range buf.fragments {
		totalSize += len(frag.Payload)
	}

	payload := make([]byte, 0, totalSize)
	for _, frag := range buf.fragments {
		payload = append(payload, frag.Payload...)
	}

	// Use header from fragment 0
	hdr := buf.fragments[0].Header
	hdr.Fragmented = false
	hdr.FragmentInfo = nil

	complete := &Message{
		Header:  hdr,
		Payload: payload,
	}

	delete(fa.pending, info.MessageID)

	return complete, nil
}

// Cleanup removes stale fragment buffers older than maxAge.
// Returns the number of buffers removed.
func (fa *FragmentAssembler) Cleanup(maxAge time.Duration) int {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	now := time.Now()
	removed := 0

	for id, buf := range fa.pending {
		if now.Sub(buf.createdAt) > maxAge {
			delete(fa.pending, id)
			removed++
		}
	}

	return removed
}
