package fr.gouv.ssi.ultrablue

import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.ViewModelProvider
import androidx.navigation.NavController
import androidx.navigation.fragment.NavHostFragment
import androidx.navigation.ui.setupActionBarWithNavController

class MainActivity : AppCompatActivity() {
    private lateinit var navController: NavController
    lateinit var viewModel: DeviceViewModel

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        viewModel = ViewModelProvider(this)[DeviceViewModel::class.java]
        setContentView(R.layout.activity_main)
        setUpActionBar()
    }

    // Replace the default action bar with the res/menu/action_bar.xml one
    private fun setUpActionBar() {
        val host = supportFragmentManager
            .findFragmentById(R.id.fragmentContainerView) as NavHostFragment? ?: return
        navController = host.navController
        setupActionBarWithNavController(navController)
    }

    // Handle the back button to pop the navigation stack
    override fun onSupportNavigateUp(): Boolean {
        return navController.navigateUp() || super.onSupportNavigateUp()
    }
}