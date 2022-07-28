package fr.gouv.ssi.ultrablue.fragments

import android.app.AlertDialog
import android.os.Bundle
import android.view.*
import android.widget.EditText
import android.widget.TextView
import androidx.core.view.MenuProvider
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

    /*
        Fragment lifecycle methods:
     */

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        val menuHost = requireActivity()
        menuHost.addMenuProvider(object: MenuProvider {
            override fun onCreateMenu(menu: Menu, menuInflater: MenuInflater) {
                menuInflater.inflate(R.menu.action_bar, menu)
            }
            override fun onMenuItemSelected(item: MenuItem): Boolean {
                return when (item.itemId) {
                    // The + button has been clicked
                    R.id.action_edit -> {
                        showDeviceRenamingDialog()
                        true
                    }
                    else -> false
                }
            }
            override fun onPrepareMenu(menu: Menu) {
                super.onPrepareMenu(menu)
                activity?.title = "${device?.name}"
                menu.findItem(R.id.action_edit).isVisible = true
                menu.findItem(R.id.action_add).isVisible = false
            }
        })
        (activity as MainActivity).showUpButton()
        device = requireArguments().getSerializable("device") as Device
        return inflater.inflate(R.layout.fragment_device, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        viewModel = (activity as MainActivity).viewModel
        displayDeviceInformation(view)
    }

    override fun onDestroyView() {
        super.onDestroyView()
        (activity as MainActivity).hideUpButton()
    }

    // Sets the view components content with the device information.
    private fun displayDeviceInformation(view: View) {
        val tv: TextView = view.findViewById(R.id.addr_value)
        tv.text = "${device?.addr}"
    }

    // Present a popup allowing the user to rename the device.
    private fun showDeviceRenamingDialog() {
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
                        activity?.title = it.name
                    }
                }
            }
            .setNegativeButton("Cancel", null)
            .show()
    }

    /*
        Fragment methods
     */

    // Checking the validity of a device name.
    // Currently length based, but could be improved, e.g. by only accepting alphanumeric characters
    private fun isNameValid(name: String) : Boolean {
        return name.length in 4..12
    }

    // Change the device name in the database.
    private fun renameDevice(dev: Device, name: String) {
        dev.name = name
        viewModel?.rename(dev, name)
    }
}