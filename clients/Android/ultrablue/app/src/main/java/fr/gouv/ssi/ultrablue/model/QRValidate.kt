package fr.gouv.ssi.ultrablue.model

import kotlinx.serialization.*
import kotlinx.serialization.json.Json
import android.util.Log

@Serializable
data class QRObject(val addr: String, val key: String)

/*
 * Checks that the passed string is a well formatted MAC address.
 * It must have this format: "ff:ff:ff:ff:ff:ff", where 'ff'
 * represents a two digits hex value. The case doesn't matters.
 *
 * Returns true if the MAC address is valid, false otherwise.
 */
private fun isMACAddressValid(addr: String): Boolean {
    // 6 groups of 2 hex digits + 5 separator characters
    if (addr.length != 17) {
        Log.d("[isMACAddressValid]", "Invalid addr size")
        return false
    }
    // Check that once split, we have 6 groups
    val bytes = addr.split(":")
    if (bytes.size != 6) {
        Log.d("[isMACAddressValid]", "Wrong field count")
        return false
    }
    // Check that each groups is composed of two hex digit
    for (byteStr in bytes) {
        if (byteStr.length != 2 || byteStr.toIntOrNull(16) == null) {
            Log.d("[isMACAddressValid]", "Invalid byte sequence")
            return false
        }
    }
    return true
}

/*
 * The key is composed of 32 bytes, hex encoded. This function
 * checks that the given string is well formatted, thus is:
 * - Of len 64
 * - Only composed of hex digits
 *
 * Returns true if the key is valid, false otherwise.
 */
private fun isKeyValid(key: String): Boolean {
    if (key.length != 64) {
        Log.d("[isKeyValid]", "Invalid key size")
        return false
    }
    for (c in key) {
        if (c.digitToIntOrNull(16) == null) {
            Log.d("[isKeyValid]", "Invalid bytes sequence")
            return false
        }
    }
    return true
}

/*
 * QRValidate takes a string (obtained through a QR code scan), and
 * assert it matches the QRObject object, JSON encoded.
 *
 * Returns the decoded object on success, null otherwise.
 */
fun qrValidate(content: String): QRObject? {
    val obj = Json.decodeFromString<QRObject>(content)
    if (!isMACAddressValid(obj.addr) || !isKeyValid(obj.key)) {
        return null
    }
    return obj
}