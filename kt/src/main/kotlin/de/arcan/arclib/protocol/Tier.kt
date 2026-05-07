package de.arcan.arclib.protocol

/** Security tier levels (0-5). Higher tiers provide more security at the cost of overhead. */
enum class Tier(val level: Int) {
    PLAINTEXT(0),
    CHECKSUM(1),
    AUTHENTICATED(2),
    ENCRYPTED(3),
    PFS(4),
    MAX_SECURITY(5);

    val info: TierInfo get() = tierInfoTable[level]
    val tierName: String get() = info.name
    fun requiresEncryption(): Boolean = level >= ENCRYPTED.level
    fun requiresAuthentication(): Boolean = level >= AUTHENTICATED.level

    companion object {
        fun fromLevel(level: Int): Tier? = entries.find { it.level == level }
        fun isValid(level: Int): Boolean = level in 0..5
    }
}

data class TierInfo(
    val name: String,
    val headerOverhead: Int = 0,
    val trailerOverhead: Int = 0,
    val totalOverhead: Int = 0,
    val encrypted: Boolean = false,
    val authenticated: Boolean = false,
)

internal val tierInfoTable = arrayOf(
    TierInfo(name = "plaintext", totalOverhead = 1),
    TierInfo(name = "checksum", trailerOverhead = 2, totalOverhead = 3),
    TierInfo(name = "authenticated", headerOverhead = 4, trailerOverhead = 8, totalOverhead = 15, authenticated = true),
    TierInfo(name = "encrypted", headerOverhead = 12, trailerOverhead = 16, totalOverhead = 31, encrypted = true, authenticated = true),
    TierInfo(name = "pfs", headerOverhead = 48, trailerOverhead = 16, totalOverhead = 67, encrypted = true, authenticated = true),
    TierInfo(name = "max_security", headerOverhead = 112, trailerOverhead = 32, totalOverhead = 147, encrypted = true, authenticated = true),
)
