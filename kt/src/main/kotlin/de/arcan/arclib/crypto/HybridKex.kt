package de.arcan.arclib.crypto

import java.security.MessageDigest

/**
 * Hybrid key exchange (X25519 + ML-KEM-768) for MyClerk Tier-3.
 *
 * Derives 88 bytes via HKDF-SHA256:
 *   IKM = x25519_shared(32) || mlkem_shared(32)
 *   info = "myclerk-session-v1-hybrid" || SHA-256(init_payload || ack_payload)
 *   → [sendKey(32) | recvKey(32) | sendNonceSeed(12) | recvNonceSeed(12)]
 *
 * Responder (server) swaps send/recv directions.
 */
object HybridKex {

    private val HKDF_INFO_PREFIX = "myclerk-session-v1-hybrid".toByteArray(Charsets.US_ASCII)

    data class SessionKeys(
        val sendKey: ByteArray,
        val recvKey: ByteArray,
        val sendNonceSeed: ByteArray,
        val recvNonceSeed: ByteArray,
    )

    data class HandshakeResult(
        val sessionKeys: SessionKeys,
        val serverX25519Pub: ByteArray,
        val mlkemCiphertext: ByteArray,
        val transcriptHash: ByteArray,
    )

    data class ClientKeyPair(
        val x25519Pub: ByteArray,
        val x25519Priv: ByteArray,
        val mlkemPub: ByteArray,
        val mlkemPriv: ByteArray,
    )

    fun generateClientKeys(): ClientKeyPair {
        val xKp = X25519.generateKeyPair()
        val mKp = MlKem.generateKeyPair()
        return ClientKeyPair(xKp.publicKey, xKp.privateKey, mKp.publicKey, mKp.privateKey)
    }

    fun serverProcessInit(
        clientX25519Pub: ByteArray,
        clientMlkemPub: ByteArray,
        initPayloadBytes: ByteArray,
        ackPayloadAssembler: (serverX25519Pub: ByteArray, mlkemCiphertext: ByteArray) -> ByteArray,
    ): Pair<HandshakeResult, ByteArray> {
        require(clientX25519Pub.size == X25519.KEY_SIZE)
        require(clientMlkemPub.size == MlKem.PUBLIC_KEY_SIZE)

        val serverX = X25519.generateKeyPair()
        val sharedX = X25519.ecdh(serverX.privateKey, clientX25519Pub)

        val encap = MlKem.encapsulate(clientMlkemPub)
        val sharedM = encap.sharedSecret
        val mlkemCt = encap.ciphertext

        val ackPayloadBytes = ackPayloadAssembler(serverX.publicKey, mlkemCt)
        val transcript = transcriptHash(initPayloadBytes, ackPayloadBytes)
        val keys = deriveKeys(sharedX, sharedM, transcript, swap = true)

        sharedX.fill(0)
        sharedM.fill(0)

        return Pair(
            HandshakeResult(keys, serverX.publicKey, mlkemCt, transcript),
            ackPayloadBytes
        )
    }

    fun clientProcessAck(
        clientKeys: ClientKeyPair,
        serverX25519Pub: ByteArray,
        mlkemCiphertext: ByteArray,
        initPayloadBytes: ByteArray,
        ackPayloadBytes: ByteArray,
    ): Pair<SessionKeys, ByteArray> {
        val sharedX = X25519.ecdh(clientKeys.x25519Priv, serverX25519Pub)
        val sharedM = MlKem.decapsulate(clientKeys.mlkemPriv, mlkemCiphertext)

        val transcript = transcriptHash(initPayloadBytes, ackPayloadBytes)
        val keys = deriveKeys(sharedX, sharedM, transcript, swap = false)

        sharedX.fill(0)
        sharedM.fill(0)

        return Pair(keys, transcript)
    }

    private fun deriveKeys(sharedX: ByteArray, sharedM: ByteArray, transcript: ByteArray, swap: Boolean): SessionKeys {
        val info = ByteArray(HKDF_INFO_PREFIX.size + transcript.size)
        System.arraycopy(HKDF_INFO_PREFIX, 0, info, 0, HKDF_INFO_PREFIX.size)
        System.arraycopy(transcript, 0, info, HKDF_INFO_PREFIX.size, transcript.size)

        val ikm = ByteArray(sharedX.size + sharedM.size)
        System.arraycopy(sharedX, 0, ikm, 0, sharedX.size)
        System.arraycopy(sharedM, 0, ikm, sharedX.size, sharedM.size)

        val okm = Hkdf.deriveKey(ikm, null, info, 88)
        ikm.fill(0)

        val sendKey = okm.copyOfRange(0, 32)
        val recvKey = okm.copyOfRange(32, 64)
        val sendSeed = okm.copyOfRange(64, 76)
        val recvSeed = okm.copyOfRange(76, 88)
        okm.fill(0)

        return if (swap) {
            SessionKeys(sendKey = recvKey, recvKey = sendKey, sendNonceSeed = recvSeed, recvNonceSeed = sendSeed)
        } else {
            SessionKeys(sendKey = sendKey, recvKey = recvKey, sendNonceSeed = sendSeed, recvNonceSeed = recvSeed)
        }
    }

    private fun transcriptHash(init: ByteArray, ack: ByteArray): ByteArray {
        val sha256 = MessageDigest.getInstance("SHA-256")
        sha256.update(init)
        sha256.update(ack)
        return sha256.digest()
    }
}
