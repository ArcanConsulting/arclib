// arclib protocol layer — wire-compatible with Go implementation.
// Header encoding is big-endian binary, tier-dependent field layout.
// Extensions are TLV (type u16, length u16, value bytes).

// --- Tiers ---

export const Tier = {
  PLAINTEXT: 0,
  CHECKSUM: 1,
  AUTHENTICATED: 2,
  ENCRYPTED: 3,
  PFS: 4,
  MAX_SECURITY: 5,
};

const tierInfo = [
  { name: 'plaintext',     headerOverhead: 0,   trailerOverhead: 0,  totalOverhead: 1,   encrypted: false, authenticated: false },
  { name: 'checksum',      headerOverhead: 0,   trailerOverhead: 2,  totalOverhead: 3,   encrypted: false, authenticated: false },
  { name: 'authenticated', headerOverhead: 4,   trailerOverhead: 8,  totalOverhead: 15,  encrypted: false, authenticated: true  },
  { name: 'encrypted',     headerOverhead: 12,  trailerOverhead: 16, totalOverhead: 31,  encrypted: true,  authenticated: true  },
  { name: 'pfs',           headerOverhead: 48,  trailerOverhead: 16, totalOverhead: 67,  encrypted: true,  authenticated: true  },
  { name: 'max_security',  headerOverhead: 112, trailerOverhead: 32, totalOverhead: 147, encrypted: true,  authenticated: true  },
];

export function getTierInfo(tier) {
  if (tier < 0 || tier > 5) return { name: 'unknown' };
  return tierInfo[tier];
}

export function tierIsValid(tier) { return tier >= 0 && tier <= 5; }
export function tierRequiresEncryption(tier) { return tier >= Tier.ENCRYPTED; }
export function tierRequiresAuthentication(tier) { return tier >= Tier.AUTHENTICATED; }

// --- Flag bits ---

const FLAG_VERSION_MASK   = 0x03;
const FLAG_TIER_MASK      = 0x1C;
const FLAG_TIER_SHIFT     = 2;
const FLAG_COMPRESSED     = 0x20;
const FLAG_FRAGMENTED     = 0x40;
const FLAG_HAS_EXTENSIONS = 0x80;

// --- Header ---

export function headerSize(tier, fragmented) {
  let size = 3; // flags(1) + opcode(2)
  if (fragmented) size += 8;
  switch (tier) {
    case Tier.PLAINTEXT: case Tier.CHECKSUM: break;
    case Tier.AUTHENTICATED: size += 4; break;
    case Tier.ENCRYPTED: size += 12; break;
    case Tier.PFS: size += 60; break;
    case Tier.MAX_SECURITY: size += 68; break;
  }
  return size;
}

export function trailerSize(tier) {
  switch (tier) {
    case Tier.PLAINTEXT: return 0;
    case Tier.CHECKSUM: return 2;
    case Tier.AUTHENTICATED: return 8;
    case Tier.ENCRYPTED: case Tier.PFS: return 16;
    case Tier.MAX_SECURITY: return 32;
    default: return 0;
  }
}

