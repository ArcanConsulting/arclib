// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk automation operation codes (category 0x0F).

package protocol

// Automation operations (category 0x0F).
const (
	OpRuleCreate          OpCode = 0x0F00 // Create rule
	OpRuleGet             OpCode = 0x0F01 // Get rule
	OpRuleUpdate          OpCode = 0x0F02 // Update rule
	OpRuleDelete          OpCode = 0x0F03 // Delete rule
	OpRuleList            OpCode = 0x0F04 // List rules
	OpRuleEnable          OpCode = 0x0F10 // Enable rule
	OpRuleDisable         OpCode = 0x0F11 // Disable rule
	OpRuleTrigger         OpCode = 0x0F12 // Manual trigger
	OpRuleTest            OpCode = 0x0F13 // Test rule
	OpTriggerEvent        OpCode = 0x0F20 // Trigger event
	OpActionExec          OpCode = 0x0F21 // Execute action
	OpRuleLogGet          OpCode = 0x0F30 // Get automation execution logs
	OpRuleLogList         OpCode = 0x0F31 // List all execution logs for family
	OpTemplateCreate      OpCode = 0x0F40 // Create automation template
	OpTemplateGet         OpCode = 0x0F41 // Get automation template
	OpTemplateUpdate      OpCode = 0x0F42 // Update automation template
	OpTemplateDelete      OpCode = 0x0F43 // Delete automation template
	OpTemplateList        OpCode = 0x0F44 // List automation templates
	OpTemplateInstantiate OpCode = 0x0F45 // Instantiate rule from template
	OpPatternEvaluate     OpCode = 0x0F50 // Evaluate state pattern against history
)
