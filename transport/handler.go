package transport

import (
	"context"
	"fmt"
	"sync"

	"arcan-it.de/arclib/protocol"
)

// Handler processes one decrypted message within a session. A non-nil returned
// message is sealed and sent back to the peer under its own opcode. Returning
// (nil, nil) means the handler has already replied itself (for example a
// streaming handler that called Session.Send for each chunk) or that no reply is
// warranted. A non-nil error is reported to the peer as an OpError frame.
type Handler interface {
	Handle(ctx context.Context, s *Session, msg *protocol.Message) (*protocol.Message, error)
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(ctx context.Context, s *Session, msg *protocol.Message) (*protocol.Message, error)

// Handle implements Handler.
func (f HandlerFunc) Handle(ctx context.Context, s *Session, msg *protocol.Message) (*protocol.Message, error) {
	return f(ctx, s, msg)
}

// HandlerRegistry maps opcodes to handlers. It is safe for concurrent use.
type HandlerRegistry struct {
	mu sync.RWMutex
	m  map[protocol.OpCode]Handler
}

// NewHandlerRegistry returns an empty registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{m: make(map[protocol.OpCode]Handler)}
}

// Register binds a handler to an opcode. It errors if h is nil or the opcode is
// already registered.
func (r *HandlerRegistry) Register(op protocol.OpCode, h Handler) error {
	if h == nil {
		return fmt.Errorf("transport: nil handler for %s", op)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.m[op]; exists {
		return fmt.Errorf("transport: handler already registered for %s", op)
	}
	r.m[op] = h
	return nil
}

// RegisterFunc binds a HandlerFunc to an opcode.
func (r *HandlerRegistry) RegisterFunc(op protocol.OpCode, f HandlerFunc) error {
	return r.Register(op, f)
}

// Has reports whether an opcode has a registered handler.
func (r *HandlerRegistry) Has(op protocol.OpCode) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.m[op]
	return ok
}

// Count returns the number of registered handlers.
func (r *HandlerRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.m)
}

// dispatch routes a message to its handler.
func (r *HandlerRegistry) dispatch(ctx context.Context, s *Session, msg *protocol.Message) (*protocol.Message, error) {
	r.mu.RLock()
	h, ok := r.m[msg.Header.OpCode]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("transport: no handler for %s", msg.Header.OpCode)
	}
	return h.Handle(ctx, s, msg)
}
