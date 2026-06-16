// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Opcode classification helpers migrated from MyClerk core.

package protocol

// IsACLOp returns true if the opcode is an ACL operation (0x0244-0x0247, 0x0249).
func (op OpCode) IsACLOp() bool {
	return (op >= 0x0244 && op <= 0x0247) || op == 0x0249
}

// IsAnalyticsOp returns true if the opcode is a Business Analytics operation (0x1F00-0x1FFF).
func (op OpCode) IsAnalyticsOp() bool {
	return op.Category() == CategoryAnalytics
}

// IsAudioOp returns true if the opcode is an Audio Pipeline operation (0x1C00-0x1CFF).
func (op OpCode) IsAudioOp() bool {
	return op.Category() == CategoryAudio
}

// IsBridgeOp returns true if the opcode is a Channel Bridge operation (0x0BB0-0x0BBC).
func (op OpCode) IsBridgeOp() bool {
	return op >= 0x0BB0 && op <= 0x0BBC
}

// IsCANOp returns true if the opcode is a CAN bus operation (0x0650-0x065F).
func (op OpCode) IsCANOp() bool {
	return op >= 0x0650 && op <= 0x065F
}

// IsCalendarGamificationOp returns true if the opcode is a Calendar Gamification operation (0x0C40-0x0C45).
func (op OpCode) IsCalendarGamificationOp() bool {
	return op >= 0x0C40 && op <= 0x0C45
}

// IsCallOp returns true if the opcode is a Call/Telephony operation (0x0B00-0x0B7F).
func (op OpCode) IsCallOp() bool {
	return op >= 0x0B00 && op <= 0x0B7F
}

// IsClientOp returns true if the opcode is a client assignment operation (0x0250-0x025F).
func (op OpCode) IsClientOp() bool {
	return op >= 0x0250 && op <= 0x025F
}

// IsCommunityFedOp returns true if the opcode is a Community Federation operation (0x2000-0x20FF).
func (op OpCode) IsCommunityFedOp() bool {
	return op.Category() == CategoryCommunityFed
}

// IsDesktopOp returns true if the opcode is a Desktop Client operation (0x1D00-0x1DFF).
func (op OpCode) IsDesktopOp() bool {
	return op.Category() == CategoryDesktop
}

// IsDeviceManagementOp returns true if the opcode is a shared device management operation (0x0200-0x020F).
func (op OpCode) IsDeviceManagementOp() bool {
	return op >= 0x0200 && op <= 0x020F
}

// IsDistributedOp returns true if the opcode is a Distributed Architecture operation (0x1140-0x11AF).
func (op OpCode) IsDistributedOp() bool {
	return op >= 0x1140 && op <= 0x11AF
}

// IsEducationOp returns true if the opcode is an Education operation (0x1380-0x13EF).
func (op OpCode) IsEducationOp() bool {
	return op >= 0x1380 && op <= 0x13EF
}

// IsEmergencyGatewayOp returns true if the opcode is an Emergency Gateway operation (0x0B70-0x0B7F).
func (op OpCode) IsEmergencyGatewayOp() bool {
	return op >= 0x0B70 && op <= 0x0B7F
}

// IsEventDetectionOp returns true if the opcode is an Event Detection operation (0x0C50-0x0C5F).
func (op OpCode) IsEventDetectionOp() bool {
	return op >= 0x0C50 && op <= 0x0C5F
}

// IsFederationOp returns true if the opcode is a federation service operation (0x0260-0x026F).
func (op OpCode) IsFederationOp() bool {
	return op >= 0x0260 && op <= 0x026F
}

// IsGPIOOp returns true if the opcode is a GPIO operation (0x0620-0x062F).
func (op OpCode) IsGPIOOp() bool {
	return op >= 0x0620 && op <= 0x062F
}

// IsGPUOp returns true if the opcode is a GPU service operation (0x0220-0x022F).
func (op OpCode) IsGPUOp() bool {
	return op >= 0x0220 && op <= 0x022F
}

// IsGSMModemOp returns true if the opcode is a GSM modem operation (0x0670-0x067F).
func (op OpCode) IsGSMModemOp() bool {
	return op >= 0x0670 && op <= 0x067F
}

// IsHardwareOp returns true if the opcode is a Hardware Passthrough operation (0x0600-0x06FF).
func (op OpCode) IsHardwareOp() bool {
	return op.Category() == CategoryHardware
}

// IsI2COp returns true if the opcode is an I2C operation (0x0630-0x063F).
func (op OpCode) IsI2COp() bool {
	return op >= 0x0630 && op <= 0x063F
}

// IsIETFStandardOp returns true if the opcode is in a standard range (0x0000-0x17FF).
func (op OpCode) IsIETFStandardOp() bool {
	cat := op.Category()
	return cat <= 0x17
}

// IsIVROp returns true if the opcode is an IVR operation (0x0BC0-0x0BCB).
func (op OpCode) IsIVROp() bool {
	return op >= 0x0BC0 && op <= 0x0BCB
}

// IsKidsChatOp returns true if the opcode is a Kids Chat operation (0x0940-0x094F).
func (op OpCode) IsKidsChatOp() bool {
	return op >= 0x0940 && op <= 0x094F
}

// IsKnowledgeOp returns true if the opcode is a Knowledge & Memory operation (0x0700-0x07FF).
func (op OpCode) IsKnowledgeOp() bool {
	return op.Category() == CategoryKnowledge
}

