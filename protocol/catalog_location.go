// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk location operation codes (category 0x15).

package protocol

// Location operations (category 0x15).
const (
	OpLocationUpdate             OpCode = 0x1500 // Update location
	OpLocationGet                OpCode = 0x1501 // Get location
	OpLocationHistory            OpCode = 0x1502 // Get location history
	OpLocationShare              OpCode = 0x1503 // Share location
	OpLocationUnshare            OpCode = 0x1504 // Stop sharing location
	OpLocationSubscribe          OpCode = 0x1505 // Subscribe to location updates
	OpNavRoutePlan               OpCode = 0x1510 // Plan route
	OpNavRouteStart              OpCode = 0x1511 // Start navigation
	OpNavRouteUpdate             OpCode = 0x1512 // Update route progress
	OpNavRouteStop               OpCode = 0x1513 // Stop navigation
	OpNavRouteGet                OpCode = 0x1514 // Get route details
	OpNavRouteList               OpCode = 0x1515 // List saved routes
	OpTransitQuery               OpCode = 0x1520 // Query public transit
	OpTransitDepartures          OpCode = 0x1521 // Get departures
	OpTransitArrivals            OpCode = 0x1522 // Get arrivals
	OpTransitAlerts              OpCode = 0x1523 // Get service alerts
	OpVehicleCreate              OpCode = 0x1530 // Register vehicle
	OpVehicleGet                 OpCode = 0x1531 // Get vehicle info
	OpVehicleUpdate              OpCode = 0x1532 // Update vehicle
	OpVehicleDelete              OpCode = 0x1533 // Remove vehicle
	OpVehicleList                OpCode = 0x1534 // List vehicles
	OpVehicleStatus              OpCode = 0x1535 // Get vehicle status
	OpVehicleLock                OpCode = 0x1536 // Lock vehicle
	OpVehicleUnlock              OpCode = 0x1537 // Unlock vehicle
	OpVehicleClimate             OpCode = 0x1538 // Control climate
	OpGeofenceCreate             OpCode = 0x1540 // Create geofence
	OpGeofenceGet                OpCode = 0x1541 // Get geofence
	OpGeofenceUpdate             OpCode = 0x1542 // Update geofence
	OpGeofenceDelete             OpCode = 0x1543 // Delete geofence
	OpGeofenceList               OpCode = 0x1544 // List geofences
	OpGeofenceTrigger            OpCode = 0x1545 // Geofence triggered event
	OpHealthAnalyticsTrend       OpCode = 0x1550 // Compute metric trend
	OpHealthAnalyticsSummary     OpCode = 0x1551 // Compute health summary
	OpHealthAnalyticsCorrelation OpCode = 0x1552 // Air quality-health correlations
	OpHealthAnalyticsAnomalies   OpCode = 0x1553 // Detect metric anomalies
	OpPropertyAssetCreate        OpCode = 0x1580 // Create asset
	OpPropertyAssetGet           OpCode = 0x1581 // Get asset details
	OpPropertyAssetUpdate        OpCode = 0x1582 // Update asset
	OpPropertyAssetDelete        OpCode = 0x1583 // Delete asset
	OpPropertyAssetList          OpCode = 0x1584 // List family assets
	OpPropertyAssetSearch        OpCode = 0x1585 // Search assets
	OpPropertyLocationCreate     OpCode = 0x1590 // Create location/room
	OpPropertyLocationGet        OpCode = 0x1591 // Get location details
	OpPropertyLocationUpdate     OpCode = 0x1592 // Update location
	OpPropertyLocationDelete     OpCode = 0x1593 // Delete location
	OpPropertyLocationList       OpCode = 0x1594 // List locations
	OpPropertyMaintCreate        OpCode = 0x15A0 // Create maintenance task
	OpPropertyMaintGet           OpCode = 0x15A1 // Get maintenance task
	OpPropertyMaintUpdate        OpCode = 0x15A2 // Update maintenance task
	OpPropertyMaintDelete        OpCode = 0x15A3 // Delete maintenance task
	OpPropertyMaintList          OpCode = 0x15A4 // List maintenance tasks
	OpPropertyMaintComplete      OpCode = 0x15A5 // Mark maintenance complete
	OpPropertyDocCreate          OpCode = 0x15B0 // Create document
	OpPropertyDocGet             OpCode = 0x15B1 // Get document
	OpPropertyDocDelete          OpCode = 0x15B2 // Delete document
	OpPropertyDocList            OpCode = 0x15B3 // List documents
	OpPropertyValuationAdd       OpCode = 0x15C0 // Add valuation
	OpPropertyValuationList      OpCode = 0x15C1 // List valuations
)
