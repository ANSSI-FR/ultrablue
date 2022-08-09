package fr.gouv.ssi.ultrablue.model

import fr.gouv.ssi.ultrablue.MainActivity
import fr.gouv.ssi.ultrablue.R
import fr.gouv.ssi.ultrablue.database.Device
import gomobile.Gomobile
import kotlinx.serialization.*
import kotlinx.serialization.cbor.ByteString
import kotlinx.serialization.cbor.Cbor
import java.security.SecureRandom

@Serializable
class RegistrationDataModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString val N: ByteArray,
    val E: UInt,
    @ByteString val Cert: ByteArray,
    val PCRExtend: Boolean,
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

@Serializable
class AttestationResponse @OptIn(ExperimentalSerializationApi::class) constructor(
    val Err: Boolean,
    @ByteString val Secret: ByteArray,
)

typealias ProtocolStep = Int

const val EK_READ = 0
const val EK_DECODE = 1
const val AUTHENTICATION_READ = 2
const val AUTHENTICATION = 3
const val AK_READ = 4
const val CREDENTIAL_ACTIVATION = 5
const val CREDENTIAL_ACTIVATION_READ = 6
const val CREDENTIAL_ACTIVATION_ASSERT = 7
const val ATTESTATION_SEND_NONCE = 8
const val ATTESTATION_READ = 9
const val VERIFY_QUOTE = 10
const val REPLAY_EVENT_LOG = 11
const val PCRS_HANDLE = 12
const val RESPONSE = 13

/*
    UltrablueProtocol is the class that drives the client - server communication.
    It implements the protocol through read/write handlers and a state machine.
    The read/write operations are handled by the caller, so that this class isn't
    aware of the communication stack. It just calls the provided read/write methods
    whenever it needs to, and gets the results back in the onRead/onWrite handlers, to
    update the state machine, and resume the protocol.
 */

