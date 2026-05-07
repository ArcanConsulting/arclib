package protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// headerVector stores a header marshal/unmarshal test vector.
type headerVector struct {
	Name         string `json:"name"`
	Tier         uint8  `json:"tier"`
	OpCode       uint16 `json:"opcode"`
	Version      uint8  `json:"version"`
	Compressed   bool   `json:"compressed"`
	Fragmented   bool   `json:"fragmented"`
	HasExt       bool   `json:"has_extensions"`
	Sequence     uint32 `json:"sequence,omitempty"`
	SessionIDHex string `json:"session_id_hex,omitempty"`
	TimestampHex string `json:"timestamp_hex,omitempty"`
	NonceHex     string `json:"nonce_hex,omitempty"`
	KeyID        uint32 `json:"key_id,omitempty"`
	ECDHHex      string `json:"ecdh_hex,omitempty"`
	HeaderHex    string `json:"header_hex"`
	HeaderSize   int    `json:"header_size"`
}

// extensionVector stores an extension marshal/parse test vector.
type extensionVector struct {
	Name      string `json:"name"`
	ReplyTo   *int   `json:"reply_to,omitempty"`
	TargetSvc string `json:"target_service,omitempty"`
	HexData   string `json:"hex_data"`
	Size      int    `json:"size"`
}

// opcodeVector stores an opcode lookup test vector.
type opcodeVector struct {
	Code     uint16 `json:"code"`
	Name     string `json:"name"`
	Category uint8  `json:"category"`
	Op       uint8  `json:"operation"`
}

// errorCodeVector stores an error code classification test vector.
type errorCodeVector struct {
	Code        uint32 `json:"code"`
	Name        string `json:"name"`
	IsSuccess   bool   `json:"is_success"`
	IsClient    bool   `json:"is_client_error"`
	IsServer    bool   `json:"is_server_error"`
	IsFed       bool   `json:"is_federation_error"`
	IsSession   bool   `json:"is_session_error"`
	IsRetryable bool   `json:"is_retryable"`
}

// protocolVectors is the top-level JSON output.
type protocolVectors struct {
	Headers    []headerVector    `json:"headers"`
	Extensions []extensionVector `json:"extensions"`
	OpCodes    []opcodeVector    `json:"opcodes"`
	ErrorCodes []errorCodeVector `json:"error_codes"`
}

