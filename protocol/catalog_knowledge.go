// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk knowledge operation codes (category 0x07).

package protocol

// Knowledge operations (category 0x07).
const (
	OpKnowledgeNodeCreate         OpCode = 0x0700 // Create knowledge node
	OpKnowledgeNodeGet            OpCode = 0x0701 // Get knowledge node
	OpKnowledgeNodeUpdate         OpCode = 0x0702 // Update knowledge node
	OpKnowledgeNodeDelete         OpCode = 0x0703 // Delete knowledge node (soft delete)
	OpKnowledgeNodeList           OpCode = 0x0704 // List knowledge nodes
	OpKnowledgeNodeRestore        OpCode = 0x0705 // Restore deleted node
	OpKnowledgeBlockAdd           OpCode = 0x0710 // Add block to node
	OpKnowledgeBlockUpdate        OpCode = 0x0711 // Update block content
	OpKnowledgeBlockDelete        OpCode = 0x0712 // Delete block
	OpKnowledgeBlockMove          OpCode = 0x0713 // Move/reorder block
	OpKnowledgeEdgeCreate         OpCode = 0x0720 // Create knowledge edge
	OpKnowledgeEdgeDelete         OpCode = 0x0721 // Delete knowledge edge
	OpKnowledgeRelatedGet         OpCode = 0x0722 // Get related nodes
	OpKnowledgeSearch             OpCode = 0x0730 // Full-text search
	OpKnowledgeSemanticSearch     OpCode = 0x0731 // Semantic/vector search
	OpKnowledgeVersionList        OpCode = 0x0740 // List node versions
	OpKnowledgeVersionRestore     OpCode = 0x0741 // Restore to version
	OpKnowledgeAccessCheck        OpCode = 0x0750 // Check access to knowledge node
	OpKnowledgeAccessPolicyCreate OpCode = 0x0751 // Create access policy
	OpKnowledgeAccessPolicyGet    OpCode = 0x0752 // Get access policy
	OpKnowledgeAccessPolicyUpdate OpCode = 0x0753 // Update access policy
	OpKnowledgeAccessPolicyDelete OpCode = 0x0754 // Delete access policy
	OpKnowledgeAccessPolicyList   OpCode = 0x0755 // List access policies
	OpKnowledgeAccessLogList      OpCode = 0x0756 // List access log entries
)
