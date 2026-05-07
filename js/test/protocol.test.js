import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import {
  Tier, getTierInfo, tierIsValid,
  marshalHeader, unmarshalHeader, headerSize, trailerSize,
  marshalExtensions, parseExtensions,
  OpCode, lookupOpCode, opCodeCategory, opCodeOperation,
  ErrorCode, lookupErrorCode, isSuccess, isClientError, isServerError,
  isFederationError, isSessionError, isRetryable,
} from '../src/protocol.js';
import { fromHex, toHex } from '../src/index.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const vectors = JSON.parse(readFileSync(join(__dirname, '../../testdata/protocol_vectors.json'), 'utf-8'));

// --- Header cross-validation ---

describe('Cross-Validation: Protocol Headers', () => {
  for (const v of vectors.headers) {
    it(`marshal ${v.name}`, () => {
      const hdr = {
        version: v.version,
        tier: v.tier,
        opcode: v.opcode,
        compressed: v.compressed,
        fragmented: v.fragmented,
        hasExtensions: v.has_extensions,
        sequence: v.sequence || 0,
        sessionID: v.session_id_hex ? BigInt('0x' + v.session_id_hex) : 0n,
        timestamp: v.timestamp_hex ? BigInt('0x' + v.timestamp_hex) : 0n,
        nonce: v.nonce_hex ? fromHex(v.nonce_hex) : new Uint8Array(12),
        keyID: v.key_id || 0,
        ecdhPublic: v.ecdh_hex ? fromHex(v.ecdh_hex) : new Uint8Array(32),
        fragmentInfo: null,
      };

      const buf = marshalHeader(hdr);
      assert.equal(toHex(buf), v.header_hex);
      assert.equal(buf.length, v.header_size);
    });

    it(`unmarshal ${v.name}`, () => {
      const data = fromHex(v.header_hex);
      const { header: hdr } = unmarshalHeader(data);

      assert.equal(hdr.version, v.version);
      assert.equal(hdr.tier, v.tier);
      assert.equal(hdr.opcode, v.opcode);
      assert.equal(hdr.compressed, v.compressed);
      assert.equal(hdr.fragmented, v.fragmented);
      assert.equal(hdr.hasExtensions, v.has_extensions);
      assert.equal(hdr.sequence, v.sequence || 0);

      if (v.session_id_hex) {
        assert.equal(hdr.sessionID, BigInt('0x' + v.session_id_hex));
      }
      if (v.timestamp_hex) {
        assert.equal(hdr.timestamp, BigInt('0x' + v.timestamp_hex));
      }
      if (v.nonce_hex) {
        assert.equal(toHex(hdr.nonce), v.nonce_hex);
      }
      if (v.key_id) {
        assert.equal(hdr.keyID, v.key_id);
      }
      if (v.ecdh_hex) {
        assert.equal(toHex(hdr.ecdhPublic), v.ecdh_hex);
      }
    });
  }
});

// --- Extension cross-validation ---

describe('Cross-Validation: Protocol Extensions', () => {
  for (const v of vectors.extensions) {
    it(`parse ${v.name}`, () => {
      const data = fromHex(v.hex_data);
      const { extensions: ext } = parseExtensions(data);

      if (v.reply_to !== undefined && v.reply_to !== null) {
        assert.equal(ext.replyTo, v.reply_to);
      }
      if (v.target_service) {
        assert.equal(ext.targetService, v.target_service);
      }
    });

    // Single-extension vectors can be marshal-verified byte-for-byte
    if (!v.target_service || !v.reply_to) {
      it(`marshal ${v.name}`, () => {
        const ext = {};
        if (v.reply_to !== undefined && v.reply_to !== null) ext.replyTo = v.reply_to;
        if (v.target_service) ext.targetService = v.target_service;

        const buf = marshalExtensions(ext);
        assert.equal(toHex(buf), v.hex_data);
        assert.equal(buf.length, v.size);
      });
    }
  }
});

// --- OpCode cross-validation ---

describe('Cross-Validation: Protocol OpCodes', () => {
  for (const v of vectors.opcodes) {
    it(`opcode 0x${v.code.toString(16).padStart(4, '0')} = ${v.name}`, () => {
      assert.equal(lookupOpCode(v.code), v.name);
      assert.equal(opCodeCategory(v.code), v.category);
      assert.equal(opCodeOperation(v.code), v.operation);
    });
  }
});

// --- ErrorCode cross-validation ---

describe('Cross-Validation: Protocol ErrorCodes', () => {
  for (const v of vectors.error_codes) {
    it(`error 0x${v.code.toString(16).padStart(2, '0')} = ${v.name}`, () => {
      assert.equal(lookupErrorCode(v.code), v.name);
      assert.equal(isSuccess(v.code), v.is_success);
      assert.equal(isClientError(v.code), v.is_client_error);
      assert.equal(isServerError(v.code), v.is_server_error);
      assert.equal(isFederationError(v.code), v.is_federation_error);
      assert.equal(isSessionError(v.code), v.is_session_error);
      assert.equal(isRetryable(v.code), v.is_retryable);
    });
  }
});

// --- Tier info ---

describe('Protocol Tier Info', () => {
  it('valid tiers return correct info', () => {
    assert.equal(getTierInfo(Tier.PLAINTEXT).name, 'plaintext');
    assert.equal(getTierInfo(Tier.MAX_SECURITY).name, 'max_security');
    assert.equal(getTierInfo(Tier.ENCRYPTED).encrypted, true);
    assert.equal(getTierInfo(Tier.CHECKSUM).encrypted, false);
  });

  it('invalid tier returns unknown', () => {
    assert.equal(getTierInfo(99).name, 'unknown');
  });

  it('tierIsValid', () => {
    assert.equal(tierIsValid(0), true);
    assert.equal(tierIsValid(5), true);
    assert.equal(tierIsValid(6), false);
  });
});

// --- Round-trip ---

describe('Protocol Header Round-Trip', () => {
  it('tier4 marshal-unmarshal round-trip', () => {
    const original = {
      version: 0, tier: Tier.PFS, opcode: OpCode.KEY_EXCHANGE_INIT,
      compressed: false, fragmented: false, hasExtensions: false,
      sequence: 12345, sessionID: 0xABCDEF0123456789n,
      nonce: new Uint8Array([1,2,3,4,5,6,7,8,9,10,11,12]),
      keyID: 0xDEAD, ecdhPublic: new Uint8Array(32).fill(0x42),
      fragmentInfo: null,
    };

    const buf = marshalHeader(original);
    const { header: parsed } = unmarshalHeader(buf);

    assert.equal(parsed.tier, original.tier);
    assert.equal(parsed.opcode, original.opcode);
    assert.equal(parsed.sequence, original.sequence);
    assert.equal(parsed.sessionID, original.sessionID);
    assert.equal(parsed.keyID, original.keyID);
    assert.deepEqual(parsed.nonce, original.nonce);
    assert.deepEqual(parsed.ecdhPublic, original.ecdhPublic);
  });
});
