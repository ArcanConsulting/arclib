package de.arcan.arclib.protocol

import java.nio.ByteBuffer
import java.nio.ByteOrder

/** Flag bits in the flags byte. */
private const val FLAG_VERSION_MASK   = 0x03
private const val FLAG_TIER_MASK      = 0x1C
private const val FLAG_TIER_SHIFT     = 2
private const val FLAG_COMPRESSED     = 0x20
private const val FLAG_FRAGMENTED     = 0x40
private const val FLAG_HAS_EXTENSIONS = 0x80

/** Protocol message header. Structure varies based on security tier. */
data class Header(
    val version: Int = 0,
    val tier: Int = 0,
    val opcode: Int = 0,
    val compressed: Boolean = false,
    val fragmented: Boolean = false,
    val hasExtensions: Boolean = false,
    val sequence: Long = 0,
    val sessionID: Long = 0,   // unsigned 64-bit treated as signed Long
    val timestamp: Long = 0,
    val nonce: ByteArray = ByteArray(12),
    val keyID: Long = 0,
    val ecdhPublic: ByteArray = ByteArray(32),
) {
    /** Compute the header size in bytes. */
    fun headerSize(): Int {
        var size = 3 // flags(1) + opcode(2)
        if (fragmented) size += 8
        when (tier) {
            Tier.PLAINTEXT.level, Tier.CHECKSUM.level -> {}
            Tier.AUTHENTICATED.level -> size += 4
            Tier.ENCRYPTED.level -> size += 12
            Tier.PFS.level -> size += 60
            Tier.MAX_SECURITY.level -> size += 68
        }
        return size
    }

    /** Compute the trailer size in bytes. */
    fun trailerSize(): Int = when (tier) {
        Tier.PLAINTEXT.level -> 0
        Tier.CHECKSUM.level -> 2
        Tier.AUTHENTICATED.level -> 8
        Tier.ENCRYPTED.level, Tier.PFS.level -> 16
        Tier.MAX_SECURITY.level -> 32
        else -> 0
    }

    /** Marshal header to big-endian bytes. */
    fun marshal(): ByteArray {
        val size = headerSize()
        val buf = ByteBuffer.allocate(size).order(ByteOrder.BIG_ENDIAN)

        // Flags byte
        var flags = version and FLAG_VERSION_MASK
        flags = flags or (((tier and 0x07) shl FLAG_TIER_SHIFT) and FLAG_TIER_MASK)
        if (compressed) flags = flags or FLAG_COMPRESSED
        if (fragmented) flags = flags or FLAG_FRAGMENTED
        if (hasExtensions) flags = flags or FLAG_HAS_EXTENSIONS
        buf.put(flags.toByte())

        // OpCode (big-endian u16)
        buf.putShort(opcode.toShort())

        // Tier-specific fields
        when (tier) {
            Tier.AUTHENTICATED.level -> {
                buf.putInt(sequence.toInt())
            }
            Tier.ENCRYPTED.level -> {
                buf.putInt(sequence.toInt())
                buf.putLong(sessionID)
            }
            Tier.PFS.level -> {
                buf.putInt(sequence.toInt())
                buf.putLong(sessionID)
                buf.put(nonce, 0, 12)
                buf.putInt(keyID.toInt())
                buf.put(ecdhPublic, 0, 32)
            }
            Tier.MAX_SECURITY.level -> {
                buf.putInt(sequence.toInt())
                buf.putLong(sessionID)
                buf.putLong(timestamp)
                buf.put(nonce, 0, 12)
                buf.putInt(keyID.toInt())
                buf.put(ecdhPublic, 0, 32)
            }
        }

        return buf.array()
    }

    // Override equals/hashCode for ByteArray fields
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is Header) return false
        return version == other.version && tier == other.tier && opcode == other.opcode &&
                compressed == other.compressed && fragmented == other.fragmented &&
                hasExtensions == other.hasExtensions && sequence == other.sequence &&
                sessionID == other.sessionID && timestamp == other.timestamp &&
                nonce.contentEquals(other.nonce) && keyID == other.keyID &&
                ecdhPublic.contentEquals(other.ecdhPublic)
    }

    override fun hashCode(): Int {
        var result = version
        result = 31 * result + tier
        result = 31 * result + opcode
        result = 31 * result + sequence.hashCode()
        return result
    }

    companion object {
        /** Unmarshal header from big-endian bytes. Returns (header, bytesRead). */
        fun unmarshal(data: ByteArray): Pair<Header, Int> {
            if (data.size < 3) throw IllegalArgumentException("message too short")
            val buf = ByteBuffer.wrap(data).order(ByteOrder.BIG_ENDIAN)

            val flags = buf.get().toInt() and 0xFF
            val version = flags and FLAG_VERSION_MASK
            val tier = (flags and FLAG_TIER_MASK) shr FLAG_TIER_SHIFT
            val compressed = (flags and FLAG_COMPRESSED) != 0
            val fragmented = (flags and FLAG_FRAGMENTED) != 0
            val hasExtensions = (flags and FLAG_HAS_EXTENSIONS) != 0

            if (!Tier.isValid(tier)) throw IllegalArgumentException("invalid security tier")

            val opcode = buf.short.toInt() and 0xFFFF

            // Create partial header to compute expected size
            val partialSize = headerSizeStatic(tier, fragmented)
            if (data.size < partialSize) throw IllegalArgumentException("message too short")

            var sequence = 0L
            var sessionID = 0L
            var timestamp = 0L
            val nonce = ByteArray(12)
            var keyID = 0L
            val ecdhPublic = ByteArray(32)

            when (tier) {
                Tier.AUTHENTICATED.level -> {
                    sequence = buf.int.toLong() and 0xFFFFFFFFL
                }
                Tier.ENCRYPTED.level -> {
                    sequence = buf.int.toLong() and 0xFFFFFFFFL
                    sessionID = buf.long
                }
                Tier.PFS.level -> {
                    sequence = buf.int.toLong() and 0xFFFFFFFFL
                    sessionID = buf.long
                    buf.get(nonce)
                    keyID = buf.int.toLong() and 0xFFFFFFFFL
                    buf.get(ecdhPublic)
                }
                Tier.MAX_SECURITY.level -> {
                    sequence = buf.int.toLong() and 0xFFFFFFFFL
                    sessionID = buf.long
                    timestamp = buf.long
                    buf.get(nonce)
                    keyID = buf.int.toLong() and 0xFFFFFFFFL
                    buf.get(ecdhPublic)
                }
            }

            val header = Header(
                version = version, tier = tier, opcode = opcode,
                compressed = compressed, fragmented = fragmented, hasExtensions = hasExtensions,
                sequence = sequence, sessionID = sessionID, timestamp = timestamp,
                nonce = nonce, keyID = keyID, ecdhPublic = ecdhPublic,
            )
            return Pair(header, buf.position())
        }

        private fun headerSizeStatic(tier: Int, fragmented: Boolean): Int {
            var size = 3
            if (fragmented) size += 8
            when (tier) {
                Tier.PLAINTEXT.level, Tier.CHECKSUM.level -> {}
                Tier.AUTHENTICATED.level -> size += 4
                Tier.ENCRYPTED.level -> size += 12
                Tier.PFS.level -> size += 60
                Tier.MAX_SECURITY.level -> size += 68
            }
            return size
        }
    }
}
