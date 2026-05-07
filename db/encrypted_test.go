package db

import (
	"context"
	"testing"

	"arcan-it.de/arclib/crypto"
)

func setupEncryptionKey(t *testing.T) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if err := SetEncryptionKey(key[:]); err != nil {
		t.Fatalf("SetEncryptionKey: %v", err)
	}
}

func TestSetEncryptionKey(t *testing.T) {
	key := make([]byte, 32)
	if err := SetEncryptionKey(key); err != nil {
		t.Fatalf("SetEncryptionKey: %v", err)
	}

	short := make([]byte, 16)
	if err := SetEncryptionKey(short); err == nil {
		t.Error("SetEncryptionKey should reject short key")
	}
}

func TestEncrypted_ValueScan_RoundTrip(t *testing.T) {
	setupEncryptionKey(t)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"normal text", "hello world"},
		{"empty string", ""},
		{"unicode", "Sch\u00f6ne Gr\u00fc\u00dfe"},
		{"long text", "The quick brown fox jumps over the lazy dog. " +
			"This is a longer text to test encryption of larger payloads."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEncrypted(tt.plaintext)
			val, err := e.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var decoded Encrypted
			if err := decoded.Scan(val); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if decoded.Plaintext != tt.plaintext {
				t.Errorf("round-trip: got %q, want %q", decoded.Plaintext, tt.plaintext)
			}
		})
	}
}

func TestEncrypted_DifferentCiphertexts(t *testing.T) {
	setupEncryptionKey(t)

	e := NewEncrypted("same plaintext")
	val1, _ := e.Value()
	val2, _ := e.Value()

	if val1 == val2 {
		t.Error("same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestEncrypted_ScanNil(t *testing.T) {
	setupEncryptionKey(t)

	var e Encrypted
	if err := e.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if e.Plaintext != "" {
		t.Errorf("Scan(nil) should set empty plaintext, got %q", e.Plaintext)
	}
}

func TestEncrypted_ScanBytes(t *testing.T) {
	setupEncryptionKey(t)

	original := NewEncrypted("byte scan test")
	val, _ := original.Value()
	encoded := val.(string)

	var e Encrypted
	if err := e.Scan([]byte(encoded)); err != nil {
		t.Fatalf("Scan([]byte): %v", err)
	}
	if e.Plaintext != "byte scan test" {
		t.Errorf("got %q, want %q", e.Plaintext, "byte scan test")
	}
}

func TestEncrypted_ScanUnsupportedType(t *testing.T) {
	setupEncryptionKey(t)

	var e Encrypted
	if err := e.Scan(12345); err == nil {
		t.Error("Scan(int) should fail")
	}
}

func TestEncrypted_NoKeySet(t *testing.T) {
	encryptionKey = nil

	e := NewEncrypted("test")
	_, err := e.Value()
	if err == nil {
		t.Error("Value without key should fail")
	}

	var d Encrypted
	if err := d.Scan("some-base64-data"); err == nil {
		t.Error("Scan without key should fail")
	}
}

func TestEncrypted_String(t *testing.T) {
	e := NewEncrypted("hello")
	if e.String() != "hello" {
		t.Errorf("String() = %q, want %q", e.String(), "hello")
	}
}

func TestEncrypted_InDatabase(t *testing.T) {
	setupEncryptionKey(t)
	conn := openTestDB(t)

	_, err := conn.ExecContext(context.Background(), "CREATE TABLE secrets (id INTEGER PRIMARY KEY, data TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	original := NewEncrypted("database secret value")
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	_, err = conn.ExecContext(context.Background(), "INSERT INTO secrets (id, data) VALUES (?, ?)", 1, val)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	var stored string
	err = conn.QueryRowContext(context.Background(), "SELECT data FROM secrets WHERE id = ?", 1).Scan(&stored)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}

	var decrypted Encrypted
	if err := decrypted.Scan(stored); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if decrypted.Plaintext != "database secret value" {
		t.Errorf("got %q, want %q", decrypted.Plaintext, "database secret value")
	}
}
