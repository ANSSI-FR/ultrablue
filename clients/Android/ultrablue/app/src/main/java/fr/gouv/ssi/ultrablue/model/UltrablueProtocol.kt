package fr.gouv.ssi.ultrablue.model

import fr.gouv.ssi.ultrablue.MainActivity
import fr.gouv.ssi.ultrablue.R
import fr.gouv.ssi.ultrablue.database.Device
import gomobile.Gomobile
import kotlinx.serialization.*
import kotlinx.serialization.cbor.ByteString
import kotlinx.serialization.cbor.Cbor
import java.security.SecureRandom
import java.security.Timestamp
import java.time.LocalDateTime
import java.time.ZoneOffset
import java.util.*

@Serializable
class RegistrationDataModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString val EKPub: ByteArray,
    val EKExp: UInt,
    @ByteString val EKCert: ByteArray,
    val PCRExtend: Boolean,
)

@Serializable
class EncryptedCredentialModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString val Credential: ByteArray,
    @ByteString val Secret: ByteArray,
)

@Serializable
class ByteArrayModel @OptIn(ExperimentalSerializationApi::class) constructor(
    @ByteString var Bytes: ByteArray,
)

@Serializable
class AttestationResponse @OptIn(ExperimentalSerializationApi::class) constructor(
    val Err: Boolean,
    @ByteString val Secret: ByteArray,
)

typealias ProtocolStep = Int

const val UUID_SEND = 1
const val ENABLE_ENCRYPTION = 2
const val AUTHENTICATION_READ = 3
const val AUTHENTICATION = 4
const val REGISTRATION_DATA_READ = 5
const val REGISTRATION_DATA_DECODE = 6
const val AK_READ = 7
const val CREDENTIAL_ACTIVATION = 8
const val CREDENTIAL_ACTIVATION_READ = 9
const val CREDENTIAL_ACTIVATION_ASSERT = 10
const val ATTESTATION_SEND_NONCE = 11
const val ATTESTATION_READ = 12
const val VERIFY_QUOTE = 13
const val REPLAY_EVENT_LOG = 14
const val PCRS_HANDLE = 15
const val RESPONSE = 16

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
    private var onCompletion: (Boolean) -> Unit,
    private var enableEncryption: () -> Unit
) {
    private var state: ProtocolStep = UUID_SEND
    private var message = byteArrayOf()

    private val rand = SecureRandom()
    private var registrationData = RegistrationDataModel(device.ekn, device.eke.toUInt(), device.ekcert, device.secret.isNotEmpty()) // If enrolling a device, this field is uninitialized, but will be after EK_READ.
    private var credentialActivationSecret: ByteArray? = null
    private var encodedAttestationKey: ByteArray? = null
    private var encodedPlatformParameters: ByteArray? = null
    private val attestationNonce = ByteArray(16)
    private var encodedPCRs: ByteArray? = null
    private var attestationResponse: AttestationResponse? = null


    fun start() {
        resume()
    }

    /*
        Some steps ends with writeMsg or readMsg. It so, the state is automatically updated, and the
        protocol resumed. If not, we need to set the state, and call resume() manually.
    */

    @OptIn(ExperimentalSerializationApi::class)
    private fun resume() {
        when (state) {
            UUID_SEND -> {
                val uuid = device.uid
                val encoded = Cbor.encodeToByteArray(ByteArrayModel(uuid.toByteArray()))
                writeMsg("UUID", encoded)
            }
            ENABLE_ENCRYPTION -> {
                enableEncryption()
                state++
                resume()
            }
            AUTHENTICATION_READ -> readMsg(activity.getString(R.string.auth_nonce))
            AUTHENTICATION -> {
                writeMsg(activity.getString(R.string.decrypted_auth_nonce), message)
            }
            REGISTRATION_DATA_READ -> {
                if (!enroll) {
                    state = AK_READ
                    resume()
                }
                readMsg(activity.getString(R.string.ek_pub_cert))
            }
            REGISTRATION_DATA_DECODE -> {
                registrationData = Cbor.decodeFromByteArray(message)
                state = AK_READ
                resume()
            }
            AK_READ -> readMsg(activity.getString(R.string.ak))
            CREDENTIAL_ACTIVATION -> {
                encodedAttestationKey = message
                logger?.push(Log("Generating credential challenge"))
                try {
                    val credentialBlob = Gomobile.makeCredential(registrationData.EKPub, registrationData.EKExp.toLong(), encodedAttestationKey)
                    // We store the secret now to validate it later.
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
                    state = ATTESTATION_SEND_NONCE
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
                    state = REPLAY_EVENT_LOG
                    resume()
                } catch (e: Exception) {
                    logger?.push(CLog("Error while verifying quote(s) signature: ${e.message}", false))
                }
            }
            REPLAY_EVENT_LOG -> {
                // From this point, an error is considered to be harmful: This isn't anymore a
                // connection error, a protocol error or whatever; It is likely that something
                // changes in the bootchain. We must make the attestation fail by sending an
                // attestation response showing the error.
                // This can be done by removing the previously set logger handler, which disconnected
                // the client in case of error, but also preventing it to get to the attestation
                // response message.
                logger?.setOnErrorHandler{}

                logger?.push(Log("Replaying event log"))
                try {
                    Gomobile.replayEventLog(encodedPlatformParameters)
                    logger?.push(CLog("Event log has been replayed", true))
                    state = PCRS_HANDLE
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
                if (!enroll) {
                    updateLastDeviceAttestation(device, !attestationResponse!!.Err)
                }
                val encodedResponse = Cbor.encodeToByteArray(attestationResponse)
                writeMsg(activity.getString(R.string.attestation_response), encodedResponse)
                onCompletion(attestationResponse?.Err == false)
            }
        }
    }

    private fun registerDevice() {
        val secret = if (registrationData.PCRExtend) {
            ByteArray(16)
        } else {
            byteArrayOf()
        }
        rand.nextBytes(secret)

        device.name = "device-" + device.uid.toString().split("-").first()
        device.ekn = registrationData.EKPub
        device.eke = registrationData.EKExp.toInt()
        device.ekcert = registrationData.EKCert
        device.encodedPCRs = encodedPCRs!!
        device.secret = secret
        logger?.push(Log("Registering device"))
        activity.viewModel.insert(device)
        logger?.push(CLog("Device has been registered", true))
    }

    private fun updateLastDeviceAttestation(device: Device, success: Boolean) {
        device.lastAttestationSuccess = success
        device.lastAttestation = System.currentTimeMillis()
        activity.viewModel.update(device)
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