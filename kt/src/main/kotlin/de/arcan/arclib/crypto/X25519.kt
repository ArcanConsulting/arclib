package de.arcan.arclib.crypto

import java.security.KeyFactory
import java.security.KeyPairGenerator
import java.security.spec.NamedParameterSpec
import java.security.spec.XECPrivateKeySpec
import java.security.spec.XECPublicKeySpec
import java.math.BigInteger
import javax.crypto.KeyAgreement

object X25519 {
    const val KEY_SIZE = 32

    data class KeyPair(val privateKey: ByteArray, val publicKey: ByteArray)

    fun generateKeyPair(): KeyPair {
        val kpg = KeyPairGenerator.getInstance("X25519")
        val javaKp = kpg.generateKeyPair()

        val privSpec = KeyFactory.getInstance("XDH")
            .getKeySpec(javaKp.private, XECPrivateKeySpec::class.java)
        val pubSpec = KeyFactory.getInstance("XDH")
            .getKeySpec(javaKp.public, XECPublicKeySpec::class.java)

        val privBytes = privSpec.scalar.reversedArray().let { padOrTrim(it, KEY_SIZE) }
        val pubBytes = pubSpec.u.toByteArray().reversedArray().let { padOrTrim(it, KEY_SIZE) }

        return KeyPair(privBytes, pubBytes)
    }

    fun ecdh(privateKey: ByteArray, peerPublicKey: ByteArray): ByteArray {
        require(privateKey.size == KEY_SIZE) { "private key must be $KEY_SIZE bytes" }
        require(peerPublicKey.size == KEY_SIZE) { "public key must be $KEY_SIZE bytes" }

        val kf = KeyFactory.getInstance("XDH")
        val privSpec = XECPrivateKeySpec(NamedParameterSpec.X25519, privateKey.reversedArray())
        val privKey = kf.generatePrivate(privSpec)

        val u = BigInteger(1, peerPublicKey.reversedArray())
        val pubSpec = XECPublicKeySpec(NamedParameterSpec.X25519, u)
        val pubKey = kf.generatePublic(pubSpec)

        val ka = KeyAgreement.getInstance("X25519")
        ka.init(privKey)
        ka.doPhase(pubKey, true)

        val secret = ka.generateSecret()
        return padOrTrim(secret, KEY_SIZE)
    }

    private fun padOrTrim(bytes: ByteArray, size: Int): ByteArray {
        return when {
            bytes.size == size -> bytes
            bytes.size > size -> bytes.copyOfRange(bytes.size - size, bytes.size)
            else -> ByteArray(size - bytes.size) + bytes
        }
    }
}
