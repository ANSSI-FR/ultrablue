package fr.gouv.ssi.ultrablue.fragments

import android.Manifest
import android.annotation.SuppressLint
import android.app.Activity
import android.bluetooth.*
import android.bluetooth.le.ScanCallback
import android.bluetooth.le.ScanResult
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.MacAddress
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.*
import androidx.core.view.MenuProvider
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.app.ActivityCompat
import androidx.fragment.app.Fragment
import fr.gouv.ssi.ultrablue.*
import fr.gouv.ssi.ultrablue.database.Device
import fr.gouv.ssi.ultrablue.model.Logger
import fr.gouv.ssi.ultrablue.model.Log
import fr.gouv.ssi.ultrablue.model.CLog
import fr.gouv.ssi.ultrablue.model.PLog
import java.util.*

const val MTU = 512

enum class State {
    ENROLLMENT,
    AUTHENTICATION
}

val ultrablueSvcUUID: UUID = UUID.fromString("ebee1789-50b3-4943-8396-16c0b7231cad")
val ultrablueChrUUID: UUID = UUID.fromString("ebee1790-50b3-4943-8396-16c0b7231cad")

/*
    This is the fragment where the attestation actually happens.
    If the passed device has a name, we start a classic attestation for it.
    Else, it means it just has been created by the calling fragment, and thus
    we start the attestation in enroll mode.

    This file contains the Bluetooth handling code, including the scanning/connecting/services and
    characteristics discovering, and also the functions to read/write on the characteristics.
    The protocol itself is defined in the UltrablueProtocol.kt file.
 */
class ProtocolFragment : Fragment() {
    private var state = State.ENROLLMENT
    private var logger: Logger? = null

    // Timer related variables
    // Limit some tasks duration, to avoid waiting forever
    private var timer = Handler(Looper.getMainLooper())
    private val timeoutPeriod: Long = 3000 // 3 seconds

    // Data structure for packet reconstruction. Used in the onCharacteristicRead callback.
    private var data = ByteArray(0)
    private var dataLen: Int = 0

    // To enhance readability of this callback heavy code, it helps to name them, and assign them
    // later, at their logical point in the code (even if we could have implemented them as methods
    // instead).
    private var onBluetoothActivationCallback: (() -> Unit)? = null
    private var onLocationPermissionGrantedCallback: (() -> Unit)? = null
    private var onDeviceConnectionCallback: ((BluetoothGatt) -> Unit)? = null
    private var onMTUChangedCallback: (() -> Unit)? = null
    private var onServicesFoundCallback: (() -> Unit)? = null

    /*
        The following variables register activities that will be started later in the program.
        They must be declared during the fragment initialization, because of the Android API.
     */

