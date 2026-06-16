// Code generated from internal/protocol/opcode.go; DO NOT EDIT.
// Migrated MyClerk presence operation codes (category 0x13).

package protocol

// Presence operations (category 0x13).
const (
	OpPresenceSet                OpCode = 0x1300 // Set presence
	OpPresenceGet                OpCode = 0x1301 // Get presence
	OpPresenceSubscribe          OpCode = 0x1310 // Subscribe to presence
	OpPresenceUnsubscribe        OpCode = 0x1311 // Unsubscribe
	OpPresenceUpdate             OpCode = 0x1312 // Presence update notification
	OpStatusSet                  OpCode = 0x1320 // Set status message
	OpStatusGet                  OpCode = 0x1321 // Get status message
	OpDetectionRuleCreate        OpCode = 0x1322 // Create detection rule
	OpDetectionRuleGet           OpCode = 0x1323 // Get detection rule
	OpDetectionRuleList          OpCode = 0x1324 // List detection rules
	OpDetectionRuleUpdate        OpCode = 0x1325 // Update detection rule
	OpDetectionRuleDelete        OpCode = 0x1326 // Delete detection rule
	OpPresenceHistory            OpCode = 0x1327 // Get presence history
	OpPresenceMoodSet            OpCode = 0x1328 // Set presence mood
	OpPresenceMoodGet            OpCode = 0x1329 // Get presence mood
	OpPresenceFamilyGet          OpCode = 0x132A // Get all family presence
	OpGamificationProfileGet     OpCode = 0x1330 // Get user profile
	OpGamificationProfileUpdate  OpCode = 0x1331 // Update profile settings
	OpGamificationXPAward        OpCode = 0x1340 // Award XP to user
	OpGamificationXPHistory      OpCode = 0x1341 // Get XP history
	OpGamificationBadgeGet       OpCode = 0x1350 // Get badge definition
	OpGamificationBadgeList      OpCode = 0x1351 // List all badges
	OpGamificationBadgeUserGet   OpCode = 0x1352 // Get user badges
	OpGamificationBadgeAward     OpCode = 0x1353 // Award badge to user
	OpGamificationBadgeCreate    OpCode = 0x1354 // Create badge definition
	OpGamificationBadgeProgress  OpCode = 0x1355 // Update badge progress
	OpGamificationLeaderboard    OpCode = 0x1360 // Get family leaderboard
	OpGamificationSettingsGet    OpCode = 0x1370 // Get family settings
	OpGamificationSettingsUpdate OpCode = 0x1371 // Update family settings
	OpEduStudentCreate           OpCode = 0x1380 // Create student profile
	OpEduStudentGet              OpCode = 0x1381 // Get student profile
	OpEduStudentUpdate           OpCode = 0x1382 // Update student profile
	OpEduStudentDelete           OpCode = 0x1383 // Delete student profile
	OpEduStudentList             OpCode = 0x1384 // List family students
	OpEduSubjectCreate           OpCode = 0x1388 // Create subject
	OpEduSubjectGet              OpCode = 0x1389 // Get subject
	OpEduSubjectUpdate           OpCode = 0x138A // Update subject
	OpEduSubjectDelete           OpCode = 0x138B // Delete subject
	OpEduSubjectList             OpCode = 0x138C // List family subjects
	OpEduGradeCreate             OpCode = 0x1390 // Add grade
	OpEduGradeGet                OpCode = 0x1391 // Get grade
	OpEduGradeUpdate             OpCode = 0x1392 // Update grade
	OpEduGradeDelete             OpCode = 0x1393 // Delete grade
	OpEduGradeList               OpCode = 0x1394 // List student grades
	OpEduAverageGet              OpCode = 0x1396 // Get subject average
	OpEduAverageList             OpCode = 0x1397 // List subject averages
	OpEduHomeworkCreate          OpCode = 0x13A0 // Create homework
	OpEduHomeworkGet             OpCode = 0x13A1 // Get homework
	OpEduHomeworkUpdate          OpCode = 0x13A2 // Update homework
	OpEduHomeworkDelete          OpCode = 0x13A3 // Delete homework
	OpEduHomeworkList            OpCode = 0x13A4 // List student homework
	OpEduHomeworkComplete        OpCode = 0x13A5 // Complete homework
	OpEduHomeworkHelp            OpCode = 0x13A6 // Request help
	OpEduHomeworkOverdue         OpCode = 0x13A7 // List overdue homework
	OpEduAttachmentAdd           OpCode = 0x13A8 // Add attachment
	OpEduAttachmentList          OpCode = 0x13A9 // List attachments
	OpEduTimetableCreate         OpCode = 0x13B0 // Add lesson
	OpEduTimetableUpdate         OpCode = 0x13B1 // Update lesson
	OpEduTimetableDelete         OpCode = 0x13B2 // Delete lesson
	OpEduTimetableList           OpCode = 0x13B3 // List lessons
	OpEduTimetableToday          OpCode = 0x13B4 // Get today's lessons
	OpEduTimetableWeek           OpCode = 0x13B5 // Get week's lessons
	OpEduMessageSend             OpCode = 0x13C0 // Send message
	OpEduMessageReceive          OpCode = 0x13C1 // Record incoming message
	OpEduMessageList             OpCode = 0x13C2 // List messages
	OpEduMessageRead             OpCode = 0x13C3 // Mark as read
	OpEduAppointmentCreate       OpCode = 0x13C8 // Create appointment
	OpEduAppointmentList         OpCode = 0x13C9 // List appointments
	OpEduAppointmentUpdate       OpCode = 0x13CA // Update appointment
	OpEduAppointmentRSVP         OpCode = 0x13CB // RSVP appointment
	OpEduAppointmentCancel       OpCode = 0x13CC // Cancel appointment
	OpEduAbsenceReport           OpCode = 0x13D0 // Report absence
	OpEduAbsenceList             OpCode = 0x13D1 // List absences
	OpEduAbsenceConfirm          OpCode = 0x13D2 // Confirm absence
	OpEduAbsenceExcuse           OpCode = 0x13D3 // Submit excuse
	OpEduAbsenceMedical          OpCode = 0x13D4 // Submit medical note
	OpEduRecommendList           OpCode = 0x13D8 // List recommendations
	OpEduRecommendSubject        OpCode = 0x13D9 // List subject recommendations
	OpEduRecommendAcknowledge    OpCode = 0x13DA // Acknowledge recommendation
)
