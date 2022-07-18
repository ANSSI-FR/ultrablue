package fr.gouv.ssi.ultrablue

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.LiveData
import androidx.lifecycle.viewModelScope
import fr.gouv.ssi.ultrablue.database.AppDatabase
import fr.gouv.ssi.ultrablue.database.Device
import fr.gouv.ssi.ultrablue.database.DeviceRepository
import kotlinx.coroutines.launch

/*
    Implements a View Model to access and update the database
 */
class DeviceViewModel(application: Application): AndroidViewModel(application) {

    val repo: DeviceRepository
    val allDevices: LiveData<List<Device>>

    init {
        val deviceDao = AppDatabase.getDatabase(application).deviceDao()
        repo = DeviceRepository(deviceDao)
        allDevices = repo.allDevices
    }

    fun insert(device: Device) = viewModelScope.launch {
        repo.insert(device)
    }

    fun delete(device: Device) = viewModelScope.launch {
        repo.delete(device)
    }

    fun rename(device: Device, name: String) {
        repo.setName(device, name)
    }
}
