// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk mdm operation codes (category 0x17).

package protocol

// Mdm operations (category 0x17).
const (
	OpMDMDeviceList             OpCode = 0x1700 // List managed devices
	OpMDMDeviceGet              OpCode = 0x1701 // Get managed device
	OpMDMDeviceRegister         OpCode = 0x1702 // Register device for MDM
	OpMDMDeviceUnregister       OpCode = 0x1703 // Unregister device from MDM
	OpMDMDeviceCommand          OpCode = 0x1704 // Send command to device
	OpMDMPolicyGet              OpCode = 0x1705 // Get device policy
	OpMDMPolicySet              OpCode = 0x1706 // Set device policy
	OpScreenTimePolicyGet       OpCode = 0x1710 // Get screen time policy
	OpScreenTimePolicySet       OpCode = 0x1711 // Set screen time policy
	OpScreenTimeUsageToday      OpCode = 0x1712 // Get today's usage
	OpScreenTimeUsageHistory    OpCode = 0x1713 // Get usage history
	OpScreenTimeRequestExtra    OpCode = 0x1714 // Request extra time
	OpScreenTimeApproveReq      OpCode = 0x1715 // Approve/deny time request
	OpScreenTimeGrantExtra      OpCode = 0x1716 // Grant extra time directly
	OpScreenTimePause           OpCode = 0x1717 // Pause device
	OpScreenTimeUnpause         OpCode = 0x1718 // Unpause device
	OpScreenTimePendingReqs     OpCode = 0x1719 // List pending time requests
	OpMDMLocationPolicyGet      OpCode = 0x1720 // Get MDM location policy
	OpMDMLocationPolicySet      OpCode = 0x1721 // Set MDM location policy
	OpMDMLocationCurrentGet     OpCode = 0x1722 // Get current device location
	OpMDMLocationHistory        OpCode = 0x1723 // Get device location history
	OpMDMLocationRequest        OpCode = 0x1724 // Request device location update
	OpMDMGeofenceCreate         OpCode = 0x1730 // Create MDM geofence
	OpMDMGeofenceUpdate         OpCode = 0x1731 // Update MDM geofence
	OpMDMGeofenceDelete         OpCode = 0x1732 // Delete MDM geofence
	OpMDMGeofenceList           OpCode = 0x1733 // List MDM geofences
	OpMDMAppPolicyGet           OpCode = 0x1740 // Get MDM app policy
	OpMDMAppPolicySet           OpCode = 0x1741 // Set MDM app policy
	OpMDMAppCatalogList         OpCode = 0x1742 // List MDM app catalog
	OpMDMAppCatalogAdd          OpCode = 0x1743 // Add to MDM app catalog
	OpMDMAppInstallRequest      OpCode = 0x1744 // Request MDM app install
	OpMDMAppInstallApprove      OpCode = 0x1745 // Approve MDM app install
	OpMDMAppInstalledList       OpCode = 0x1746 // List MDM installed apps
	OpMDMKioskEnable            OpCode = 0x1750 // Enable MDM kiosk mode
	OpMDMKioskDisable           OpCode = 0x1751 // Disable MDM kiosk mode
	OpMDMPrivilegePointsGet     OpCode = 0x1760 // Get MDM privilege points
	OpMDMPrivilegePointsAward   OpCode = 0x1761 // Award MDM privilege points
	OpMDMPrivilegeListAvailable OpCode = 0x1762 // List available MDM privileges
	OpMDMPrivilegeUnlock        OpCode = 0x1763 // Unlock MDM privilege
	OpMDMPrivilegeRevoke        OpCode = 0x1764 // Revoke MDM privilege
	OpMDMPrivilegeChildStatus   OpCode = 0x1765 // Get MDM child status
)
