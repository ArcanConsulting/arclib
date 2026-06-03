package crypto

import (
	"bytes"
	"testing"
)

// fastParams keeps the tests quick while exercising the same algorithm.
func fastParams() Argon2idParams {
	return Argon2idParams{Time: 1, Memory: 8 * 1024, Threads: 1, KeyLen: KeySize}
}

func TestDeriveKeyArgon2id_Deterministic(t *testing.T) {
	pw := []byte("correct horse battery staple")
	salt := []byte("0123456789abcdef")
	a := DeriveKeyArgon2id(pw, salt, fastParams())
	b := DeriveKeyArgon2id(pw, salt, fastParams())
	if !bytes.Equal(a, b) {
		t.Fatal("same password+salt+params must derive the same key")
	}
	if len(a) != int(fastParams().KeyLen) {
		t.Fatalf("key length = %d, want %d", len(a), fastParams().KeyLen)
	}
}

func TestDeriveKeyArgon2id_DifferentSalt(t *testing.T) {
	pw := []byte("password")
	a := DeriveKeyArgon2id(pw, []byte("salt-one-0123456"), fastParams())
	b := DeriveKeyArgon2id(pw, []byte("salt-two-0123456"), fastParams())
	if bytes.Equal(a, b) {
		t.Fatal("different salts must derive different keys")
	}
}

func TestDeriveKeyArgon2id_DifferentPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	a := DeriveKeyArgon2id([]byte("password-a"), salt, fastParams())
	b := DeriveKeyArgon2id([]byte("password-b"), salt, fastParams())
	if bytes.Equal(a, b) {
		t.Fatal("different passwords must derive different keys")
	}
}

func TestDefaultArgon2idParams(t *testing.T) {
	p := DefaultArgon2idParams()
	if p.Time < 1 || p.Memory < 8*1024 || p.Threads < 1 || p.KeyLen != KeySize {
		t.Fatalf("default params look unsafe: %+v", p)
	}
}
