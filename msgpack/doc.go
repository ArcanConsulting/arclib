// Package msgpack provides MessagePack serialization for the MyClerk Protocol.
//
// It wraps github.com/vmihailenco/msgpack/v5 behind a thin, consistent API
// used throughout arclib for encoding protocol payloads (ErrorPayload,
// SessionInit, etc.).
//
// Basic usage:
//
//	data, err := msgpack.Marshal(payload)
//	if err != nil { ... }
//
//	var out MyStruct
//	if err := msgpack.Unmarshal(data, &out); err != nil { ... }
//
// Streaming usage:
//
//	enc := msgpack.NewEncoder(conn)
//	enc.Encode(payload)
//
//	dec := msgpack.NewDecoder(conn)
//	dec.Decode(&out)
package msgpack
