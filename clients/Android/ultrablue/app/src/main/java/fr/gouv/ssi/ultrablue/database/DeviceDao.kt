package fr.gouv.ssi.ultrablue.database

import androidx.lifecycle.LiveData
import androidx.room.*

/*
    Queries to manage device_table
 */
@Dao
interface DeviceDao {

    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun addDevice(device: Device)

    @Query("SELECT * FROM device_table")
    fun getAll(): LiveData<List<Device>>

    @Query("SELECT * FROM device_table WHERE uid=:uid")
    fun get(uid: Int): Device

    @Query("SELECT * FROM device_table WHERE addr=:addr")
    fun get(addr: String): Device

    @Query("UPDATE device_table SET name=:newName WHERE uid=:uid")
    fun setName(uid: Int, newName: String)

    @Delete
    suspend fun removeDevice(device: Device)
}
