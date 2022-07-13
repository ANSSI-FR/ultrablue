package com.example.ultrablue.fragments

import android.app.AlertDialog
import com.example.ultrablue.R
import android.os.Bundle
import android.util.Log
import android.view.*
import androidx.fragment.app.Fragment
import androidx.navigation.NavHostController
import androidx.navigation.findNavController
import com.example.ultrablue.ConnData
import io.github.g00fy2.quickie.QRResult
import io.github.g00fy2.quickie.ScanCustomCode
import io.github.g00fy2.quickie.config.BarcodeFormat
import io.github.g00fy2.quickie.config.ScannerConfig

/*
* This fragment displays a list of registered devices.
* */
class DeviceListFragment : Fragment() {
    private val scanner = registerForActivityResult(ScanCustomCode(), ::onQRCodeScannerResult)

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        setHasOptionsMenu(true)
        return inflater.inflate(R.layout.fragment_device_list, container, false)
    }

    override fun onCreateOptionsMenu(menu: Menu, inflater: MenuInflater) {
        inflater.inflate(R.menu.action_bar, menu)
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
            // The + button has been clicked
            R.id.action_add -> {
                showQRCodeScanner()
                true
            }
            else -> super.onOptionsItemSelected(item)
        }
    }

    private fun showQRCodeScanner() {
        scanner.launch(
            ScannerConfig.build {
                setBarcodeFormats(listOf(BarcodeFormat.FORMAT_QR_CODE))
                setOverlayStringRes(R.string.qrcode_scanner_subtitle)
            }
        )
    }

    private fun showErrorPopup(title: String, message: String) {
        val alertDialogBuilder = AlertDialog.Builder(activity)
        alertDialogBuilder
            .setTitle(title)
            .setMessage(message)
            .setPositiveButton("Ok", null)
            .show()
    }

    /*
        When receiving QR code data, this function chekcs for potential
        errors, which can be:
            - Scanning error
            - Invalid received data
        If an error occurred, an alert is displayed.
        Otherwise, we navigate to the protocol fragment.
     */
    private fun onQRCodeScannerResult(result: QRResult) {
        when(result) {
            is QRResult.QRSuccess -> {
                val cd: ConnData?
                cd = ConnData.parse(result.content.rawValue)
                if (cd == null) {
                    showErrorPopup("Invalid QR code", getString(R.string.qrcode_error_invalid_message))
                } else {
                    val nc = activity?.findNavController(R.id.fragmentContainerView) as NavHostController
                    nc.navigate(R.id.action_deviceListFragment_to_protocolFragment)
                }
            }
            is QRResult.QRError ->
                showErrorPopup("Invalid QR code", getString(R.string.qrcode_error_failure_message))
            is QRResult.QRMissingPermission ->
                showErrorPopup("Missing camera permission", getString(R.string.qrcode_error_camera_permission_message))
            is QRResult.QRUserCanceled -> { }
        }
    }
}