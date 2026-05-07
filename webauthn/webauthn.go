package webauthn

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
)

// Config configures the WebAuthn relying party.
type Config struct {
	RPDisplayName string   // Human-readable RP name, e.g. "ArcHub".
	RPID          string   // RP ID (effective domain), e.g. "localhost".
	RPOrigins     []string // Allowed origins, e.g. ["https://localhost:5680"].
}

// Authenticator wraps the go-webauthn library and exposes a simplified API for
// registration and login ceremonies.
type Authenticator struct {
	wa *gowebauthn.WebAuthn
}

// New creates an Authenticator from the given Config.
func New(cfg Config) (*Authenticator, error) {
	wc := &gowebauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     cfg.RPOrigins,
	}

	wa, err := gowebauthn.New(wc)
	if err != nil {
		return nil, fmt.Errorf("webauthn: create instance: %w", err)
	}

	return &Authenticator{wa: wa}, nil
}

// ---------- Registration ----------------------------------------------------

// BeginRegistration starts the WebAuthn credential registration ceremony.
// It returns JSON-encoded PublicKeyCredentialCreationOptions (to be forwarded
// to the browser) and opaque session data that must be stored server-side
// until FinishRegistration is called.
func (a *Authenticator) BeginRegistration(user User) (options, sessionData []byte, err error) {
	opts, session, err := a.wa.BeginRegistration(&user)
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn: begin registration: %w", err)
	}

	optionsJSON, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn: marshal options: %w", err)
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn: marshal session: %w", err)
	}

	return optionsJSON, sessionJSON, nil
}

// FinishRegistration completes the registration ceremony by validating the
// authenticator response from the browser. On success it returns the new
// Credential that should be persisted.
func (a *Authenticator) FinishRegistration(user User, sessionData, response []byte) (*Credential, error) {
	var session gowebauthn.SessionData
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, fmt.Errorf("webauthn: unmarshal session: %w", err)
	}

	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(response))
	if err != nil {
		return nil, fmt.Errorf("webauthn: parse credential response: %w", err)
	}

	cred, err := a.wa.CreateCredential(&user, session, parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("webauthn: create credential: %w", err)
	}

	return credentialFromLib(cred), nil
}

// ---------- Login -----------------------------------------------------------

// BeginLogin starts the WebAuthn authentication ceremony. It returns
// JSON-encoded PublicKeyCredentialRequestOptions and opaque session data.
func (a *Authenticator) BeginLogin(user User) (options, sessionData []byte, err error) {
	opts, session, err := a.wa.BeginLogin(&user)
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn: begin login: %w", err)
	}

	optionsJSON, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn: marshal options: %w", err)
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, nil, fmt.Errorf("webauthn: marshal session: %w", err)
	}

	return optionsJSON, sessionJSON, nil
}

// FinishLogin completes the authentication ceremony by validating the
// assertion response from the browser. On success the credential's
// sign counter should be updated in storage.
func (a *Authenticator) FinishLogin(user User, sessionData, response []byte) error {
	var session gowebauthn.SessionData
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return fmt.Errorf("webauthn: unmarshal session: %w", err)
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(response))
	if err != nil {
		return fmt.Errorf("webauthn: parse assertion response: %w", err)
	}

	_, err = a.wa.ValidateLogin(&user, session, parsedResponse)
	if err != nil {
		return fmt.Errorf("webauthn: validate login: %w", err)
	}

	return nil
}
