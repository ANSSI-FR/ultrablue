package fr.gouv.ssi.ultrablue.database

import androidx.room.*
import java.io.Serializable
import java.sql.Timestamp
import java.util.*

/*
    This table stores the registered devices.
 */
@Entity(tableName = "device_table")
data class Device (
    @PrimaryKey(autoGenerate = false)
    val uid: UUID, // UUID used to enroll the phone device to the server
    var name: String, // Device name, defaults to: "device-xxxxxxxx" where xxs are replaced with first uuid bytes
    var addr: String, // MAC address
    var ekn: ByteArray, // Public part of the Endorsement Key
    var eke: Int, // Exponent of the Endorsement Key
    var ekcert: ByteArray, // Raw certificate for the Endorsement Key
    var encodedPCRs: ByteArray, // PCRs we got at enrollment
    var secret: ByteArray, // Secret to send to the attester on attestation success, in order to extend a PCR
    var lastAttestation: Long, // Time of the last attestation (time since epoch) - Set to 0 after an enrollment
    var lastAttestationSuccess: Boolean // true for success, false otherwise
) : Serializable {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as Device

        if (uid != other.uid) return false

        return true
    }

    override fun hashCode(): Int {
        return uid.hashCode()
    }
}