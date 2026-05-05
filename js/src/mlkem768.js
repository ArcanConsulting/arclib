// FIPS 203: Module-Lattice-Based Key-Encapsulation Mechanism (ML-KEM),
// parameter set ML-KEM-768.
//
// Algorithms below follow the FIPS 203 spec section/algorithm numbering
// where possible (e.g. `K-PKE.Encrypt` matches FIPS 203 Algorithm 14).
// The K-PKE core is wrapped by the FO-style transform in §7 to make
// ML-KEM IND-CCA secure.
//
// Parameters (FIPS 203 §8 Table 2 for ML-KEM-768):
//   n=256, q=3329, k=3, η1=2, η2=2, du=10, dv=4
// Key/CT sizes:
//   pk=1184, sk=2400, ct=1088, ss=32
//
// G(x)        = SHA3-512(x)
// H(x)        = SHA3-256(x)
// J(x)        = SHAKE256(x, 32)
// XOF(rho,j,i)= SHAKE128(rho || j || i)         absorb-then-squeeze pattern
// PRF_η(s, b) = SHAKE256(s || b, 64*η)

import { sha3_256, sha3_512, shake256, shake128xof } from './sha3.js';
import { concat } from './utils.js';

const N = 256;
const Q = 3329;
const K = 3;
const ETA1 = 2;
const ETA2 = 2;
const DU = 10;
const DV = 4;
const Q_INV_NEG = 3327;     // -q^-1 mod 2^16 (used in Montgomery reduction; not needed in this impl)

export const PUBLICKEY_BYTES = 1184;     // 384*K + 32
export const SECRETKEY_BYTES = 2400;     // 768*K + 96
export const CIPHERTEXT_BYTES = 1088;    // 32*(DU*K + DV)
export const SHARED_BYTES = 32;

// ---------------------------------------------------------------------
// Modular arithmetic helpers
// ---------------------------------------------------------------------

function modQ(x) {
  let r = x % Q;
  if (r < 0) r += Q;
  return r;
}

// Barrett-style reduce of arbitrary signed int into [0, Q).
function reduce(x) { return modQ(x); }

// ---------------------------------------------------------------------
// NTT (Number Theoretic Transform)
// FIPS 203 Algorithm 9 / 10. Operates in-place on a length-256 array.
// ---------------------------------------------------------------------

// ζ = 17 is the chosen primitive 256th root of unity mod 3329.
// Precompute zetas[i] = ζ^BitRev7(i) mod 3329 for i = 1..127.
const ZETAS = (() => {
  const zetas = new Int16Array(128);
  let z = 1;
  // BitRev7(i): reverse the bottom 7 bits of i.
  function bitrev7(i) {
    let r = 0;
    for (let b = 0; b < 7; b++) if (i & (1 << b)) r |= (1 << (6 - b));
    return r;
  }
  // Compute ζ^k mod q for k = 0..127, store at index BitRev7(k).
  // (Equivalent to the table in FIPS 203 Appendix A.)
  let pow = 1;
  const tmp = new Int16Array(128);
  for (let k = 0; k < 128; k++) {
    tmp[k] = pow;
    pow = (pow * 17) % Q;
  }
  for (let k = 0; k < 128; k++) zetas[k] = tmp[bitrev7(k)];
  return zetas;
})();

// Forward NTT: in-place transform of f (length 256) to its NTT
// representation. FIPS 203 Algorithm 9.
function ntt(f) {
  let i = 1;
  for (let len = 128; len >= 2; len >>= 1) {
    for (let start = 0; start < N; start += 2 * len) {
      const zeta = ZETAS[i++];
      for (let j = start; j < start + len; j++) {
        const t = (zeta * f[j + len]) % Q;
        f[j + len] = reduce(f[j] - t);
        f[j]       = reduce(f[j] + t);
      }
    }
  }
}

