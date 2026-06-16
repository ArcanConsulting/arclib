// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk calendar operation codes (category 0x0C).

package protocol

// Calendar operations (category 0x0C).
const (
	OpEventCreate            OpCode = 0x0C00 // Create event
	OpEventGet               OpCode = 0x0C01 // Get event
	OpEventUpdate            OpCode = 0x0C02 // Update event
	OpEventDelete            OpCode = 0x0C03 // Delete event
	OpEventList              OpCode = 0x0C04 // List events
	OpEventSearch            OpCode = 0x0C05 // Search events
	OpEventRSVP              OpCode = 0x0C10 // RSVP to event
	OpEventRemind            OpCode = 0x0C11 // Set/delete reminder
	OpEventShare             OpCode = 0x0C12 // Share event
	OpEventAttendees         OpCode = 0x0C13 // List event attendees
	OpEventSync              OpCode = 0x0C20 // Sync calendar
	OpCalendarCreate         OpCode = 0x0C30 // Create calendar
	OpCalendarGet            OpCode = 0x0C31 // Get calendar
	OpCalendarUpdate         OpCode = 0x0C32 // Update calendar
	OpCalendarDelete         OpCode = 0x0C33 // Delete calendar
	OpCalendarList           OpCode = 0x0C34 // List calendars
	OpCalendarStats          OpCode = 0x0C35 // Get calendar statistics
	OpCalendarShare          OpCode = 0x0C36 // Share calendar with user
	OpCalendarUnshare        OpCode = 0x0C37 // Remove calendar share
	OpCalendarPermissions    OpCode = 0x0C38 // List calendar permissions
	OpCalendarRSVPReward     OpCode = 0x0C40 // Award XP for RSVP
	OpCalendarEventTaskLink  OpCode = 0x0C41 // Link event to task
	OpCalendarFamilyActivity OpCode = 0x0C42 // Get family activity summary
	OpEventSuggest           OpCode = 0x0C50 // Create event suggestion
	OpEventSuggestionList    OpCode = 0x0C51 // List pending suggestions
	OpEventSuggestionAccept  OpCode = 0x0C52 // Accept event suggestion
	OpEventSuggestionDecline OpCode = 0x0C53 // Decline event suggestion
	OpEventDetectionConfig   OpCode = 0x0C54 // Get/set detection config
)
