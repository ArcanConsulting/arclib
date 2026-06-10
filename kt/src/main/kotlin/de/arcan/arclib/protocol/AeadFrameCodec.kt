package de.arcan.arclib.protocol

import de.arcan.arclib.crypto.Aead

/**
 * Tier-3 AEAD frame encoder/decoder for the MyClerk protocol.
 *
 * Wire layout: [Header 15B] [Ciphertext] [Poly1305 Tag 16B]
 * The header is used as AAD (authenticated but not encrypted).
 */
object AeadFrameCodec {

    private const val FLAGS_T3_AEAD: Byte = 0x8C.toByte() // tier=3 + hasExtensions bit7=1
    private const val HEADER_SIZE = 15
    private const val TAG_SIZE = 16

    data class DecodedFrame(val opcode: Int, val payloadBytes: ByteArray)

    fun encode(opcode: Int, plaintext: ByteArray, session: AeadSession): ByteArray {
        val seq = session.nextSendSeq()
        val header = buildHeader(opcode, seq, session.sessionId)
        val nonce = AeadSession.deriveNonce(session.keys.sendNonceSeed, seq)
        val ctWithTag = Aead.encrypt(session.keys.sendKey, nonce, plaintext, header)

        val frame = ByteArray(HEADER_SIZE + ctWithTag.size)
        System.arraycopy(header, 0, frame, 0, HEADER_SIZE)
        System.arraycopy(ctWithTag, 0, frame, HEADER_SIZE, ctWithTag.size)
        return frame
    }

    fun decode(frame: ByteArray, session: AeadSession): DecodedFrame {
        require(frame.size >= HEADER_SIZE + TAG_SIZE) { "Frame too short: ${frame.size}" }
        require(frame[0] == FLAGS_T3_AEAD) { "Not Tier-3 AEAD: 0x${(frame[0].toInt() and 0xFF).toString(16)}" }

        val opcode = ((frame[1].toInt() and 0xFF) shl 8) or (frame[2].toInt() and 0xFF)
        val seq = readSeq(frame)
        val sid = readSessionId(frame)

        require(sid == session.sessionId) { "Session ID mismatch: $sid != ${session.sessionId}" }
        if (session.lastRecvSeq >= 0 && seq <= session.lastRecvSeq) {
            throw IllegalStateException("Replayed sequence: $seq <= ${session.lastRecvSeq}")
        }
        session.lastRecvSeq = seq

        val aad = frame.copyOfRange(0, HEADER_SIZE)
        val ctWithTag = frame.copyOfRange(HEADER_SIZE, frame.size)
        val nonce = AeadSession.deriveNonce(session.keys.recvNonceSeed, seq.toInt())
        val plaintext = Aead.decrypt(session.keys.recvKey, nonce, ctWithTag, aad)

        return DecodedFrame(opcode, plaintext)
    }

    private fun buildHeader(opcode: Int, seq: Int, sessionId: Long): ByteArray {
        val h = ByteArray(HEADER_SIZE)
        h[0] = FLAGS_T3_AEAD
        h[1] = ((opcode shr 8) and 0xFF).toByte()
        h[2] = (opcode and 0xFF).toByte()
        h[3] = ((seq ushr 24) and 0xFF).toByte()
        h[4] = ((seq ushr 16) and 0xFF).toByte()
        h[5] = ((seq ushr 8) and 0xFF).toByte()
        h[6] = (seq and 0xFF).toByte()
        for (i in 0 until 8) {
            h[7 + i] = ((sessionId ushr ((7 - i) * 8)) and 0xFF).toByte()
        }
        return h
    }

    private fun readSeq(frame: ByteArray): Long =
        (((frame[3].toInt() and 0xFF).toLong() shl 24) or
         ((frame[4].toInt() and 0xFF).toLong() shl 16) or
         ((frame[5].toInt() and 0xFF).toLong() shl 8) or
          (frame[6].toInt() and 0xFF).toLong()) and 0xFFFFFFFFL

    private fun readSessionId(frame: ByteArray): Long {
        var sid = 0L
        for (i in 0 until 8) sid = (sid shl 8) or (frame[7 + i].toLong() and 0xFF)
        return sid
    }
}
