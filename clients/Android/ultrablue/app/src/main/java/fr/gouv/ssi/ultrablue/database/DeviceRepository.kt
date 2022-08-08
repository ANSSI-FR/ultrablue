package fr.gouv.ssi.ultrablue.database

import androidx.lifecycle.LiveData

/*
    This class provides an abstraction layer to manage the device_table.
 */
class DeviceRepository(private val deviceDao: DeviceDao) {
    val allDevices: LiveData<List<Device>> = deviceDao.getAll()

    suspend fun insert(device: Device) {
        deviceDao.addDevice(device)
    }

    fun get(id: Int) : Device {
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

    fun setPCRs(device: Device, newPCRs: ByteArray) {
        deviceDao.setPCRs(device.uid, newPCRs)
    }
}
