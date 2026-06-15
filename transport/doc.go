// Package transport implements the MyClerk-Protocol transport layer: a WebSocket
// Server and Client that perform the hybrid X25519 + ML-KEM-768 handshake
// (SESSION_INIT / SESSION_ACK per draft-myclerk-protocol Section 6.1) and then
// exchange Tier-3 (ChaCha20-Poly1305) frames carrying MessagePack payloads.
//
// It is the single generic transport shared by MyClerk core and mayaservices,
// built on arclib/protocol (codec, AeadSession) and arclib/crypto. Application
// opcodes and payload types live in their own packages (e.g. arclib/llm); this
// package is payload-agnostic — handlers receive decrypted (opcode, payload)
// pairs and reply with their own opcode, optionally streaming via Session.Send.
package transport
