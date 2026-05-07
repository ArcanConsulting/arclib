package de.arcan.arclib.protocol

import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import java.io.File

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class CrossValidationTest {
    private val gson = Gson()
    private val vectors: Map<String, Any>

    init {
        val baseDir = File(System.getProperty("user.dir")).parentFile.resolve("testdata")
        vectors = gson.fromJson(
            baseDir.resolve("protocol_vectors.json").readText(),
            object : TypeToken<Map<String, Any>>() {}.type,
        )
    }

    private fun fromHex(hex: String): ByteArray {
        if (hex.isEmpty()) return ByteArray(0)
        return hex.chunked(2).map { it.toInt(16).toByte() }.toByteArray()
    }

    private fun toHex(bytes: ByteArray): String =
        bytes.joinToString("") { "%02x".format(it) }

    // --- Header cross-validation ---

    @Test
    fun `header marshal cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val headers = vectors["headers"] as List<Map<String, Any>>
        for (v in headers) {
            val tier = (v["tier"] as Double).toInt()
            val opcode = (v["opcode"] as Double).toInt()
            val version = (v["version"] as Double).toInt()
            val compressed = v["compressed"] as Boolean
            val fragmented = v["fragmented"] as Boolean
            val hasExt = v["has_extensions"] as Boolean
            val sequence = ((v["sequence"] as? Double) ?: 0.0).toLong()
            val sessionIDHex = v["session_id_hex"] as? String ?: ""
            val sessionID = if (sessionIDHex.isNotEmpty()) java.lang.Long.parseUnsignedLong(sessionIDHex, 16) else 0L
            val timestampHex = v["timestamp_hex"] as? String ?: ""
            val timestamp = if (timestampHex.isNotEmpty()) java.lang.Long.parseUnsignedLong(timestampHex, 16) else 0L
            val nonceHex = v["nonce_hex"] as? String ?: ""
            val keyID = ((v["key_id"] as? Double) ?: 0.0).toLong()
            val ecdhHex = v["ecdh_hex"] as? String ?: ""
            val expectedHex = v["header_hex"] as String
            val expectedSize = (v["header_size"] as Double).toInt()

            val hdr = Header(
                version = version, tier = tier, opcode = opcode,
                compressed = compressed, fragmented = fragmented, hasExtensions = hasExt,
                sequence = sequence, sessionID = sessionID, timestamp = timestamp,
                nonce = if (nonceHex.isNotEmpty()) fromHex(nonceHex) else ByteArray(12),
                keyID = keyID,
                ecdhPublic = if (ecdhHex.isNotEmpty()) fromHex(ecdhHex) else ByteArray(32),
            )

            val buf = hdr.marshal()
            assertEquals(expectedHex, toHex(buf), "marshal ${v["name"]}")
            assertEquals(expectedSize, buf.size, "size ${v["name"]}")
        }
    }

    @Test
    fun `header unmarshal cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val headers = vectors["headers"] as List<Map<String, Any>>
        for (v in headers) {
            val data = fromHex(v["header_hex"] as String)
            val (hdr, _) = Header.unmarshal(data)

            assertEquals((v["version"] as Double).toInt(), hdr.version, "version ${v["name"]}")
            assertEquals((v["tier"] as Double).toInt(), hdr.tier, "tier ${v["name"]}")
            assertEquals((v["opcode"] as Double).toInt(), hdr.opcode, "opcode ${v["name"]}")
            assertEquals(v["compressed"] as Boolean, hdr.compressed, "compressed ${v["name"]}")
            assertEquals(v["fragmented"] as Boolean, hdr.fragmented, "fragmented ${v["name"]}")
            assertEquals(v["has_extensions"] as Boolean, hdr.hasExtensions, "hasExt ${v["name"]}")

            val expectedSeq = ((v["sequence"] as? Double) ?: 0.0).toLong()
            assertEquals(expectedSeq, hdr.sequence, "sequence ${v["name"]}")

            val expectedSessionHex = v["session_id_hex"] as? String
            if (!expectedSessionHex.isNullOrEmpty()) {
                assertEquals(java.lang.Long.parseUnsignedLong(expectedSessionHex, 16), hdr.sessionID, "sessionID ${v["name"]}")
            }
            val expectedTimestampHex = v["timestamp_hex"] as? String
            if (!expectedTimestampHex.isNullOrEmpty()) {
                assertEquals(java.lang.Long.parseUnsignedLong(expectedTimestampHex, 16), hdr.timestamp, "timestamp ${v["name"]}")
            }
            val nonceHex = v["nonce_hex"] as? String
            if (!nonceHex.isNullOrEmpty()) {
                assertEquals(nonceHex, toHex(hdr.nonce), "nonce ${v["name"]}")
            }
            val keyIDVal = (v["key_id"] as? Double)?.toLong()
            if (keyIDVal != null && keyIDVal != 0L) {
                assertEquals(keyIDVal, hdr.keyID, "keyID ${v["name"]}")
            }
            val ecdhHex = v["ecdh_hex"] as? String
            if (!ecdhHex.isNullOrEmpty()) {
                assertEquals(ecdhHex, toHex(hdr.ecdhPublic), "ecdh ${v["name"]}")
            }
        }
    }

    // --- Extension cross-validation ---

    @Test
    fun `extension parse cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val extensions = vectors["extensions"] as List<Map<String, Any>>
        for (v in extensions) {
            val data = fromHex(v["hex_data"] as String)
            val (ext, _) = Extensions.parse(data)

            val expectedReplyTo = (v["reply_to"] as? Double)?.toLong()
            if (expectedReplyTo != null) {
                assertEquals(expectedReplyTo, ext.replyTo, "replyTo ${v["name"]}")
            }
            val expectedSvc = v["target_service"] as? String
            if (!expectedSvc.isNullOrEmpty()) {
                assertEquals(expectedSvc, ext.targetService, "targetService ${v["name"]}")
            }
        }
    }

    @Test
    fun `extension marshal single cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val extensions = vectors["extensions"] as List<Map<String, Any>>
        for (v in extensions) {
            val replyTo = (v["reply_to"] as? Double)?.toLong()
            val targetService = v["target_service"] as? String
            // Only test single-extension vectors for byte-exact marshal
            if (replyTo != null && !targetService.isNullOrEmpty()) continue

            val ext = Extensions(
                replyTo = replyTo,
                targetService = if (targetService.isNullOrEmpty()) null else targetService,
            )
            val buf = ext.marshal()
            assertNotNull(buf, "marshal should produce bytes for ${v["name"]}")
            assertEquals(v["hex_data"] as String, toHex(buf!!), "marshal ${v["name"]}")
            assertEquals((v["size"] as Double).toInt(), buf.size, "size ${v["name"]}")
        }
    }

    // --- OpCode cross-validation ---

    @Test
    fun `opcode lookup cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val opcodes = vectors["opcodes"] as List<Map<String, Any>>
        for (v in opcodes) {
            val code = (v["code"] as Double).toInt()
            assertEquals(v["name"] as String, OpCode.lookup(code), "name $code")
            assertEquals((v["category"] as Double).toInt(), OpCode.category(code), "category $code")
            assertEquals((v["operation"] as Double).toInt(), OpCode.operation(code), "operation $code")
        }
    }

    // --- ErrorCode cross-validation ---

    @Test
    fun `error code classification cross-validation`() {
        @Suppress("UNCHECKED_CAST")
        val errorCodes = vectors["error_codes"] as List<Map<String, Any>>
        for (v in errorCodes) {
            val code = (v["code"] as Double).toInt()
            val name = v["name"] as String
            assertEquals(name, ErrorCode.lookup(code), "name $code")
            assertEquals(v["is_success"] as Boolean, ErrorCode.isSuccess(code), "isSuccess $name")
            assertEquals(v["is_client_error"] as Boolean, ErrorCode.isClientError(code), "isClient $name")
            assertEquals(v["is_server_error"] as Boolean, ErrorCode.isServerError(code), "isServer $name")
            assertEquals(v["is_federation_error"] as Boolean, ErrorCode.isFederationError(code), "isFed $name")
            assertEquals(v["is_session_error"] as Boolean, ErrorCode.isSessionError(code), "isSession $name")
            assertEquals(v["is_retryable"] as Boolean, ErrorCode.isRetryable(code), "isRetryable $name")
        }
    }
}