// Inverse NTT: FIPS 203 Algorithm 10.
function invNtt(f) {
  let i = 127;
  for (let len = 2; len <= 128; len <<= 1) {
    for (let start = 0; start < N; start += 2 * len) {
      const zeta = ZETAS[i--];
      for (let j = start; j < start + len; j++) {
        const t = f[j];
        f[j]       = reduce(t + f[j + len]);
        f[j + len] = reduce(zeta * (f[j + len] - t));
      }
    }
  }
  // Multiply by 3303 = 128^-1 mod 3329.
  for (let j = 0; j < N; j++) f[j] = reduce(f[j] * 3303);
}

// Pointwise multiplication in NTT representation.
// FIPS 203 Algorithm 11 / 12 (BaseCaseMultiply).
function nttMul(f, g) {
  const out = new Int16Array(N);
  for (let i = 0; i < 64; i++) {
    const zeta = ZETAS[64 + i];
    const a0 = f[4 * i], a1 = f[4 * i + 1];
    const b0 = g[4 * i], b1 = g[4 * i + 1];
    out[4 * i]     = reduce(a0 * b0 + a1 * b1 * zeta);
    out[4 * i + 1] = reduce(a0 * b1 + a1 * b0);
    const a2 = f[4 * i + 2], a3 = f[4 * i + 3];
    const b2 = g[4 * i + 2], b3 = g[4 * i + 3];
    out[4 * i + 2] = reduce(a2 * b2 - a3 * b3 * zeta);
    out[4 * i + 3] = reduce(a2 * b3 + a3 * b2);
  }
  return out;
}

function polyAdd(a, b) {
  const c = new Int16Array(N);
  for (let i = 0; i < N; i++) c[i] = reduce(a[i] + b[i]);
  return c;
}

function polySub(a, b) {
  const c = new Int16Array(N);
  for (let i = 0; i < N; i++) c[i] = reduce(a[i] - b[i]);
  return c;
}

// Vector-vector inner product in NTT domain (returns a single polynomial).
function vecMul(uVec, vVec) {
  let acc = nttMul(uVec[0], vVec[0]);
  for (let i = 1; i < uVec.length; i++) acc = polyAdd(acc, nttMul(uVec[i], vVec[i]));
  return acc;
}

// Matrix × vector: A is k×k of polynomials (NTT), s is k of polynomials (NTT).
// Returns k of polynomials (NTT).
function matVecMul(A, s) {
  const out = new Array(K);
  for (let i = 0; i < K; i++) {
    let acc = nttMul(A[i][0], s[0]);
    for (let j = 1; j < K; j++) acc = polyAdd(acc, nttMul(A[i][j], s[j]));
    out[i] = acc;
  }
  return out;
}

function vecAdd(a, b) {
  const out = new Array(a.length);
  for (let i = 0; i < a.length; i++) out[i] = polyAdd(a[i], b[i]);
  return out;
}

// ---------------------------------------------------------------------
// Sampling
// ---------------------------------------------------------------------

// FIPS 203 Algorithm 6 — SampleNTT (rejection sample from XOF stream).
function sampleNTT(rho, j, i) {
  const seed = new Uint8Array(34);
  seed.set(rho);
  seed[32] = j; seed[33] = i;
  const xof = shake128xof(seed);
  const out = new Int16Array(N);
  let count = 0;
  while (count < N) {
    const buf = xof.squeeze(3);
    const d1 = buf[0] | ((buf[1] & 0x0f) << 8);
    const d2 = (buf[1] >> 4) | (buf[2] << 4);
    if (d1 < Q && count < N) out[count++] = d1;
    if (d2 < Q && count < N) out[count++] = d2;
  }
  return out;
}

