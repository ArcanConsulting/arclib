// Package webauthn provides WebAuthn/Passkeys (FIDO2) registration and
// authentication flows built on top of the go-webauthn library. It is designed
// as a reusable building block: the caller provides user/credential storage
// while this package handles the cryptographic ceremony.
package webauthn

import (
	"encoding/base64"
	"time"

	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
)

// User represents a WebAuthn-capable user. The caller must populate ID, Name,
// and Credentials before passing User to registration or login flows. User
// implements the webauthn.User interface required by go-webauthn.
type User struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []Credential
}

// Credential represents a single stored WebAuthn credential (passkey,
// security key, platform authenticator, etc.).
type Credential struct {
	ID              []byte
	PublicKey       []byte
	AttestationType string
	AAGUID          []byte
	SignCount       uint32
	CreatedAt       time.Time
}

// --- webauthn.User interface ------------------------------------------------

// WebAuthnID returns the user handle used in WebAuthn ceremonies.
func (u User) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName returns the human-readable account name (e.g. email).
func (u User) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName returns the display name shown during ceremonies.
func (u User) WebAuthnDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Name
}

// WebAuthnCredentials converts the stored credentials into the format
// required by go-webauthn.
func (u User) WebAuthnCredentials() []gowebauthn.Credential {
	out := make([]gowebauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		out[i] = gowebauthn.Credential{
			ID:              c.ID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Authenticator: gowebauthn.Authenticator{
				AAGUID:    c.AAGUID,
				SignCount: c.SignCount,
			},
		}
	}
	return out
}

// credentialFromLib converts a go-webauthn Credential into our Credential type.
func credentialFromLib(c *gowebauthn.Credential) *Credential {
	attestation := c.AttestationType
	if attestation == "" {
		attestation = "none"
	}
	return &Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: attestation,
		AAGUID:          c.Authenticator.AAGUID,
		SignCount:       c.Authenticator.SignCount,
		CreatedAt:       time.Now().UTC(),
	}
}

// CredentialIDBase64 returns the credential ID as a base64url string without
// padding, which is the standard encoding used by WebAuthn.
func (c Credential) CredentialIDBase64() string {
	return base64.RawURLEncoding.EncodeToString(c.ID)
}
