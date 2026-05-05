package protocol

import "encoding/binary"

// ExtensionType identifies the type of an extension header.
type ExtensionType uint16

// Defined extension types.
const (
	ExtTerminator    ExtensionType = 0x0000 // Sentinel: end of extensions
	ExtReplyTo       ExtensionType = 0x0001 // Request/response correlation (uint32 sequence)
	ExtTargetService ExtensionType = 0x0002 // Service routing name (UTF-8 string)
)

// Extensions holds parsed extension headers as a type-to-value map.
type Extensions struct {
	entries map[ExtensionType][]byte
}

// ReplyTo returns the correlated sequence number, if present.
func (e *Extensions) ReplyTo() (uint32, bool) {
	v, ok := e.entries[ExtReplyTo]
	if !ok || len(v) != 4 {
		return 0, false
	}
	return binary.BigEndian.Uint32(v), true
}

// SetReplyTo sets the ReplyTo extension to the given sequence number.
func (e *Extensions) SetReplyTo(seq uint32) {
	if e.entries == nil {
		e.entries = make(map[ExtensionType][]byte)
	}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, seq)
	e.entries[ExtReplyTo] = buf
}

// TargetService returns the routing service name, if present.
func (e *Extensions) TargetService() (string, bool) {
	v, ok := e.entries[ExtTargetService]
	if !ok || len(v) == 0 {
		return "", false
	}
	return string(v), true
}

// SetTargetService sets the TargetService extension.
func (e *Extensions) SetTargetService(name string) {
	if e.entries == nil {
		e.entries = make(map[ExtensionType][]byte)
	}
	e.entries[ExtTargetService] = []byte(name)
}

// Size returns the total wire size of the serialized extensions.
// Each entry is 4 bytes (type + length) plus the value length.
// Returns 0 when there are no extensions (no bytes emitted).
func (e *Extensions) Size() int {
	if len(e.entries) == 0 {
		return 0
	}
	size := 0
	for _, v := range e.entries {
		size += 4 + len(v) // type(2) + length(2) + value
	}
	return size
}

// Marshal serializes extensions into TLV wire format.
// Returns nil when there are no extensions.
func (e *Extensions) Marshal() []byte {
	if len(e.entries) == 0 {
		return nil
	}

	buf := make([]byte, e.Size())
	offset := 0
	for typ, val := range e.entries {
		binary.BigEndian.PutUint16(buf[offset:], uint16(typ))
		offset += 2
		binary.BigEndian.PutUint16(buf[offset:], uint16(len(val))) //nolint:gosec // extension values are bounded by TLV 16-bit length field
		offset += 2
		copy(buf[offset:], val)
		offset += len(val)
	}
	return buf
}

// ParseExtensions parses a TLV extension chain from data.
// It returns the parsed extensions, the number of bytes consumed, and any error.
// Parsing stops when a terminator type (0x0000) is encountered or all bytes are consumed.
func ParseExtensions(data []byte) (Extensions, int, error) {
	ext := Extensions{entries: make(map[ExtensionType][]byte)}
	offset := 0

	for offset+4 <= len(data) {
		typ := ExtensionType(binary.BigEndian.Uint16(data[offset:]))
		if typ == ExtTerminator {
			offset += 4 // consume the full terminator TLV header
			break
		}

		length := binary.BigEndian.Uint16(data[offset+2:])
		offset += 4

		if offset+int(length) > len(data) {
			return Extensions{}, 0, ErrMessageTooShort
		}

		val := make([]byte, length)
		copy(val, data[offset:offset+int(length)])
		ext.entries[typ] = val
		offset += int(length)
	}

	return ext, offset, nil
}
