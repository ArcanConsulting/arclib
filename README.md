# arclib

Shared foundation library for the Arc ecosystem.

## What is arclib?

arclib provides common components used across Arc projects (ArcHub, MyClerk, ArcShell):

- **Crypto** — ChaCha20-Poly1305, HKDF-SHA256, X25519, ML-KEM-768 (post-quantum), ZSTD
- **DB** — Database abstraction for PostgreSQL, SQLite, MariaDB/MySQL with encrypted fields
- **Protocol** — MyClerk Protocol layer (tiers, sessions, fragmentation, VFS)
- **And more** — Rate limiting, WebAuthn, JSON5, structured logging, TUI framework

## Three Languages

arclib is implemented in Go (reference), JavaScript, and Kotlin. All three pass against shared test vectors via cross-validation.

```
go get arcan-it.de/arclib          # Go
npm install @arcan/arclib           # JavaScript
implementation 'de.arcan.arclib'    # Kotlin
```

## Status

| Component | Go | JS | Kotlin |
|-----------|----|----|--------|
| Crypto (SHA-256, CRC, HMAC, ChaCha20, HKDF, X25519, ML-KEM-768, Key Rotation, ZSTD) | Done | In Progress | Planned |
| DB Abstraction | Planned | — | — |
| Protocol Layer | Planned | Planned | Planned |
| Extras (Rate Limiter, WebAuthn, JSON5, TUI, Logging) | Planned | — | — |

Current version: **unreleased** (targeting `v0.1.0-alpha.1` after JS + Kotlin crypto cross-validation)

## Repository

- **Gitea**: `gitea.rackpc.de/ck/arclib`
- **Module path**: `arcan-it.de/arclib`

## License

Apache 2.0