export function marshalHeader(hdr) {
  const size = headerSize(hdr.tier, hdr.fragmented);
  const buf = new Uint8Array(size);
  const dv = new DataView(buf.buffer);
  let off = 0;

  // Flags byte
  let flags = (hdr.version || 0) & FLAG_VERSION_MASK;
  flags |= ((hdr.tier & 0x07) << FLAG_TIER_SHIFT) & FLAG_TIER_MASK;
  if (hdr.compressed) flags |= FLAG_COMPRESSED;
  if (hdr.fragmented) flags |= FLAG_FRAGMENTED;
  if (hdr.hasExtensions) flags |= FLAG_HAS_EXTENSIONS;
  buf[off++] = flags;

  // OpCode (big-endian u16)
  dv.setUint16(off, hdr.opcode, false);
  off += 2;

  // Fragment info
  if (hdr.fragmented && hdr.fragmentInfo) {
    dv.setUint32(off, hdr.fragmentInfo.messageID, false); off += 4;
    dv.setUint16(off, hdr.fragmentInfo.fragmentIndex, false); off += 2;
    dv.setUint16(off, hdr.fragmentInfo.totalFragments, false); off += 2;
  }

  // Tier-specific fields
  switch (hdr.tier) {
    case Tier.AUTHENTICATED:
      dv.setUint32(off, hdr.sequence || 0, false);
      break;

    case Tier.ENCRYPTED:
      dv.setUint32(off, hdr.sequence || 0, false); off += 4;
      setBigUint64(dv, off, hdr.sessionID || 0n); off += 8;
      break;

    case Tier.PFS:
      dv.setUint32(off, hdr.sequence || 0, false); off += 4;
      setBigUint64(dv, off, hdr.sessionID || 0n); off += 8;
      buf.set(hdr.nonce || new Uint8Array(12), off); off += 12;
      dv.setUint32(off, hdr.keyID || 0, false); off += 4;
      buf.set(hdr.ecdhPublic || new Uint8Array(32), off);
      break;

    case Tier.MAX_SECURITY:
      dv.setUint32(off, hdr.sequence || 0, false); off += 4;
      setBigUint64(dv, off, hdr.sessionID || 0n); off += 8;
      setBigUint64(dv, off, hdr.timestamp || 0n); off += 8;
      buf.set(hdr.nonce || new Uint8Array(12), off); off += 12;
      dv.setUint32(off, hdr.keyID || 0, false); off += 4;
      buf.set(hdr.ecdhPublic || new Uint8Array(32), off);
      break;
  }

  return buf;
}

export function unmarshalHeader(data) {
  if (data.length < 3) throw new Error('message too short');
  const dv = new DataView(data.buffer, data.byteOffset, data.byteLength);
  let off = 0;

  const flags = data[off++];
  const version = flags & FLAG_VERSION_MASK;
  const tier = (flags & FLAG_TIER_MASK) >> FLAG_TIER_SHIFT;
  const compressed = (flags & FLAG_COMPRESSED) !== 0;
  const fragmented = (flags & FLAG_FRAGMENTED) !== 0;
  const hasExtensions = (flags & FLAG_HAS_EXTENSIONS) !== 0;

  if (!tierIsValid(tier)) throw new Error('invalid security tier');

  const opcode = dv.getUint16(off, false); off += 2;

  const hdr = { version, tier, compressed, fragmented, hasExtensions, opcode,
    sequence: 0, sessionID: 0n, timestamp: 0n,
    nonce: new Uint8Array(12), keyID: 0, ecdhPublic: new Uint8Array(32),
    fragmentInfo: null };

  const expectedSize = headerSize(tier, fragmented);
  if (data.length < expectedSize) throw new Error('message too short');

  if (fragmented) {
    hdr.fragmentInfo = {
      messageID: dv.getUint32(off, false),
      fragmentIndex: dv.getUint16(off + 4, false),
      totalFragments: dv.getUint16(off + 6, false),
    };
    off += 8;
  }

  switch (tier) {
    case Tier.AUTHENTICATED:
      hdr.sequence = dv.getUint32(off, false); off += 4;
      break;

    case Tier.ENCRYPTED:
      hdr.sequence = dv.getUint32(off, false); off += 4;
      hdr.sessionID = getBigUint64(dv, off); off += 8;
      break;

    case Tier.PFS:
      hdr.sequence = dv.getUint32(off, false); off += 4;
      hdr.sessionID = getBigUint64(dv, off); off += 8;
      hdr.nonce = data.slice(off, off + 12); off += 12;
      hdr.keyID = dv.getUint32(off, false); off += 4;
      hdr.ecdhPublic = data.slice(off, off + 32); off += 32;
      break;

    case Tier.MAX_SECURITY:
      hdr.sequence = dv.getUint32(off, false); off += 4;
      hdr.sessionID = getBigUint64(dv, off); off += 8;
      hdr.timestamp = getBigUint64(dv, off); off += 8;
      hdr.nonce = data.slice(off, off + 12); off += 12;
      hdr.keyID = dv.getUint32(off, false); off += 4;
      hdr.ecdhPublic = data.slice(off, off + 32); off += 32;
      break;
  }

  return { header: hdr, bytesRead: off };
}

