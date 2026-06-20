package protocol

// ArcMail Mail Service OpCodes (0x3100-0x31FF).
//
// ArcMail is a mail client/service that carries its pkg/core API over the
// MyClerk transport (Tier 3). It reserves the 0x31 category byte — a fresh
// top-level application range, following the SilverWidget Market Data (0x30)
// precedent rather than squatting in an already-assigned range. See
// draft-myclerk-arcmail (the dedicated ArcMail section of the MyClerk Protocol).
// Sub-blocks: 0x310x Accounts, 0x311x Mailboxes, 0x312x Messages, 0x313x Outbox,
// 0x314x Identities, 0x315x Drafts, 0x316x Crypto, 0x317x Quarantine, 0x318x SmartFolders,
// 0x319x Tags, 0x31Ax Sieve, 0x31Bx Rules.
const (
	// Accounts (0x310x).
	OpArcMailAccountList OpCode = 0x3101 // Accounts.List

	// Mailboxes (0x311x).
	OpArcMailMailboxList OpCode = 0x3110 // Mailboxes.List

	// Messages (0x312x).
	OpArcMailMessageGet         OpCode = 0x3120 // Messages.Get
	OpArcMailMessageList        OpCode = 0x3121 // Messages.List
	OpArcMailMessagePage        OpCode = 0x3122 // Messages.Page
	OpArcMailMessageSetFlags    OpCode = 0x3123 // Messages.SetFlags
	OpArcMailMessageBody        OpCode = 0x3124 // Messages.Body
	OpArcMailMessageAttachments OpCode = 0x3125 // Messages.Attachments (forward, D-8-11)
	OpArcMailMessageAttachment  OpCode = 0x3126 // Messages.Attachment (open one + scan, D-10-5)
	OpArcMailMessageSearch      OpCode = 0x3127 // Messages.Search (FTS5 full-text search, D-6-5)

	// Outbox (0x313x): the send queue (D-8-3).
	OpArcMailOutboxSend   OpCode = 0x3130 // Outbox.Send (enqueue an outgoing message)
	OpArcMailOutboxList   OpCode = 0x3131 // Outbox.List
	OpArcMailOutboxRetry  OpCode = 0x3132 // Outbox.Retry (requeue a failed item)
	OpArcMailOutboxCancel OpCode = 0x3133 // Outbox.Cancel (discard a queued item)

	// Identities (0x314x): sending identities / From addresses (D-4-2).
	OpArcMailIdentityList        OpCode = 0x3140 // Identities.List
	OpArcMailIdentitySelectReply OpCode = 0x3141 // Identities.SelectReply (Auto-From, D-4-6)

	// Drafts (0x315x): unsent messages in the server-side IMAP Drafts folder (D-8-12).
	OpArcMailDraftSave   OpCode = 0x3150 // Drafts.Save (APPEND to the Drafts mailbox)
	OpArcMailDraftDelete OpCode = 0x3151 // Drafts.Delete (UID-EXPUNGE a draft)

	// Crypto (0x316x): message-level crypto key management (PGP/S-MIME, D-10-11).
	OpArcMailCryptoImportKey       OpCode = 0x3160 // Crypto.ImportKey (import a PGP key / S-MIME cert or identity)
	OpArcMailCryptoListKeys        OpCode = 0x3161 // Crypto.ListKeys (enumerate known keys)
	OpArcMailCryptoDeleteKey       OpCode = 0x3162 // Crypto.DeleteKey (remove a key by fingerprint, D-10-13)
	OpArcMailCryptoSetTrust        OpCode = 0x3163 // Crypto.SetCertTrust (mark an S/MIME cert user-trusted, D-10-19)
	OpArcMailCryptoRecipientStatus OpCode = 0x3164 // Crypto.RecipientKeyStatus (per-recipient encryption-key availability, D-10-21)
	OpArcMailCryptoDiscoverKeys    OpCode = 0x3165 // Crypto.DiscoverKeys (fetch missing correspondent keys: GnuPG/WKD/keyserver/Autocrypt, D-10-24/D-10-2)

	// Quarantine (0x317x): malware attachments retained for inspection/release (D-10-23).
	OpArcMailQuarantineList    OpCode = 0x3170 // Quarantine.List (enumerate quarantined attachments)
	OpArcMailQuarantineRelease OpCode = 0x3171 // Quarantine.Release (return a quarantined attachment's withheld bytes)
	OpArcMailQuarantineDelete  OpCode = 0x3172 // Quarantine.Delete (discard a quarantined attachment)

	// Smart folders (0x318x): saved searches shown as virtual folders (D-6-3/D-6-5).
	OpArcMailSmartFolderList   OpCode = 0x3180 // SmartFolders.List (enumerate an account's saved searches)
	OpArcMailSmartFolderSave   OpCode = 0x3181 // SmartFolders.Save (create/update a saved search by name)
	OpArcMailSmartFolderDelete OpCode = 0x3182 // SmartFolders.Delete (remove a saved search)

	// Tags (0x319x): colour-coded IMAP keywords, multiple per message (D-9-2).
	OpArcMailTagList     OpCode = 0x3190 // Tags.List (enumerate an account's tags + colours)
	OpArcMailTagAdd      OpCode = 0x3191 // Tags.Add (attach a keyword to a message, write-through)
	OpArcMailTagRemove   OpCode = 0x3192 // Tags.Remove (detach a keyword from a message)
	OpArcMailTagSetColor OpCode = 0x3193 // Tags.SetColor (set a tag's local colour token)

	// Sieve (0x31Ax): server-side filter scripts over ManageSieve (RFC 5804, D-9-1).
	OpArcMailSieveList      OpCode = 0x31A0 // Sieve.List (enumerate an account's scripts + active)
	OpArcMailSieveGet       OpCode = 0x31A1 // Sieve.Get (fetch a script's body)
	OpArcMailSievePut       OpCode = 0x31A2 // Sieve.Put (create/replace a script; server compiles)
	OpArcMailSieveSetActive OpCode = 0x31A3 // Sieve.SetActive (activate a script, or deactivate all)
	OpArcMailSieveDelete    OpCode = 0x31A4 // Sieve.Delete (remove a script)
	OpArcMailSieveCheck     OpCode = 0x31A5 // Sieve.Check (validate syntax without storing)

	// Rules (0x31Bx): client-side filter rules evaluated in the core on new mail
	// post-sync (D-9-5).
	OpArcMailRuleList   OpCode = 0x31B0 // Rules.List (enumerate an account's rules)
	OpArcMailRuleSave   OpCode = 0x31B1 // Rules.Save (create/replace a rule)
	OpArcMailRuleDelete OpCode = 0x31B2 // Rules.Delete (remove a rule)

	// Transport control (0x31Fx).
	OpArcMailEvent OpCode = 0x31F0 // Server->client push event (carries no ExtReplyTo)
)

