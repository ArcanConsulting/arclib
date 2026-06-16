// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk smarthome operation codes (category 0x0E).

package protocol

// Smarthome operations (category 0x0E).
const (
	OpSmartDeviceAdd             OpCode = 0x0E00 // Add smart device
	OpSmartDeviceRemove          OpCode = 0x0E01 // Remove smart device
	OpSmartDeviceList            OpCode = 0x0E02 // List smart devices
	OpSmartDeviceGet             OpCode = 0x0E03 // Get smart device
	OpSmartDeviceControl         OpCode = 0x0E10 // Control device
	OpSmartDeviceState           OpCode = 0x0E11 // Get state
	OpSmartDeviceSet             OpCode = 0x0E12 // Set state
	OpSmartGroupCreate           OpCode = 0x0E20 // Create group
	OpSmartGroupGet              OpCode = 0x0E21 // Get group
	OpSmartGroupUpdate           OpCode = 0x0E22 // Update group
	OpSmartGroupDelete           OpCode = 0x0E23 // Delete group
	OpSmartGroupList             OpCode = 0x0E24 // List groups
	OpSmartGroupControl          OpCode = 0x0E25 // Control group
	OpSmartGroupAddDevice        OpCode = 0x0E26 // Add device to group
	OpSmartGroupRemoveDevice     OpCode = 0x0E27 // Remove device from group
	OpSmartSceneCreate           OpCode = 0x0E30 // Create scene
	OpSmartSceneGet              OpCode = 0x0E31 // Get scene
	OpSmartSceneUpdate           OpCode = 0x0E32 // Update scene
	OpSmartSceneDelete           OpCode = 0x0E33 // Delete scene
	OpSmartSceneList             OpCode = 0x0E34 // List scenes
	OpSmartSceneActivate         OpCode = 0x0E35 // Activate scene
	OpSmartSceneDeactivate       OpCode = 0x0E36 // Deactivate scene
	OpSmartRoomCreate            OpCode = 0x0E40 // Create room
	OpSmartRoomGet               OpCode = 0x0E41 // Get room
	OpSmartRoomUpdate            OpCode = 0x0E42 // Update room
	OpSmartRoomDelete            OpCode = 0x0E43 // Delete room
	OpSmartRoomList              OpCode = 0x0E44 // List rooms
	OpEnergySystemGet            OpCode = 0x0E50 // Get energy system
	OpEnergySystemUpdate         OpCode = 0x0E51 // Update energy system
	OpEnergyReadingRecord        OpCode = 0x0E52 // Record energy reading
	OpEnergyReadingHistory       OpCode = 0x0E53 // Get reading history
	OpTariffGetCurrent           OpCode = 0x0E54 // Get current tariff price
	OpTariffGetPrices            OpCode = 0x0E55 // Get tariff prices for period
	OpLoadSchedule               OpCode = 0x0E56 // Schedule a load
	OpLoadGetSchedule            OpCode = 0x0E57 // Get scheduled loads
	OpEnergyOptimize             OpCode = 0x0E58 // Get optimization recommendations
	OpAdapterConfigGet           OpCode = 0x0E59 // Get adapter config
	OpAdapterConfigSet           OpCode = 0x0E5A // Set adapter config
	OpAdapterConfigList          OpCode = 0x0E5B // List adapter configs
	OpSmartDeviceDiscover        OpCode = 0x0E5C // Discover devices
	OpDeviceStateHistory         OpCode = 0x0E5D // Get device state history
	OpDeviceFamilyStateHistory   OpCode = 0x0E5E // Get family device state history
	OpDeviceStateHistoryClean    OpCode = 0x0E5F // Clean old state history
	OpDashboardOverview          OpCode = 0x0E60 // Get full dashboard overview
	OpDashboardDeviceSummary     OpCode = 0x0E61 // Get device summary stats
	OpDashboardEnergySummary     OpCode = 0x0E62 // Get energy overview
	OpDashboardEnergyChart       OpCode = 0x0E63 // Get energy chart data
	OpDashboardAutomationSummary OpCode = 0x0E64 // Get automation summary
	OpDashboardRecentAlerts      OpCode = 0x0E65 // Get recent alerts
	OpDashboardWidgetGet         OpCode = 0x0E66 // Get widget configuration
	OpDashboardWidgetUpdate      OpCode = 0x0E67 // Update widget configuration
	OpDashboardDevicesByRoom     OpCode = 0x0E68 // Get devices grouped by room
	OpDashboardStateChart        OpCode = 0x0E69 // Get device state chart data
	OpDashboardAlertAck          OpCode = 0x0E6A // Acknowledge alert
	OpDashboardAlertConfig       OpCode = 0x0E6B // Get/update alert config
)
