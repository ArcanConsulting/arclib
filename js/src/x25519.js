// RFC 7748 X25519 — Diffie-Hellman on Curve25519.
//
// Field arithmetic uses BigInt over the prime p = 2^255 − 19. JS engines
// optimise BigInt well enough for our throughput (a handful of handshakes
// per minute). For a high-perf JIT-friendly impl, swap in a packed-limb
// representation (radix-2^25.5 or radix-2^51) — but verified-correct is
// the priority here.
//
// All operations on secret material run in the same shape regardless of
// secret bits (Montgomery ladder with conditional swap), but JS cannot
// guarantee constant-time at the engine level. See
// CLIENT-IMPL-JS-HYBRID-PQ.md §1.2.

const P = (1n << 255n) - 19n;
const A24 = 121665n;  // (486662 - 2) / 4

// Encode/decode 32 little-endian bytes <-> bigint (RFC 7748 §5).
function decodeUCoord(buf) {
  const tmp = new Uint8Array(buf);
  tmp[31] &= 0x7f;  // clear top bit per §5
  let x = 0n;
  for (let i = 31; i >= 0; i--) x = (x << 8n) | BigInt(tmp[i]);
  return x % P;
}

function encodeUCoord(x) {
  const out = new Uint8Array(32);
  let v = ((x % P) + P) % P;
  for (let i = 0; i < 32; i++) { out[i] = Number(v & 0xffn); v >>= 8n; }
  return out;
}

function decodeScalar(buf) {
  const tmp = new Uint8Array(buf);
  // RFC 7748 §5 clamping for X25519
  tmp[0] &= 248;
  tmp[31] &= 127;
  tmp[31] |= 64;
  let s = 0n;
  for (let i = 31; i >= 0; i--) s = (s << 8n) | BigInt(tmp[i]);
  return s;
}

// Modular inverse via Fermat's little theorem: a^(p-2) mod p.
function modInv(a) {
  return modPow(a, P - 2n, P);
}

function modPow(b, e, m) {
  let r = 1n; b %= m; if (b < 0n) b += m;
  while (e > 0n) {
    if (e & 1n) r = (r * b) % m;
    e >>= 1n;
    b = (b * b) % m;
  }
  return r;
}

// Conditional swap: if swap == 1, exchange (x_2, z_2) and (x_3, z_3).
// Implemented via mask to avoid a branch on the secret bit.
function cswap(swap, a, b) {
  const mask = swap === 1n ? -1n : 0n;
  const m = mask & ((1n << 256n) - 1n);
  const dummy = m & (a ^ b);
  return [a ^ dummy, b ^ dummy];
}

// X25519 scalar multiplication via Montgomery ladder (RFC 7748 §5).
function x25519Scalarmult(scalar, uCoord) {
  const k = decodeScalar(scalar);
  const u = decodeUCoord(uCoord);

  let x_1 = u;
  let x_2 = 1n, z_2 = 0n;
  let x_3 = u, z_3 = 1n;
  let swap = 0n;

  for (let t = 254; t >= 0; t--) {
    const k_t = (k >> BigInt(t)) & 1n;
    swap ^= k_t;
    [x_2, x_3] = cswap(swap, x_2, x_3);
    [z_2, z_3] = cswap(swap, z_2, z_3);
    swap = k_t;

    const A = (x_2 + z_2) % P;
    const AA = (A * A) % P;
    const B = (x_2 - z_2 + P) % P;
    const BB = (B * B) % P;
    const E = (AA - BB + P) % P;
    const C = (x_3 + z_3) % P;
    const D = (x_3 - z_3 + P) % P;
    const DA = (D * A) % P;
    const CB = (C * B) % P;
    x_3 = ((DA + CB) * (DA + CB)) % P;
    z_3 = (x_1 * ((DA - CB + P) * (DA - CB + P)) % P) % P;
    x_2 = (AA * BB) % P;
    z_2 = (E * (AA + (A24 * E) % P)) % P;
  }
  [x_2, x_3] = cswap(swap, x_2, x_3);
  [z_2, z_3] = cswap(swap, z_2, z_3);

  const result = (x_2 * modInv(z_2)) % P;
  return encodeUCoord(result);
}

// Public API: scalarmult(scalar, u) -> 32 bytes
export function x25519(scalar, uCoord) {
  if (scalar.length !== 32) throw new Error('x25519: scalar must be 32 bytes');
  if (uCoord.length !== 32) throw new Error('x25519: u-coord must be 32 bytes');
  return x25519Scalarmult(scalar, uCoord);
}

// Generate a public key from a private scalar (multiply by base point u=9).
const BASE = new Uint8Array([9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                              0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]);

export function x25519PublicKey(privateScalar) {
  return x25519(privateScalar, BASE);
}

// Generate a fresh keypair: random 32-byte private + derived public.
export function x25519GenerateKeyPair(randomBytes) {
  const priv = new Uint8Array(32);
  if (randomBytes) priv.set(randomBytes);
  else if (typeof crypto !== 'undefined' && crypto.getRandomValues) crypto.getRandomValues(priv);
  else throw new Error('x25519: no random source available');
  const pub = x25519PublicKey(priv);
  return { privateKey: priv, publicKey: pub };
}
