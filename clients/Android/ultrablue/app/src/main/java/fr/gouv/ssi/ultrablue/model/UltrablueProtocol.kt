package fr.gouv.ssi.ultrablue.model

import fr.gouv.ssi.ultrablue.MainActivity
import fr.gouv.ssi.ultrablue.database.Device
import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.Serializable
import kotlinx.serialization.cbor.ByteString
import kotlinx.serialization.cbor.Cbor
import kotlinx.serialization.decodeFromByteArray

@Serializable
class EKModel(@ByteString val N: ByteArray, val E: UInt, @ByteString val Cert: ByteArray)

typealias ProtocolStep = Int

const val REGISTRATION_READ = 0
const val REGISTRATION = 1
const val AUTHENTICATION_READ = 2
const val AUTHENTICATION = 3
const val AK_READ = 4
const val CREDENTIAL_ACTIVATION = 5
const val CREDENTIAL_ACTIVATION_ASSERT = 6

/*
    UltrablueProtocol is the class that drives the client - server communication.
    It implements the protocol through read/write handlers and a state machine.
    The read/write operations are handled by the caller, so that this class isn't
    aware of the communication stack. It just calls the provided read/write methods
    whenever it needs to, and gets the results back in the onRead/onWrite handlers, to
    update the state machine, and resume the protocol.
 */

class UltrablueProtocol(private val activity: MainActivity, private var device: Device, private val readMsg: (String) -> Unit, private var writeMsg: (String, ByteArray) -> Unit) {
    private var state: ProtocolStep = REGISTRATION_READ
    private var message = byteArrayOf()

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
                activity.viewModel.insert(device)
                writeMsg("registration confirmation", byteArrayOf(-10)) // Registration success, CBOR encoded
            }
            AUTHENTICATION_READ -> readMsg("authentication nonce")
            AUTHENTICATION -> {
                // When encryption will be implemented, we'll first need to decrypt the nonce here.
                writeMsg("Nonce", message)
            }
            AK_READ -> readMsg("attestation key")
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