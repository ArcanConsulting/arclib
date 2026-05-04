# arclib — Project Configuration

## Overview

arclib is the shared foundation library for the Arc ecosystem (ArcHub, MyClerk, ArcShell).
Module path: `arcan-it.de/arclib`
Repository: `gitea.rackpc.de/ck/arclib`
License: Apache 2.0

## Architecture

Three-language library: Go (reference), JavaScript, Kotlin.
Go is the source of truth. JS and Kotlin must pass against Go-generated test vectors.

```
arclib/
├── crypto/      → ChaCha20-Poly1305, HKDF, X25519, ML-KEM-768, etc.
├── db/          → DB Abstraction (PostgreSQL, SQLite, MariaDB/MySQL)
├── protocol/    → MyClerk Protocol Layer
├── ratelimit/   → Token-Bucket, Sliding-Window
├── webauthn/    → Passkey Registration & Auth
├── msgpack/     → MessagePack Serialization
├── tui/         → Bubbletea TUI Framework
├── log/         → Structured Logging
├── json5/       → JSON5 Parser (self-implemented)
├── config/      → Config Loader + Validation
├── js/          → JavaScript Package (@arcan/arclib)
├── kt/          → Kotlin Package (de.arcan.arclib)
└── testdata/    → Shared test vectors (JSON)
```

## Critical Rules

- NO TODOs in code
- NO mocks in integration tests
- 90%+ test coverage
- All crypto via arclib — no direct stdlib crypto calls from consuming projects
- Cross-validation (Go↔JS↔Kotlin) is a CI gate for crypto and protocol
- `golangci-lint` and `gosec` must pass with 0 findings

## Crypto Rules

- All randomness via `crypto/rand`
- Key material zeroed after use
- No hardcoded keys, salts, or nonces
- ChaCha20-Poly1305 for AEAD (not AES)
- HKDF-SHA256 for key derivation
- Hybrid ML-KEM-768 + X25519 for key exchange

## Test Vectors

Test vectors in `testdata/` are JSON files with input/output pairs.
All three languages must produce identical output for identical input.
Format:
```json
{
  "algorithm": "chacha20-poly1305",
  "vectors": [
    {
      "key": "hex-encoded",
      "nonce": "hex-encoded",
      "plaintext": "hex-encoded",
      "aad": "hex-encoded",
      "ciphertext": "hex-encoded",
      "tag": "hex-encoded"
    }
  ]
}
```

## Workflow

- Feature branches: `feature/<name>`
- Commit messages: imperative, short
- Every PR needs green CI
- No force push on main
