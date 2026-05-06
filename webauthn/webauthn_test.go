package webauthn

import (
	"encoding/json"
	"testing"
)

func TestNewAuthenticator(t *testing.T) {
	auth, err := New(Config{
		RPDisplayName: "ArcHub",
		RPID:          "localhost",
		RPOrigins:     []string{"https://localhost:5680"},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if auth == nil {
		t.Fatal("expected non-nil Authenticator")
	}
	if auth.wa == nil {
		t.Fatal("expected non-nil inner webauthn instance")
	}
}

func TestNewAuthenticatorInvalidConfig(t *testing.T) {
	// Empty RPID should fail.
	_, err := New(Config{
		RPDisplayName: "",
		RPID:          "",
		RPOrigins:     nil,
	})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestBeginRegistrationReturnsValidJSON(t *testing.T) {
	auth, err := New(Config{
		RPDisplayName: "TestApp",
		RPID:          "localhost",
		RPOrigins:     []string{"https://localhost"},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	user := User{
		ID:          []byte("user-123"),
		Name:        "alice@example.com",
		DisplayName: "Alice",
	}

	options, sessionData, err := auth.BeginRegistration(user)
	if err != nil {
		t.Fatalf("BeginRegistration() failed: %v", err)
	}

	// Options must be valid JSON.
	if !json.Valid(options) {
		t.Fatal("options is not valid JSON")
	}

	// Session data must be valid JSON.
	if !json.Valid(sessionData) {
		t.Fatal("sessionData is not valid JSON")
	}

	// Options must contain publicKey field.
	var opts map[string]any
	if err := json.Unmarshal(options, &opts); err != nil {
		t.Fatalf("unmarshal options: %v", err)
	}

	pubKey, ok := opts["publicKey"]
	if !ok {
		t.Fatal("options missing 'publicKey' field")
	}

	pkMap, ok := pubKey.(map[string]any)
	if !ok {
		t.Fatal("publicKey is not an object")
	}

	// Must contain challenge.
	if _, ok := pkMap["challenge"]; !ok {
		t.Fatal("publicKey missing 'challenge' field")
	}

	// Must contain rp info.
	rpInfo, ok := pkMap["rp"]
	if !ok {
		t.Fatal("publicKey missing 'rp' field")
	}

	rpMap, ok := rpInfo.(map[string]any)
	if !ok {
		t.Fatal("rp is not an object")
	}

	if rpMap["name"] != "TestApp" {
		t.Fatalf("expected rp.name='TestApp', got %v", rpMap["name"])
	}
	if rpMap["id"] != "localhost" {
		t.Fatalf("expected rp.id='localhost', got %v", rpMap["id"])
	}
}

func TestBeginLoginRequiresCredentials(t *testing.T) {
	auth, err := New(Config{
		RPDisplayName: "TestApp",
		RPID:          "localhost",
		RPOrigins:     []string{"https://localhost"},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// A user with no credentials should cause BeginLogin to fail
	// because the library requires at least one credential for
	// a non-discoverable login flow.
	user := User{
		ID:   []byte("user-456"),
		Name: "bob@example.com",
	}

	_, _, err = auth.BeginLogin(user)
	if err == nil {
		t.Fatal("expected error when user has no credentials")
	}
}

func TestUserImplementsInterface(t *testing.T) {
	u := User{
		ID:          []byte("test-id"),
		Name:        "testuser",
		DisplayName: "Test User",
		Credentials: []Credential{
			{
				ID:              []byte("cred-1"),
				PublicKey:       []byte("pk-1"),
				AttestationType: "none",
				SignCount:       5,
			},
		},
	}

	if string(u.WebAuthnID()) != "test-id" {
		t.Fatalf("unexpected ID: %s", u.WebAuthnID())
	}
	if u.WebAuthnName() != "testuser" {
		t.Fatalf("unexpected name: %s", u.WebAuthnName())
	}
	if u.WebAuthnDisplayName() != "Test User" {
		t.Fatalf("unexpected display name: %s", u.WebAuthnDisplayName())
	}

	creds := u.WebAuthnCredentials()
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
	if string(creds[0].ID) != "cred-1" {
		t.Fatalf("unexpected credential ID: %s", creds[0].ID)
	}
	if creds[0].Authenticator.SignCount != 5 {
		t.Fatalf("unexpected sign count: %d", creds[0].Authenticator.SignCount)
	}
}

func TestUserDisplayNameFallback(t *testing.T) {
	u := User{
		ID:   []byte("x"),
		Name: "fallback-name",
	}
	if u.WebAuthnDisplayName() != "fallback-name" {
		t.Fatalf("expected fallback to Name, got %s", u.WebAuthnDisplayName())
	}
}

func TestCredentialIDBase64(t *testing.T) {
	c := Credential{
		ID: []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}
	encoded := c.CredentialIDBase64()
	if encoded == "" {
		t.Fatal("expected non-empty base64 encoding")
	}
	// base64url of 0xDEADBEEF = "3q2-7w"
	if encoded != "3q2-7w" {
		t.Fatalf("expected '3q2-7w', got %q", encoded)
	}
}
