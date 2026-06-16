// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk mobility operation codes (category 0x19).

package protocol

// Mobility operations (category 0x19).
const (
	OpMobilityVehicleCreate    OpCode = 0x1900 // Create vehicle
	OpMobilityVehicleGet       OpCode = 0x1901 // Get vehicle
	OpMobilityVehicleUpdate    OpCode = 0x1902 // Update vehicle
	OpMobilityVehicleDelete    OpCode = 0x1903 // Delete vehicle
	OpMobilityVehicleList      OpCode = 0x1904 // List family vehicles
	OpMobilityMaintCreate      OpCode = 0x1908 // Create maintenance entry
	OpMobilityMaintGet         OpCode = 0x1909 // Get maintenance entry
	OpMobilityMaintUpdate      OpCode = 0x190A // Update maintenance entry
	OpMobilityMaintDelete      OpCode = 0x190B // Delete maintenance entry
	OpMobilityMaintList        OpCode = 0x190C // List maintenance for vehicle
	OpMobilityMaintComplete    OpCode = 0x190D // Mark maintenance complete
	OpMobilityFuelAdd          OpCode = 0x1910 // Add fuel/charge entry
	OpMobilityFuelDelete       OpCode = 0x1911 // Delete fuel entry
	OpMobilityFuelList         OpCode = 0x1912 // List fuel entries
	OpMobilityFuelStats        OpCode = 0x1913 // Get fuel statistics
	OpMobilityBookingCreate    OpCode = 0x1918 // Create booking
	OpMobilityBookingGet       OpCode = 0x1919 // Get booking
	OpMobilityBookingUpdate    OpCode = 0x191A // Update booking
	OpMobilityBookingDelete    OpCode = 0x191B // Delete booking
	OpMobilityBookingList      OpCode = 0x191C // List bookings
	OpMobilityBookingConflicts OpCode = 0x191D // Check booking conflicts
	OpMobilityCarpoolCreate    OpCode = 0x1920 // Create carpool
	OpMobilityCarpoolGet       OpCode = 0x1921 // Get carpool
	OpMobilityCarpoolUpdate    OpCode = 0x1922 // Update carpool
	OpMobilityCarpoolDelete    OpCode = 0x1923 // Delete carpool
	OpMobilityCarpoolList      OpCode = 0x1924 // List carpools
	OpMobilityCarpoolJoin      OpCode = 0x1925 // Join carpool
	OpMobilityCarpoolLeave     OpCode = 0x1926 // Leave carpool
	OpMobilityCarpoolRotation  OpCode = 0x1927 // Get/advance rotation
	OpMobilityTripCreate       OpCode = 0x1930 // Create trip entry
	OpMobilityTripGet          OpCode = 0x1931 // Get trip entry
	OpMobilityTripUpdate       OpCode = 0x1932 // Update trip entry
	OpMobilityTripDelete       OpCode = 0x1933 // Delete trip entry
	OpMobilityTripList         OpCode = 0x1934 // List trips
	OpMobilityTripStats        OpCode = 0x1935 // Get trip statistics
)
