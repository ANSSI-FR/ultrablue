package com.example.ultrablue

/*
    ConnData stores the information the client application
    needs to connect to the Ultrablue server.
 */
class ConnData private constructor(data: String) {

    private val addr: String = data

    fun getAddr(): String {
        return this.addr
    }

    companion object {
        fun parse(data: String): ConnData? {
            return if (isAMACAddress(data)) {
                ConnData(data)
            } else {
                null
            }
        }
    }
}

fun isAMACAddress(data: String): Boolean {
    // TODO: Better validation
    return data.length == 18
}