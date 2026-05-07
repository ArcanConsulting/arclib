package de.arcan.arclib.protocol

import java.nio.ByteBuffer
import java.nio.ByteOrder

/** Extension type identifiers. */
object ExtType {
    const val TERMINATOR: Int = 0x0000
    const val REPLY_TO: Int = 0x0001
    const val TARGET_SERVICE: Int = 0x0002
}

/** Parsed extension headers. */
data class Extensions(
    val replyTo: Long? = null,
    val targetService: String? = null,
    val raw: Map<Int, ByteArray> = emptyMap(),
) {
    /** Compute wire size of extensions. */
    fun size(): Int {
        var s = 0
        if (replyTo != null) s += 8       // type(2) + len(2) + value(4)
        if (targetService != null) s += 4 + targetService.toByteArray(Charsets.UTF_8).size
        return s
    }

    /** Marshal extensions to TLV wire format. Returns null if empty. */
    fun marshal(): ByteArray? {
        val sz = size()
        if (sz == 0) return null
        val buf = ByteBuffer.allocate(sz).order(ByteOrder.BIG_ENDIAN)

        if (replyTo != null) {
            buf.putShort(ExtType.REPLY_TO.toShort())
            buf.putShort(4.toShort())
            buf.putInt(replyTo.toInt())
        }
        if (targetService != null) {
            val bytes = targetService.toByteArray(Charsets.UTF_8)
            buf.putShort(ExtType.TARGET_SERVICE.toShort())
            buf.putShort(bytes.size.toShort())
            buf.put(bytes)
        }

        return buf.array()
    }

    companion object {
        /** Parse TLV extension chain. Returns (extensions, bytesRead). */
        fun parse(data: ByteArray): Pair<Extensions, Int> {
            val buf = ByteBuffer.wrap(data).order(ByteOrder.BIG_ENDIAN)
            var replyTo: Long? = null
            var targetService: String? = null
            val raw = mutableMapOf<Int, ByteArray>()

            while (buf.remaining() >= 4) {
                val type = buf.short.toInt() and 0xFFFF
                if (type == ExtType.TERMINATOR) {
                    buf.short // consume length field of terminator
                    break
                }
                val length = buf.short.toInt() and 0xFFFF
                if (buf.remaining() < length) throw IllegalArgumentException("message too short")

                val value = ByteArray(length)
                buf.get(value)
                raw[type] = value

                when (type) {
                    ExtType.REPLY_TO -> {
                        if (length == 4) {
                            replyTo = ByteBuffer.wrap(value).order(ByteOrder.BIG_ENDIAN).int.toLong() and 0xFFFFFFFFL
                        }
                    }
                    ExtType.TARGET_SERVICE -> {
                        if (length > 0) targetService = String(value, Charsets.UTF_8)
                    }
                }
            }

            return Pair(Extensions(replyTo, targetService, raw), buf.position())
        }
    }
}
