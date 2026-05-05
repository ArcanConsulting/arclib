package de.arcan.arclib.crypto

import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Assertions.assertNotEquals
import org.junit.jupiter.api.Test

class X25519Test {
    @Test
    fun `key pair generation produces valid keys`() {
        val kp = X25519.generateKeyPair()
        assert(kp.privateKey.size == 32)
        assert(kp.publicKey.size == 32)
        assert(!kp.privateKey.contentEquals(ByteArray(32)))
        assert(!kp.publicKey.contentEquals(ByteArray(32)))
    }

    @Test
    fun `ECDH shared secret agreement`() {
        val alice = X25519.generateKeyPair()
        val bob = X25519.generateKeyPair()

        val sharedAlice = X25519.ecdh(alice.privateKey, bob.publicKey)
        val sharedBob = X25519.ecdh(bob.privateKey, alice.publicKey)

        assertArrayEquals(sharedAlice, sharedBob)
        assert(!sharedAlice.contentEquals(ByteArray(32)))
    }

    @Test
    fun `different peers produce different secrets`() {
        val alice = X25519.generateKeyPair()
        val bob = X25519.generateKeyPair()
        val carol = X25519.generateKeyPair()

        val sharedAB = X25519.ecdh(alice.privateKey, bob.publicKey)
        val sharedAC = X25519.ecdh(alice.privateKey, carol.publicKey)

        assertNotEquals(sharedAB.toList(), sharedAC.toList())
    }

    @Test
    fun `ECDH is deterministic`() {
        val alice = X25519.generateKeyPair()
        val bob = X25519.generateKeyPair()

        val s1 = X25519.ecdh(alice.privateKey, bob.publicKey)
        val s2 = X25519.ecdh(alice.privateKey, bob.publicKey)

        assertArrayEquals(s1, s2)
    }

    @Test
    fun `full pipeline X25519 to HKDF to AEAD`() {
        val alice = X25519.generateKeyPair()
        val bob = X25519.generateKeyPair()

        val shared = X25519.ecdh(alice.privateKey, bob.publicKey)
        val key = Hkdf.deriveKey(shared, null, "test".toByteArray(), 32)

        val nonce = ByteArray(12).also { java.security.SecureRandom().nextBytes(it) }
        val plaintext = "post-quantum ready".toByteArray()

        val ct = Aead.encrypt(key, nonce, plaintext)
        val pt = Aead.decrypt(key, nonce, ct)

        assertArrayEquals(plaintext, pt)
    }
}