// --- Extensions (TLV) ---

export const ExtType = {
  TERMINATOR: 0x0000,
  REPLY_TO: 0x0001,
  TARGET_SERVICE: 0x0002,
};

export function marshalExtensions(ext) {
  let size = 0;
  const entries = [];
  if (ext.replyTo !== undefined && ext.replyTo !== null) {
    const val = new Uint8Array(4);
    new DataView(val.buffer).setUint32(0, ext.replyTo, false);
    entries.push({ type: ExtType.REPLY_TO, value: val });
    size += 8;
  }
  if (ext.targetService) {
    const val = new TextEncoder().encode(ext.targetService);
    entries.push({ type: ExtType.TARGET_SERVICE, value: val });
    size += 4 + val.length;
  }
  if (entries.length === 0) return null;

  const buf = new Uint8Array(size);
  const dv = new DataView(buf.buffer);
  let off = 0;
  for (const e of entries) {
    dv.setUint16(off, e.type, false); off += 2;
    dv.setUint16(off, e.value.length, false); off += 2;
    buf.set(e.value, off); off += e.value.length;
  }
  return buf;
}

export function parseExtensions(data) {
  const ext = { replyTo: null, targetService: null, raw: new Map() };
  let off = 0;
  while (off + 4 <= data.length) {
    const dv = new DataView(data.buffer, data.byteOffset, data.byteLength);
    const type = dv.getUint16(off, false);
    if (type === ExtType.TERMINATOR) { off += 4; break; }
    const length = dv.getUint16(off + 2, false);
    off += 4;
    if (off + length > data.length) throw new Error('message too short');
    const value = data.slice(off, off + length);
    off += length;

    ext.raw.set(type, value);
    if (type === ExtType.REPLY_TO && length === 4) {
      ext.replyTo = new DataView(value.buffer, value.byteOffset, 4).getUint32(0, false);
    } else if (type === ExtType.TARGET_SERVICE && length > 0) {
      ext.targetService = new TextDecoder().decode(value);
    }
  }
  return { extensions: ext, bytesRead: off };
}

// --- OpCodes ---

export const OpCode = {
  NOP: 0x0000, KEEPALIVE: 0x0001, KEEPALIVE_ACK: 0x0002,
  SESSION_INIT: 0x0003, SESSION_ACK: 0x0004,
  SESSION_CLOSE: 0x0005, SESSION_CLOSE_ACK: 0x0006,
  SESSION_RESUME: 0x0007, SESSION_RESUMED: 0x0008,
  KEY_EXCHANGE_INIT: 0x0010, KEY_EXCHANGE_RESPONSE: 0x0011, KEY_EXCHANGE_COMPLETE: 0x0012,
  SESSION_ROTATE: 0x0016, SESSION_REVOKE: 0x0017,
  ERROR: 0x00FF,
};

const opCodeNames = new Map([
  [0x0000, 'NOP'], [0x0001, 'KEEPALIVE'], [0x0002, 'KEEPALIVE_ACK'],
  [0x0003, 'SESSION_INIT'], [0x0004, 'SESSION_ACK'],
  [0x0005, 'SESSION_CLOSE'], [0x0006, 'SESSION_CLOSE_ACK'],
  [0x0007, 'SESSION_RESUME'], [0x0008, 'SESSION_RESUMED'],
  [0x0010, 'KEY_EXCHANGE_INIT'], [0x0011, 'KEY_EXCHANGE_RESPONSE'], [0x0012, 'KEY_EXCHANGE_COMPLETE'],
  [0x0016, 'SESSION_ROTATE'], [0x0017, 'SESSION_REVOKE'],
  [0x00FF, 'ERROR'],
]);

export function lookupOpCode(code) {
  return opCodeNames.get(code) || '';
}

export function opCodeCategory(code) { return (code >> 8) & 0xFF; }
export function opCodeOperation(code) { return code & 0xFF; }