func TestGenerateProtocolVectors(t *testing.T) {
	if os.Getenv("GENERATE_VECTORS") == "" {
		t.Skip("set GENERATE_VECTORS=1 to generate protocol test vectors")
	}

	vectors := protocolVectors{}

	// --- Header vectors ---
	headerCases := []struct {
		name   string
		header Header
	}{
		{
			name: "tier0_plaintext_nop",
			header: Header{
				Version: 0, Tier: TierPlaintext, OpCode: OpNop,
			},
		},
		{
			name: "tier0_keepalive",
			header: Header{
				Version: 0, Tier: TierPlaintext, OpCode: OpKeepalive,
			},
		},
		{
			name: "tier1_checksum",
			header: Header{
				Version: 0, Tier: TierChecksum, OpCode: OpSessionInit,
			},
		},
		{
			name: "tier2_authenticated",
			header: Header{
				Version: 0, Tier: TierAuthenticated, OpCode: OpSessionAck,
				Sequence: 42,
			},
		},
		{
			name: "tier3_encrypted",
			header: Header{
				Version: 0, Tier: TierEncrypted, OpCode: OpKeyExchangeInit,
				Sequence: 100, SessionID: 0xDEADBEEFCAFE0001,
			},
		},
		{
			name: "tier4_pfs",
			header: Header{
				Version: 0, Tier: TierPFS, OpCode: OpKeyExchangeResponse,
				Sequence: 7, SessionID: 0x0102030405060708,
				Nonce: [12]byte{0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15},
				KeyID: 0xAABBCCDD,
				ECDHPublic: [32]byte{
					0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
					0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
					0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
					0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20,
				},
			},
		},
		{
			name: "tier5_max_security",
			header: Header{
				Version: 0, Tier: TierMaxSecurity, OpCode: OpError,
				Sequence: 999, SessionID: 0xFFFFFFFFFFFFFFFF,
				Timestamp: 1700000000000,
				Nonce:     [12]byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44},
				KeyID:     0x12345678,
				ECDHPublic: [32]byte{
					0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11,
					0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
					0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11,
					0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
				},
			},
		},
		{
			name: "tier0_with_extensions_flag",
			header: Header{
				Version: 0, Tier: TierPlaintext, OpCode: OpSessionInit,
				HasExtensions: true,
			},
		},
		{
			name: "tier0_compressed",
			header: Header{
				Version: 0, Tier: TierPlaintext, OpCode: OpNop,
				Compressed: true,
			},
		},
	}

	for _, tc := range headerCases {
		hdr := tc.header
		buf := hdr.MarshalHeader()

		hv := headerVector{
			Name:       tc.name,
			Tier:       uint8(hdr.Tier),
			OpCode:     uint16(hdr.OpCode),
			Version:    hdr.Version,
			Compressed: hdr.Compressed,
			Fragmented: hdr.Fragmented,
			HasExt:     hdr.HasExtensions,
			Sequence:   hdr.Sequence,
			KeyID:      hdr.KeyID,
			HeaderHex:  hex.EncodeToString(buf),
			HeaderSize: len(buf),
		}
		if hdr.SessionID != 0 {
			hv.SessionIDHex = fmt.Sprintf("%016x", hdr.SessionID)
		}
		if hdr.Timestamp != 0 {
			hv.TimestampHex = fmt.Sprintf("%016x", hdr.Timestamp)
		}
		if hdr.Nonce != [12]byte{} {
			hv.NonceHex = hex.EncodeToString(hdr.Nonce[:])
		}
		if hdr.ECDHPublic != [32]byte{} {
			hv.ECDHHex = hex.EncodeToString(hdr.ECDHPublic[:])
		}

		vectors.Headers = append(vectors.Headers, hv)
	}

	// --- Extension vectors ---
	extCases := []struct {
		name      string
		replyTo   *uint32
		targetSvc string
	}{
		{
			name:    "reply_to_only",
			replyTo: ptrU32(42),
		},
		{
			name:      "target_service_only",
			targetSvc: "chat",
		},
		{
			name:      "both_extensions",
			replyTo:   ptrU32(100),
			targetSvc: "vfs",
		},
	}

	for _, tc := range extCases {
		ext := &Extensions{entries: make(map[ExtensionType][]byte)}
		var replyVal *int
		if tc.replyTo != nil {
			ext.SetReplyTo(*tc.replyTo)
			v := int(*tc.replyTo)
			replyVal = &v
		}
		if tc.targetSvc != "" {
			ext.SetTargetService(tc.targetSvc)
		}

		data := ext.Marshal()
		vectors.Extensions = append(vectors.Extensions, extensionVector{
			Name:      tc.name,
			ReplyTo:   replyVal,
			TargetSvc: tc.targetSvc,
			HexData:   hex.EncodeToString(data),
			Size:      len(data),
		})
	}

	// --- OpCode vectors ---
	opcodes := []OpCode{
		OpNop, OpKeepalive, OpKeepaliveAck,
		OpSessionInit, OpSessionAck, OpSessionClose, OpSessionCloseAck,
		OpSessionResume, OpSessionResumed,
		OpKeyExchangeInit, OpKeyExchangeResponse, OpKeyExchangeComplete,
		OpSessionRotate, OpSessionRevoke,
		OpError,
	}

	for _, op := range opcodes {
		vectors.OpCodes = append(vectors.OpCodes, opcodeVector{
			Code:     uint16(op),
			Name:     op.String(),
			Category: op.Category(),
			Op:       op.Operation(),
		})
	}

	// --- ErrorCode vectors ---
	errorCodes := []ErrorCode{
		CodeOK, CodeOKAsync, CodeOKPartial,
		CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeNotFound,
		CodeConflict, CodeGone, CodeTooLarge, CodeInvalidTier,
		CodeInvalidVersion, CodeInvalidSequence, CodeRateLimited,
		CodeInternalError, CodeServiceUnavailable, CodeTimeout,
		CodeOverloaded, CodeNotImplemented,
		CodeNodeUnreachable, CodeClusterPartition, CodeSyncFailed,
		CodeFederationDenied, CodeVersionMismatch, CodeQuorumUnavailable,
		CodeSessionExpired, CodeSessionInvalid, CodeKeyExpired,
		CodeHandshakeFailed, CodeReplayDetected,
	}

	for _, ec := range errorCodes {
		vectors.ErrorCodes = append(vectors.ErrorCodes, errorCodeVector{
			Code:        uint32(ec),
			Name:        ec.String(),
			IsSuccess:   ec.IsSuccess(),
			IsClient:    ec.IsClientError(),
			IsServer:    ec.IsServerError(),
			IsFed:       ec.IsFederationError(),
			IsSession:   ec.IsSessionError(),
			IsRetryable: ec.IsRetryable(),
		})
	}

	// Write output
	outDir := "../testdata"
	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		t.Fatalf("marshaling vectors: %v", err)
	}

	outPath := outDir + "/protocol_vectors.json"
	if err := os.WriteFile(outPath, data, 0o600); err != nil {
		t.Fatalf("writing vectors: %v", err)
	}

	t.Logf("wrote %d bytes to %s", len(data), outPath)
}

func ptrU32(v uint32) *uint32 { return &v }
