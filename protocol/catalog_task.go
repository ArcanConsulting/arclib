// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk task operation codes (category 0x0D).

package protocol

// Task operations (category 0x0D).
const (
	OpTaskCreate      OpCode = 0x0D00 // Create task
	OpTaskGet         OpCode = 0x0D01 // Get task
	OpTaskUpdate      OpCode = 0x0D02 // Update task
	OpTaskDelete      OpCode = 0x0D03 // Delete task
	OpTaskList        OpCode = 0x0D04 // List tasks
	OpTaskSkip        OpCode = 0x0D05 // Skip task
	OpTaskUnassign    OpCode = 0x0D06 // Unassign task
	OpTaskAgenda      OpCode = 0x0D07 // Daily agenda
	OpTaskStats       OpCode = 0x0D08 // Task stats
	OpTaskComplete    OpCode = 0x0D10 // Mark complete
	OpTaskAssign      OpCode = 0x0D11 // Assign task
	OpTaskPriority    OpCode = 0x0D12 // Set priority
	OpListCreate      OpCode = 0x0D20 // Create list
	OpListGet         OpCode = 0x0D21 // Get list
	OpListUpdate      OpCode = 0x0D22 // Update list
	OpListDelete      OpCode = 0x0D23 // Delete list
	OpListShare       OpCode = 0x0D24 // Share list
	OpListList        OpCode = 0x0D25 // List all lists
	OpListUnshare     OpCode = 0x0D26 // Remove share
	OpListArchive     OpCode = 0x0D27 // Archive list
	OpItemCreate      OpCode = 0x0D30 // Add item to list
	OpItemGet         OpCode = 0x0D31 // Get item
	OpItemUpdate      OpCode = 0x0D32 // Update item
	OpItemDelete      OpCode = 0x0D33 // Delete item
	OpItemCheck       OpCode = 0x0D34 // Check/uncheck item
	OpItemMove        OpCode = 0x0D35 // Move item to another list
	OpItemList        OpCode = 0x0D36 // List items in a list
	OpChallengeCreate OpCode = 0x0D40 // Create challenge
	OpChallengeGet    OpCode = 0x0D41 // Get challenge
	OpChallengeList   OpCode = 0x0D42 // List challenges
	OpChallengeJoin   OpCode = 0x0D43 // Join challenge
	OpRewardCreate    OpCode = 0x0D50 // Create reward
	OpRewardGet       OpCode = 0x0D51 // Get reward
	OpRewardList      OpCode = 0x0D52 // List rewards
	OpRewardClaim     OpCode = 0x0D53 // Claim reward
	OpRewardFulfill   OpCode = 0x0D54 // Fulfill reward claim
	OpRotationCreate  OpCode = 0x0D60 // Create rotation schedule
	OpRotationGet     OpCode = 0x0D61 // Get rotation schedule
	OpRotationUpdate  OpCode = 0x0D62 // Update rotation schedule
	OpRotationDelete  OpCode = 0x0D63 // Delete rotation schedule
	OpRotationList    OpCode = 0x0D64 // List rotation schedules
	OpRotationNext    OpCode = 0x0D65 // Get next assignee
	OpRotationAssign  OpCode = 0x0D66 // Assign next rotation
	OpRotationSkip    OpCode = 0x0D67 // Skip rotation turn
	OpRotationSwap    OpCode = 0x0D68 // Swap rotation turns
	OpRotationStats   OpCode = 0x0D69 // Get rotation statistics
)
