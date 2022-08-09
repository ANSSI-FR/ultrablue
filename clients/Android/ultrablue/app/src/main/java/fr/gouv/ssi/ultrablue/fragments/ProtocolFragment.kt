package fr.gouv.ssi.ultrablue.fragments

import android.Manifest
import android.annotation.SuppressLint
import android.app.Activity
import android.bluetooth.*
import android.bluetooth.le.ScanCallback
import android.bluetooth.le.ScanResult
import android.content.Context
import android.content.Intent
import android.net.MacAddress
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.*
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.view.MenuProvider
import androidx.fragment.app.Fragment
import fr.gouv.ssi.ultrablue.MainActivity
import fr.gouv.ssi.ultrablue.R
import fr.gouv.ssi.ultrablue.database.Device
import fr.gouv.ssi.ultrablue.model.*
import java.util.*

const val MTU = 512

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
    private var enroll = false
    private var logger: Logger? = null

    private var protocol: UltrablueProtocol? = null

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
    private var gatt: BluetoothGatt? = null

    private var onDeviceDisconnectionCallback: (() -> Unit) = {
        logger?.push(CLog("Disconnected from device", false))
    }

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
    private val permissionsRequest = registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { perms ->
        var granted = true
        for (perm in perms) {
            if (perm.value) {
                logger?.push(CLog("${perm.key.split('.')[2]} permission granted", true))
            } else {
                logger?.push(CLog("${perm.key.split('.')[2]} permission denied", false))
                granted = false
            }
        }
        if (granted) {
            onLocationPermissionGrantedCallback?.invoke()
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
                onDeviceDisconnectionCallback.invoke()
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
                protocol?.let {
                    val bytes = characteristic.value
                    if (dataLen == 0) {
                        dataLen = byteArrayToInt(bytes.take(4))
                        data = bytes.drop(4).toByteArray()
                        logger?.push(PLog(dataLen))
                    } else {
                        data += bytes
                    }
                    logger?.update(data.size)

                    if (data.size < dataLen) {
                        gatt.readCharacteristic(characteristic)
                    } else {
                        val msg = data
                        dataLen = 0
                        data = byteArrayOf()
                        it.onMessageRead(msg)
                    }
                }
            }
        }

        /*
            As messages the client sends will never be bigger than the MTU, we don't need to care
            about chunking them.
         */
        override fun onCharacteristicWrite(gatt: BluetoothGatt?, characteristic: BluetoothGattCharacteristic?, status: Int) {
            super.onCharacteristicWrite(gatt, characteristic, status)
            protocol?.let { proto ->
                characteristic?.let {
                    proto.onMessageWrite()
                }
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
                val device = requireArguments().getSerializable("device") as Device
                activity?.title = if (device.name.isEmpty()) {
                    "Enrollment in progress"
                } else {
                    "Attestation in progress"
                }
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

    @SuppressLint("MissingPermission")
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		val device = requireArguments().getSerializable("device") as Device
        if (device.name.isEmpty()) {
            enroll = true
        }
        logger = Logger(
            activity as MainActivity?,
            view.findViewById(R.id.logger_text_view),
            view.findViewById(R.id.logger_scroll_view),
            onError = {
                // TODO: Disconnect device
            }
        )

        /*
            As the following operations are not blocking, we let them run and give them
            a callback to call when completed.
            Briefly, those operations start Bluetooth, search for the attesting device,
            connect to it, and then start the Ultrablue Protocol.

            We deliberately break indentation to make this callback-based code read like
            direct style.
         */
        askForBluetoothPermissions(onSuccess = {
            val btAdapter = getBluetoothAdapter()
            turnBluetoothOn(btAdapter, onSuccess = {
            scanForDevice(btAdapter, MacAddress.fromString(device.addr), onDeviceFound = { btDevice ->
            connectToDevice(btDevice, onSuccess = { gatt ->
            this.gatt = gatt
            requestMTU(gatt, MTU, onSuccess = {
            searchForUltrablueService(gatt, onServiceFound = { service ->
            val chr = service.getCharacteristic(ultrablueChrUUID)
            protocol = UltrablueProtocol(
                (activity as MainActivity), enroll, device, logger,
                readMsg = { tag ->
                    logger?.push(Log("Getting $tag"))
                    gatt.readCharacteristic(chr)
                },
                writeMsg = { tag, msg ->
                    val prepended = intToByteArray(msg.size) + msg
                    if (prepended.size > MTU) {
                        logger?.push(
                            CLog("$tag doesn't fit in one packet: message size = ${prepended.size}", false)
                        )
                    } else {
                        logger?.push(Log("Sending $tag"))
                        chr.value = prepended
                        gatt.writeCharacteristic(chr)
                    }
                },
                onCompletion = { success ->
                    if (success) {
                        logger?.push(CLog("Attestation success", true))
                    } else {
                        logger?.push(CLog("Attestation failure", false))
                    }
                    logger?.push(Log("Closing connection"))
                    onDeviceDisconnectionCallback = {
                        logger?.push(CLog("Device disconnected", true))
                    }
                    gatt.disconnect()
                    if (success && (enroll || device.secret.isNotEmpty())) {
                        (activity as MainActivity).onSupportNavigateUp()
                    }
                }
            )
            protocol?.start()
            }) }) }) }) }) })
    }

    override fun onDestroyView() {
        super.onDestroyView()
        gatt?.disconnect()
        logger?.reset()
        (activity as MainActivity).hideUpButton()
    }

    /*
        The methods below implement each step of the initial setup.
        Most of them call asynchronous APIs and set up callbacks to plumb the results to.
     */

    private fun askForBluetoothPermissions(onSuccess: () -> Unit) {
        onLocationPermissionGrantedCallback = onSuccess
        var perms = arrayOf(Manifest.permission.ACCESS_FINE_LOCATION)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            perms += Manifest.permission.BLUETOOTH_CONNECT
            perms += Manifest.permission.BLUETOOTH_SCAN
        }
        permissionsRequest.launch(perms)
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
        var runnable: Runnable? = null

        val scanCallback = object: ScanCallback() {
            var deviceFound = false

            override fun onScanResult(callbackType: Int, result: ScanResult?) {
                super.onScanResult(callbackType, result)
                result?.let { scanResult ->
                    if (MacAddress.fromString(scanResult.device.address) == address && !deviceFound) {
                        deviceFound = true
                        logger?.push(CLog("Found attesting device", true))
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

        logger?.push(Log("Scanning for attesting device"))
        runnable = startTimer(timeoutPeriod, onTimeout = {
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

    /*
        Break an integer into it's little endian bytes representation.
        Useful to prepend the size of a message before sending.
     */
    private fun intToByteArray(value: Int): ByteArray {
        var n = value
        val bytes = ByteArray(4)
        for (i in 0..3) {
            bytes[i] = (n and 0xffff).toByte()
            n = n ushr 8
        }
        return bytes
    }
}