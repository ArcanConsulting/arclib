package protocol

import (
	"errors"
	"testing"
)

func TestErrorCodeString(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{CodeOK, "OK"},
		{CodeOKAsync, "OK_ASYNC"},
		{CodeOKPartial, "OK_PARTIAL"},
		{CodeBadRequest, "BAD_REQUEST"},
		{CodeUnauthorized, "UNAUTHORIZED"},
		{CodeForbidden, "FORBIDDEN"},
		{CodeNotFound, "NOT_FOUND"},
		{CodeConflict, "CONFLICT"},
		{CodeGone, "GONE"},
		{CodeTooLarge, "TOO_LARGE"},
		{CodeInvalidTier, "INVALID_TIER"},
		{CodeInvalidVersion, "INVALID_VERSION"},
		{CodeInvalidSequence, "INVALID_SEQUENCE"},
		{CodeRateLimited, "RATE_LIMITED"},
		{CodeInternalError, "INTERNAL_ERROR"},
		{CodeServiceUnavailable, "SERVICE_UNAVAILABLE"},
		{CodeTimeout, "TIMEOUT"},
		{CodeOverloaded, "OVERLOADED"},
		{CodeNotImplemented, "NOT_IMPLEMENTED"},
		{CodeNodeUnreachable, "NODE_UNREACHABLE"},
		{CodeClusterPartition, "CLUSTER_PARTITION"},
		{CodeSyncFailed, "SYNC_FAILED"},
		{CodeFederationDenied, "FEDERATION_DENIED"},
		{CodeVersionMismatch, "VERSION_MISMATCH"},
		{CodeQuorumUnavailable, "QUORUM_UNAVAILABLE"},
		{CodeSessionExpired, "SESSION_EXPIRED"},
		{CodeSessionInvalid, "SESSION_INVALID"},
		{CodeKeyExpired, "KEY_EXPIRED"},
		{CodeHandshakeFailed, "HANDSHAKE_FAILED"},
		{CodeReplayDetected, "REPLAY_DETECTED"},
	}

	for _, tt := range tests {
		if got := tt.code.String(); got != tt.expected {
			t.Errorf("ErrorCode(0x%02X).String() = %q, want %q", uint32(tt.code), got, tt.expected)
		}
	}
}

func TestErrorCodeStringUnknown(t *testing.T) {
	code := ErrorCode(0xFE)
	got := code.String()
	if got != "UNKNOWN(0xFE)" {
		t.Errorf("ErrorCode(0xFE).String() = %q, want %q", got, "UNKNOWN(0xFE)")
	}
}

func TestErrorCodeIsSuccess(t *testing.T) {
	successCodes := []ErrorCode{CodeOK, CodeOKAsync, CodeOKPartial}
	for _, code := range successCodes {
		if !code.IsSuccess() {
			t.Errorf("ErrorCode(0x%02X).IsSuccess() = false, want true", uint32(code))
		}
	}

	nonSuccessCodes := []ErrorCode{CodeBadRequest, CodeInternalError, CodeSessionExpired}
	for _, code := range nonSuccessCodes {
		if code.IsSuccess() {
			t.Errorf("ErrorCode(0x%02X).IsSuccess() = true, want false", uint32(code))
		}
	}
}

func TestErrorCodeIsClientError(t *testing.T) {
	clientCodes := []ErrorCode{
		CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeNotFound,
		CodeRateLimited,
	}
	for _, code := range clientCodes {
		if !code.IsClientError() {
			t.Errorf("ErrorCode(0x%02X).IsClientError() = false, want true", uint32(code))
		}
	}

	if CodeOK.IsClientError() {
		t.Error("CodeOK should not be a client error")
	}
	if CodeInternalError.IsClientError() {
		t.Error("CodeInternalError should not be a client error")
	}
}

func TestErrorCodeIsServerError(t *testing.T) {
	serverCodes := []ErrorCode{
		CodeInternalError, CodeServiceUnavailable, CodeTimeout, CodeOverloaded,
		CodeNotImplemented,
	}
	for _, code := range serverCodes {
		if !code.IsServerError() {
			t.Errorf("ErrorCode(0x%02X).IsServerError() = false, want true", uint32(code))
		}
	}

	if CodeBadRequest.IsServerError() {
		t.Error("CodeBadRequest should not be a server error")
	}
}

