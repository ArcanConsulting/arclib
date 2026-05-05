package de.arcan.arclib.crypto

import java.security.MessageDigest
import java.util.zip.CRC32 as JavaCRC32

object Hash {
    fun sha256(data: ByteArray): ByteArray {
        return MessageDigest.getInstance("SHA-256").digest(data)
    }

    fun crc16(data: ByteArray): Int {
        var crc = 0xFFFF
        for (b in data) {
            crc = crc xor (b.toInt() and 0xFF)
            repeat(8) {
                crc = if (crc and 1 != 0) {
                    (crc ushr 1) xor 0xA001
                } else {
                    crc ushr 1
                }
            }
        }
        return crc and 0xFFFF
    }

    fun crc32(data: ByteArray): Long {
        val crc = JavaCRC32()
        crc.update(data)
        return crc.value
    }
}
