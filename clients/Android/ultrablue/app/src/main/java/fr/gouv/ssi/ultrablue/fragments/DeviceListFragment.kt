package fr.gouv.ssi.ultrablue.fragments

import android.app.AlertDialog
import android.os.Bundle
import android.view.*
import androidx.core.os.bundleOf
import androidx.core.view.MenuProvider
import androidx.fragment.app.Fragment
import androidx.lifecycle.Lifecycle
import androidx.navigation.NavHostController
import androidx.navigation.findNavController
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import fr.gouv.ssi.ultrablue.*
import fr.gouv.ssi.ultrablue.database.Device
import fr.gouv.ssi.ultrablue.database.DeviceViewModel
import io.github.g00fy2.quickie.QRResult
import io.github.g00fy2.quickie.ScanCustomCode
import io.github.g00fy2.quickie.config.BarcodeFormat
import io.github.g00fy2.quickie.config.ScannerConfig

/*
* This fragment displays a list of registered devices.
* */
class DeviceListFragment : Fragment(R.layout.fragment_device_list), ItemClickListener {
    private val scanner = registerForActivityResult(ScanCustomCode(), ::onQRCodeScannerResult)
    private var viewModel: DeviceViewModel? = null

    /*
        Fragment lifecycle methods:
     */

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        val menuHost = requireActivity()
        menuHost.addMenuProvider(object: MenuProvider {
            override fun onCreateMenu(menu: Menu, menuInflater: MenuInflater) {
                activity?.title = "Your devices"
                menuInflater.inflate(R.menu.action_bar, menu)
            }
            override fun onMenuItemSelected(item: MenuItem): Boolean {
                return when (item.itemId) {
                    // The + button has been clicked
                    R.id.action_add -> {
                        showQRCodeScanner()
                        true
                    }
                    else -> false
                }
            }
            override fun onPrepareMenu(menu: Menu) {
                super.onPrepareMenu(menu)
                menu.findItem(R.id.action_add).isVisible = true
                menu.findItem(R.id.action_edit).isVisible = false
            }
        })
        return inflater.inflate(R.layout.fragment_device_list, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        viewModel = (activity as MainActivity).viewModel
        setUpDeviceRecyclerView(view)
    }

    // Handle clicks on a specific device card.
    override fun onClick(id: ItemClickListener.Target, device: Device) {
        when (id) {
            ItemClickListener.Target.CARD_VIEW -> {
                val nc = activity?.findNavController(R.id.fragmentContainerView) as NavHostController
                val bundle = bundleOf("device" to device)
                nc.navigate(R.id.action_deviceListFragment_to_deviceFragment, bundle)
            }
            ItemClickListener.Target.ATTESTATION_BUTTON  -> {
                val nc = activity?.findNavController(R.id.fragmentContainerView) as NavHostController
                val bundle = bundleOf("device" to device)
                nc.navigate(R.id.action_deviceListFragment_to_protocolFragment, bundle)
            }
            ItemClickListener.Target.TRASH_BUTTON  -> {
                val alertDialogBuilder = AlertDialog.Builder(activity)
                alertDialogBuilder
                    .setTitle(R.string.delete_device_dialog_title)
                    .setMessage(getString(R.string.delete_device_dialog_body, device.name))
                    .setPositiveButton(R.string.delete_label) { _, _ ->
                        viewModel?.delete(device)
                    }
                    .setNegativeButton(R.string.cancel_label, null)
                    .show()
            }
        }
    }

    /*
        Fragment methods
     */

    private fun setUpDeviceRecyclerView(view: View) {
        val recyclerview = view.findViewById<RecyclerView>(R.id.recyclerview)
        val adapter = DeviceAdapter(this)
        recyclerview.layoutManager = LinearLayoutManager(requireContext())
        viewModel?.allDevices?.observe(viewLifecycleOwner) { items ->
            adapter.setRegisteredDevices(items)
        }
        recyclerview.adapter = adapter
    }

    private fun showQRCodeScanner() {
        scanner.launch(
            ScannerConfig.build {
                setBarcodeFormats(listOf(BarcodeFormat.FORMAT_QR_CODE))
                setOverlayStringRes(R.string.qrcode_scanner_subtitle)
            }
        )
    }

    /*
        When receiving QR code data, this function checks for potential
        errors, which can be:
            - Scanning error
            - Invalid received data
        If an error occurred, an alert is displayed.
        Otherwise, we navigate to the protocol fragment.
     */
    private fun onQRCodeScannerResult(result: QRResult) {
        when(result) {
            is QRResult.QRSuccess -> {
                if (isMACAddressValid(result.content.rawValue.trim())) {
                    val device = Device(0, "", result.content.rawValue.trim(), byteArrayOf(), 0, byteArrayOf())
                    val nc = activity?.findNavController(R.id.fragmentContainerView) as NavHostController
                    val bundle = bundleOf("device" to device)
                    nc.navigate(R.id.action_deviceListFragment_to_protocolFragment, bundle)
                } else {
                    showErrorPopup(getString(R.string.qrcode_error_invalid_title), getString(R.string.qrcode_error_invalid_message))
                }
            }
            is QRResult.QRError ->
                showErrorPopup(getString(R.string.qrcode_error_failure_title), getString(R.string.qrcode_error_failure_message))
            is QRResult.QRMissingPermission ->
                showErrorPopup(getString(R.string.qrcode_error_camera_permission_title), getString(R.string.qrcode_error_camera_permission_message))
            is QRResult.QRUserCanceled -> { }
        }
    }

    private fun showErrorPopup(title: String, message: String) {
        val alertDialogBuilder = AlertDialog.Builder(activity)
        alertDialogBuilder
            .setTitle(title)
            .setMessage(message)
            .setPositiveButton("Ok", null)
            .show()
    }
}