// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk room operation codes (category 0x09).

package protocol

// Room operations (category 0x09).
const (
	OpRoomCreate             OpCode = 0x0900 // Create room
	OpRoomGet                OpCode = 0x0901 // Get room info
	OpRoomUpdate             OpCode = 0x0902 // Update room
	OpRoomDelete             OpCode = 0x0903 // Delete room
	OpRoomList               OpCode = 0x0904 // List rooms
	OpRoomJoin               OpCode = 0x0910 // Join room
	OpRoomLeave              OpCode = 0x0911 // Leave room
	OpRoomInvite             OpCode = 0x0912 // Invite to room
	OpRoomKick               OpCode = 0x0913 // Kick from room
	OpRoomBan                OpCode = 0x0914 // Ban from room
	OpRoomUnban              OpCode = 0x0915 // Unban from room
	OpRoomMute               OpCode = 0x0916 // Mute room
	OpRoomUnmute             OpCode = 0x0917 // Unmute room
	OpRoomMembers            OpCode = 0x0920 // List members
	OpRoomMemberUpdate       OpCode = 0x0921 // Update member
	OpRoomSettings           OpCode = 0x0930 // Room settings
	OpKidsChatRoomCreate     OpCode = 0x0940 // Create kids chat room
	OpKidsChatRoomGet        OpCode = 0x0941 // Get kids chat room
	OpKidsChatRoomUpdate     OpCode = 0x0942 // Update kids chat room
	OpKidsChatRoomDelete     OpCode = 0x0943 // Delete kids chat room
	OpKidsChatRoomList       OpCode = 0x0944 // List kids chat rooms
	OpKidsChatRoomLock       OpCode = 0x0945 // Lock kids chat room
	OpKidsChatRoomUnlock     OpCode = 0x0946 // Unlock kids chat room
	OpKidsChatSendMessage    OpCode = 0x0947 // Send kids chat message
	OpKidsChatGetMessages    OpCode = 0x0948 // Get kids chat messages
	OpKidsChatReport         OpCode = 0x0949 // Report kids chat message
	OpKidsChatReviewReport   OpCode = 0x094A // Review kids chat report
	OpKidsChatSettings       OpCode = 0x094B // Get kids chat settings
	OpKidsChatSettingsUpdate OpCode = 0x094C // Update kids chat settings
)
