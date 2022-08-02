package fr.gouv.ssi.ultrablue.model

import fr.gouv.ssi.ultrablue.MainActivity
import fr.gouv.ssi.ultrablue.database.Device
import gomobile.Gomobile
import kotlinx.serialization.*
import kotlinx.serialization.cbor.ByteString
import kotlinx.serialization.cbor.Cbor
import java.security.SecureRandom

@Serializable
class EKModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString val N: ByteArray,
    val E: UInt,
    @ByteString val Cert: ByteArray,
)

@Serializable
class EncryptedCredentialModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString val Credential: ByteArray,
    @ByteString val Secret: ByteArray,
)

@Serializable
class ByteArrayModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString val Bytes: ByteArray,
)

typealias ProtocolStep = Int

const val REGISTRATION_READ = 0
const val REGISTRATION = 1
const val AUTHENTICATION_READ = 2
const val AUTHENTICATION = 3
const val AK_READ = 4
const val CREDENTIAL_ACTIVATION = 5
const val CREDENTIAL_ACTIVATION_READ = 6
const val CREDENTIAL_ACTIVATION_ASSERT = 7
const val ATTESTATION_SEND_NONCE = 8
const val ATTESTATION_READ = 9
const val ATTESTATION_PERFORM = 10

/*
    UltrablueProtocol is the class that drives the client - server communication.
    It implements the protocol through read/write handlers and a state machine.
    The read/write operations are handled by the caller, so that this class isn't
    aware of the communication stack. It just calls the provided read/write methods
    whenever it needs to, and gets the results back in the onRead/onWrite handlers, to
    update the state machine, and resume the protocol.
 */

class UltrablueProtocol(private val activity: MainActivity, private var device: Device, private val logger: Logger?, private val readMsg: (String) -> Unit, private var writeMsg: (String, ByteArray) -> Unit) {
    private var state: ProtocolStep = REGISTRATION_READ
    private var message = byteArrayOf()

    private val rand = SecureRandom()
    private var credentialActivationSecret: ByteArray? = null
    private var encodedAttestationKey: ByteArray? = null
    private val attestationNonce = ByteArray(16)

    init {
        resume()
    }

    @OptIn(ExperimentalSerializationApi::class)
    private fun resume() {
        when (state) {
            REGISTRATION_READ -> readMsg("EkPub and EkCert")
            REGISTRATION -> {
                val key = Cbor.decodeFromByteArray<EKModel>(message)
                device.name = "device" + device.uid
                device.ekn = key.N
                device.eke = key.E.toInt()
                device.ekcert = key.Cert
                logger?.push(Log("Registering device"))
                activity.viewModel.insert(device)
                logger?.push(CLog("Device has been registered", true))
                writeMsg("registration confirmation", byteArrayOf(-10)) // Registration success, CBOR encoded
            }
            AUTHENTICATION_READ -> readMsg("authentication nonce")
            AUTHENTICATION -> {
                val authNonce = Cbor.decodeFromByteArray<ByteArrayModel>(message)
                //TODO: When encryption will be implemented, we'll need to decrypt the nonce here.
                val encodedAuthNonce = Cbor.encodeToByteArray(authNonce)
                writeMsg("Nonce", encodedAuthNonce)
            }
            AK_READ -> readMsg("attestation key")
            CREDENTIAL_ACTIVATION -> {
                encodedAttestationKey = message
                logger?.push(Log("Generating credential challenge"))
                val credentialBlob = Gomobile.makeCredential(device.ekn, device.eke.toLong(), encodedAttestationKey)
                credentialActivationSecret = credentialBlob.secret
                val encryptedCredential = EncryptedCredentialModel(credentialBlob.cred, credentialBlob.credSecret)
                val encodedCredential = Cbor.encodeToByteArray(encryptedCredential)
                logger?.push(CLog("Credential generated", true))
                writeMsg("encrypted credential", encodedCredential)
            }
            CREDENTIAL_ACTIVATION_READ -> readMsg("decrypted credential")
            CREDENTIAL_ACTIVATION_ASSERT -> {
                val decryptedCredential = Cbor.decodeFromByteArray<ByteArrayModel>(message)
                logger?.push(Log("Comparing received credential"))
                if (decryptedCredential.Bytes.contentEquals(credentialActivationSecret)) {
                    logger?.push(CLog("Credential matches the generated one", true))
                    state += 1
                    resume()
                } else {
                    logger?.push(CLog("Credential doesn't match the generated one", false))
                }
            }
            ATTESTATION_SEND_NONCE -> {
                rand.nextBytes(attestationNonce)
                logger?.push(CLog("Generated anit replay nonce", true))
                val encoded = Cbor.encodeToByteArray(ByteArrayModel(attestationNonce))
                writeMsg("anti replay nonce", encoded)
            }
            ATTESTATION_READ -> readMsg("Attestation data")
            ATTESTATION_PERFORM -> {
                logger?.push(Log("Verifying quotes signature"))
                try {
                    Gomobile.checkQuotesSignature(message, encodedAttestationKey, attestationNonce)
                } catch (e: Exception) {
                    logger?.push(CLog("${e.message}", false))
                    return
                }
                logger?.push(CLog("Quotes signature are valid", true))
            }
        }
    }

    fun onMessageRead(message: ByteArray) {
        state += 1
        this.message = message
        resume()
    }

    fun onMessageWrite() {
        state += 1
        resume()
    }

}