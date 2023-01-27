package fr.gouv.ssi.ultrablue.database

import androidx.lifecycle.LiveData
import java.util.*

/*
    This class provides an abstraction layer to manage the device_table.
 */
class DeviceRepository(private val deviceDao: DeviceDao) {
    val allDevices: LiveData<List<Device>> = deviceDao.getAll()

    suspend fun insert(device: Device) {
        deviceDao.addDevice(device)
    }

    fun get(id: UUID) : Device {
        return deviceDao.get(id)
    }

    fun get(addr: String) : Device {
        return deviceDao.get(addr)
    }

    suspend fun delete(device: Device) {
        deviceDao.removeDevice(device)
    }

    fun setName(device: Device, newName: String) {
        deviceDao.setName(device.uid, newName)
    }

    fun update(device: Device) {
        deviceDao.update(device.uid, device.lastAttestation, device.lastAttestationSuccess)
    }

    // TODO: This function is meant to be used when values of some PCRs changed
    //  but we now it is the result of a trusted action. We want to allow the user
    //  to take the new PCR values as reference values.
    fun setPCRs(device: Device, newPCRs: ByteArray) {
        deviceDao.setPCRs(device.uid, newPCRs)
    }
}
