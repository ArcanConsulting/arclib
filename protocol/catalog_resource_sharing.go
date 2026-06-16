// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk resource sharing operation codes (category 0x02).

package protocol

// Resource Sharing operations (category 0x02).
const (
	OpDeviceList           OpCode = 0x0200 // List shared devices
	OpDeviceInfo           OpCode = 0x0201 // Get shared device details
	OpDeviceSubscribe      OpCode = 0x0202 // Subscribe to device data
	OpDeviceUnsubscribe    OpCode = 0x0203 // Unsubscribe from device
	OpDeviceLock           OpCode = 0x0204 // Lock device for exclusive access
	OpDeviceUnlock         OpCode = 0x0205 // Unlock device
	OpDeviceConfigure      OpCode = 0x0206 // Configure device parameters
	OpDeviceCapabilities   OpCode = 0x0207 // Query device capabilities
	OpDeviceWrite          OpCode = 0x0208 // Write to device
	OpDeviceData           OpCode = 0x0209 // Data from device (push)
	OpDeviceQueue          OpCode = 0x020A // Queue command
	OpDeviceDiscover       OpCode = 0x020B // Discover devices (incl. unshared)
	OpDeviceQueueStatus    OpCode = 0x020C // Get queue status
	OpStreamStart          OpCode = 0x0210 // Start video/audio stream
	OpStreamStop           OpCode = 0x0211 // Stop stream
	OpStreamData           OpCode = 0x0212 // Stream data chunk
	OpStreamConfigure      OpCode = 0x0213 // Configure stream parameters
	OpStreamQuality        OpCode = 0x0214 // Adjust quality settings
	OpStreamPause          OpCode = 0x0215 // Pause stream
	OpStreamResume         OpCode = 0x0216 // Resume paused stream
	OpGPURequest           OpCode = 0x0220 // GPU inference request
	OpGPUResponse          OpCode = 0x0221 // GPU inference response
	OpGPUStatus            OpCode = 0x0222 // GPU availability status
	OpGPUCancel            OpCode = 0x0223 // Cancel pending GPU request
	OpGPUQueueInfo         OpCode = 0x0224 // GPU queue position and ETA
	OpRouteCreate          OpCode = 0x0240 // Create device route
	OpRouteDelete          OpCode = 0x0241 // Delete device route
	OpRouteList            OpCode = 0x0242 // List active routes
	OpRouteModify          OpCode = 0x0243 // Modify existing route
	OpACLSet               OpCode = 0x0244 // Set access control list
	OpACLGet               OpCode = 0x0245 // Get access control list
	OpACLGrant             OpCode = 0x0246 // Grant access to resource
	OpACLRevoke            OpCode = 0x0247 // Revoke access from resource
	OpRoutePriority        OpCode = 0x0248 // Set routing priority
	OpACLList              OpCode = 0x0249 // List all ACL grants for a resource
	OpClientRegister       OpCode = 0x0250 // Register client at hub
	OpClientHeartbeat      OpCode = 0x0251 // Client keepalive
	OpClientAssign         OpCode = 0x0252 // Assign device to client
	OpClientRevoke         OpCode = 0x0253 // Revoke device access
	OpClientStatus         OpCode = 0x0254 // Query client status
	OpClientList           OpCode = 0x0255 // List registered clients
	OpClientInfo           OpCode = 0x0256 // Get client information
	OpDeviceRequestCreate  OpCode = 0x0257 // Request device sharing (consent-first)
	OpDeviceRequestList    OpCode = 0x0258 // List pending/active share requests
	OpDeviceRequestDetails OpCode = 0x0259 // Get request details
	OpDeviceRequestRespond OpCode = 0x025A // Accept/reject share request
	OpDeviceRequestCancel  OpCode = 0x025B // Cancel pending request
	OpFedAdvertise         OpCode = 0x0260 // Push service advertisement
	OpFedAdvertiseAck      OpCode = 0x0261 // Advertisement acknowledgment
	OpFedRequest           OpCode = 0x0262 // Federated service request (LLM/TTS/STT)
	OpFedResponse          OpCode = 0x0263 // Federated service response
	OpFedStreamChunk       OpCode = 0x0264 // Streaming response chunk
	OpFedStreamEnd         OpCode = 0x0265 // End of stream
	OpFedCancel            OpCode = 0x0266 // Cancel in-flight request
	OpFedPreempt           OpCode = 0x0267 // Preemption notification
	OpFedPreemptAck        OpCode = 0x0268 // Preemption acknowledgment
	OpFedBenchmark         OpCode = 0x0269 // Benchmark request
	OpFedBenchmarkResult   OpCode = 0x026A // Benchmark result
	OpFedGrantSync         OpCode = 0x026B // Synchronize grant configuration
	OpFedGrantSyncAck      OpCode = 0x026C // Grant sync acknowledgment
	OpFedBillingSync       OpCode = 0x026D // Billing reconciliation
	OpFedBillingSyncAck    OpCode = 0x026E // Billing reconciliation ack
	OpFedGPUReport         OpCode = 0x026F // GPU inventory report
)
