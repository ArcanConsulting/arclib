// FIPS 180-4 SHA-256.
//
// Streaming hasher (init/update/finalize) plus a one-shot helper.
// All multi-block accumulation goes through `update()`; padding is added
// by `finalize()`. State is held as 8 × uint32 in a Uint32Array.

import { loadU32BE, storeU32BE, rotr32 } from './utils.js';

const K = new Uint32Array([
  0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
  0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
  0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
  0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
  0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
  0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
  0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
  0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
]);

const IV = new Uint32Array([
  0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
  0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
]);

export class SHA256 {
  constructor() {
    this.h = new Uint32Array(IV);
    this.buf = new Uint8Array(64);
    this.bufLen = 0;
    this.totalBytes = 0;
  }

  // 64-byte block compression function (FIPS 180-4 §6.2.2).
  _block(block, off) {
    const w = new Uint32Array(64);
    for (let t = 0; t < 16; t++) w[t] = loadU32BE(block, off + t * 4);
    for (let t = 16; t < 64; t++) {
      const s0 = rotr32(w[t - 15], 7) ^ rotr32(w[t - 15], 18) ^ (w[t - 15] >>> 3);
      const s1 = rotr32(w[t - 2], 17) ^ rotr32(w[t - 2], 19) ^ (w[t - 2] >>> 10);
      w[t] = (w[t - 16] + s0 + w[t - 7] + s1) >>> 0;
    }
    let a = this.h[0], b = this.h[1], c = this.h[2], d = this.h[3];
    let e = this.h[4], f = this.h[5], g = this.h[6], h = this.h[7];
    for (let t = 0; t < 64; t++) {
      const S1 = rotr32(e, 6) ^ rotr32(e, 11) ^ rotr32(e, 25);
      const ch = (e & f) ^ ((~e >>> 0) & g);
      const t1 = (h + S1 + ch + K[t] + w[t]) >>> 0;
      const S0 = rotr32(a, 2) ^ rotr32(a, 13) ^ rotr32(a, 22);
      const mj = (a & b) ^ (a & c) ^ (b & c);
      const t2 = (S0 + mj) >>> 0;
      h = g; g = f; f = e; e = (d + t1) >>> 0;
      d = c; c = b; b = a; a = (t1 + t2) >>> 0;
    }
    this.h[0] = (this.h[0] + a) >>> 0;
    this.h[1] = (this.h[1] + b) >>> 0;
    this.h[2] = (this.h[2] + c) >>> 0;
    this.h[3] = (this.h[3] + d) >>> 0;
    this.h[4] = (this.h[4] + e) >>> 0;
    this.h[5] = (this.h[5] + f) >>> 0;
    this.h[6] = (this.h[6] + g) >>> 0;
    this.h[7] = (this.h[7] + h) >>> 0;
  }

  update(data) {
    if (typeof data === 'string') data = new TextEncoder().encode(data);
    let off = 0;
    this.totalBytes += data.length;
    if (this.bufLen > 0) {
      const need = 64 - this.bufLen;
      const take = Math.min(need, data.length);
      this.buf.set(data.subarray(0, take), this.bufLen);
      this.bufLen += take;
      off += take;
      if (this.bufLen === 64) {
        this._block(this.buf, 0);
        this.bufLen = 0;
      }
    }
    while (data.length - off >= 64) {
      this._block(data, off);
      off += 64;
    }
    if (off < data.length) {
      this.buf.set(data.subarray(off), 0);
      this.bufLen = data.length - off;
    }
    return this;
  }

  digest() {
    // Padding (FIPS 180-4 §5.1.1): append 0x80, zeros, 64-bit BE length.
    const bitsHi = Math.floor(this.totalBytes / 0x20000000) >>> 0;        // bits >> 32
    const bitsLo = (this.totalBytes * 8) >>> 0;
    this.buf[this.bufLen++] = 0x80;
    if (this.bufLen > 56) {
      while (this.bufLen < 64) this.buf[this.bufLen++] = 0;
      this._block(this.buf, 0);
      this.bufLen = 0;
    }
    while (this.bufLen < 56) this.buf[this.bufLen++] = 0;
    storeU32BE(this.buf, 56, bitsHi);
    storeU32BE(this.buf, 60, bitsLo);
    this._block(this.buf, 0);
    const out = new Uint8Array(32);
    for (let i = 0; i < 8; i++) storeU32BE(out, i * 4, this.h[i]);
    return out;
  }
}

export function sha256(data) {
  return new SHA256().update(data).digest();
}