// FIPS 203 Algorithm 7 — SamplePolyCBD_η (centered binomial distribution).
function samplePolyCBD(eta, prfOutput) {
  // prfOutput length = 64*eta bytes.
  const out = new Int16Array(N);
  for (let i = 0; i < N; i++) {
    let a = 0, b = 0;
    for (let j = 0; j < eta; j++) {
      const bitPos = 2 * i * eta + j;
      a += (prfOutput[bitPos >> 3] >> (bitPos & 7)) & 1;
    }
    for (let j = 0; j < eta; j++) {
      const bitPos = 2 * i * eta + eta + j;
      b += (prfOutput[bitPos >> 3] >> (bitPos & 7)) & 1;
    }
    out[i] = reduce(a - b);
  }
  return out;
}

function prf(eta, seed, b) {
  const input = new Uint8Array(33);
  input.set(seed);
  input[32] = b;
  return shake256(input, 64 * eta);
}

// ---------------------------------------------------------------------
// Compression / Decompression (FIPS 203 §4.2.1 / Definition 4.7)
// ---------------------------------------------------------------------

function compress(d, x) {
  // Compress_d(x) = round((2^d / q) * x) mod 2^d
  const mod = 1 << d;
  return Math.floor((x * mod + Q / 2) / Q) % mod;
}

function decompress(d, y) {
  // Decompress_d(y) = round((q / 2^d) * y)
  const div = 1 << d;
  return Math.floor((y * Q + (div >> 1)) / div);
}

function polyCompress(d, p) {
  const out = new Int16Array(N);
  for (let i = 0; i < N; i++) out[i] = compress(d, p[i] >= 0 ? p[i] : p[i] + Q);
  return out;
}

function polyDecompress(d, p) {
  const out = new Int16Array(N);
  for (let i = 0; i < N; i++) out[i] = decompress(d, p[i]);
  return out;
}

// ---------------------------------------------------------------------
// Byte encoding / decoding
// ---------------------------------------------------------------------

// FIPS 203 Algorithm 4 — ByteEncode_d.
function byteEncode(d, p) {
  const out = new Uint8Array(32 * d);
  let bitPos = 0;
  for (let i = 0; i < N; i++) {
    let val = p[i] >= 0 ? p[i] : p[i] + Q;
    for (let bit = 0; bit < d; bit++) {
      if (val & 1) out[bitPos >> 3] |= 1 << (bitPos & 7);
      bitPos++;
      val >>= 1;
    }
  }
  return out;
}

// FIPS 203 Algorithm 5 — ByteDecode_d.
function byteDecode(d, bytes) {
  const out = new Int16Array(N);
  let bitPos = 0;
  const mask = (1 << d) - 1;
  for (let i = 0; i < N; i++) {
    let val = 0;
    for (let bit = 0; bit < d; bit++) {
      val |= ((bytes[bitPos >> 3] >> (bitPos & 7)) & 1) << bit;
      bitPos++;
    }
    out[i] = val & mask;
  }
  return out;
}

// ---------------------------------------------------------------------
// K-PKE (FIPS 203 §6 — IND-CPA core)
// ---------------------------------------------------------------------

// Algorithm 13: K-PKE.KeyGen(d).
function kpkeKeyGen(d) {
  const G = sha3_512(concat(d, new Uint8Array([K])));
  const rho = G.subarray(0, 32);
  const sigma = G.subarray(32, 64);

  // Generate matrix A (k×k of NTT-domain polynomials).
  const A = new Array(K);
  for (let i = 0; i < K; i++) {
    A[i] = new Array(K);
    for (let j = 0; j < K; j++) A[i][j] = sampleNTT(rho, j, i);
  }

  // Sample s and e from CBD_η1.
  const s = new Array(K), e = new Array(K);
  for (let i = 0; i < K; i++) s[i] = samplePolyCBD(ETA1, prf(ETA1, sigma, i));
  for (let i = 0; i < K; i++) e[i] = samplePolyCBD(ETA1, prf(ETA1, sigma, K + i));

  // NTT-transform s and e.
  for (let i = 0; i < K; i++) ntt(s[i]);
  for (let i = 0; i < K; i++) ntt(e[i]);

  // t = A·s + e (in NTT domain).
  const t = vecAdd(matVecMul(A, s), e);

  // Encode public key: t || rho.
  const pkBytes = new Uint8Array(PUBLICKEY_BYTES);
  for (let i = 0; i < K; i++) pkBytes.set(byteEncode(12, t[i]), 384 * i);
  pkBytes.set(rho, 384 * K);

  // Encode secret key: s.
  const skBytes = new Uint8Array(384 * K);
  for (let i = 0; i < K; i++) skBytes.set(byteEncode(12, s[i]), 384 * i);

  return { pkBytes, skBytes };
}

