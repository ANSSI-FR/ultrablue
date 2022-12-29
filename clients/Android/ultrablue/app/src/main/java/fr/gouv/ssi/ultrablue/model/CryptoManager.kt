package fr.gouv.ssi.ultrablue.model

import android.security.keystore.KeyProperties
import android.security.keystore.KeyProtection
import java.security.KeyStore
import java.security.KeyStore.SecretKeyEntry
import javax.crypto.Cipher
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.IvParameterSpec

class CryptoManager {
    private val keyStore = KeyStore.getInstance("AndroidKeyStore").apply {
        this.load(null)
    }
    private var key: SecretKeyEntry

    constructor(key: SecretKey) {
        this.key = SecretKeyEntry(key)
    }

    constructor(alias: String) {
        this.key = keyStore.getEntry(alias, null) as SecretKeyEntry
    }

    fun persistKey(alias: String) {
        keyStore.setEntry(
            alias,
            key,
            KeyProtection.Builder(PURPOSES)
                .setBlockModes(BLOCK_MODE)
                .setEncryptionPaddings(PADDING)
                .setRandomizedEncryptionRequired(true)
                .build()
        )
    }

    fun encrypt(data: ByteArray) : ByteArray {
        val cipher = Cipher.getInstance(TRANSFORMATION).apply {
            this.init(Cipher.ENCRYPT_MODE, key.secretKey)
        }
        return cipher.iv + cipher.doFinal(data)
    }

    fun decrypt(data: ByteArray) : ByteArray {
        val cipher = Cipher.getInstance(TRANSFORMATION).apply {
            this.init(Cipher.DECRYPT_MODE, key.secretKey, GCMParameterSpec(128, data.take(12).toByteArray()))
        }
        return cipher.doFinal(data.drop(12).toByteArray())
    }

    companion object {
        private const val ALGORITHM = KeyProperties.KEY_ALGORITHM_AES
        private const val BLOCK_MODE = KeyProperties.BLOCK_MODE_GCM
        private const val PADDING = KeyProperties.ENCRYPTION_PADDING_NONE
        private const val TRANSFORMATION = "$ALGORITHM/$BLOCK_MODE/$PADDING"
        private const val PURPOSES = KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
    }

}