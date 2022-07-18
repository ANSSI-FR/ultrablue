package com.example.ultrablue.fragments

import android.os.Bundle
import android.view.*
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.navigation.findNavController
import com.example.ultrablue.MainActivity
import com.example.ultrablue.R
import com.example.ultrablue.database.Device

/*
    This fragment displays the details about a specific Device.
 */
class DeviceFragment : Fragment() {
    private var device: Device? = null

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        setHasOptionsMenu(true)
        device = requireArguments().getSerializable("device") as Device
        return inflater.inflate(R.layout.fragment_device, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        (requireActivity() as MainActivity).supportActionBar?.title = device?.name
    }

    override fun onCreateOptionsMenu(menu: Menu, inflater: MenuInflater) {
        inflater.inflate(R.menu.action_bar, menu)
    }

    override fun onPrepareOptionsMenu(menu: Menu) {
        menu.findItem(R.id.action_add).isVisible = false
        super.onPrepareOptionsMenu(menu)
    }
}