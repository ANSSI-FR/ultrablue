package fr.gouv.ssi.ultrablue.fragments

import android.app.AlertDialog
import android.os.Bundle
import android.view.*
import android.widget.EditText
import android.widget.TextView
import androidx.fragment.app.Fragment
import fr.gouv.ssi.ultrablue.database.DeviceViewModel
import fr.gouv.ssi.ultrablue.MainActivity
import fr.gouv.ssi.ultrablue.R
import fr.gouv.ssi.ultrablue.database.Device

/*
    This fragment displays the details about a specific Device.
 */
class DeviceFragment : Fragment() {
    private var viewModel: DeviceViewModel? = null
    private var device: Device? = null

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        setHasOptionsMenu(true)
        device = requireArguments().getSerializable("device") as Device
        return inflater.inflate(R.layout.fragment_device, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        viewModel = (activity as MainActivity).viewModel
        (requireActivity() as MainActivity).supportActionBar?.title = device?.name
        displayDeviceInformations(view)
    }

    override fun onCreateOptionsMenu(menu: Menu, inflater: MenuInflater) {
        inflater.inflate(R.menu.action_bar, menu)
    }

    override fun onPrepareOptionsMenu(menu: Menu) {
        menu.findItem(R.id.action_add).isVisible = false
        super.onPrepareOptionsMenu(menu)
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
            // Displays an alert dialog, asking for the new device name.
            R.id.action_edit -> {
                val nameField = EditText(requireContext())
                nameField.hint = "name"
                nameField.width = 150
                nameField.setPadding(30, 30, 30, 30)
                val alertDialogBuilder = AlertDialog.Builder(activity)
                alertDialogBuilder
                    .setTitle(R.string.rename_device_dialog_title)
                    .setView(nameField)
                    .setPositiveButton("Ok") { _, _ ->
                        device?.let {
                            if (isNameValid(nameField.text.toString())) {
                                renameDevice(it, nameField.text.toString())
                            }
                        }
                    }
                    .setNegativeButton("Cancel", null)
                    .show()
                true
            }
            else -> super.onOptionsItemSelected(item)
        }
    }

    private fun displayDeviceInformations(view: View) {
        val tv: TextView = view.findViewById(R.id.addr_value)
        tv.text = "${device?.addr}"
    }

    private fun isNameValid(name: String) : Boolean {
        return name.length in 4..12
    }

    private fun renameDevice(dev: Device, name: String) {
        dev.name = name
        viewModel?.rename(dev, name)
        (requireActivity() as MainActivity).supportActionBar?.title = dev.name
    }
}