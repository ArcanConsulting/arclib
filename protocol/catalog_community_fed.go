// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk community fed operation codes (category 0x20).

package protocol

// Community Fed operations (category 0x20).
const (
	OpCommAlertCreate         OpCode = 0x2000 // Create/broadcast community alert
	OpCommAlertGet            OpCode = 0x2001 // Get alert by ID
	OpCommAlertList           OpCode = 0x2002 // List active community alerts
	OpCommAlertConfirm        OpCode = 0x2003 // Confirm a detected alert
	OpCommAlertResolve        OpCode = 0x2004 // Resolve an alert
	OpCommResourceOffer       OpCode = 0x2005 // Offer resource for an alert
	OpCommResourceList        OpCode = 0x2006 // List resources for an alert
	OpCommResourceMatch       OpCode = 0x2007 // Match resources to alert needs
	OpRegistryRegister        OpCode = 0x200A // Register instance in directory
	OpRegistryUnregister      OpCode = 0x200B // Remove from directory
	OpRegistryUpdate          OpCode = 0x200C // Update registration config
	OpRegistrySearchGeo       OpCode = 0x200D // Search by geography
	OpRegistrySearchTopic     OpCode = 0x200E // Search by topic
	OpRegistryGetInstance     OpCode = 0x200F // Get instance by ID
	OpCollectiveGetConfig     OpCode = 0x2010 // Get collective intelligence config
	OpCollectiveUpdateConfig  OpCode = 0x2011 // Update collective config
	OpCollectiveShareInsight  OpCode = 0x2012 // Share insight with community
	OpCollectiveGetInsights   OpCode = 0x2013 // Get community insights by category
	OpCollectiveParticipate   OpCode = 0x2014 // Participate in federated learning round
	OpCollectiveReceiveUpdate OpCode = 0x2015 // Receive model update
	OpBridgeGetConfig         OpCode = 0x2016 // Get platform config
	OpBridgeUpdateConfig      OpCode = 0x2017 // Update platform config
	OpBridgeConnect           OpCode = 0x2018 // Connect to external platform
	OpBridgeDisconnect        OpCode = 0x2019 // Disconnect platform
	OpBridgeStatus            OpCode = 0x201A // Get connection status
	OpBridgeSync              OpCode = 0x201B // Trigger manual sync
	OpBridgeAuditLog          OpCode = 0x201C // Get integration audit log
)