func init() {
	RegisterOpCodes("arcmail", map[OpCode]string{
		OpArcMailAccountList:           "ARCMAIL_ACCOUNT_LIST",
		OpArcMailMailboxList:           "ARCMAIL_MAILBOX_LIST",
		OpArcMailMessageGet:            "ARCMAIL_MESSAGE_GET",
		OpArcMailMessageList:           "ARCMAIL_MESSAGE_LIST",
		OpArcMailMessagePage:           "ARCMAIL_MESSAGE_PAGE",
		OpArcMailMessageSetFlags:       "ARCMAIL_MESSAGE_SETFLAGS",
		OpArcMailMessageBody:           "ARCMAIL_MESSAGE_BODY",
		OpArcMailMessageAttachments:    "ARCMAIL_MESSAGE_ATTACHMENTS",
		OpArcMailMessageAttachment:     "ARCMAIL_MESSAGE_ATTACHMENT",
		OpArcMailMessageSearch:         "ARCMAIL_MESSAGE_SEARCH",
		OpArcMailOutboxSend:            "ARCMAIL_OUTBOX_SEND",
		OpArcMailOutboxList:            "ARCMAIL_OUTBOX_LIST",
		OpArcMailOutboxRetry:           "ARCMAIL_OUTBOX_RETRY",
		OpArcMailOutboxCancel:          "ARCMAIL_OUTBOX_CANCEL",
		OpArcMailIdentityList:          "ARCMAIL_IDENTITY_LIST",
		OpArcMailIdentitySelectReply:   "ARCMAIL_IDENTITY_SELECT_REPLY",
		OpArcMailDraftSave:             "ARCMAIL_DRAFT_SAVE",
		OpArcMailDraftDelete:           "ARCMAIL_DRAFT_DELETE",
		OpArcMailCryptoImportKey:       "ARCMAIL_CRYPTO_IMPORT_KEY",
		OpArcMailCryptoListKeys:        "ARCMAIL_CRYPTO_LIST_KEYS",
		OpArcMailCryptoDeleteKey:       "ARCMAIL_CRYPTO_DELETE_KEY",
		OpArcMailCryptoSetTrust:        "ARCMAIL_CRYPTO_SET_TRUST",
		OpArcMailCryptoRecipientStatus: "ARCMAIL_CRYPTO_RECIPIENT_STATUS",
		OpArcMailCryptoDiscoverKeys:    "ARCMAIL_CRYPTO_DISCOVER_KEYS",
		OpArcMailQuarantineList:        "ARCMAIL_QUARANTINE_LIST",
		OpArcMailQuarantineRelease:     "ARCMAIL_QUARANTINE_RELEASE",
		OpArcMailQuarantineDelete:      "ARCMAIL_QUARANTINE_DELETE",
		OpArcMailSmartFolderList:       "ARCMAIL_SMARTFOLDER_LIST",
		OpArcMailSmartFolderSave:       "ARCMAIL_SMARTFOLDER_SAVE",
		OpArcMailSmartFolderDelete:     "ARCMAIL_SMARTFOLDER_DELETE",
		OpArcMailTagList:               "ARCMAIL_TAG_LIST",
		OpArcMailTagAdd:                "ARCMAIL_TAG_ADD",
		OpArcMailTagRemove:             "ARCMAIL_TAG_REMOVE",
		OpArcMailTagSetColor:           "ARCMAIL_TAG_SET_COLOR",
		OpArcMailSieveList:             "ARCMAIL_SIEVE_LIST",
		OpArcMailSieveGet:              "ARCMAIL_SIEVE_GET",
		OpArcMailSievePut:              "ARCMAIL_SIEVE_PUT",
		OpArcMailSieveSetActive:        "ARCMAIL_SIEVE_SET_ACTIVE",
		OpArcMailSieveDelete:           "ARCMAIL_SIEVE_DELETE",
		OpArcMailSieveCheck:            "ARCMAIL_SIEVE_CHECK",
		OpArcMailRuleList:              "ARCMAIL_RULE_LIST",
		OpArcMailRuleSave:              "ARCMAIL_RULE_SAVE",
		OpArcMailRuleDelete:            "ARCMAIL_RULE_DELETE",
		OpArcMailEvent:                 "ARCMAIL_EVENT",
	})
}

// IsArcMailOp reports whether op is an ArcMail Mail Service operation
// (0x3100-0x31FF).
func (op OpCode) IsArcMailOp() bool {
	return op.Category() == 0x31
}
