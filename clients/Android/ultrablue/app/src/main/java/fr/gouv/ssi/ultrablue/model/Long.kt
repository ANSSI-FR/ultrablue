package fr.gouv.ssi.ultrablue.model

import java.sql.Timestamp
import java.text.SimpleDateFormat
import java.util.*

/*
    Converts a long (assuming it represents the timestamp since epoch), to a date string
    representation of the following format: "dd/MM/yyyy"
 */
fun Long.toDateFmt() : String {
    val date = Date(this)

    val format = SimpleDateFormat("dd/MM/yyyy")
    return format.format(date)
}

/*
    Converts a long (assuming it represents the timestamp since epoch), to a date string
    representation of the following format: "dd/MM/yyyy"
 */
fun Long.toDateTimeFmt() : String {
    val dateFormatter = SimpleDateFormat("dd/MM/yyyy")
    val timeFormatter = SimpleDateFormat("HH:mm")
    val date = Date.from(Timestamp(this).toInstant())
    return dateFormatter.format(date) + " at " + timeFormatter.format(date)
}