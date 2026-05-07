package crypto

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

type aeadVector struct {
	Key        string `json:"key"`
	Nonce      string `json:"nonce"`
	Plaintext  string `json:"plaintext"`
	AAD        string `json:"aad"`
	Ciphertext string `json:"ciphertext"`
}

type hkdfVector struct {
	IKM    string `json:"ikm"`
	Salt   string `json:"salt"`
	Info   string `json:"info"`
	Length int    `json:"length"`
	OKM    string `json:"okm"`
}

type x25519Vector struct {
	AlicePrivate string `json:"alice_private"`
	AlicePublic  string `json:"alice_public"`
	BobPrivate   string `json:"bob_private"`
	BobPublic    string `json:"bob_public"`
	SharedSecret string `json:"shared_secret"`
}

type rotationVector struct {
	BaseKey    string `json:"base_key"`
	Counter    uint32 `json:"counter"`
	DerivedKey string `json:"derived_key"`
}

type vectorFile struct {
	AEAD     []aeadVector     `json:"chacha20_poly1305"`
	HKDF     []hkdfVector     `json:"hkdf_sha256"`
	X25519   []x25519Vector   `json:"x25519"`
	Rotation []rotationVector `json:"key_rotation"`
}

func TestGenerateVectors(t *testing.T) {
	if os.Getenv("GENERATE_VECTORS") == "" {
		t.Skip("set GENERATE_VECTORS=1 to regenerate test vectors")
	}

	vectors := vectorFile{}

	// AEAD vectors
	for i := 0; i < 3; i++ {
		key, _ := GenerateKey()
		nonce, _ := GenerateNonce()
		plaintexts := []string{"", "Hello, MyClerk!", "The quick brown fox jumps over the lazy dog"}
		aads := []string{"", "aad-data", "additional authenticated data for testing"}

		plaintext := []byte(plaintexts[i])
		aad := []byte(aads[i])

		aead, _ := NewAEAD(key[:])
		ct, _ := aead.Encrypt(nonce[:], plaintext, aad)

		vectors.AEAD = append(vectors.AEAD, aeadVector{
			Key:        hex.EncodeToString(key[:]),
			Nonce:      hex.EncodeToString(nonce[:]),
			Plaintext:  hex.EncodeToString(plaintext),
			AAD:        hex.EncodeToString(aad),
			Ciphertext: hex.EncodeToString(ct),
		})
	}

	// HKDF vectors (using RFC 5869 TC1 + custom)
	hkdfCases := []struct {
		ikm, salt, info string
		length          int
	}{
		{"0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b", "000102030405060708090a0b0c", "f0f1f2f3f4f5f6f7f8f9", 42},
		{"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "", "6d79636c65726b2d73657373696f6e2d76312d687962726964", 88},
	}
	for _, tc := range hkdfCases {
		ikm, _ := hex.DecodeString(tc.ikm)
		var salt []byte
		if tc.salt != "" {
			salt, _ = hex.DecodeString(tc.salt)
		}
		info, _ := hex.DecodeString(tc.info)
		okm, _ := HKDFDeriveKey(ikm, salt, info, tc.length)

		vectors.HKDF = append(vectors.HKDF, hkdfVector{
			IKM:    tc.ikm,
			Salt:   tc.salt,
			Info:   tc.info,
			Length: tc.length,
			OKM:    hex.EncodeToString(okm),
		})
	}

	// X25519 vectors
	for i := 0; i < 3; i++ {
		alice, _ := GenerateX25519KeyPair()
		bob, _ := GenerateX25519KeyPair()
		shared, _ := alice.ECDH(bob.Public)

		vectors.X25519 = append(vectors.X25519, x25519Vector{
			AlicePrivate: hex.EncodeToString(alice.Private[:]),
			AlicePublic:  hex.EncodeToString(alice.Public[:]),
			BobPrivate:   hex.EncodeToString(bob.Private[:]),
			BobPublic:    hex.EncodeToString(bob.Public[:]),
			SharedSecret: hex.EncodeToString(shared[:]),
		})
	}

	// Key rotation vectors
	baseKey, _ := GenerateKey()
	for counter := uint32(0); counter < 3; counter++ {
		derived, _ := RotateKey(baseKey, counter)
		vectors.Rotation = append(vectors.Rotation, rotationVector{
			BaseKey:    hex.EncodeToString(baseKey[:]),
			Counter:    counter,
			DerivedKey: hex.EncodeToString(derived[:]),
		})
	}

	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		t.Fatalf("marshal vectors: %v", err)
	}

	err = os.WriteFile("../testdata/crypto_vectors.json", data, 0o600)
	if err != nil {
		t.Fatalf("write vectors: %v", err)
	}

	t.Logf("wrote %d AEAD + %d HKDF + %d X25519 + %d rotation vectors",
		len(vectors.AEAD), len(vectors.HKDF), len(vectors.X25519), len(vectors.Rotation))
}
