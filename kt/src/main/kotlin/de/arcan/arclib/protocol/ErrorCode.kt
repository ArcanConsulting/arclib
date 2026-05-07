package de.arcan.arclib.protocol

/** Protocol error codes with classification helpers. */
object ErrorCode {
    // Success codes (0x00-0x0F)
    const val OK         = 0x00
    const val OK_ASYNC   = 0x01
    const val OK_PARTIAL = 0x02

    // Client errors (0x10-0x1F)
    const val BAD_REQUEST       = 0x10
    const val UNAUTHORIZED      = 0x11
    const val FORBIDDEN         = 0x12
    const val NOT_FOUND         = 0x13
    const val CONFLICT          = 0x14
    const val GONE              = 0x15
    const val TOO_LARGE         = 0x16
    const val INVALID_TIER      = 0x17
    const val INVALID_VERSION   = 0x18
    const val INVALID_SEQUENCE  = 0x19
    const val RATE_LIMITED      = 0x1A

    // Server errors (0x20-0x2F)
    const val INTERNAL_ERROR       = 0x20
    const val SERVICE_UNAVAILABLE  = 0x21
    const val TIMEOUT              = 0x22
    const val OVERLOADED           = 0x23
    const val NOT_IMPLEMENTED      = 0x24

    // Federation errors (0x30-0x3F)
    const val NODE_UNREACHABLE    = 0x30
    const val CLUSTER_PARTITION   = 0x31
    const val SYNC_FAILED         = 0x32
    const val FEDERATION_DENIED   = 0x33
    const val VERSION_MISMATCH    = 0x34
    const val QUORUM_UNAVAILABLE  = 0x35

    // Session errors (0x40-0x4F)
    const val SESSION_EXPIRED   = 0x40
    const val SESSION_INVALID   = 0x41
    const val KEY_EXPIRED       = 0x42
    const val HANDSHAKE_FAILED  = 0x43
    const val REPLAY_DETECTED   = 0x44

    private val names = mapOf(
        OK to "OK", OK_ASYNC to "OK_ASYNC", OK_PARTIAL to "OK_PARTIAL",
        BAD_REQUEST to "BAD_REQUEST", UNAUTHORIZED to "UNAUTHORIZED", FORBIDDEN to "FORBIDDEN",
        NOT_FOUND to "NOT_FOUND", CONFLICT to "CONFLICT", GONE to "GONE",
        TOO_LARGE to "TOO_LARGE", INVALID_TIER to "INVALID_TIER",
        INVALID_VERSION to "INVALID_VERSION", INVALID_SEQUENCE to "INVALID_SEQUENCE",
        RATE_LIMITED to "RATE_LIMITED",
        INTERNAL_ERROR to "INTERNAL_ERROR", SERVICE_UNAVAILABLE to "SERVICE_UNAVAILABLE",
        TIMEOUT to "TIMEOUT", OVERLOADED to "OVERLOADED", NOT_IMPLEMENTED to "NOT_IMPLEMENTED",
        NODE_UNREACHABLE to "NODE_UNREACHABLE", CLUSTER_PARTITION to "CLUSTER_PARTITION",
        SYNC_FAILED to "SYNC_FAILED", FEDERATION_DENIED to "FEDERATION_DENIED",
        VERSION_MISMATCH to "VERSION_MISMATCH", QUORUM_UNAVAILABLE to "QUORUM_UNAVAILABLE",
        SESSION_EXPIRED to "SESSION_EXPIRED", SESSION_INVALID to "SESSION_INVALID",
        KEY_EXPIRED to "KEY_EXPIRED", HANDSHAKE_FAILED to "HANDSHAKE_FAILED",
        REPLAY_DETECTED to "REPLAY_DETECTED",
    )

    fun lookup(code: Int): String = names[code] ?: "UNKNOWN(0x${code.toString(16).padStart(2, '0')})"
    fun isSuccess(code: Int): Boolean = code <= 0x0F
    fun isClientError(code: Int): Boolean = code in 0x10..0x1F
    fun isServerError(code: Int): Boolean = code in 0x20..0x2F
    fun isFederationError(code: Int): Boolean = code in 0x30..0x3F
    fun isSessionError(code: Int): Boolean = code in 0x40..0x4F
    fun isRetryable(code: Int): Boolean = code in setOf(
        SERVICE_UNAVAILABLE, TIMEOUT, OVERLOADED,
        NODE_UNREACHABLE, CLUSTER_PARTITION, QUORUM_UNAVAILABLE,
    )
}
