package crypto

import (
	"crypto/aes"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
)

// aesKeyWrapIV is the default initial value (the "Alternative Initial Value"
// A6A6A6A6A6A6A6A6, RFC 3394 §2.2.3.1) prepended as the integrity check
// register of the AES Key Wrap.
var aesKeyWrapIV = [8]byte{0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6}

// AESKeyWrap wraps keyData with the key-encryption key kek using the AES Key
// Wrap algorithm (RFC 3394 §2.2.1). The kek length selects AES-128/192/256
// (16/24/32 bytes). keyData must be a whole number of 64-bit blocks and at
// least 16 bytes (two blocks), as RFC 3394 §2 requires — the case used by CMS
// ECDH (wrapping a content-encryption key, RFC 5753). The result is
// len(keyData)+8 bytes (the integrity register prepended).
func AESKeyWrap(kek, keyData []byte) ([]byte, error) {
	if len(keyData) < 16 || len(keyData)%8 != 0 {
		return nil, fmt.Errorf("crypto: AES key wrap: key data must be a multiple of 8 bytes and at least 16, got %d", len(keyData))
	}
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("crypto: AES key wrap: %w", err)
	}
	n := len(keyData) / 8
	// R holds the n 64-bit registers R[1..n]; a is the 64-bit integrity register A.
	r := make([]byte, len(keyData))
	copy(r, keyData)
	a := aesKeyWrapIV
	var b [16]byte
	for j := 0; j < 6; j++ {
		for i := 1; i <= n; i++ {
			copy(b[:8], a[:])
			copy(b[8:], r[(i-1)*8:i*8])
			block.Encrypt(b[:], b[:])
			// A = MSB64(B) XOR t, where t = (n*j)+i; R[i] = LSB64(B).
			t := uint64(n*j + i) //nolint:gosec // n,j,i are small loop bounds, no overflow
			binary.BigEndian.PutUint64(a[:], binary.BigEndian.Uint64(b[:8])^t)
			copy(r[(i-1)*8:i*8], b[8:])
		}
	}
	out := make([]byte, 0, len(keyData)+8)
	out = append(out, a[:]...)
	out = append(out, r...)
	return out, nil
}

// AESKeyUnwrap reverses AESKeyWrap (RFC 3394 §2.2.2): it recovers the wrapped
// key data and verifies the integrity check value in constant time. wrapped
// must be a whole number of 64-bit blocks and at least 24 bytes. A failed
// integrity check returns an error and no key material (the recovered buffer is
// zeroed first), so a tampered or wrong-KEK input never yields a usable key.
func AESKeyUnwrap(kek, wrapped []byte) ([]byte, error) {
	if len(wrapped) < 24 || len(wrapped)%8 != 0 {
		return nil, fmt.Errorf("crypto: AES key unwrap: wrapped data must be a multiple of 8 bytes and at least 24, got %d", len(wrapped))
	}
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("crypto: AES key unwrap: %w", err)
	}
	n := len(wrapped)/8 - 1
	var a [8]byte
	copy(a[:], wrapped[:8])
	r := make([]byte, len(wrapped)-8)
	copy(r, wrapped[8:])
	var b [16]byte
	for j := 5; j >= 0; j-- {
		for i := n; i >= 1; i-- {
			// A = A XOR t, then B = AES-Decrypt(A | R[i]); A = MSB64(B); R[i] = LSB64(B).
			t := uint64(n*j + i) //nolint:gosec // n,j,i are small loop bounds, no overflow
			binary.BigEndian.PutUint64(a[:], binary.BigEndian.Uint64(a[:])^t)
			copy(b[:8], a[:])
			copy(b[8:], r[(i-1)*8:i*8])
			block.Decrypt(b[:], b[:])
			copy(a[:], b[:8])
			copy(r[(i-1)*8:i*8], b[8:])
		}
	}
	if subtle.ConstantTimeCompare(a[:], aesKeyWrapIV[:]) != 1 {
		ZeroBytes(r)
		return nil, fmt.Errorf("crypto: AES key unwrap: integrity check failed")
	}
	return r, nil
}
