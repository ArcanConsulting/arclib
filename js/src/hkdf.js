// RFC 5869 HKDF-SHA256.
//
// Two-step construction: Extract pseudorandom key from input keying
// material + salt, then Expand into the requested output length using
// counter-mode HMAC.

import { hmacSha256 } from './hmac.js';
import { concat } from './utils.js';

const HASH_LEN = 32;

// HKDF-Extract(salt, IKM) -> PRK
export function hkdfExtract(salt, ikm) {
  if (!salt || salt.length === 0) salt = new Uint8Array(HASH_LEN);
  return hmacSha256(salt, ikm);
}

// HKDF-Expand(PRK, info, L) -> OKM
export function hkdfExpand(prk, info, length) {
  const N = Math.ceil(length / HASH_LEN);
  if (N > 255) throw new Error('hkdf: requested output too large (>255 * HashLen)');
  if (!info) info = new Uint8Array(0);
  const out = new Uint8Array(length);
  let prev = new Uint8Array(0);
  let off = 0;
  for (let i = 1; i <= N; i++) {
    const t = hmacSha256(prk, concat(prev, info, new Uint8Array([i])));
    const take = Math.min(HASH_LEN, length - off);
    out.set(t.subarray(0, take), off);
    off += take;
    prev = t;
  }
  return out;
}

// One-shot HKDF-SHA256.
export function hkdf(salt, ikm, info, length) {
  return hkdfExpand(hkdfExtract(salt, ikm), info, length);
}
