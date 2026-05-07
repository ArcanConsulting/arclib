package de.arcan.arclib.crypto

import java.security.MessageDigest
import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

object Hmac {
    fun sha256(key: ByteArray, data: ByteArray): ByteArray {
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(SecretKeySpec(key, "HmacSHA256"))
        return mac.doFinal(data)
    }

    fun sha256Truncated(key: ByteArray, data: ByteArray, n: Int): ByteArray {
        require(n in 1..32) { "truncation must be 1-32 bytes" }
        return sha256(key, data).copyOfRange(0, n)
    }

    fun verifySha256(key: ByteArray, data: ByteArray, expectedMac: ByteArray): Boolean {
        val computed = sha256(key, data)
        return MessageDigest.isEqual(computed, expectedMac)
    }
}
