// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk system operation codes (category 0x00).

package protocol

// System operations (category 0x00).
const (
	OpDiscover      OpCode = 0x0020 // Service discovery
	OpDiscoverReply OpCode = 0x0021 // Discovery response
	OpCapabilities  OpCode = 0x0022 // Query capabilities
	OpVersion       OpCode = 0x0023 // Protocol version query
)
