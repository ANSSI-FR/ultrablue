package fr.gouv.ssi.ultrablue.model

import java.sql.Timestamp
import java.text.SimpleDateFormat
import java.util.*

/*
    Converts a long (assuming it represents the timestamp since epoch), to a date string
    representation of the following format: "dd/MM/yyyy"
 */
fun Long.toDateFmt() : String {
    val formatter = SimpleDateFormat("dd/MM/yyyy")
    val date = Date.from(Timestamp(this).toInstant())
    return formatter.format(date)
}