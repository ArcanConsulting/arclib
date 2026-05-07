import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { sha256, hmacSha256, hkdf, aeadEncrypt, aeadDecrypt, x25519, crc16, crc32, fromHex, toHex } from '../src/index.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const vectors = JSON.parse(readFileSync(join(__dirname, '../../testdata/crypto_vectors.json'), 'utf-8'));
const hashVectors = JSON.parse(readFileSync(join(__dirname, '../../testdata/hash_vectors.json'), 'utf-8'));

describe('Cross-Validation: SHA-256', () => {
  for (const v of hashVectors.sha256) {
    it(`sha256("${v.input}")`, () => {
      const input = v.hex_input ? fromHex(v.hex_input) : new TextEncoder().encode(v.input);
      const result = toHex(sha256(input));
      assert.equal(result, v.output);
    });
  }
});

describe('Cross-Validation: CRC-16-IBM', () => {
  for (const v of hashVectors.crc16_ibm.vectors) {
    it(`crc16(${v.hex_input || '""'})`, () => {
      const input = v.hex_input ? fromHex(v.hex_input) : new Uint8Array(0);
      const result = crc16(input);
      const expected = parseInt(v.output, 16);
      assert.equal(result, expected);
    });
  }
});

describe('Cross-Validation: CRC-32-IEEE', () => {
  for (const v of hashVectors.crc32_ieee.vectors) {
    it(`crc32(${v.hex_input || '""'})`, () => {
      const input = v.hex_input ? fromHex(v.hex_input) : new Uint8Array(0);
      const result = crc32(input);
      const expected = parseInt(v.output, 16);
      assert.equal(result, expected);
    });
  }
});

describe('Cross-Validation: HMAC-SHA256', () => {
  for (const v of hashVectors.hmac_sha256.vectors) {
    it(v.name, () => {
      const key = fromHex(v.key);
      const data = fromHex(v.data);
      const result = toHex(hmacSha256(key, data));
      assert.equal(result, v.output);
    });
  }
});

describe('Cross-Validation: ChaCha20-Poly1305', () => {
  for (const [i, v] of vectors.chacha20_poly1305.entries()) {
    it(`vector ${i}`, () => {
      const key = fromHex(v.key);
      const nonce = fromHex(v.nonce);
      const plaintext = fromHex(v.plaintext);
      const aad = fromHex(v.aad);
      const expectedCiphertext = fromHex(v.ciphertext);

      // Encrypt and compare
      const encrypted = aeadEncrypt(key, nonce, plaintext, aad);
      assert.equal(toHex(encrypted), v.ciphertext);

      // Decrypt and verify round-trip
      const decrypted = aeadDecrypt(key, nonce, expectedCiphertext, aad);
      assert.equal(toHex(decrypted), v.plaintext);
    });
  }
});

describe('Cross-Validation: HKDF-SHA256', () => {
  for (const [i, v] of vectors.hkdf_sha256.entries()) {
    it(`vector ${i} (length=${v.length})`, () => {
      const ikm = fromHex(v.ikm);
      const salt = v.salt ? fromHex(v.salt) : null;
      const info = v.info ? fromHex(v.info) : null;
      const result = hkdf(salt, ikm, info, v.length);
      assert.equal(toHex(result), v.okm);
    });
  }
});

describe('Cross-Validation: X25519', () => {
  for (const [i, v] of vectors.x25519.entries()) {
    it(`vector ${i}: Alice ECDH(Bob)`, () => {
      const alicePrivate = fromHex(v.alice_private);
      const bobPublic = fromHex(v.bob_public);
      const shared = x25519(alicePrivate, bobPublic);
      assert.equal(toHex(shared), v.shared_secret);
    });

    it(`vector ${i}: Bob ECDH(Alice)`, () => {
      const bobPrivate = fromHex(v.bob_private);
      const alicePublic = fromHex(v.alice_public);
      const shared = x25519(bobPrivate, alicePublic);
      assert.equal(toHex(shared), v.shared_secret);
    });
  }
});

describe('Cross-Validation: Key Rotation', () => {
  for (const [i, v] of vectors.key_rotation.entries()) {
    it(`vector ${i}: counter=${v.counter}`, () => {
      const baseKey = fromHex(v.base_key);
      // Key rotation: HKDF(IKM=baseKey, salt="rotate", info=counter_be32, length=32)
      const info = new Uint8Array(4);
      const dv = new DataView(info.buffer);
      dv.setUint32(0, v.counter, false); // big-endian
      const salt = new TextEncoder().encode('rotate');
      const derived = hkdf(salt, baseKey, info, 32);
      assert.equal(toHex(derived), v.derived_key);
    });
  }
});
