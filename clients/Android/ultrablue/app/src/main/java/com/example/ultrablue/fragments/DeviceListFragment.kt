package com.example.ultrablue.fragments

import com.example.ultrablue.R
import android.os.Bundle
import android.util.Log
import android.view.*
import androidx.fragment.app.Fragment
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
        activity?.title = "Your devices"
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
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

    private fun onQRCodeScannerResult(result: QRResult) {
        Log.d("DEBUG", result.toString())
    }
}