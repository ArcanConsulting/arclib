// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk tenant operation codes (category 0x1E).

package protocol

// Tenant operations (category 0x1E).
const (
	OpTenantCreate         OpCode = 0x1E00 // Create tenant
	OpTenantGet            OpCode = 0x1E01 // Get tenant info
	OpTenantUpdate         OpCode = 0x1E02 // Update tenant
	OpTenantDelete         OpCode = 0x1E03 // Delete (soft) tenant
	OpTenantList           OpCode = 0x1E04 // List user's tenants
	OpTenantSwitch         OpCode = 0x1E05 // Switch active tenant (returns new JWT)
	OpTenantSuspend        OpCode = 0x1E06 // Suspend tenant
	OpTenantActivate       OpCode = 0x1E07 // Activate tenant
	OpTenantMemberAdd      OpCode = 0x1E10 // Add member to tenant
	OpTenantMemberRemove   OpCode = 0x1E11 // Remove member from tenant
	OpTenantMemberUpdate   OpCode = 0x1E12 // Update member role
	OpTenantMemberList     OpCode = 0x1E13 // List tenant members
	OpTenantInviteCreate   OpCode = 0x1E20 // Create invitation
	OpTenantInviteAccept   OpCode = 0x1E21 // Accept invitation
	OpTenantInviteDecline  OpCode = 0x1E22 // Decline invitation
	OpTenantInviteList     OpCode = 0x1E23 // List invitations
	OpTenantInviteRevoke   OpCode = 0x1E24 // Revoke invitation
	OpTenantQuotaGet       OpCode = 0x1E30 // Get quota limits
	OpTenantQuotaUpdate    OpCode = 0x1E31 // Update quota (admin)
	OpTenantUsageGet       OpCode = 0x1E32 // Get current usage
	OpTenantBrandingGet    OpCode = 0x1E40 // Get branding config
	OpTenantBrandingUpdate OpCode = 0x1E41 // Update branding
)
