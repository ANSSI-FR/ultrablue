package com.example.ultrablue.fragments

import android.os.Bundle
import android.view.*
import androidx.fragment.app.Fragment
import com.example.ultrablue.MainActivity
import com.example.ultrablue.R
import com.example.ultrablue.database.Device

/*
    This fragment performs the Ultrablue attestation protocol.
    It takes an optional Device parameter. If no Device is passed,
    an enrollment is performed. Else, it means we already know about
    the device to attest, and jump straight to the attestation part.
 */
class ProtocolFragment : Fragment() {
    var device: Device? = null

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        setHasOptionsMenu(true)
        device = requireArguments().getSerializable("device") as Device?
        return inflater.inflate(R.layout.fragment_device_list, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        (requireActivity() as MainActivity).supportActionBar?.title = if (device != null) {
            "Attestation in progress"
        } else {
            "Enrollment in progress"
        }
    }

    override fun onCreateOptionsMenu(menu: Menu, inflater: MenuInflater) {
        inflater.inflate(R.menu.action_bar, menu)
    }

    override fun onPrepareOptionsMenu(menu: Menu) {
        menu.findItem(R.id.action_add).isVisible = false
        menu.findItem(R.id.action_edit).isVisible = false
        super.onPrepareOptionsMenu(menu)
    }
}