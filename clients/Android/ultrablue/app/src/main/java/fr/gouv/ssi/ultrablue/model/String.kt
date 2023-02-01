package fr.gouv.ssi.ultrablue.model

/*
 * Converts a hex string to a byte array and returns it
 * NOTE: The string must be of even length, or the function will throw
 */
fun String.toByteArray(): ByteArray {
    return chunked(2)
        .map { it.toInt(16).toByte() }
        .toByteArray()
}