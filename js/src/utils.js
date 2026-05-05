// Constant-time helpers + byte-array conversions used across crypto modules.
//
// All comparisons of secret material MUST go through `eqCT` — never `===` on
// arrays and never short-circuit on first differing byte. JavaScript engines
// don't guarantee constant-time semantics, but using bitwise reduction
// removes the most obvious branch-on-secret class of bugs.

export function eqCT(a, b) {
  if (a.length !== b.length) return false;
  let diff = 0;
  for (let i = 0; i < a.length; i++) diff |= a[i] ^ b[i];
  return diff === 0;
}

export function fromHex(hex) {
  if (hex.length % 2 !== 0) throw new Error('fromHex: odd-length input');
  const out = new Uint8Array(hex.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = parseInt(hex.substr(i * 2, 2), 16);
  }
  return out;
}

export function toHex(buf) {
  let out = '';
  for (let i = 0; i < buf.length; i++) out += buf[i].toString(16).padStart(2, '0');
  return out;
}

export function concat(...arrays) {
  let total = 0;
  for (const a of arrays) total += a.length;
  const out = new Uint8Array(total);
  let off = 0;
  for (const a of arrays) { out.set(a, off); off += a.length; }
  return out;
}

// Wipe a byte array in place. JS makes no guarantees the underlying memory
// is reused (JIT can keep copies), but explicit zeroing closes the obvious
// "leftover plaintext in the heap dump" hole.
export function wipe(...arrays) {
  for (const a of arrays) if (a && a.fill) a.fill(0);
}

// 32-bit big-endian load/store. Used by SHA-2.
export function loadU32BE(buf, off) {
  return ((buf[off] << 24) | (buf[off + 1] << 16) | (buf[off + 2] << 8) | buf[off + 3]) >>> 0;
}

export function storeU32BE(buf, off, v) {
  buf[off]     = (v >>> 24) & 0xff;
  buf[off + 1] = (v >>> 16) & 0xff;
  buf[off + 2] = (v >>> 8)  & 0xff;
  buf[off + 3] = v          & 0xff;
}

// 32-bit little-endian load/store. Used by ChaCha20.
export function loadU32LE(buf, off) {
  return ((buf[off]) | (buf[off + 1] << 8) | (buf[off + 2] << 16) | (buf[off + 3] << 24)) >>> 0;
}

export function storeU32LE(buf, off, v) {
  buf[off]     = v          & 0xff;
  buf[off + 1] = (v >>> 8)  & 0xff;
  buf[off + 2] = (v >>> 16) & 0xff;
  buf[off + 3] = (v >>> 24) & 0xff;
}

// 32-bit rotate-right and rotate-left.
export function rotr32(x, n) { return ((x >>> n) | (x << (32 - n))) >>> 0; }
export function rotl32(x, n) { return ((x << n)  | (x >>> (32 - n))) >>> 0; }
