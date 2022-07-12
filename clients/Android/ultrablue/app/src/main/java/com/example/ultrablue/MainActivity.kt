package com.example.ultrablue

import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import com.example.ultrablue.fragments.DeviceListFragment

class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.fragment_device_list)
        supportFragmentManager.beginTransaction()
            .add(R.id.fragment_device_list, DeviceListFragment())
            .disallowAddToBackStack()
            .commit()
    }
}