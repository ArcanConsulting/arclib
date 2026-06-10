package de.arcan.arclib.msgpack

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.Assertions.assertTimeoutPreemptively
import org.junit.jupiter.api.Test
import java.time.Duration

/**
 * Regression tests for the MsgPack decoder DoS hardening (parser audit, 2026-05-30).
 *
 * A malicious MyClerk peer must not be able to (a) exhaust the JVM stack via deep nesting, nor
 * (b) force a multi-gigabyte allocation via a forged length prefix. Both classes manifest as a
 * java.lang.Error (StackOverflowError / OutOfMemoryError) that callers' catch(Exception) would NOT
 * contain, so each case must be turned into a plain rejectable exception. Each is wrapped in a
 * preemptive timeout because the un-hardened code would crash/hang rather than fail cleanly.
 */
class MsgPackDosTest {

    private val limit: Duration = Duration.ofSeconds(5)

    @Test
    fun deeplyNestedArraysAreRejectedNotStackOverflowed() {
        // 5000 x 0x91 (fixarray of length 1, each element another array). Unbounded recursion would
        // blow the stack; the depth limit must throw first.
        assertTimeoutPreemptively(limit) {
            assertThrows(IllegalArgumentException::class.java) {
                MsgPack.decode(ByteArray(5000) { 0x91.toByte() })
            }
        }
    }

    @Test
    fun bin32WithHugeLengthIsRejectedBeforeAllocating() {
        // 0xC6 = bin32, length 0x7FFFFFFF (~2 GB), no body follows.
        assertTimeoutPreemptively(limit) {
            assertThrows(IllegalArgumentException::class.java) {
                MsgPack.decode(byteArrayOf(0xC6.toByte(), 0x7F.toByte(), 0xFF.toByte(), 0xFF.toByte(), 0xFF.toByte()))
            }
        }
    }

    @Test
    fun array32WithHugeCountIsRejectedBeforeAllocating() {
        // 0xDD = array32, count 0x7FFFFFFF (~2.1 billion) — would pre-allocate ArrayList(count).
        assertTimeoutPreemptively(limit) {
            assertThrows(IllegalArgumentException::class.java) {
                MsgPack.decode(byteArrayOf(0xDD.toByte(), 0x7F.toByte(), 0xFF.toByte(), 0xFF.toByte(), 0xFF.toByte()))
            }
        }
    }

    @Test
    fun str32WithHugeLengthIsRejectedBeforeAllocating() {
        // 0xDB = str32, length 0x7FFFFFFF.
        assertTimeoutPreemptively(limit) {
            assertThrows(IllegalArgumentException::class.java) {
                MsgPack.decode(byteArrayOf(0xDB.toByte(), 0x7F.toByte(), 0xFF.toByte(), 0xFF.toByte(), 0xFF.toByte()))
            }
        }
    }

    @Test
    fun legitimateNestedStructureStillDecodes() {
        val encoded = MsgPack.encode(mapOf("a" to listOf(1, 2, 3), "b" to "hello"))
        val (decoded, _) = MsgPack.decode(encoded)
        @Suppress("UNCHECKED_CAST")
        val map = decoded as Map<Any?, Any?>
        assertEquals(listOf(1, 2, 3), map["a"])
        assertEquals("hello", map["b"])
    }
}
