// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk audio operation codes (category 0x1C).

package protocol

// Audio operations (category 0x1C).
const (
	OpAudioZoneCreate           OpCode = 0x1C00 // Create audio zone
	OpAudioZoneGet              OpCode = 0x1C01 // Get audio zone
	OpAudioZoneUpdate           OpCode = 0x1C02 // Update audio zone
	OpAudioZoneDelete           OpCode = 0x1C03 // Delete audio zone
	OpAudioZoneList             OpCode = 0x1C04 // List zones by family
	OpAudioZoneListByType       OpCode = 0x1C05 // List zones by type
	OpAudioZoneAddNode          OpCode = 0x1C06 // Add node to zone
	OpAudioZoneRemoveNode       OpCode = 0x1C07 // Remove node from zone
	OpAudioZoneSetMaster        OpCode = 0x1C08 // Set zone master node
	OpAudioZoneSetVolume        OpCode = 0x1C09 // Set zone volume
	OpAudioZoneSetMute          OpCode = 0x1C0A // Set zone mute state
	OpAudioNodeRegister         OpCode = 0x1C10 // Register audio node
	OpAudioNodeGet              OpCode = 0x1C11 // Get audio node
	OpAudioNodeGetByDevice      OpCode = 0x1C12 // Get node by device ID
	OpAudioNodeUpdate           OpCode = 0x1C13 // Update audio node
	OpAudioNodeDelete           OpCode = 0x1C14 // Delete audio node
	OpAudioNodeList             OpCode = 0x1C15 // List nodes by family
	OpAudioNodeListByZone       OpCode = 0x1C16 // List nodes in zone
	OpAudioNodeListByRoom       OpCode = 0x1C17 // List nodes in room
	OpAudioNodeListOnline       OpCode = 0x1C18 // List online nodes
	OpAudioNodeUpdateStatus     OpCode = 0x1C19 // Update node status
	OpAudioNodeHeartbeat        OpCode = 0x1C1A // Node heartbeat
	OpAudioNodeSetZone          OpCode = 0x1C1B // Assign node to zone
	OpAudioNodeSetVolume        OpCode = 0x1C1C // Set node volume
	OpAudioStreamCreate         OpCode = 0x1C20 // Create audio stream
	OpAudioStreamGet            OpCode = 0x1C21 // Get audio stream
	OpAudioStreamUpdate         OpCode = 0x1C22 // Update audio stream
	OpAudioStreamDelete         OpCode = 0x1C23 // Delete audio stream
	OpAudioStreamListZone       OpCode = 0x1C24 // List streams in zone
	OpAudioStreamListState      OpCode = 0x1C25 // List streams by state
	OpAudioStreamActive         OpCode = 0x1C26 // List active streams in zone
	OpAudioStreamSetState       OpCode = 0x1C27 // Update stream state
	OpAudioStreamSetPos         OpCode = 0x1C28 // Update stream position
	OpAudioStreamPurge          OpCode = 0x1C29 // Purge old streams
	OpAudioStreamControl        OpCode = 0x1C2A // Stream control (play/pause/stop/seek)
	OpAudioStreamVolume         OpCode = 0x1C2B // Set stream volume
	OpAudioTTSCacheGet          OpCode = 0x1C30 // Get TTS cache entry
	OpAudioTTSCacheByVoice      OpCode = 0x1C31 // List cache by voice
	OpAudioTTSCacheAccess       OpCode = 0x1C32 // Update cache access
	OpAudioTTSCacheDelete       OpCode = 0x1C33 // Delete cache entry
	OpAudioTTSCachePurge        OpCode = 0x1C34 // Purge old cache entries
	OpAudioTTSCacheStats        OpCode = 0x1C35 // Get cache statistics
	OpAudioTTSCacheCreate       OpCode = 0x1C36 // Create cache entry
	OpAudioTransCreate          OpCode = 0x1C40 // Create transcription
	OpAudioTransGet             OpCode = 0x1C41 // Get transcription
	OpAudioTransListByUser      OpCode = 0x1C42 // List transcriptions by user
	OpAudioTransListByFamily    OpCode = 0x1C43 // List transcriptions by family
	OpAudioTransListByNode      OpCode = 0x1C44 // List transcriptions by node
	OpAudioTransDelete          OpCode = 0x1C45 // Delete transcription
	OpAudioTransPurge           OpCode = 0x1C46 // Purge old transcriptions
	OpAudioVoiceEventCreate     OpCode = 0x1C50 // Create voice event
	OpAudioVoiceEventListByNode OpCode = 0x1C51 // List voice events by node
	OpAudioVoiceEventListByType OpCode = 0x1C52 // List voice events by type
	OpAudioVoiceEventPurge      OpCode = 0x1C53 // Purge old voice events
)
