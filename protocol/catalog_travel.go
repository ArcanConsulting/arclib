// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk travel operation codes (category 0x1A).

package protocol

// Travel operations (category 0x1A).
const (
	OpTravelTripCreate     OpCode = 0x1A00 // Create trip
	OpTravelTripGet        OpCode = 0x1A01 // Get trip
	OpTravelTripUpdate     OpCode = 0x1A02 // Update trip
	OpTravelTripDelete     OpCode = 0x1A03 // Delete trip
	OpTravelTripList       OpCode = 0x1A04 // List trips
	OpTravelDestCreate     OpCode = 0x1A08 // Add destination to trip
	OpTravelDestGet        OpCode = 0x1A09 // Get destination
	OpTravelDestUpdate     OpCode = 0x1A0A // Update destination
	OpTravelDestDelete     OpCode = 0x1A0B // Remove destination
	OpTravelDestList       OpCode = 0x1A0C // List destinations for trip
	OpTravelBookingCreate  OpCode = 0x1A10 // Create booking
	OpTravelBookingGet     OpCode = 0x1A11 // Get booking
	OpTravelBookingUpdate  OpCode = 0x1A12 // Update booking
	OpTravelBookingDelete  OpCode = 0x1A13 // Delete booking
	OpTravelBookingList    OpCode = 0x1A14 // List bookings for trip
	OpTravelBookingStats   OpCode = 0x1A15 // Get booking cost summary
	OpTravelPackListCreate OpCode = 0x1A18 // Create packing list
	OpTravelPackListGet    OpCode = 0x1A19 // Get packing list
	OpTravelPackListDelete OpCode = 0x1A1A // Delete packing list
	OpTravelPackItemAdd    OpCode = 0x1A1B // Add item to list
	OpTravelPackItemRemove OpCode = 0x1A1C // Remove item
	OpTravelPackItemToggle OpCode = 0x1A1D // Toggle packed status
	OpTravelPackListList   OpCode = 0x1A1E // List packing lists for trip
	OpTravelBudgetSet      OpCode = 0x1A20 // Set trip budget
	OpTravelBudgetGet      OpCode = 0x1A21 // Get trip budget
	OpTravelExpenseAdd     OpCode = 0x1A22 // Add expense
	OpTravelExpenseUpdate  OpCode = 0x1A23 // Update expense
	OpTravelExpenseDelete  OpCode = 0x1A24 // Delete expense
	OpTravelExpenseList    OpCode = 0x1A25 // List expenses
	OpTravelDocCreate      OpCode = 0x1A28 // Add travel document
	OpTravelDocUpdate      OpCode = 0x1A29 // Update travel document
	OpTravelDocDelete      OpCode = 0x1A2A // Delete travel document
	OpTravelDocList        OpCode = 0x1A2B // List travel documents
)