class UltrablueProtocol(
    private val activity: MainActivity,
    private var enroll: Boolean = false,
    private var device: Device,
    private val logger: Logger?,
    private val readMsg: (String) -> Unit,
    private var writeMsg: (String, ByteArray) -> Unit,
    private var onCompletion: (Boolean) -> Unit
) {
    private var state: ProtocolStep = if (enroll) { EK_READ } else { AUTHENTICATION_READ }
    private var message = byteArrayOf()

    private val rand = SecureRandom()
    private var ek = RegistrationDataModel(device.ekn, device.eke.toUInt(), device.ekcert, device.secret.isNotEmpty()) // If enrolling a device, this field is uninitialized, but will be after EK_READ.
    private var credentialActivationSecret: ByteArray? = null
    private var encodedAttestationKey: ByteArray? = null
    private var encodedPlatformParameters: ByteArray? = null
    private val attestationNonce = ByteArray(16)
    private var encodedPCRs: ByteArray? = null
    private var attestationResponse: AttestationResponse? = null


    fun start() {
        resume()
    }

    @OptIn(ExperimentalSerializationApi::class)
    private fun resume() {
        when (state) {
            EK_READ -> readMsg(activity.getString(R.string.ek_pub_cert))
            EK_DECODE -> {
                ek = Cbor.decodeFromByteArray(message)
                state += 1
                resume()
            }
            AUTHENTICATION_READ -> readMsg(activity.getString(R.string.auth_nonce))
            AUTHENTICATION -> {
                if (message.size != 24) {
                    logger?.push(CLog("Invalid nonce length. Make sure you ran the attestation server without the --enroll flag.", false))
                    return
                }
                val authNonce = Cbor.decodeFromByteArray<ByteArrayModel>(message)
                //TODO: When encryption will be implemented, we'll need to decrypt the nonce here.
                val encodedAuthNonce = Cbor.encodeToByteArray(authNonce)
                writeMsg(activity.getString(R.string.decrypted_auth_nonce), encodedAuthNonce)
            }
            AK_READ -> readMsg(activity.getString(R.string.ak))
            CREDENTIAL_ACTIVATION -> {
                encodedAttestationKey = message
                logger?.push(Log("Generating credential challenge"))
                try {
                    val credentialBlob = Gomobile.makeCredential(ek.N, ek.E.toLong(), encodedAttestationKey)
                    credentialActivationSecret = credentialBlob.secret
                    val encryptedCredential = EncryptedCredentialModel(credentialBlob.cred, credentialBlob.credSecret)
                    val encodedCredential = Cbor.encodeToByteArray(encryptedCredential)
                    logger?.push(CLog("Credential generated", true))
                    writeMsg(activity.getString(R.string.encrypted_cred), encodedCredential)
                } catch (e: Exception) {
                    logger?.push(CLog("Failed to generate credential: ${e.message}", false))
                }
            }
            CREDENTIAL_ACTIVATION_READ -> readMsg(activity.getString(R.string.decrypted_cred))
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
                logger?.push(CLog("Generated anti replay nonce", true))
                val encoded = Cbor.encodeToByteArray(ByteArrayModel(attestationNonce))
                writeMsg("anti replay nonce", encoded)
            }
            ATTESTATION_READ -> readMsg(activity.getString(R.string.attestation_data))
            VERIFY_QUOTE -> {
                encodedPlatformParameters = message
                logger?.push(Log("Verifying quotes signature"))
                try {
                    Gomobile.checkQuotesSignature(encodedPlatformParameters, encodedAttestationKey, attestationNonce)
                    logger?.push(CLog("Quotes signature are valid", true))
                    state += 1
                    resume()
                } catch (e: Exception) {
                    logger?.push(CLog("Error while verifying quote(s) signature: ${e.message}", false))
                }
            }
            REPLAY_EVENT_LOG -> {
                logger?.push(Log("Replaying event log"))
                try {
                    Gomobile.replayEventLog(encodedPlatformParameters)
                    logger?.push(CLog("Event log has been replayed", true))
                    state += 1
                    resume()
                } catch (e: Exception) {
                    logger?.push(CLog("${e.message}", false))
                    attestationResponse = AttestationResponse(true, byteArrayOf())
                    state = RESPONSE
                    resume()
                }
            }
            PCRS_HANDLE -> {
                try {
                    logger?.push(Log("Getting PCRs"))
                    encodedPCRs = Gomobile.getPCRs(encodedPlatformParameters).data
                    logger?.push(CLog("Got PCRs", true))
                } catch (e: Exception) {
                    logger?.push(CLog("Error while getting PCRs: ${e.message}", false))
                    return
                }
                attestationResponse = if (enroll) {
                    logger?.push(Log("Storing new attester entry"))
                    registerDevice()
                    AttestationResponse(false, device.secret)
                } else {
                    logger?.push(Log("Comparing PCRs"))
                    if (device.encodedPCRs.contentEquals(encodedPCRs)) {
                        logger?.push(CLog("PCRs entries match the stored ones", true))
                        AttestationResponse(false, device.secret)
                    } else {
                        // TODO: Need deeper investigation to determine which PCR changed and why.
                        logger?.push(CLog("PCRs don't match", false))
                        AttestationResponse(true, byteArrayOf())
                    }
                }
                state = RESPONSE
                resume()
            }
            RESPONSE -> {
                val encodedResponse = Cbor.encodeToByteArray(attestationResponse)
                writeMsg(activity.getString(R.string.attestation_response), encodedResponse)
                onCompletion(attestationResponse?.Err == false)
            }
        }
    }

    private fun registerDevice() {
        val secret = if (ek.PCRExtend) {
            ByteArray(16)
        } else {
            byteArrayOf()
        }
        rand.nextBytes(secret)

        device.name = "device" + device.uid
        device.ekn = ek.N
        device.eke = ek.E.toInt()
        device.ekcert = ek.Cert
        device.encodedPCRs = encodedPCRs!!
        device.secret = secret
        logger?.push(Log("Registering device"))
        activity.viewModel.insert(device)
        logger?.push(CLog("Device has been registered", true))
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