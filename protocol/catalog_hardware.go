// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk hardware operation codes (category 0x06).

package protocol

// Hardware operations (category 0x06).
const (
	OpUSBDeviceList          OpCode = 0x0600 // List USB devices
	OpUSBDeviceAttach        OpCode = 0x0601 // Attach USB device
	OpUSBDeviceDetach        OpCode = 0x0602 // Detach USB device
	OpUSBControlTransfer     OpCode = 0x0603 // USB control transfer
	OpUSBBulkTransfer        OpCode = 0x0604 // USB bulk transfer
	OpUSBInterruptTransfer   OpCode = 0x0605 // USB interrupt transfer
	OpUSBIsochronousTransfer OpCode = 0x0606 // USB isochronous transfer
	OpUSBGetDescriptor       OpCode = 0x0607 // Get USB descriptor
	OpUSBSetConfiguration    OpCode = 0x0608 // Set USB configuration
	OpUSBSetInterface        OpCode = 0x0609 // Set USB interface
	OpUSBClearHalt           OpCode = 0x060A // Clear endpoint halt
	OpUSBReset               OpCode = 0x060B // Reset USB device
	OpSerialPortList         OpCode = 0x0610 // List serial ports
	OpSerialPortOpen         OpCode = 0x0611 // Open serial port
	OpSerialPortClose        OpCode = 0x0612 // Close serial port
	OpSerialPortConfigure    OpCode = 0x0613 // Configure baud, parity, etc.
	OpSerialDataWrite        OpCode = 0x0614 // Write to serial port
	OpSerialDataRead         OpCode = 0x0615 // Read from serial port
	OpSerialControlSet       OpCode = 0x0616 // Set control lines (DTR, RTS)
	OpSerialControlGet       OpCode = 0x0617 // Get control line status
	OpSerialBreak            OpCode = 0x0618 // Send break condition
	OpGPIOChipList           OpCode = 0x0620 // List GPIO chips
	OpGPIOLineInfo           OpCode = 0x0621 // Get line information
	OpGPIOLineRequest        OpCode = 0x0622 // Request GPIO line
	OpGPIOLineRelease        OpCode = 0x0623 // Release GPIO line
	OpGPIOLineSet            OpCode = 0x0624 // Set line value
	OpGPIOLineGet            OpCode = 0x0625 // Get line value
	OpGPIOLineWatch          OpCode = 0x0626 // Watch for edge events
	OpGPIOEvent              OpCode = 0x0627 // GPIO edge event notification
	OpGPIOPWMConfigure       OpCode = 0x0628 // Configure PWM output
	OpGPIOPWMSet             OpCode = 0x0629 // Set PWM duty cycle
	OpI2CBusList             OpCode = 0x0630 // List I2C buses
	OpI2CBusScan             OpCode = 0x0631 // Scan for devices on bus
	OpI2CTransfer            OpCode = 0x0632 // I2C read/write transfer
	OpI2CWriteByte           OpCode = 0x0633 // Write single byte
	OpI2CReadByte            OpCode = 0x0634 // Read single byte
	OpI2CWriteBlock          OpCode = 0x0635 // Write data block
	OpI2CReadBlock           OpCode = 0x0636 // Read data block
	OpI2CSMBusCommand        OpCode = 0x0637 // SMBus protocol command
	OpSPIBusList             OpCode = 0x0640 // List SPI buses
	OpSPIDeviceOpen          OpCode = 0x0641 // Open SPI device
	OpSPIDeviceClose         OpCode = 0x0642 // Close SPI device
	OpSPIConfigure           OpCode = 0x0643 // Configure mode, speed, bits
	OpSPITransfer            OpCode = 0x0644 // Full-duplex SPI transfer
	OpSPIWrite               OpCode = 0x0645 // SPI write-only
	OpSPIRead                OpCode = 0x0646 // SPI read-only
	OpCANInterfaceList       OpCode = 0x0650 // List CAN interfaces
	OpCANInterfaceOpen       OpCode = 0x0651 // Open CAN interface
	OpCANInterfaceClose      OpCode = 0x0652 // Close CAN interface
	OpCANConfigure           OpCode = 0x0653 // Configure bitrate, mode
	OpCANFrameSend           OpCode = 0x0654 // Send CAN frame
	OpCANFrameReceive        OpCode = 0x0655 // Receive CAN frame
	OpCANFilterSet           OpCode = 0x0656 // Set receive filter
	OpCANErrorStatus         OpCode = 0x0657 // Get error counters
	OpOneWireBusList         OpCode = 0x0660 // List 1-Wire buses
	OpOneWireSearch          OpCode = 0x0661 // Search for devices
	OpOneWireReset           OpCode = 0x0662 // Reset bus
	OpOneWireReadROM         OpCode = 0x0663 // Read device ROM
	OpOneWireMatchROM        OpCode = 0x0664 // Select device by ROM
	OpOneWireSkipROM         OpCode = 0x0665 // Skip ROM (single device)
	OpOneWireRead            OpCode = 0x0666 // Read bytes
	OpOneWireWrite           OpCode = 0x0667 // Write bytes
	OpGSMModemInit           OpCode = 0x0670 // Initialize GSM modem
	OpGSMModemStatus         OpCode = 0x0671 // Get modem status (manufacturer, model, IMEI)
	OpGSMNetworkReg          OpCode = 0x0672 // Network registration status
	OpGSMSignalQuality       OpCode = 0x0673 // Signal quality (RSSI, BER)
	OpGSMSMSSend             OpCode = 0x0674 // Send SMS via AT commands
	OpGSMSMSReceive          OpCode = 0x0675 // Receive SMS notification
	OpGSMSMSList             OpCode = 0x0676 // List SMS messages
	OpGSMSMSDelete           OpCode = 0x0677 // Delete SMS from SIM
	OpGSMUSSDSend            OpCode = 0x0678 // Send USSD command
	OpGSMUSSDResponse        OpCode = 0x0679 // USSD response
	OpGSMGNSSEnable          OpCode = 0x067A // Enable/disable GNSS module
	OpGSMGNSSPosition        OpCode = 0x067B // Get GNSS position fix
	OpGSMGNSSConfig          OpCode = 0x067C // Configure GNSS (GPS, GLONASS, BeiDou)
	OpGSMVoiceDial           OpCode = 0x067D // Initiate voice call
	OpGSMVoiceHangup         OpCode = 0x067E // Hangup voice call
	OpGSMVoiceAnswer         OpCode = 0x067F // Answer incoming call
	OpGSMGNSSHistory         OpCode = 0x0680 // Get GNSS position history
)
