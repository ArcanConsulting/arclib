package de.arcan.arclib.crypto

import java.math.BigInteger
import java.security.KeyFactory
import java.security.KeyPairGenerator
import java.security.spec.NamedParameterSpec
import java.security.spec.XECPrivateKeySpec
import java.security.spec.XECPublicKeySpec
import javax.crypto.KeyAgreement

object X25519 {
    const val KEY_SIZE = 32

    data class KeyPair(val privateKey: ByteArray, val publicKey: ByteArray)

    fun generateKeyPair(): KeyPair {
        val kpg = KeyPairGenerator.getInstance("X25519")
        val javaKp = kpg.generateKeyPair()

        val kf = KeyFactory.getInstance("XDH")
        val privSpec = kf.getKeySpec(javaKp.private, XECPrivateKeySpec::class.java)
        val pubSpec = kf.getKeySpec(javaKp.public, XECPublicKeySpec::class.java)

        val privBytes = privSpec.scalar
        val pubBytes = uToLittleEndian(pubSpec.u)

        return KeyPair(privBytes, pubBytes)
    }

    fun ecdh(privateKey: ByteArray, peerPublicKey: ByteArray): ByteArray {
        require(privateKey.size == KEY_SIZE) { "private key must be $KEY_SIZE bytes" }
        require(peerPublicKey.size == KEY_SIZE) { "public key must be $KEY_SIZE bytes" }

        val kf = KeyFactory.getInstance("XDH")
        val privSpec = XECPrivateKeySpec(NamedParameterSpec.X25519, privateKey)
        val privKey = kf.generatePrivate(privSpec)

        val u = littleEndianToU(peerPublicKey)
        val pubSpec = XECPublicKeySpec(NamedParameterSpec.X25519, u)
        val pubKey = kf.generatePublic(pubSpec)

        val ka = KeyAgreement.getInstance("X25519")
        ka.init(privKey)
        ka.doPhase(pubKey, true)

        val secret = ka.generateSecret()
        return toLittleEndian32(secret)
    }

    private fun uToLittleEndian(u: BigInteger): ByteArray {
        val be = u.toByteArray()
        val le = ByteArray(KEY_SIZE)
        val start = if (be.size > KEY_SIZE) be.size - KEY_SIZE else 0
        val len = minOf(be.size, KEY_SIZE)
        for (i in 0 until len) {
            le[i] = be[be.size - 1 - i - start + start]
        }
        // Simple: reverse the big-endian bytes, trim/pad to 32
        val trimmed = be.dropWhile { it == 0.toByte() && be.size > KEY_SIZE }.toByteArray()
        val result = ByteArray(KEY_SIZE)
        for (i in trimmed.indices.reversed()) {
            val leIdx = trimmed.size - 1 - i
            if (leIdx < KEY_SIZE) result[leIdx] = trimmed[i]
        }
        return result
    }

    private fun littleEndianToU(le: ByteArray): BigInteger {
        val be = ByteArray(le.size)
        for (i in le.indices) {
            be[le.size - 1 - i] = le[i]
        }
        return BigInteger(1, be)
    }

    private fun toLittleEndian32(secret: ByteArray): ByteArray {
        // JDK KeyAgreement.generateSecret() for X25519 returns the raw
        // little-endian shared secret directly (no BigInteger conversion needed)
        return if (secret.size == KEY_SIZE) secret
        else if (secret.size > KEY_SIZE) secret.copyOf(KEY_SIZE)
        else ByteArray(KEY_SIZE).also { System.arraycopy(secret, 0, it, 0, secret.size) }
    }
}