    // Asks the user to enable bluetooth if it is not already turned on.
    private val bluetoothActivationRequest = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == Activity.RESULT_OK) {
            logger?.push(CLog("Bluetooth turned on", true))
            onBluetoothActivationCallback?.invoke()
        } else {
            logger?.push(CLog("Failed to turn Bluetooth on", false))
        }
    }

    // Asks the user to grant location permission if not already granted.
    // To enhance readability, we set onLocationPermissionGrantedCallback later.
    private val locationPermissionRequest = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted ->
        when (granted) {
            true -> {
                logger?.push(CLog("Location permission granted", true))
                onLocationPermissionGrantedCallback?.invoke()
            }
            false -> logger?.push(CLog("Location permission denied", false))
        }
    }

    /*
        Once we're connected to a device, this object's methods will be called back
        for every Bluetooth action, thus it implements all the logic we
        need such as read/write operations, connection updates, etc.
    */
    private val gattHandler = object : BluetoothGattCallback() {
        @SuppressLint("MissingPermission")
        override fun onConnectionStateChange(gatt: BluetoothGatt?, status: Int, newState: Int) {
            super.onConnectionStateChange(gatt, status, newState)
            if (newState == BluetoothProfile.STATE_CONNECTED) {
                logger?.push(CLog("Connected to device", true))
                onDeviceConnectionCallback?.invoke(gatt!!)
            } else if (newState == BluetoothProfile.STATE_DISCONNECTED) {
                logger?.push(CLog("Disconnected from device", false))
            }
        }

        override fun onMtuChanged(gatt: BluetoothGatt?, mtu: Int, status: Int) {
            super.onMtuChanged(gatt, mtu, status)
            if (status == BluetoothGatt.GATT_SUCCESS) {
                logger?.push(CLog("MTU has been updated", true))
                onMTUChangedCallback?.invoke()
            }
        }

        override fun onServicesDiscovered(gatt: BluetoothGatt?, status: Int) {
            super.onServicesDiscovered(gatt, status)
            if (status == BluetoothGatt.GATT_SUCCESS) {
                onServicesFoundCallback?.invoke()
            }
        }

        // As packets will arrive by chunks, we need to reconstruct them.
        // The size of the full message is encoded in little endian, and
        // is put at the start of the first packet, on 4 bytes.
        // When the full message is read, we call the protocol's onMessageRead
        // callback.
        @SuppressLint("MissingPermission")
        override fun onCharacteristicRead(gatt: BluetoothGatt, characteristic: BluetoothGattCharacteristic, status: Int) {
            super.onCharacteristicRead(gatt, characteristic, status)

            if (status == BluetoothGatt.GATT_SUCCESS) {
                val bytes = characteristic.value
                if (dataLen == 0 && bytes.size >= 4) {
                    dataLen = byteArrayToInt(bytes.take(4))
                    data = bytes.drop(4).toByteArray()
                    logger?.push(PLog(dataLen))
                } else if (data.isNotEmpty()) {
                    data += bytes
                } else {
                    logger?.push(CLog("Invalid packet: data length: ${dataLen}, data size: ${data.size}, received: ${bytes.size}", false))
                }
                logger?.update(PLog(data.size))

                if (data.size < dataLen) {
                    gatt.readCharacteristic(characteristic)
                } else {
                    dataLen = 0
                    data = byteArrayOf()
                    // Retrieve the message from the protocol
                }
            }
        }

        // The writing operation is much simpler than the read operation because
        // we never need to chunk.
        override fun onCharacteristicWrite(gatt: BluetoothGatt?, characteristic: BluetoothGattCharacteristic?, status: Int) {
            super.onCharacteristicWrite(gatt, characteristic, status)
            characteristic?.let {
                // handle write operation from the protocol
            }
        }
    }

    /*
        Fragment lifecycle methods:
     */

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        val menuHost = requireActivity()
        menuHost.addMenuProvider(object: MenuProvider {
            override fun onCreateMenu(menu: Menu, menuInflater: MenuInflater) {
                menuInflater.inflate(R.menu.action_bar, menu)
            }
            override fun onPrepareMenu(menu: Menu) {
                super.onPrepareMenu(menu)
                menu.findItem(R.id.action_edit).isVisible = false
                menu.findItem(R.id.action_add).isVisible = false
            }
            override fun onMenuItemSelected(menuItem: MenuItem): Boolean {
                return false
            }
        })
        (activity as MainActivity).showUpButton()
        return inflater.inflate(R.layout.fragment_protocol, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		val device = requireArguments().getSerializable("device") as Device
        if (device.name.isEmpty()) {
            state = State.ENROLLMENT
            activity?.title = "Enrollment in progress"
        } else {
            state = State.AUTHENTICATION
            activity?.title = "Attestation in progress"
        }
        logger = Logger(activity as MainActivity?, view.findViewById(R.id.logger_text_view), onError = {
            // TODO: Ask the user if he wants to inspect the logs or go back to the menu
        })

        if (!hasBluetoothPermission()) {
            return
        }

        /*
            As the following operations are not blocking, we let them run and give them
            a callback to call when completed.
            Briefly, those operations start Bluetooth, search for the attesting device,
            connect to it, and then start the Ultrablue Protocol.

            We deliberately break indentation to make this callback-based code read like
            direct style.
         */
        askForLocationPermission(onSuccess = {
        val btAdapter = getBluetoothAdapter()
        turnBluetoothOn(btAdapter, onSuccess = {
        scanForDevice(btAdapter, MacAddress.fromString(device.addr), onDeviceFound = { device ->
        connectToDevice(device, onSuccess = { gatt ->
        requestMTU(gatt, MTU, onSuccess = {
        searchForUltrablueService(gatt, onServiceFound = { service ->
            val chr = service.getCharacteristic(ultrablueChrUUID)
        }) }) }) }) }) })
    }

    override fun onDestroyView() {
        super.onDestroyView()
        (activity as MainActivity).hideUpButton()
    }

    /*
        The methods below implement each step of the initial setup.
        Most of them call asynchronous APIs and set up callbacks to plumb the results to.
     */

    private fun hasBluetoothPermission(): Boolean {
        val permission = if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S) {
            Manifest.permission.BLUETOOTH
        } else {
            Manifest.permission.BLUETOOTH_CONNECT
        }
        return ActivityCompat.checkSelfPermission(
            requireContext(),
            permission
        ) == PackageManager.PERMISSION_GRANTED
    }

    private fun askForLocationPermission(onSuccess: () -> Unit) {
        onLocationPermissionGrantedCallback = onSuccess
        locationPermissionRequest.launch(Manifest.permission.ACCESS_FINE_LOCATION)
    }

    private fun getBluetoothAdapter(): BluetoothAdapter {
        logger?.push(Log("Getting Bluetooth adapter"))
        val manager = requireContext().getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager
        val adapter = manager.adapter
        logger?.push(CLog("Got Bluetooth adapter", true))
        return adapter
    }

    private fun turnBluetoothOn(adapter: BluetoothAdapter, onSuccess: () -> Unit) {
        logger?.push(Log("Checking for Bluetooth"))
        if (adapter.isEnabled) {
            logger?.push(CLog("Bluetooth is on", true))
            onSuccess()
        } else {
            val intent = Intent(BluetoothAdapter.ACTION_REQUEST_ENABLE)
            logger?.push(Log("Bluetooth is off. Turning on"))
            onBluetoothActivationCallback = onSuccess
            bluetoothActivationRequest.launch(intent)
        }
    }

    @SuppressLint("MissingPermission")
    private fun scanForDevice(adapter: BluetoothAdapter, address: MacAddress, onDeviceFound: (device: BluetoothDevice) -> Unit) {
        var deviceFound = false
        var scanning = false

        val scanCallback = object: ScanCallback() {
            var runnable: Runnable? = null

            override fun onScanResult(callbackType: Int, result: ScanResult?) {
                super.onScanResult(callbackType, result)
                result?.let { scanResult ->
                    if (MacAddress.fromString(scanResult.device.address) == address && !deviceFound) {
                        deviceFound = true
                        logger?.push(CLog("Found attesting device", true))
                        scanning = false
                        adapter.bluetoothLeScanner.stopScan(this)
                        stopTimer(runnable!!)
                        onDeviceFound(scanResult.device)
                    }
                }
            }
            override fun onScanFailed(errorCode: Int) {
                super.onScanFailed(errorCode)
                logger?.push(CLog("Failed to scan", false))
            }
        }

        Handler(Looper.getMainLooper()).postDelayed({
            if (scanning && !deviceFound) {
                logger?.push(CLog("Device not found", false))
                scanning = false
                adapter.bluetoothLeScanner.stopScan(scanCallback)
            }
        }, timeoutPeriod)
        logger?.push(Log("Scanning for attesting device"))
        scanCallback.runnable = startTimer(timeoutPeriod, onTimeout = {
            adapter.bluetoothLeScanner.stopScan(scanCallback)
        })
        adapter.bluetoothLeScanner.startScan(scanCallback)
    }

    @SuppressLint("MissingPermission") // Permission has already been granted
    private fun connectToDevice(device: BluetoothDevice, onSuccess: (BluetoothGatt) -> Unit) {
        logger?.push(Log("Trying to connect"))
        val runnable = startTimer(timeoutPeriod, onTimeout = { /* let if fail, nothing more to do */})
        onDeviceConnectionCallback = { gatt ->
            stopTimer(runnable)
            onSuccess(gatt)
        }
        device.connectGatt(context, false, gattHandler, BluetoothDevice.TRANSPORT_LE)
    }

    @SuppressLint("MissingPermission")
    private fun requestMTU(gatt: BluetoothGatt, mtu: Int, onSuccess: () -> Unit) {
        logger?.push(Log("Request MTU of $mtu"))
        val runnable = startTimer(timeoutPeriod, onTimeout = { /* let if fail, nothing more to do */})
        onMTUChangedCallback = {
            stopTimer(runnable)
            onSuccess()
        }
        gatt.requestMtu(mtu)
    }

    @SuppressLint("MissingPermission")
    private fun searchForUltrablueService(gatt: BluetoothGatt, onServiceFound: (BluetoothGattService) -> Unit) {
        val runnable = startTimer(timeoutPeriod, onTimeout = { /* let if fail, nothing more to do */})
        onServicesFoundCallback = {
            val ultrablueSvc = gatt.getService(ultrablueSvcUUID)
            logger?.push(CLog("Found Ultrablue service", true))
            stopTimer(runnable)
            onServiceFound(ultrablueSvc)
        }
        logger?.push(Log("Searching for Ultrablue service"))
        gatt.discoverServices()
    }

    private fun startTimer(period: Long, onTimeout: () -> Unit) : Runnable {
        val timeoutCallback = Runnable {
            logger?.push(CLog("Timed out", false))
            onTimeout()
        }
        timer.postDelayed(timeoutCallback, period)
        return timeoutCallback
    }

    private fun stopTimer(runnable: Runnable) {
        timer.removeCallbacks(runnable)
    }

    /*
        Reconstructs an Int from a little endian array of bytes.
        In particular, when reading the first packet of a BLE message.
     */
    private fun byteArrayToInt(bytes: List<Byte>) : Int {
        var result = 0
        for (i in bytes.indices) {
            var n = bytes[i].toInt()
            if (n < 0) {
                n += 256
            }
            result = result or (n shl 8 * i)
        }
        return result
    }
}