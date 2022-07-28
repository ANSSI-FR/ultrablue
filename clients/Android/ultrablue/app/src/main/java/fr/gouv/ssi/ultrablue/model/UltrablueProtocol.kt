package fr.gouv.ssi.ultrablue.model

import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.Serializable
import kotlinx.serialization.cbor.ByteString
import kotlinx.serialization.cbor.Cbor
import kotlinx.serialization.decodeFromByteArray

typealias ProtocolStep = Int

const val REGISTRATION_READ = 0
const val REGISTRATION = 1
const val AUTHENTICATION_GET_NONCE = 2
const val AUTHENTICATION_SEND_NONCE_BACK = 3

/*
    UltrablueProtocol is the class that drives the client - server communication.
    It implements the protocol through read/write handlers and a state machine.
    The read/write operations are handled by the caller, so that this class isn't
    aware of the communication stack. It just calls the provided read/write methods
    whenever it needs to, and gets the results back in the onRead/onWrite handlers, to
    update the state machine, and resume the protocol.
 */

class UltrablueProtocol(private val readMsg: (String) -> Unit, private var writeMsg: (String, ByteArray) -> Unit) {
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
                Cbor.decodeFromByteArray<Int>(message)
                writeMsg("registration confirmation", byteArrayOf(0x01, 0x00, 0x00, 0x00, -10)) // Registration success, CBOR encoded
            }
            AUTHENTICATION_GET_NONCE -> {
                // Read nonce
            }
            AUTHENTICATION_SEND_NONCE_BACK -> {
                // Send back the decrypted nonce (well, its currently not encrypted. That makes things easier.)
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