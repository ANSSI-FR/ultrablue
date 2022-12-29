package fr.gouv.ssi.ultrablue.model

import java.nio.ByteBuffer
import java.util.*

/*
    Converts and returns a byte array from an UUID
 */
fun UUID.toByteArray() : ByteArray {
    val bytes = ByteBuffer.wrap(ByteArray(16))
    bytes.putLong(this.mostSignificantBits)
    bytes.putLong(this.leastSignificantBits)
    return bytes.array()
}