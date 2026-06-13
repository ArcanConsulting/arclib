package protocol

import "testing"

func TestArcMailOpCodesRegistered(t *testing.T) {
	tests := []struct {
		op   OpCode
		name string
	}{
		{OpArcMailAccountList, "ARCMAIL_ACCOUNT_LIST"},
		{OpArcMailMailboxList, "ARCMAIL_MAILBOX_LIST"},
		{OpArcMailMessageGet, "ARCMAIL_MESSAGE_GET"},
		{OpArcMailMessageList, "ARCMAIL_MESSAGE_LIST"},
		{OpArcMailMessagePage, "ARCMAIL_MESSAGE_PAGE"},
		{OpArcMailMessageSetFlags, "ARCMAIL_MESSAGE_SETFLAGS"},
		{OpArcMailMessageBody, "ARCMAIL_MESSAGE_BODY"},
		{OpArcMailCryptoImportKey, "ARCMAIL_CRYPTO_IMPORT_KEY"},
		{OpArcMailCryptoListKeys, "ARCMAIL_CRYPTO_LIST_KEYS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LookupOpCode(tt.op); got != tt.name {
				t.Errorf("LookupOpCode(0x%04X) = %q, want %q", uint16(tt.op), got, tt.name)
			}
			if s := tt.op.String(); s != tt.name {
				t.Errorf("OpCode(0x%04X).String() = %q, want %q", uint16(tt.op), s, tt.name)
			}
			// ArcMail occupies the 0x31 category byte exclusively.
			if cat := tt.op.Category(); cat != 0x31 {
				t.Errorf("OpCode(0x%04X).Category() = 0x%02X, want 0x31", uint16(tt.op), cat)
			}
			if !tt.op.IsArcMailOp() {
				t.Errorf("OpCode(0x%04X).IsArcMailOp() = false, want true", uint16(tt.op))
			}
		})
	}
}