// IsMDMOp returns true if the opcode is an MDM operation (0x1700-0x17FF).
func (op OpCode) IsMDMOp() bool {
	return op.Category() == CategoryMDM
}

// IsMarketDataOp returns true if the opcode is a Market Data operation (0x3000-0x30FF).
func (op OpCode) IsMarketDataOp() bool {
	return op.Category() == CategoryMarketData
}

// IsOneWireOp returns true if the opcode is a 1-Wire operation (0x0660-0x066F).
func (op OpCode) IsOneWireOp() bool {
	return op >= 0x0660 && op <= 0x066F
}

// IsPluginOp returns true if the opcode is a plugin operation (0x1610-0x162F).
func (op OpCode) IsPluginOp() bool {
	return op >= 0x1610 && op <= 0x162F
}

// IsResourceSharingOp returns true if the opcode is a Resource Sharing operation (0x0200-0x02FF).
func (op OpCode) IsResourceSharingOp() bool {
	return op.Category() == CategoryResourceSharing
}

// IsRotationOp returns true if the opcode is a Task Rotation operation (0x0D60-0x0D69).
func (op OpCode) IsRotationOp() bool {
	return op >= 0x0D60 && op <= 0x0D69
}

// IsRouteOp returns true if the opcode is a routing operation (0x0240-0x0243, 0x0248).
func (op OpCode) IsRouteOp() bool {
	return (op >= 0x0240 && op <= 0x0243) || op == 0x0248
}

// IsSMSGatewayOp returns true if the opcode is an SMS Gateway operation (0x0B50-0x0B6F).
func (op OpCode) IsSMSGatewayOp() bool {
	return op >= 0x0B50 && op <= 0x0B6F
}

// IsSPIOp returns true if the opcode is an SPI operation (0x0640-0x064F).
func (op OpCode) IsSPIOp() bool {
	return op >= 0x0640 && op <= 0x064F
}

// IsScreeningOp returns true if the opcode is a Call Screening operation (0x0B40-0x0B4F).
func (op OpCode) IsScreeningOp() bool {
	return op >= 0x0B40 && op <= 0x0B4F
}

// IsSerialOp returns true if the opcode is a Serial operation (0x0610-0x061F).
func (op OpCode) IsSerialOp() bool {
	return op >= 0x0610 && op <= 0x061F
}

// IsStreamOp returns true if the opcode is a stream operation (0x0210-0x021F).
func (op OpCode) IsStreamOp() bool {
	return op >= 0x0210 && op <= 0x021F
}

// IsTenantOp returns true if the opcode is a Tenant Management operation (0x1E00-0x1EFF).
func (op OpCode) IsTenantOp() bool {
	return op.Category() == CategoryTenant
}

// IsUSBOp returns true if the opcode is a USB operation (0x0600-0x060F).
func (op OpCode) IsUSBOp() bool {
	return op >= 0x0600 && op <= 0x060F
}

// IsVFSBasicOp returns true if the opcode is a VFS basic operation (0x0500-0x050F).
func (op OpCode) IsVFSBasicOp() bool {
	return op >= 0x0500 && op <= 0x050F
}

// IsVFSChunkOp returns true if the opcode is a VFS chunk operation (0x0510-0x051F).
func (op OpCode) IsVFSChunkOp() bool {
	return op >= 0x0510 && op <= 0x051F
}

// IsVFSEmergencyOp returns true if the opcode is a VFS emergency operation (0x05C0-0x05CF).
func (op OpCode) IsVFSEmergencyOp() bool {
	return op >= 0x05C0 && op <= 0x05CF
}

// IsVFSFragmentOp returns true if the opcode is a VFS fragment operation (0x0520-0x052F).
func (op OpCode) IsVFSFragmentOp() bool {
	return op >= 0x0520 && op <= 0x052F
}

// IsVFSMetadataOp returns true if the opcode is a VFS metadata operation (0x0530-0x053F).
func (op OpCode) IsVFSMetadataOp() bool {
	return op >= 0x0530 && op <= 0x053F
}

// IsVFSOp returns true if the opcode is a VFS operation (0x0500-0x05CF).
// This includes all VFS categories: Basic, Chunk, Fragment, Metadata,
// Sharing, and Emergency operations.
func (op OpCode) IsVFSOp() bool {
	return op.Category() == CategoryVFS
}

// IsVFSShareOp returns true if the opcode is a VFS sharing operation (0x0550-0x055F).
func (op OpCode) IsVFSShareOp() bool {
	return op >= 0x0550 && op <= 0x055F
}

// IsVendorOp returns true if the opcode is a vendor extension (0xF000-0xFEFF).
func (op OpCode) IsVendorOp() bool {
	cat := op.Category()
	return cat >= 0xF0 && cat <= 0xFE
}

// IsVideoCallOp returns true if the opcode is a Video Call operation (0x0BD0-0x0BD9).
func (op OpCode) IsVideoCallOp() bool {
	return op >= 0x0BD0 && op <= 0x0BD9
}

// IsVoicemailOp returns true if the opcode is a Voicemail operation (0x0B30-0x0B3F).
func (op OpCode) IsVoicemailOp() bool {
	return op >= 0x0B30 && op <= 0x0B3F
}
