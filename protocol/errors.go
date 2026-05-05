package protocol

import (
	"errors"
	"fmt"
)

// ErrorCode represents protocol error codes.
type ErrorCode uint32

// Success codes (0x00-0x0F).
const (
	CodeOK        ErrorCode = 0x00 // Success
	CodeOKAsync   ErrorCode = 0x01 // Async operation started
	CodeOKPartial ErrorCode = 0x02 // Partial success
)

// Client errors (0x10-0x1F).
const (
	CodeBadRequest      ErrorCode = 0x10 // Malformed request
	CodeUnauthorized    ErrorCode = 0x11 // Authentication required
	CodeForbidden       ErrorCode = 0x12 // Permission denied
	CodeNotFound        ErrorCode = 0x13 // Resource not found
	CodeConflict        ErrorCode = 0x14 // Resource conflict
	CodeGone            ErrorCode = 0x15 // Resource no longer available
	CodeTooLarge        ErrorCode = 0x16 // Payload too large
	CodeInvalidTier     ErrorCode = 0x17 // Invalid security tier
	CodeInvalidVersion  ErrorCode = 0x18 // Protocol version mismatch
	CodeInvalidSequence ErrorCode = 0x19 // Invalid sequence number
	CodeRateLimited     ErrorCode = 0x1A // Too many requests
)

// Server errors (0x20-0x2F).
const (
	CodeInternalError      ErrorCode = 0x20 // Internal server error
	CodeServiceUnavailable ErrorCode = 0x21 // Service temporarily unavailable
	CodeTimeout            ErrorCode = 0x22 // Operation timeout
	CodeOverloaded         ErrorCode = 0x23 // Server overloaded
	CodeNotImplemented     ErrorCode = 0x24 // Operation not implemented
)

// Federation errors (0x30-0x3F).
const (
	CodeNodeUnreachable   ErrorCode = 0x30 // Remote node unreachable
	CodeClusterPartition  ErrorCode = 0x31 // Cluster partition detected
	CodeSyncFailed        ErrorCode = 0x32 // Synchronization failed
	CodeFederationDenied  ErrorCode = 0x33 // Federation not permitted
	CodeVersionMismatch   ErrorCode = 0x34 // Protocol version mismatch
	CodeQuorumUnavailable ErrorCode = 0x35 // Quorum not available
)

// Session errors (0x40-0x4F).
const (
	CodeSessionExpired  ErrorCode = 0x40 // Session expired
	CodeSessionInvalid  ErrorCode = 0x41 // Session invalid
	CodeKeyExpired      ErrorCode = 0x42 // Encryption key expired
	CodeHandshakeFailed ErrorCode = 0x43 // Key exchange failed
	CodeReplayDetected  ErrorCode = 0x44 // Replay attack detected
)

// errorCodeNames maps error codes to their string representation.
var errorCodeNames = map[ErrorCode]string{
	CodeOK:                 "OK",
	CodeOKAsync:            "OK_ASYNC",
	CodeOKPartial:          "OK_PARTIAL",
	CodeBadRequest:         "BAD_REQUEST",
	CodeUnauthorized:       "UNAUTHORIZED",
	CodeForbidden:          "FORBIDDEN",
	CodeNotFound:           "NOT_FOUND",
	CodeConflict:           "CONFLICT",
	CodeGone:               "GONE",
	CodeTooLarge:           "TOO_LARGE",
	CodeInvalidTier:        "INVALID_TIER",
	CodeInvalidVersion:     "INVALID_VERSION",
	CodeInvalidSequence:    "INVALID_SEQUENCE",
	CodeRateLimited:        "RATE_LIMITED",
	CodeInternalError:      "INTERNAL_ERROR",
	CodeServiceUnavailable: "SERVICE_UNAVAILABLE",
	CodeTimeout:            "TIMEOUT",
	CodeOverloaded:         "OVERLOADED",
	CodeNotImplemented:     "NOT_IMPLEMENTED",
	CodeNodeUnreachable:    "NODE_UNREACHABLE",
	CodeClusterPartition:   "CLUSTER_PARTITION",
	CodeSyncFailed:         "SYNC_FAILED",
	CodeFederationDenied:   "FEDERATION_DENIED",
	CodeVersionMismatch:    "VERSION_MISMATCH",
	CodeQuorumUnavailable:  "QUORUM_UNAVAILABLE",
	CodeSessionExpired:     "SESSION_EXPIRED",
	CodeSessionInvalid:     "SESSION_INVALID",
	CodeKeyExpired:         "KEY_EXPIRED",
	CodeHandshakeFailed:    "HANDSHAKE_FAILED",
	CodeReplayDetected:     "REPLAY_DETECTED",
}

// String returns the error code name.
func (c ErrorCode) String() string {
	if name, ok := errorCodeNames[c]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(0x%02X)", uint32(c))
}

// IsSuccess returns true if the code indicates success.
func (c ErrorCode) IsSuccess() bool {
	return c <= 0x0F
}

// IsClientError returns true if the code indicates a client error.
func (c ErrorCode) IsClientError() bool {
	return c >= 0x10 && c <= 0x1F
}

// IsServerError returns true if the code indicates a server error.
func (c ErrorCode) IsServerError() bool {
	return c >= 0x20 && c <= 0x2F
}

// IsFederationError returns true if the code indicates a federation error.
func (c ErrorCode) IsFederationError() bool {
	return c >= 0x30 && c <= 0x3F
}

// IsSessionError returns true if the code indicates a session error.
func (c ErrorCode) IsSessionError() bool {
	return c >= 0x40 && c <= 0x4F
}

// IsRetryable returns true if the error is retryable.
func (c ErrorCode) IsRetryable() bool {
	switch c {
	case CodeServiceUnavailable, CodeTimeout, CodeOverloaded,
		CodeNodeUnreachable, CodeClusterPartition, CodeQuorumUnavailable:
		return true
	default:
		return false
	}
}

// Error represents an error in the protocol layer.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches a target error code.
func (e *Error) Is(target error) bool {
	var pe *Error
	if errors.As(target, &pe) {
		return e.Code == pe.Code
	}
	return false
}

// NewError creates a new protocol error.
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithCause creates a new protocol error with a cause.
func NewErrorWithCause(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// CodeFromError extracts the error code from an error.
// Returns CodeInternalError if the error is not a protocol Error.
func CodeFromError(err error) ErrorCode {
	var pe *Error
	if errors.As(err, &pe) {
		return pe.Code
	}
	return CodeInternalError
}
