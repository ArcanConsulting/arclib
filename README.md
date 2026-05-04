# arclib

Shared foundation library for the Arc ecosystem.

## What is arclib?

arclib provides common components used across Arc projects (ArcHub, MyClerk, ArcShell):

- **Crypto** — ChaCha20-Poly1305, HKDF-SHA256, X25519, ML-KEM-768 (post-quantum)
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

## License

Apache 2.0
