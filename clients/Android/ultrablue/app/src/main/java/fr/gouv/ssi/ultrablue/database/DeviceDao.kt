package fr.gouv.ssi.ultrablue.database

import androidx.lifecycle.LiveData
import androidx.room.*
import java.util.*

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
    fun get(uid: UUID): Device

    @Query("SELECT * FROM device_table WHERE addr=:addr")
    fun get(addr: String): Device

    @Query("UPDATE device_table SET name=:newName WHERE uid=:uid")
    fun setName(uid: UUID, newName: String)

    @Query("UPDATE device_table SET encodedPCRs=:newPCRs WHERE uid=:uid")
    fun setPCRs(uid: UUID, newPCRs: ByteArray)

    @Query("UPDATE device_table SET lastAttestation=:ts, lastAttestationSuccess=:success WHERE uid=:uid")
    fun update(uid: UUID, ts: Long, success: Boolean)

    @Delete
    suspend fun removeDevice(device: Device)
}