// --- Error Codes ---

export const ErrorCode = {
  OK: 0x00, OK_ASYNC: 0x01, OK_PARTIAL: 0x02,
  BAD_REQUEST: 0x10, UNAUTHORIZED: 0x11, FORBIDDEN: 0x12, NOT_FOUND: 0x13,
  CONFLICT: 0x14, GONE: 0x15, TOO_LARGE: 0x16, INVALID_TIER: 0x17,
  INVALID_VERSION: 0x18, INVALID_SEQUENCE: 0x19, RATE_LIMITED: 0x1A,
  INTERNAL_ERROR: 0x20, SERVICE_UNAVAILABLE: 0x21, TIMEOUT: 0x22,
  OVERLOADED: 0x23, NOT_IMPLEMENTED: 0x24,
  NODE_UNREACHABLE: 0x30, CLUSTER_PARTITION: 0x31, SYNC_FAILED: 0x32,
  FEDERATION_DENIED: 0x33, VERSION_MISMATCH: 0x34, QUORUM_UNAVAILABLE: 0x35,
  SESSION_EXPIRED: 0x40, SESSION_INVALID: 0x41, KEY_EXPIRED: 0x42,
  HANDSHAKE_FAILED: 0x43, REPLAY_DETECTED: 0x44,
};

const errorCodeNames = new Map([
  [0x00, 'OK'], [0x01, 'OK_ASYNC'], [0x02, 'OK_PARTIAL'],
  [0x10, 'BAD_REQUEST'], [0x11, 'UNAUTHORIZED'], [0x12, 'FORBIDDEN'], [0x13, 'NOT_FOUND'],
  [0x14, 'CONFLICT'], [0x15, 'GONE'], [0x16, 'TOO_LARGE'], [0x17, 'INVALID_TIER'],
  [0x18, 'INVALID_VERSION'], [0x19, 'INVALID_SEQUENCE'], [0x1A, 'RATE_LIMITED'],
  [0x20, 'INTERNAL_ERROR'], [0x21, 'SERVICE_UNAVAILABLE'], [0x22, 'TIMEOUT'],
  [0x23, 'OVERLOADED'], [0x24, 'NOT_IMPLEMENTED'],
  [0x30, 'NODE_UNREACHABLE'], [0x31, 'CLUSTER_PARTITION'], [0x32, 'SYNC_FAILED'],
  [0x33, 'FEDERATION_DENIED'], [0x34, 'VERSION_MISMATCH'], [0x35, 'QUORUM_UNAVAILABLE'],
  [0x40, 'SESSION_EXPIRED'], [0x41, 'SESSION_INVALID'], [0x42, 'KEY_EXPIRED'],
  [0x43, 'HANDSHAKE_FAILED'], [0x44, 'REPLAY_DETECTED'],
]);

export function lookupErrorCode(code) {
  return errorCodeNames.get(code) || `UNKNOWN(0x${code.toString(16).padStart(2, '0')})`;
}

export function isSuccess(code) { return code <= 0x0F; }
export function isClientError(code) { return code >= 0x10 && code <= 0x1F; }
export function isServerError(code) { return code >= 0x20 && code <= 0x2F; }
export function isFederationError(code) { return code >= 0x30 && code <= 0x3F; }
export function isSessionError(code) { return code >= 0x40 && code <= 0x4F; }
export function isRetryable(code) {
  return code === 0x21 || code === 0x22 || code === 0x23 ||
         code === 0x30 || code === 0x31 || code === 0x35;
}

// --- BigInt helpers for 64-bit fields ---

function setBigUint64(dv, off, val) {
  const big = BigInt(val);
  dv.setUint32(off, Number((big >> 32n) & 0xFFFFFFFFn), false);
  dv.setUint32(off + 4, Number(big & 0xFFFFFFFFn), false);
}

function getBigUint64(dv, off) {
  const hi = BigInt(dv.getUint32(off, false));
  const lo = BigInt(dv.getUint32(off + 4, false));
  return (hi << 32n) | lo;
}
