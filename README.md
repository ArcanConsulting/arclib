# arclib

Shared foundation library for the Arc ecosystem.

## What is arclib?

arclib provides common components used across Arc projects (ArcHub, MyClerk, ArcShell):

- **Crypto** — ChaCha20-Poly1305, HKDF-SHA256, X25519, ML-KEM-768 (post-quantum), ZSTD
- **DB** — Database abstraction for PostgreSQL, SQLite, MariaDB/MySQL with encrypted fields
- **Protocol** — MyClerk Protocol layer (tiers, sessions, fragmentation, VFS)
- **And more** — Rate limiting, WebAuthn, JSON5, structured logging, config loader

## Three Languages

arclib is implemented in Go (reference), JavaScript, and Kotlin. All three pass against shared test vectors via cross-validation to guarantee byte-identical output.

```
go get arcan-it.de/arclib          # Go
npm install @arcan/arclib           # JavaScript
implementation 'de.arcan.arclib'    # Kotlin
```

## Status

| Component | Go | JS | Kotlin |
|-----------|----|----|--------|
| Crypto (ChaCha20, HKDF, X25519, ML-KEM-768, HMAC, CRC, ZSTD) | Done | Done | Done |
| DB Abstraction (PostgreSQL, SQLite, MariaDB, encrypted fields) | Done | — | — |
| Protocol (tiers, codec, extensions, fragmentation, VFS) | Done | Done | Done |
| Extras (rate limiter, JSON5, logging, config, msgpack, WebAuthn) | Done | — | — |

Current version: **v1.3.0**

## Module path

`arcan-it.de/arclib` — hosted on [GitHub](https://github.com/arcanconsulting/arclib)

## License

Apache 2.0
