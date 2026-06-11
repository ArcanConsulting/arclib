package crypto

import (
	"bytes"
	"testing"
)

func TestSQLiteVFSCipherKeyLength(t *testing.T) {
	for _, n := range []int{0, 16, 31, 33, 64} {
		if SQLiteVFSCipher(make([]byte, n)) != nil {
			t.Errorf("SQLiteVFSCipher(%d-byte key) must be nil", n)
		}
	}
	if SQLiteVFSCipher(make([]byte, SQLiteVFSKeySize)) == nil {
		t.Fatal("SQLiteVFSCipher(32-byte key) must build a cipher")
	}
}

func TestSQLiteVFSCipherRoundTrip(t *testing.T) {
	key := make([]byte, SQLiteVFSKeySize)
	for i := range key {
		key[i] = byte(i)
	}
	tweak := []byte{1, 2, 3, 4, 5, 6, 7, 8} // the VFS uses the file offset as tweak
	pt := bytes.Repeat([]byte("arcmail-page!"), 316)[:4096]

	c := SQLiteVFSCipher(key)
	ct := c.Encrypt(append([]byte(nil), pt...), tweak)
	if bytes.Equal(ct, pt) {
		t.Fatal("ciphertext must differ from plaintext")
	}
	dt := c.Decrypt(append([]byte(nil), ct...), tweak)
	if !bytes.Equal(dt, pt) {
		t.Fatal("decrypt did not recover the plaintext")
	}

	// Determinism across instances: the same key+tweak must reproduce the same
	// ciphertext, otherwise databases written earlier could not be decrypted.
	ct2 := SQLiteVFSCipher(key).Encrypt(append([]byte(nil), pt...), tweak)
	if !bytes.Equal(ct, ct2) {
		t.Fatal("cipher is not deterministic for the same key/tweak")
	}

	// A different tweak (page offset) yields different ciphertext.
	other := SQLiteVFSCipher(key).Encrypt(append([]byte(nil), pt...), []byte{9, 9, 9, 9, 9, 9, 9, 9})
	if bytes.Equal(ct, other) {
		t.Error("ciphertext must depend on the tweak")
	}
}
