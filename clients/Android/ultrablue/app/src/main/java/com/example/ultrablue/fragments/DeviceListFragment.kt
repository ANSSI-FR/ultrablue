package com.example.ultrablue.fragments

import com.example.ultrablue.R
import android.os.Bundle
import android.util.Log
import android.view.*
import androidx.fragment.app.Fragment

/*
* This fragment displays a list of registered devices.
* */
class DeviceListFragment : Fragment() {
    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        return inflater.inflate(R.layout.fragment_device_list, container, false)
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setHasOptionsMenu(true)
    }

    override fun onCreateOptionsMenu(menu: Menu, inflater: MenuInflater) {
        inflater.inflate(R.menu.action_bar, menu)
        activity?.title = "Your devices"
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
            R.id.action_add -> {
                Log.d("DEBUG", "Hello from action add")
                true
            }
            else -> super.onOptionsItemSelected(item)
        }
    }
}