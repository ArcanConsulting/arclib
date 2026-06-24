# arclib

Shared foundation library for the Arc ecosystem.

## What is arclib?

arclib provides the common building blocks used across Arc projects (ArcHub, MyClerk, ArcShell, ArcMail):

- **Crypto** — ChaCha20-Poly1305 (AEAD); X25519 and ML-KEM-768 (post-quantum) key exchange; Ed25519 and ML-DSA-65 (post-quantum) signatures; hybrid identity keys; HKDF-SHA256, X9.63-KDF, Argon2; HMAC-SHA256, SHA-256, CRC-16/32; key rotation and key wrapping; an at-rest SQLite cipher; ZSTD compression.
- **Protocol** — the MyClerk Protocol layer: tiered message framing (tiers 0–5), opcodes/categories/extensions, fragmentation, a hybrid post-quantum handshake (X25519 + ML-KEM-768), Tier-5 sessions with automatic key rotation, and mutual identity authentication. MyClerk is published as an IETF Internet-Draft; arclib is its reference implementation.
- **Transport** — MyClerk over WebSocket (sessions, handshake, dispatch).
- **LLM** — a generic LLM-over-MyClerk contract plus Ollama and OpenAI-compatible clients.
- **DB** — database abstraction for PostgreSQL, SQLite, MariaDB/MySQL with encrypted fields.
- **And more** — rate limiting, WebAuthn/Passkeys, JSON5, structured logging, config loader, MessagePack.

## Languages

Go is the reference implementation; the Kotlin (`kt/`) and JavaScript/TypeScript (`js/`) mirrors
track it against shared cross-validation test vectors for byte-identical output. The most recent
additions (hybrid post-quantum handshake, Tier-5 sessions, identity authentication) land in Go
first; the mirrors follow.

```
go get arcan-it.de/arclib
```

## Module path

`arcan-it.de/arclib` — hosted on [GitHub](https://github.com/arcanconsulting/arclib).
Current version: **v1.4.0**.

## License

arclib is dual-licensed:

- The **`crypto/`** package is licensed under the **Apache License 2.0** (see
  [`crypto/LICENSE`](crypto/LICENSE)) — permissively reusable.
- **Everything else** is licensed under the **Business Source License 1.1** (see
  [`LICENSE`](LICENSE)).

Under BSL 1.1, production use is **free for organizations with annual revenue below
EUR 1,000,000**; larger organizations require a commercial license. Each released version
automatically converts to the **Apache License 2.0** four years after its release.

The Business Source License is not an OSI-approved open-source license, but every version becomes
Apache-2.0 on its change date. For commercial licensing, contact Arcan Consulting.
