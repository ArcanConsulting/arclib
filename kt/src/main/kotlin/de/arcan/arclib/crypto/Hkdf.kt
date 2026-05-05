package de.arcan.arclib.crypto

object Hkdf {
    private const val HASH_LEN = 32

    fun deriveKey(ikm: ByteArray, salt: ByteArray?, info: ByteArray?, length: Int): ByteArray {
        val prk = extract(salt ?: ByteArray(HASH_LEN), ikm)
        return expand(prk, info ?: ByteArray(0), length)
    }

    fun extract(salt: ByteArray, ikm: ByteArray): ByteArray {
        return Hmac.sha256(salt, ikm)
    }

    fun expand(prk: ByteArray, info: ByteArray, length: Int): ByteArray {
        val n = (length + HASH_LEN - 1) / HASH_LEN
        require(n <= 255) { "HKDF expand: output too large" }

        val output = ByteArray(length)
        var prev = ByteArray(0)
        var offset = 0

        for (i in 1..n) {
            val input = prev + info + byteArrayOf(i.toByte())
            prev = Hmac.sha256(prk, input)
            val take = minOf(HASH_LEN, length - offset)
            prev.copyInto(output, offset, 0, take)
            offset += take
        }
        return output
    }
}
