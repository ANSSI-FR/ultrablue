package fr.gouv.ssi.ultrablue

fun isMACAddressValid(data: String): Boolean {
    val trimmed = data.trim()
    // 6 groups of 2 hex digits + 5 separator characters
    if (trimmed.length != 17) {
        return false
    }
    // Check that once split, we have 6 groups
    val bytes = trimmed.split(":")
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