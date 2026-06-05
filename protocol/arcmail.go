package protocol

// ArcMail Mail Service OpCodes (0x3100-0x31FF).
//
// ArcMail is a mail client/service that carries its pkg/core API over the
// MyClerk transport (Tier 3). It reserves the 0x31 category byte — a fresh
// top-level application range, following the SilverWidget Market Data (0x30)
// precedent rather than squatting in an already-assigned range. See
// draft-myclerk-arcmail (the dedicated ArcMail section of the MyClerk Protocol).
// Sub-blocks: 0x310x Accounts, 0x311x Mailboxes, 0x312x Messages.
const (
	// Accounts (0x310x).
	OpArcMailAccountList OpCode = 0x3101 // Accounts.List

	// Mailboxes (0x311x).
	OpArcMailMailboxList OpCode = 0x3110 // Mailboxes.List

	// Messages (0x312x).
	OpArcMailMessageGet      OpCode = 0x3120 // Messages.Get
	OpArcMailMessageList     OpCode = 0x3121 // Messages.List
	OpArcMailMessagePage     OpCode = 0x3122 // Messages.Page
	OpArcMailMessageSetFlags OpCode = 0x3123 // Messages.SetFlags
	OpArcMailMessageBody     OpCode = 0x3124 // Messages.Body

	// Transport control (0x31Fx).
	OpArcMailEvent OpCode = 0x31F0 // Server->client push event (carries no ExtReplyTo)
)

func init() {
	RegisterOpCodes("arcmail", map[OpCode]string{
		OpArcMailAccountList:     "ARCMAIL_ACCOUNT_LIST",
		OpArcMailMailboxList:     "ARCMAIL_MAILBOX_LIST",
		OpArcMailMessageGet:      "ARCMAIL_MESSAGE_GET",
		OpArcMailMessageList:     "ARCMAIL_MESSAGE_LIST",
		OpArcMailMessagePage:     "ARCMAIL_MESSAGE_PAGE",
		OpArcMailMessageSetFlags: "ARCMAIL_MESSAGE_SETFLAGS",
		OpArcMailMessageBody:     "ARCMAIL_MESSAGE_BODY",
		OpArcMailEvent:           "ARCMAIL_EVENT",
	})
}

// IsArcMailOp reports whether op is an ArcMail Mail Service operation
// (0x3100-0x31FF).
func (op OpCode) IsArcMailOp() bool {
	return op.Category() == 0x31
}
