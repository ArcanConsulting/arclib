package de.arcan.arclib.crypto

import java.nio.ByteBuffer

object KeyRotation {
    fun rotateKey(currentKey: ByteArray, counter: Int): ByteArray {
        val info = ByteBuffer.allocate(4).putInt(counter).array()
        return Hkdf.deriveKey(currentKey, "rotate".toByteArray(), info, 32)
    }

    fun rotateKeyN(currentKey: ByteArray, startCounter: Int, n: Int): ByteArray {
        var key = currentKey
        repeat(n) { i ->
            key = rotateKey(key, startCounter + i)
        }
        return key
    }
}
