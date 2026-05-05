// RFC 8439: ChaCha20 stream cipher + Poly1305 MAC, combined as AEAD.
//
// Used for every Tier-3 frame after handshake.
// Nonce is 12 bytes (RFC 8439 §2.3, IETF variant), key is 32 bytes.
// Tag is 16 bytes appended to ciphertext.

import { loadU32LE, storeU32LE, rotl32, eqCT } from './utils.js';

const CONSTANTS = new Uint32Array([0x61707865, 0x3320646e, 0x79622d32, 0x6b206574]); // "expand 32-byte k"

// One ChaCha20 quarter-round on state words a, b, c, d (RFC 8439 §2.1).
function qr(s, a, b, c, d) {
  s[a] = (s[a] + s[b]) >>> 0; s[d] = rotl32(s[d] ^ s[a], 16);
  s[c] = (s[c] + s[d]) >>> 0; s[b] = rotl32(s[b] ^ s[c], 12);
  s[a] = (s[a] + s[b]) >>> 0; s[d] = rotl32(s[d] ^ s[a],  8);
  s[c] = (s[c] + s[d]) >>> 0; s[b] = rotl32(s[b] ^ s[c],  7);
}

// One ChaCha20 block: state -> 64 bytes of keystream (RFC 8439 §2.3).
function chachaBlock(key, counter, nonce, out, outOff) {
  const s = new Uint32Array(16);
  s.set(CONSTANTS, 0);
  for (let i = 0; i < 8; i++) s[4 + i] = loadU32LE(key, i * 4);
  s[12] = counter >>> 0;
  s[13] = loadU32LE(nonce, 0);
  s[14] = loadU32LE(nonce, 4);
  s[15] = loadU32LE(nonce, 8);
  const w = new Uint32Array(s);
  for (let r = 0; r < 10; r++) {
    qr(w, 0, 4,  8, 12); qr(w, 1, 5,  9, 13); qr(w, 2, 6, 10, 14); qr(w, 3, 7, 11, 15);
    qr(w, 0, 5, 10, 15); qr(w, 1, 6, 11, 12); qr(w, 2, 7,  8, 13); qr(w, 3, 4,  9, 14);
  }
  for (let i = 0; i < 16; i++) {
    const v = (w[i] + s[i]) >>> 0;
    storeU32LE(out, outOff + i * 4, v);
  }
}

// Encrypt/decrypt `data` with ChaCha20 keystream (xor in place into a new buffer).
// `initialCounter`: 1 for AEAD payload (0 is reserved for Poly1305 key derivation).
export function chacha20(key, nonce, data, initialCounter = 1) {
  const out = new Uint8Array(data.length);
  const block = new Uint8Array(64);
  for (let off = 0, ctr = initialCounter; off < data.length; off += 64, ctr++) {
    chachaBlock(key, ctr, nonce, block, 0);
    const take = Math.min(64, data.length - off);
    for (let i = 0; i < take; i++) out[off + i] = data[off + i] ^ block[i];
  }
  return out;
}

// Poly1305-MAC over `data`. Key is 32 bytes (r || s).
// Implementation following RFC 8439 §2.5.1: accumulator over 130-bit prime
// 2^130 - 5. We use BigInt for clarity; performance is ample for our
// frame sizes (≤64 KB per frame). Constant-time is not achievable in JS
// regardless of strategy — see CLIENT-IMPL-JS-HYBRID-PQ.md §1.2.
function poly1305(key, data) {
  const P = (1n << 130n) - 5n;

  // Clamp r per RFC 8439 §2.5: clear top 4 bits of bytes 3, 7, 11, 15
  // and bottom 2 bits of bytes 4, 8, 12.
  const r = new Uint8Array(16);
  r.set(key.subarray(0, 16));
  r[3] &= 15; r[7] &= 15; r[11] &= 15; r[15] &= 15;
  r[4] &= 252; r[8] &= 252; r[12] &= 252;
  let rBig = 0n;
  for (let i = 15; i >= 0; i--) rBig = (rBig << 8n) | BigInt(r[i]);

  let sBig = 0n;
  for (let i = 15; i >= 0; i--) sBig = (sBig << 8n) | BigInt(key[16 + i]);

  let acc = 0n;
  for (let off = 0; off < data.length; off += 16) {
    const remaining = data.length - off;
    const take = Math.min(16, remaining);
    let n = 0n;
    for (let i = take - 1; i >= 0; i--) n = (n << 8n) | BigInt(data[off + i]);
    n |= 1n << BigInt(8 * take);    // append the high "1" bit per RFC §2.5.1
    acc = ((acc + n) * rBig) % P;
  }
  acc = (acc + sBig) & ((1n << 128n) - 1n);

  const tag = new Uint8Array(16);
  for (let i = 0; i < 16; i++) {
    tag[i] = Number(acc & 0xffn);
    acc >>= 8n;
  }
  return tag;
}

