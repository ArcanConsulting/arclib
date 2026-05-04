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
