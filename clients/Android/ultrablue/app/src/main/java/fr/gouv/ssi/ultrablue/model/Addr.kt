package fr.gouv.ssi.ultrablue

// Checks that the passed string is a well formatted MAC address.
// It must have this format: "ff:ff:ff:ff:ff:ff", where 'ff'
// represents a two digits hex value. The case doesn't matters.
fun isMACAddressValid(data: String): Boolean {
    // 6 groups of 2 hex digits + 5 separator characters
    if (data.length != 17) {
        return false
    }
    // Check that once split, we have 6 groups
    val bytes = data.split(":")
    if (bytes.size != 6) {
        return false
    }
    // Check that each groups is composed of two hex digit
    for (byteStr in bytes) {
        if (byteStr.length != 2 || byteStr.toIntOrNull(16) == null) {
            return false
        }
    }
    return true
}