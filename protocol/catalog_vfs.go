// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk vfs operation codes (category 0x05).

package protocol

// Vfs operations (category 0x05).
const (
	OpVFSMkdir                  OpCode = 0x050A // Create directory
	OpVFSRmdir                  OpCode = 0x050B // Remove directory
	OpVFSTruncate               OpCode = 0x050C // Truncate file
	OpVFSCopy                   OpCode = 0x050D // Copy file
	OpVFSQuota                  OpCode = 0x050E // Get quota info
	OpVFSWatch                  OpCode = 0x050F // Watch for changes
	OpFragmentStore             OpCode = 0x0520 // Store fragment
	OpFragmentRetrieve          OpCode = 0x0521 // Retrieve fragment
	OpFragmentStatus            OpCode = 0x0522 // Get fragment status
	OpFragmentRedistribute      OpCode = 0x0523 // Redistribute fragments
	OpFragmentHealthReport      OpCode = 0x0524 // Fragment health report
	OpMetaSyncRequest           OpCode = 0x0530 // Request metadata sync
	OpMetaSyncDiff              OpCode = 0x0531 // Send metadata diff
	OpMetaConflictResolve       OpCode = 0x0532 // Resolve metadata conflict
	OpMetaSnapshot              OpCode = 0x0533 // Create metadata snapshot
	OpMetaRestore               OpCode = 0x0534 // Restore from snapshot
	OpVFSShareCreate            OpCode = 0x0550 // Create share
	OpVFSShareRevoke            OpCode = 0x0551 // Revoke share
	OpVFSShareList              OpCode = 0x0552 // List shares
	OpVFSShareAccess            OpCode = 0x0553 // Access shared resource
	OpVFSShareSync              OpCode = 0x0554 // Sync shared resource
	OpVFSShareUpdate            OpCode = 0x0555 // Update share permissions
	OpVFSShareAccept            OpCode = 0x0556 // Accept share invitation
	OpVFSCommonFolderCreate     OpCode = 0x0557 // Create common folder
	OpVFSCommonFolderJoin       OpCode = 0x0558 // Join common folder
	OpVFSCommonFolderLeave      OpCode = 0x0559 // Leave common folder
	OpVFSCommonFolderSync       OpCode = 0x055A // Force common folder sync
	OpVFSEmergencyRevoke        OpCode = 0x05C0 // Immediate access revocation
	OpVFSEmergencyKeyInvalidate OpCode = 0x05C1 // Invalidate all keys
	OpVFSEmergencyCacheWipe     OpCode = 0x05C2 // Request cache wipe
	OpVFSEmergencyConfirm       OpCode = 0x05C3 // Client confirmation
	OpVFSEmergencyStatus        OpCode = 0x05C4 // Query cut-off status
	OpVFSEmergencyRestore       OpCode = 0x05C5 // Restore access
)