func TestErrorCodeIsFederationError(t *testing.T) {
	fedCodes := []ErrorCode{
		CodeNodeUnreachable, CodeClusterPartition, CodeSyncFailed,
		CodeFederationDenied, CodeVersionMismatch, CodeQuorumUnavailable,
	}
	for _, code := range fedCodes {
		if !code.IsFederationError() {
			t.Errorf("ErrorCode(0x%02X).IsFederationError() = false, want true", uint32(code))
		}
	}
}

func TestErrorCodeIsSessionError(t *testing.T) {
	sessionCodes := []ErrorCode{
		CodeSessionExpired, CodeSessionInvalid, CodeKeyExpired,
		CodeHandshakeFailed, CodeReplayDetected,
	}
	for _, code := range sessionCodes {
		if !code.IsSessionError() {
			t.Errorf("ErrorCode(0x%02X).IsSessionError() = false, want true", uint32(code))
		}
	}
}

func TestErrorCodeIsRetryable(t *testing.T) {
	retryable := []ErrorCode{
		CodeServiceUnavailable, CodeTimeout, CodeOverloaded,
		CodeNodeUnreachable, CodeClusterPartition, CodeQuorumUnavailable,
	}
	for _, code := range retryable {
		if !code.IsRetryable() {
			t.Errorf("ErrorCode(0x%02X).IsRetryable() = false, want true", uint32(code))
		}
	}

	notRetryable := []ErrorCode{
		CodeOK, CodeBadRequest, CodeForbidden, CodeNotFound,
		CodeInternalError, CodeSessionExpired,
	}
	for _, code := range notRetryable {
		if code.IsRetryable() {
			t.Errorf("ErrorCode(0x%02X).IsRetryable() = true, want false", uint32(code))
		}
	}
}

func TestErrorInterface(t *testing.T) {
	err := NewError(CodeBadRequest, "invalid payload")
	if err.Error() != "BAD_REQUEST: invalid payload" {
		t.Errorf("Error() = %q, want %q", err.Error(), "BAD_REQUEST: invalid payload")
	}
}

func TestErrorWithCause(t *testing.T) {
	cause := errors.New("connection reset")
	err := NewErrorWithCause(CodeTimeout, "request failed", cause)

	if !errors.Is(err, cause) {
		t.Error("errors.Is(err, cause) should be true")
	}

	expected := "TIMEOUT: request failed (connection reset)"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestErrorIs(t *testing.T) {
	err1 := NewError(CodeNotFound, "user not found")
	err2 := NewError(CodeNotFound, "device not found")
	err3 := NewError(CodeForbidden, "access denied")

	if !errors.Is(err1, err2) {
		t.Error("errors.Is should match same error code")
	}
	if errors.Is(err1, err3) {
		t.Error("errors.Is should not match different error codes")
	}
}

func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying")
	err := NewErrorWithCause(CodeInternalError, "wrapped", cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestCodeFromError(t *testing.T) {
	protoErr := NewError(CodeForbidden, "nope")
	if got := CodeFromError(protoErr); got != CodeForbidden {
		t.Errorf("CodeFromError(proto) = 0x%02X, want 0x%02X", uint32(got), uint32(CodeForbidden))
	}

	stdErr := errors.New("standard error")
	if got := CodeFromError(stdErr); got != CodeInternalError {
		t.Errorf("CodeFromError(std) = 0x%02X, want 0x%02X", uint32(got), uint32(CodeInternalError))
	}
}

func TestErrorCodeRanges(t *testing.T) {
	// Verify code values are in expected ranges
	if CodeOK != 0x00 {
		t.Errorf("CodeOK = 0x%02X, want 0x00", uint32(CodeOK))
	}
	if CodeBadRequest != 0x10 {
		t.Errorf("CodeBadRequest = 0x%02X, want 0x10", uint32(CodeBadRequest))
	}
	if CodeInternalError != 0x20 {
		t.Errorf("CodeInternalError = 0x%02X, want 0x20", uint32(CodeInternalError))
	}
	if CodeNodeUnreachable != 0x30 {
		t.Errorf("CodeNodeUnreachable = 0x%02X, want 0x30", uint32(CodeNodeUnreachable))
	}
	if CodeSessionExpired != 0x40 {
		t.Errorf("CodeSessionExpired = 0x%02X, want 0x40", uint32(CodeSessionExpired))
	}
}
