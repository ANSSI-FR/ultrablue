package com.example.ultrablue.database

import androidx.room.*

/*
    This table stores the registered devices.
 */
@Entity(tableName = "device_table")
data class Device (
    @PrimaryKey(autoGenerate = true)
    val uid: Int,
    val name: String, // user-defined device name
    val addr: String, // MAC address
)