// Algorithm 14: K-PKE.Encrypt(pk, m, r).
function kpkeEncrypt(pkBytes, m, r) {
  // Decode t and rho from pk.
  const t = new Array(K);
  for (let i = 0; i < K; i++) t[i] = byteDecode(12, pkBytes.subarray(384 * i, 384 * (i + 1)));
  const rho = pkBytes.subarray(384 * K, 384 * K + 32);

  // Regenerate matrix A^T (use sampleNTT(rho, i, j) with swapped indices).
  const AT = new Array(K);
  for (let i = 0; i < K; i++) {
    AT[i] = new Array(K);
    for (let j = 0; j < K; j++) AT[i][j] = sampleNTT(rho, i, j);
  }

  // Sample r vec, e1 vec, e2 poly from CBD.
  const rVec = new Array(K), e1 = new Array(K);
  for (let i = 0; i < K; i++) rVec[i] = samplePolyCBD(ETA1, prf(ETA1, r, i));
  for (let i = 0; i < K; i++) e1[i] = samplePolyCBD(ETA2, prf(ETA2, r, K + i));
  const e2 = samplePolyCBD(ETA2, prf(ETA2, r, 2 * K));

  // NTT-transform rVec.
  for (let i = 0; i < K; i++) ntt(rVec[i]);

  // u = invNTT(A^T · rVec) + e1
  const u = matVecMul(AT, rVec);
  for (let i = 0; i < K; i++) invNtt(u[i]);
  const uFinal = vecAdd(u, e1);

  // v = invNTT(t · rVec) + e2 + decompress(1, decode(1, m))
  let v = vecMul(t, rVec);
  invNtt(v);
  v = polyAdd(v, e2);
  const mPoly = polyDecompress(1, byteDecode(1, m));
  v = polyAdd(v, mPoly);

  // Compress and encode ciphertext.
  const c1Bytes = new Uint8Array(32 * DU * K);
  for (let i = 0; i < K; i++) {
    c1Bytes.set(byteEncode(DU, polyCompress(DU, uFinal[i])), 32 * DU * i);
  }
  const c2Bytes = byteEncode(DV, polyCompress(DV, v));
  return concat(c1Bytes, c2Bytes);
}

// Algorithm 15: K-PKE.Decrypt(sk, c).
function kpkeDecrypt(skBytes, ct) {
  // Decode u (compressed) and v (compressed).
  const c1Bytes = ct.subarray(0, 32 * DU * K);
  const c2Bytes = ct.subarray(32 * DU * K);
  const u = new Array(K);
  for (let i = 0; i < K; i++) {
    const enc = byteDecode(DU, c1Bytes.subarray(32 * DU * i, 32 * DU * (i + 1)));
    u[i] = polyDecompress(DU, enc);
  }
  const v = polyDecompress(DV, byteDecode(DV, c2Bytes));

  // s in NTT domain.
  const s = new Array(K);
  for (let i = 0; i < K; i++) s[i] = byteDecode(12, skBytes.subarray(384 * i, 384 * (i + 1)));

  // m_poly = v - invNTT(s · NTT(u))
  for (let i = 0; i < K; i++) ntt(u[i]);
  let su = vecMul(s, u);
  invNtt(su);
  const mPoly = polySub(v, su);
  return byteEncode(1, polyCompress(1, mPoly));
}

