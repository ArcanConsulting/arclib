// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk family operation codes (category 0x03).

package protocol

// Family operations (category 0x03).
const (
	OpFamilyCreate    OpCode = 0x0300 // Create family
	OpFamilyGet       OpCode = 0x0301 // Get family info
	OpFamilyUpdate    OpCode = 0x0302 // Update family
	OpFamilyDelete    OpCode = 0x0303 // Delete family
	OpFamilyList      OpCode = 0x0304 // List families
	OpMemberAdd       OpCode = 0x0310 // Add member
	OpMemberRemove    OpCode = 0x0311 // Remove member
	OpMemberUpdate    OpCode = 0x0312 // Update member role
	OpMemberList      OpCode = 0x0313 // List members
	OpMemberInvite    OpCode = 0x0314 // Invite to family
	OpMemberAccept    OpCode = 0x0315 // Accept invitation
	OpMemberDecline   OpCode = 0x0316 // Decline invitation
	OpRoleCreate      OpCode = 0x0320 // Create role
	OpRoleUpdate      OpCode = 0x0321 // Update role
	OpRoleDelete      OpCode = 0x0322 // Delete role
	OpRoleList        OpCode = 0x0323 // List roles
	OpRoleAssign      OpCode = 0x0324 // Assign role
	OpPermissionCheck OpCode = 0x0330 // Check permission
)
