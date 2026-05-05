package protocol

import "testing"

func TestTierIsValid(t *testing.T) {
	tests := []struct {
		tier  Tier
		valid bool
	}{
		{TierPlaintext, true},
		{TierChecksum, true},
		{TierAuthenticated, true},
		{TierEncrypted, true},
		{TierPFS, true},
		{TierMaxSecurity, true},
		{Tier(6), false},
		{Tier(255), false},
	}

	for _, tt := range tests {
		if got := tt.tier.IsValid(); got != tt.valid {
			t.Errorf("Tier(%d).IsValid() = %v, want %v", tt.tier, got, tt.valid)
		}
	}
}

func TestTierString(t *testing.T) {
	tests := []struct {
		tier Tier
		name string
	}{
		{TierPlaintext, "plaintext"},
		{TierChecksum, "checksum"},
		{TierAuthenticated, "authenticated"},
		{TierEncrypted, "encrypted"},
		{TierPFS, "pfs"},
		{TierMaxSecurity, "max_security"},
		{Tier(6), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.tier.String(); got != tt.name {
			t.Errorf("Tier(%d).String() = %q, want %q", tt.tier, got, tt.name)
		}
	}
}

func TestTierInfo(t *testing.T) {
	info := TierPFS.Info()
	if !info.Encrypted {
		t.Error("TierPFS should be encrypted")
	}
	if !info.Authenticated {
		t.Error("TierPFS should be authenticated")
	}
	if !info.HasSequence {
		t.Error("TierPFS should have sequence")
	}
	if !info.HasSessionID {
		t.Error("TierPFS should have session ID")
	}
	if !info.HasNonce {
		t.Error("TierPFS should have nonce")
	}
	if !info.HasKeyID {
		t.Error("TierPFS should have key ID")
	}
	if !info.HasECDH {
		t.Error("TierPFS should have ECDH")
	}
	if info.HasTimestamp {
		t.Error("TierPFS should not have timestamp")
	}
}

func TestTierInfoPlaintext(t *testing.T) {
	info := TierPlaintext.Info()
	if info.Encrypted {
		t.Error("TierPlaintext should not be encrypted")
	}
	if info.Authenticated {
		t.Error("TierPlaintext should not be authenticated")
	}
	if info.HeaderOverhead != 0 {
		t.Errorf("TierPlaintext HeaderOverhead = %d, want 0", info.HeaderOverhead)
	}
	if info.TrailerOverhead != 0 {
		t.Errorf("TierPlaintext TrailerOverhead = %d, want 0", info.TrailerOverhead)
	}
}

func TestTierInfoMaxSecurity(t *testing.T) {
	info := TierMaxSecurity.Info()
	if !info.HasTimestamp {
		t.Error("TierMaxSecurity should have timestamp")
	}
	if !info.HasECDH {
		t.Error("TierMaxSecurity should have ECDH")
	}
	if info.TotalOverhead != 147 {
		t.Errorf("TierMaxSecurity TotalOverhead = %d, want 147", info.TotalOverhead)
	}
}

func TestTierRequiresEncryption(t *testing.T) {
	tests := []struct {
		tier     Tier
		expected bool
	}{
		{TierPlaintext, false},
		{TierChecksum, false},
		{TierAuthenticated, false},
		{TierEncrypted, true},
		{TierPFS, true},
		{TierMaxSecurity, true},
	}

	for _, tt := range tests {
		if got := tt.tier.RequiresEncryption(); got != tt.expected {
			t.Errorf("Tier(%d).RequiresEncryption() = %v, want %v", tt.tier, got, tt.expected)
		}
	}
}

func TestTierRequiresAuthentication(t *testing.T) {
	tests := []struct {
		tier     Tier
		expected bool
	}{
		{TierPlaintext, false},
		{TierChecksum, false},
		{TierAuthenticated, true},
		{TierEncrypted, true},
		{TierPFS, true},
		{TierMaxSecurity, true},
	}

	for _, tt := range tests {
		if got := tt.tier.RequiresAuthentication(); got != tt.expected {
			t.Errorf("Tier(%d).RequiresAuthentication() = %v, want %v", tt.tier, got, tt.expected)
		}
	}
}

func TestTierInfoInvalid(t *testing.T) {
	info := Tier(99).Info()
	if info.Name != "unknown" {
		t.Errorf("Invalid tier info Name = %q, want %q", info.Name, "unknown")
	}
}
