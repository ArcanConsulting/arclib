package de.arcan.arclib.protocol

/** Protocol operation codes. Format: 8-bit category + 8-bit operation. */
object OpCode {
    // Core Session Operations
    const val NOP               = 0x0000
    const val KEEPALIVE         = 0x0001
    const val KEEPALIVE_ACK     = 0x0002
    const val SESSION_INIT      = 0x0003
    const val SESSION_ACK       = 0x0004
    const val SESSION_CLOSE     = 0x0005
    const val SESSION_CLOSE_ACK = 0x0006
    const val SESSION_RESUME    = 0x0007
    const val SESSION_RESUMED   = 0x0008

    // Key Management Operations
    const val KEY_EXCHANGE_INIT     = 0x0010
    const val KEY_EXCHANGE_RESPONSE = 0x0011
    const val KEY_EXCHANGE_COMPLETE = 0x0012
    const val SESSION_ROTATE        = 0x0016
    const val SESSION_REVOKE        = 0x0017

    // Error
    const val ERROR = 0x00FF

    private val names = mapOf(
        NOP to "NOP", KEEPALIVE to "KEEPALIVE", KEEPALIVE_ACK to "KEEPALIVE_ACK",
        SESSION_INIT to "SESSION_INIT", SESSION_ACK to "SESSION_ACK",
        SESSION_CLOSE to "SESSION_CLOSE", SESSION_CLOSE_ACK to "SESSION_CLOSE_ACK",
        SESSION_RESUME to "SESSION_RESUME", SESSION_RESUMED to "SESSION_RESUMED",
        KEY_EXCHANGE_INIT to "KEY_EXCHANGE_INIT", KEY_EXCHANGE_RESPONSE to "KEY_EXCHANGE_RESPONSE",
        KEY_EXCHANGE_COMPLETE to "KEY_EXCHANGE_COMPLETE",
        SESSION_ROTATE to "SESSION_ROTATE", SESSION_REVOKE to "SESSION_REVOKE",
        ERROR to "ERROR",
    )

    /** Look up the name for an opcode. Returns empty string if unknown. */
    fun lookup(code: Int): String = names[code] ?: ""

    /** Extract category byte (high byte). */
    fun category(code: Int): Int = (code shr 8) and 0xFF

    /** Extract operation byte (low byte). */
    fun operation(code: Int): Int = code and 0xFF
}