// Derive the Poly1305 one-time key from the (key, nonce) pair: ChaCha20
// block 0 truncated to 32 bytes (RFC 8439 §2.6).
function poly1305Key(key, nonce) {
  const block = new Uint8Array(64);
  chachaBlock(key, 0, nonce, block, 0);
  return block.subarray(0, 32);
}

// Pad to a multiple of 16 (per RFC 8439 §2.8.1).
function pad16Length(len) { return len % 16 === 0 ? 0 : 16 - (len % 16); }

// AEAD construction (RFC 8439 §2.8.1):
//   tag = Poly1305(otkey, AAD || pad16 || ciphertext || pad16 || len(AAD)_u64le || len(ct)_u64le)
function aeadMacInput(aad, ct) {
  const padA = pad16Length(aad.length);
  const padC = pad16Length(ct.length);
  const total = aad.length + padA + ct.length + padC + 16;
  const buf = new Uint8Array(total);
  let off = 0;
  buf.set(aad, off); off += aad.length;
  off += padA;
  buf.set(ct, off); off += ct.length;
  off += padC;
  // 8-byte LE lengths
  let aLen = aad.length, cLen = ct.length;
  for (let i = 0; i < 8; i++) { buf[off + i] = aLen & 0xff; aLen = Math.floor(aLen / 256); }
  off += 8;
  for (let i = 0; i < 8; i++) { buf[off + i] = cLen & 0xff; cLen = Math.floor(cLen / 256); }
  return buf;
}

// AEAD encrypt: returns ciphertext || tag (RFC 8439 §2.8 wraps these).
export function aeadEncrypt(key, nonce, plaintext, aad = new Uint8Array(0)) {
  if (key.length !== 32) throw new Error('aead: key must be 32 bytes');
  if (nonce.length !== 12) throw new Error('aead: nonce must be 12 bytes');
  const otk = poly1305Key(key, nonce);
  const ct = chacha20(key, nonce, plaintext, 1);
  const tag = poly1305(otk, aeadMacInput(aad, ct));
  const out = new Uint8Array(ct.length + 16);
  out.set(ct, 0);
  out.set(tag, ct.length);
  return out;
}

// AEAD decrypt: takes ciphertext-with-tag (last 16 bytes are the tag),
// verifies tag (constant-time), returns plaintext or throws.
export function aeadDecrypt(key, nonce, ciphertextWithTag, aad = new Uint8Array(0)) {
  if (key.length !== 32) throw new Error('aead: key must be 32 bytes');
  if (nonce.length !== 12) throw new Error('aead: nonce must be 12 bytes');
  if (ciphertextWithTag.length < 16) throw new Error('aead: input shorter than tag');
  const ct = ciphertextWithTag.subarray(0, ciphertextWithTag.length - 16);
  const tag = ciphertextWithTag.subarray(ciphertextWithTag.length - 16);
  const otk = poly1305Key(key, nonce);
  const expected = poly1305(otk, aeadMacInput(aad, ct));
  if (!eqCT(tag, expected)) throw new Error('aead: authentication failed');
  return chacha20(key, nonce, ct, 1);
}
