// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk sync operation codes (category 0x11).

package protocol

// Sync operations (category 0x11).
const (
	OpSyncRequest             OpCode = 0x1100 // Request sync
	OpSyncStart               OpCode = 0x1101 // Start sync
	OpSyncData                OpCode = 0x1102 // Sync data chunk
	OpSyncComplete            OpCode = 0x1103 // Sync complete
	OpSyncConflict            OpCode = 0x1110 // Conflict detected
	OpSyncResolve             OpCode = 0x1111 // Resolve conflict
	OpSyncStatus              OpCode = 0x1120 // Sync status
	OpSyncReset               OpCode = 0x1121 // Reset sync state
	OpVFSFederationConnect    OpCode = 0x1130 // Connect to federated VFS
	OpVFSFederationDisconnect OpCode = 0x1131 // Disconnect from federated VFS
	OpVFSFederationSync       OpCode = 0x1132 // Sync with federated VFS
	OpVFSFederationStatus     OpCode = 0x1133 // Get federation status
	OpVFSFederationInvite     OpCode = 0x1134 // Invite to federation
	OpVFSFederationAccept     OpCode = 0x1135 // Accept federation invite
	OpVFSFederationReject     OpCode = 0x1136 // Reject federation invite
	OpVFSFederationLeave      OpCode = 0x1137 // Leave federation
	OpSpaceCreate             OpCode = 0x1140 // Create space
	OpSpaceGet                OpCode = 0x1141 // Get space details
	OpSpaceUpdate             OpCode = 0x1142 // Update space
	OpSpaceDelete             OpCode = 0x1143 // Delete space
	OpSpaceList               OpCode = 0x1144 // List spaces
	OpSpaceMembers            OpCode = 0x1145 // List space members
	OpFederationConnect       OpCode = 0x1150 // Establish instance connection
	OpFederationDisconnect    OpCode = 0x1151 // Remove instance connection
	OpFederationList          OpCode = 0x1152 // List connections
	OpFederationStatus        OpCode = 0x1153 // Get connection status
	OpFederationAnnounce      OpCode = 0x1154 // Announce capabilities
	OpFederationHeartbeat     OpCode = 0x1155 // Instance heartbeat
	OpSatelliteRegister       OpCode = 0x1160 // Register satellite
	OpSatelliteGet            OpCode = 0x1161 // Get satellite config
	OpSatelliteUpdate         OpCode = 0x1162 // Update satellite config
	OpSatelliteRemove         OpCode = 0x1163 // Remove satellite
	OpSatelliteList           OpCode = 0x1164 // List satellites
	OpSatelliteStatus         OpCode = 0x1165 // Get satellite status
	OpResourceRegister        OpCode = 0x1170 // Register shared resource
	OpResourceGet             OpCode = 0x1171 // Get resource details
	OpResourceUpdate          OpCode = 0x1172 // Update resource
	OpResourceRemove          OpCode = 0x1173 // Remove resource
	OpResourceList            OpCode = 0x1174 // List resources
	OpResourceRequest         OpCode = 0x1175 // Request resource allocation
	OpResourceRelease         OpCode = 0x1176 // Release resource allocation
	OpSharingRuleCreate       OpCode = 0x1178 // Create sharing rule
	OpSharingRuleDelete       OpCode = 0x1179 // Delete sharing rule
	OpSharingRuleList         OpCode = 0x117A // List sharing rules
	OpHAConfigGet             OpCode = 0x1180 // Get HA configuration
	OpHAConfigUpdate          OpCode = 0x1181 // Update HA configuration
	OpHAStatus                OpCode = 0x1182 // Get HA status
	OpHAFailover              OpCode = 0x1183 // Trigger failover
	OpHAFailback              OpCode = 0x1184 // Trigger failback
	OpHABackupAdd             OpCode = 0x1185 // Add backup instance
	OpHABackupRemove          OpCode = 0x1186 // Remove backup instance
	OpHABackupList            OpCode = 0x1187 // List backup instances
	OpDiscoveryStart          OpCode = 0x1190 // Start LAN discovery
	OpDiscoveryStop           OpCode = 0x1191 // Stop LAN discovery
	OpDiscoveryList           OpCode = 0x1192 // List discovered instances
	OpPairingInitiate         OpCode = 0x1193 // Start pairing
	OpPairingConfirm          OpCode = 0x1194 // Confirm pairing with code
	OpPairingDeny             OpCode = 0x1195 // Deny pairing request
	OpPairingStatus           OpCode = 0x1196 // Get pairing status
	OpWireGuardSetup          OpCode = 0x11A0 // Setup WireGuard tunnel for connection
	OpWireGuardTeardown       OpCode = 0x11A1 // Tear down WireGuard tunnel
	OpWireGuardStatus         OpCode = 0x11A2 // Get WireGuard tunnel status
	OpWireGuardKeyExchange    OpCode = 0x11A3 // Exchange public keys between instances
	OpWireGuardEndpointUpdate OpCode = 0x11A4 // Update remote endpoint (roaming)
	OpWireGuardList           OpCode = 0x11A5 // List WireGuard tunnels
)
