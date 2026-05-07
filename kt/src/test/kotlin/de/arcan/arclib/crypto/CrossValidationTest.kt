package de.arcan.arclib.crypto

import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import java.io.File

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class CrossValidationTest {
    private val gson = Gson()
    private val hashVectors: Map<String, Any>
    private val cryptoVectors: Map<String, Any>

    init {
        val baseDir = File(System.getProperty("user.dir")).parentFile.resolve("testdata")
        hashVectors = gson.fromJson(
            baseDir.resolve("hash_vectors.json").readText(),
            object : TypeToken<Map<String, Any>>() {}.type
        )
        cryptoVectors = gson.fromJson(
            baseDir.resolve("crypto_vectors.json").readText(),
            object : TypeToken<Map<String, Any>>() {}.type
        )
    }

    private fun fromHex(hex: String): ByteArray {
        if (hex.isEmpty()) return ByteArray(0)
        return hex.chunked(2).map { it.toInt(16).toByte() }.toByteArray()
    }

    private fun toHex(bytes: ByteArray): String {
        return bytes.joinToString("") { "%02x".format(it) }
    }

    @Test
    fun `SHA-256 cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val vectors = hashVectors["sha256"] as List<Map<String, String>>
        for (v in vectors) {
            val input = fromHex(v["hex_input"]!!)
            val expected = v["output"]!!
            assertEquals(expected, toHex(Hash.sha256(input)))
        }
    }

    @Test
    fun `CRC-16-IBM cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val section = hashVectors["crc16_ibm"] as Map<String, Any>
        @Suppress("UNCHECKED_CAST")
        val vectors = section["vectors"] as List<Map<String, String>>
        for (v in vectors) {
            val input = fromHex(v["hex_input"]!!)
            val expected = Integer.parseInt(v["output"]!!.removePrefix("0x"), 16)
            assertEquals(expected, Hash.crc16(input))
        }
    }

    @Test
    fun `CRC-32-IEEE cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val section = hashVectors["crc32_ieee"] as Map<String, Any>
        @Suppress("UNCHECKED_CAST")
        val vectors = section["vectors"] as List<Map<String, String>>
        for (v in vectors) {
            val input = fromHex(v["hex_input"]!!)
            val expected = java.lang.Long.parseLong(v["output"]!!.removePrefix("0x"), 16)
            assertEquals(expected, Hash.crc32(input))
        }
    }

    @Test
    fun `HMAC-SHA256 cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val section = hashVectors["hmac_sha256"] as Map<String, Any>
        @Suppress("UNCHECKED_CAST")
        val vectors = section["vectors"] as List<Map<String, String>>
        for (v in vectors) {
            val key = fromHex(v["key"]!!)
            val data = fromHex(v["data"]!!)
            val expected = v["output"]!!
            assertEquals(expected, toHex(Hmac.sha256(key, data)))
        }
    }

    @Test
    fun `ChaCha20-Poly1305 cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val vectors = cryptoVectors["chacha20_poly1305"] as List<Map<String, String>>
        for (v in vectors) {
            val key = fromHex(v["key"]!!)
            val nonce = fromHex(v["nonce"]!!)
            val plaintext = fromHex(v["plaintext"]!!)
            val aad = fromHex(v["aad"]!!)
            val expectedCt = v["ciphertext"]!!

            val encrypted = Aead.encrypt(key, nonce, plaintext, aad)
            assertEquals(expectedCt, toHex(encrypted))

            val decrypted = Aead.decrypt(key, nonce, fromHex(expectedCt), aad)
            assertArrayEquals(plaintext, decrypted)
        }
    }

    @Test
    fun `HKDF-SHA256 cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val vectors = cryptoVectors["hkdf_sha256"] as List<Map<String, Any>>
        for (v in vectors) {
            val ikm = fromHex(v["ikm"] as String)
            val salt = (v["salt"] as String).let { if (it.isEmpty()) null else fromHex(it) }
            val info = (v["info"] as String).let { if (it.isEmpty()) null else fromHex(it) }
            val length = (v["length"] as Double).toInt()
            val expected = v["okm"] as String

            val result = Hkdf.deriveKey(ikm, salt, info, length)
            assertEquals(expected, toHex(result))
        }
    }

    @Test
    fun `Key Rotation cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val vectors = cryptoVectors["key_rotation"] as List<Map<String, Any>>
        for (v in vectors) {
            val baseKey = fromHex(v["base_key"] as String)
            val counter = (v["counter"] as Double).toInt()
            val expected = v["derived_key"] as String

            val derived = KeyRotation.rotateKey(baseKey, counter)
            assertEquals(expected, toHex(derived))
        }
    }
}
