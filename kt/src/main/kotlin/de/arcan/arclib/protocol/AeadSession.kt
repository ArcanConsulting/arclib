package de.arcan.arclib.protocol

import de.arcan.arclib.crypto.HybridKex
import java.util.concurrent.atomic.AtomicInteger

/**
 * Per-connection AEAD session state: keys, directional sequence counters, nonce seeds.
 * Created after a successful hybrid key exchange handshake.
 */
class AeadSession(val sessionId: Long, val keys: HybridKex.SessionKeys) {
    init {
        require(keys.sendKey.size == 32 && keys.recvKey.size == 32)
        require(keys.sendNonceSeed.size == 12 && keys.recvNonceSeed.size == 12)
    }

    private val sendSeqCounter = AtomicInteger(0)
    @Volatile var lastRecvSeq: Long = -1

    fun nextSendSeq(): Int = sendSeqCounter.getAndIncrement()

    companion object {
        fun deriveNonce(seed: ByteArray, seq: Int): ByteArray {
            val n = seed.copyOf()
            n[0] = (n[0].toInt() xor ((seq ushr 24) and 0xFF)).toByte()
            n[1] = (n[1].toInt() xor ((seq ushr 16) and 0xFF)).toByte()
            n[2] = (n[2].toInt() xor ((seq ushr 8) and 0xFF)).toByte()
            n[3] = (n[3].toInt() xor (seq and 0xFF)).toByte()
            return n
        }
    }
}
