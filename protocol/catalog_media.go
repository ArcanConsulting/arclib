// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk media operation codes (category 0x0A).

package protocol

// Media operations (category 0x0A).
const (
	OpMediaUpload           OpCode = 0x0A00 // Upload media
	OpMediaDownload         OpCode = 0x0A01 // Download media
	OpMediaGet              OpCode = 0x0A02 // Get media info
	OpMediaDelete           OpCode = 0x0A03 // Delete media
	OpMediaList             OpCode = 0x0A04 // List media
	OpMediaShare            OpCode = 0x0A10 // Share media
	OpMediaUnshare          OpCode = 0x0A11 // Unshare media
	OpMediaThumbnail        OpCode = 0x0A20 // Get thumbnail
	OpMediaStream           OpCode = 0x0A30 // Stream media
	OpMediaStreamStart      OpCode = 0x0A31 // Start streaming
	OpMediaStreamStop       OpCode = 0x0A32 // Stop streaming
	OpMediaTranscode        OpCode = 0x0A40 // Transcode request
	OpMediaAlbumCreate      OpCode = 0x0A50 // Create album
	OpMediaAlbumGet         OpCode = 0x0A51 // Get album
	OpMediaAlbumUpdate      OpCode = 0x0A52 // Update album
	OpMediaAlbumDelete      OpCode = 0x0A53 // Delete album
	OpMediaAlbumList        OpCode = 0x0A54 // List albums
	OpMediaAlbumAddMedia    OpCode = 0x0A55 // Add media to album
	OpMediaAlbumRemMedia    OpCode = 0x0A56 // Remove media from album
	OpMediaAlbumGetMedia    OpCode = 0x0A57 // Get media in album
	OpMediaPersonCreate     OpCode = 0x0A60 // Create person
	OpMediaPersonGet        OpCode = 0x0A61 // Get person
	OpMediaPersonUpdate     OpCode = 0x0A62 // Update person
	OpMediaPersonDelete     OpCode = 0x0A63 // Delete person
	OpMediaPersonList       OpCode = 0x0A64 // List persons
	OpMediaFaceDetect       OpCode = 0x0A70 // Detect faces in media
	OpMediaFaceGet          OpCode = 0x0A71 // Get face detection
	OpMediaFaceAssign       OpCode = 0x0A72 // Assign face to person
	OpMediaFaceListByPerson OpCode = 0x0A73 // List faces by person
	OpMediaTagAdd           OpCode = 0x0A80 // Add tag to media
	OpMediaTagRemove        OpCode = 0x0A81 // Remove tag from media
	OpMediaTagList          OpCode = 0x0A82 // List tags for media
	OpMediaTagSearch        OpCode = 0x0A83 // Search media by tag
	OpMediaShareLinkCreate  OpCode = 0x0A90 // Create share link
	OpMediaShareLinkGet     OpCode = 0x0A91 // Get share link
	OpMediaShareLinkRevoke  OpCode = 0x0A92 // Revoke share link
	OpMediaShareLinkAccess  OpCode = 0x0A93 // Access share link
	OpMediaTimeline         OpCode = 0x0AA0 // Get media timeline
	OpMediaSearch           OpCode = 0x0AA1 // Search media
	OpMediaMemories         OpCode = 0x0AA2 // Get memories/on-this-day
)
