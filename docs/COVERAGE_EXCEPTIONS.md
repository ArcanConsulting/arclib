# Coverage Exceptions

Coverage gate: 95% per package. Exceptions must be approved individually.

## Approved Exceptions

### crypto (93.7%)

**Affected functions:**
- `GenerateKey` (75%) — error path requires `crypto/rand.Read` failure (broken OS entropy)
- `GenerateNonce` (75%) — same as above
- `NewAEAD` (83.3%) — error path requires `chacha20poly1305.New` failure (unreachable with correct KeySize validation)
- `HKDFDeriveKey` (80%) — error path requires `io.ReadFull` failure on HKDF reader (unreachable in practice)
- `HKDFDeriveKeys` (93.3%) — same as above

**Reason:** These error paths only trigger when the OS cryptographic random source is broken or when Go's standard crypto libraries fail internally. Mocking `crypto/rand` to test these paths would require dependency injection that adds complexity without value — the correct behavior on `crypto/rand` failure is to propagate the error, which the code does.

**Approved:** 2026-05-04

### crypto — NewCompressor (62.5%)

**Affected function:**
- `NewCompressor` (62.5%) — error paths require `zstd.NewWriter` or `zstd.NewReader` to fail internally, plus the cleanup path (close encoder when decoder creation fails)

**Reason:** The klauspost/zstd library does not return errors with the parameters used (level 3, nil writer). The error handling is defensive but unreachable in practice.

**Approved:** 2026-05-04

### transport (93.3%)

**Affected functions:**
- `clientHandshake` (86.5%) / `serverHandshake` (85.3%) — error paths require `crypto/rand` keygen failures (`GenerateX25519KeyPair`, `GenerateMLKEMKeyPair`, `GenerateSessionSalt`) and `NewAeadSession` failure; all unreachable with valid entropy and validated inputs.
- `writeTier0` (75%) / `encodeError` (75%) — `Message.Marshal` / `msgpack.Marshal` of valid, fixed-shape values never error.
- `Session.Send` (75%) — `AeadSession.Encode` only errors on 2^32 sequence exhaustion (a reconnect condition); the write-error branch needs a connection that dies mid-stream.
- `client.readLoop` (88.9%) — the client-side decrypt-error branch requires a correct server to emit an undecryptable frame, which it never does.
- `server.readLoop` (88.9%) — response-send-failure branches require the peer socket to fail between dispatch and reply.

**Reason:** These are defensive error guards around `crypto/rand`, MessagePack encoding of fixed structs, and mid-frame socket death. The happy paths and all reachable error paths (handshake rejection, AEAD decrypt failure, timeouts, marshal failure on caller payloads, server shutdown propagation) are covered by the end-to-end loopback tests. Fault-injecting `crypto/rand` or the socket would add dependency-injection complexity without testing real behavior.

**Approved:** 2026-06-16
