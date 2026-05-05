// RFC 2104 HMAC, parameterised by hash. Currently only HMAC-SHA256 is
// needed (HKDF dependency); generic enough that HMAC-SHA512 could plug in.
//
// HMAC(K, m) = H((K' xor opad) || H((K' xor ipad) || m))
// where K' = K if |K|<=B else H(K), then padded to B with zeros.

import { SHA256 } from './sha256.js';

const SHA256_BLOCK = 64;
const SHA256_DIGEST = 32;

export function hmacSha256(key, msg) {
  if (typeof msg === 'string') msg = new TextEncoder().encode(msg);
  if (typeof key === 'string') key = new TextEncoder().encode(key);

  let k = key;
  if (k.length > SHA256_BLOCK) k = new SHA256().update(k).digest();
  if (k.length < SHA256_BLOCK) {
    const padded = new Uint8Array(SHA256_BLOCK);
    padded.set(k);
    k = padded;
  }

  const ipad = new Uint8Array(SHA256_BLOCK);
  const opad = new Uint8Array(SHA256_BLOCK);
  for (let i = 0; i < SHA256_BLOCK; i++) {
    ipad[i] = k[i] ^ 0x36;
    opad[i] = k[i] ^ 0x5c;
  }

  const inner = new SHA256().update(ipad).update(msg).digest();
  return new SHA256().update(opad).update(inner).digest();
}
