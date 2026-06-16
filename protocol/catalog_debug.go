// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk debug operation codes (category 0x16).

package protocol

// Debug operations (category 0x16).
const (
	OpDebugEcho             OpCode = 0x1600 // Echo request
	OpDebugDump             OpCode = 0x1601 // Dump state
	OpDebugLog              OpCode = 0x1602 // Log message
	OpDebugTrace            OpCode = 0x1603 // Start/stop tracing
	OpDebugStats            OpCode = 0x1604 // Get debug statistics
	OpDebugConfig           OpCode = 0x1605 // Get/set debug config
	OpPluginInstall         OpCode = 0x1610 // Install a plugin
	OpPluginUninstall       OpCode = 0x1611 // Uninstall a plugin
	OpPluginStart           OpCode = 0x1612 // Start a plugin
	OpPluginStop            OpCode = 0x1613 // Stop a plugin
	OpPluginRestart         OpCode = 0x1614 // Restart a plugin
	OpPluginUpdate          OpCode = 0x1615 // Update a plugin's WASM binary
	OpPluginGet             OpCode = 0x1618 // Get plugin details
	OpPluginList            OpCode = 0x1619 // List plugins
	OpPluginGetLogs         OpCode = 0x161A // Get plugin execution logs
	OpPluginGetUsage        OpCode = 0x161B // Get plugin resource usage
	OpPluginPermGrant       OpCode = 0x1620 // Grant a plugin permission
	OpPluginPermRevoke      OpCode = 0x1621 // Revoke a plugin permission
	OpPluginPermList        OpCode = 0x1622 // List plugin permissions
	OpPluginStorageList     OpCode = 0x1624 // List plugin storage keys
	OpPluginWidgetCreate    OpCode = 0x1625 // Create a widget
	OpPluginWidgetGet       OpCode = 0x1626 // Get widget details
	OpPluginWidgetList      OpCode = 0x1627 // List plugin widgets
	OpPluginWidgetDelete    OpCode = 0x1628 // Delete a widget
	OpPluginWidgetConfigGet OpCode = 0x1629 // Get user widget config
	OpPluginWidgetConfigSet OpCode = 0x162A // Set user widget config
	OpPluginSubCreate       OpCode = 0x162B // Subscribe to event
	OpPluginSubDelete       OpCode = 0x162C // Unsubscribe from event
	OpPluginSubList         OpCode = 0x162D // List subscriptions
)