// ---------------------------------------------------------------------
// ML-KEM (FIPS 203 §7 — IND-CCA wrap)
// ---------------------------------------------------------------------

// Algorithm 19: ML-KEM.KeyGen(d, z).
// d (32 bytes) seeds K-PKE.KeyGen; z (32 bytes) is the implicit-rejection
// key embedded in the secret key.
export function keygen(seed) {
  if (!seed) {
    seed = new Uint8Array(64);
    if (typeof crypto !== 'undefined' && crypto.getRandomValues) crypto.getRandomValues(seed);
    else throw new Error('mlkem: no random source available');
  }
  if (seed.length !== 64) throw new Error('mlkem: keygen seed must be 64 bytes');
  const d = seed.subarray(0, 32);
  const z = seed.subarray(32, 64);

  const { pkBytes, skBytes: skKpke } = kpkeKeyGen(d);
  const hPk = sha3_256(pkBytes);

  // sk = sk_kpke || pk || H(pk) || z
  const sk = new Uint8Array(SECRETKEY_BYTES);
  sk.set(skKpke, 0);
  sk.set(pkBytes, 384 * K);
  sk.set(hPk, 384 * K + PUBLICKEY_BYTES);
  sk.set(z, 384 * K + PUBLICKEY_BYTES + 32);

  return { publicKey: pkBytes, secretKey: sk };
}

// Algorithm 20: ML-KEM.Encaps(pk, m).
// m (32 bytes) is the message randomness; in tests it comes from the KAT
// vector, in production from `crypto.getRandomValues`.
export function encaps(pk, m) {
  if (pk.length !== PUBLICKEY_BYTES) throw new Error('mlkem: bad pk length');
  if (!m) {
    m = new Uint8Array(32);
    if (typeof crypto !== 'undefined' && crypto.getRandomValues) crypto.getRandomValues(m);
    else throw new Error('mlkem: no random source available');
  }
  if (m.length !== 32) throw new Error('mlkem: m must be 32 bytes');

  const hPk = sha3_256(pk);
  const G = sha3_512(concat(m, hPk));
  const sharedSecret = G.subarray(0, 32);
  const r = G.subarray(32, 64);
  const ct = kpkeEncrypt(pk, m, r);
  return { sharedSecret: new Uint8Array(sharedSecret), ciphertext: ct };
}

// Algorithm 21: ML-KEM.Decaps(sk, c).
export function decaps(sk, ct) {
  if (sk.length !== SECRETKEY_BYTES) throw new Error('mlkem: bad sk length');
  if (ct.length !== CIPHERTEXT_BYTES) throw new Error('mlkem: bad ct length');

  const skKpke = sk.subarray(0, 384 * K);
  const pk     = sk.subarray(384 * K, 384 * K + PUBLICKEY_BYTES);
  const hPk    = sk.subarray(384 * K + PUBLICKEY_BYTES, 384 * K + PUBLICKEY_BYTES + 32);
  const z      = sk.subarray(384 * K + PUBLICKEY_BYTES + 32);

  const m_prime = kpkeDecrypt(skKpke, ct);
  const G = sha3_512(concat(m_prime, hPk));
  const K_prime = G.subarray(0, 32);
  const r_prime = G.subarray(32, 64);

  // Re-encrypt with derived randomness; constant-time compare.
  const c_prime = kpkeEncrypt(pk, m_prime, r_prime);

  // Implicit rejection: K_bar = J(z || c) if compare fails.
  let diff = 0;
  for (let i = 0; i < CIPHERTEXT_BYTES; i++) diff |= ct[i] ^ c_prime[i];
  if (diff === 0) {
    return new Uint8Array(K_prime);
  }
  return shake256(concat(z, ct), 32);
}
