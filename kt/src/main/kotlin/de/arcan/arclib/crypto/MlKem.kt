package de.arcan.arclib.crypto

/**
 * ML-KEM-768 post-quantum key encapsulation mechanism.
 * arclib defines the interface; implementations are injected by the consuming project
 * (e.g. BouncyCastle on Android, liboqs on server).
 */
object MlKem {
    const val PUBLIC_KEY_SIZE = 1184
    const val CIPHERTEXT_SIZE = 1088
    const val SHARED_SECRET_SIZE = 32

    data class KeyPair(val publicKey: ByteArray, val privateKey: ByteArray)
    data class Encapsulation(val ciphertext: ByteArray, val sharedSecret: ByteArray)

    interface Provider {
        fun generateKeyPair(): KeyPair
        fun encapsulate(publicKey: ByteArray): Encapsulation
        fun decapsulate(privateKey: ByteArray, ciphertext: ByteArray): ByteArray
    }

    @Volatile
    var provider: Provider? = null

    fun generateKeyPair(): KeyPair {
        val p = provider ?: throw IllegalStateException("MlKem.provider not set — call MlKem.provider = BouncyCastleMlKemProvider() first")
        return p.generateKeyPair()
    }

    fun encapsulate(publicKey: ByteArray): Encapsulation {
        require(publicKey.size == PUBLIC_KEY_SIZE) { "ML-KEM-768 public key must be $PUBLIC_KEY_SIZE bytes, got ${publicKey.size}" }
        val p = provider ?: throw IllegalStateException("MlKem.provider not set")
        return p.encapsulate(publicKey)
    }

    fun decapsulate(privateKey: ByteArray, ciphertext: ByteArray): ByteArray {
        require(ciphertext.size == CIPHERTEXT_SIZE) { "ML-KEM-768 ciphertext must be $CIPHERTEXT_SIZE bytes, got ${ciphertext.size}" }
        val p = provider ?: throw IllegalStateException("MlKem.provider not set")
        return p.decapsulate(privateKey, ciphertext)
    }
}
