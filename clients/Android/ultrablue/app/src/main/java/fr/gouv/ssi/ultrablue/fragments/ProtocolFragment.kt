package fr.gouv.ssi.ultrablue.fragments

import android.os.Bundle
import android.view.*
import androidx.core.view.MenuProvider
import androidx.fragment.app.Fragment
import fr.gouv.ssi.ultrablue.*
import fr.gouv.ssi.ultrablue.database.Device
import fr.gouv.ssi.ultrablue.model.Logger

enum class State {
    ENROLLMENT,
    AUTHENTICATION
}

/*
    This fragment performs the Ultrablue attestation protocol.
    It takes an optional Device parameter. If no Device is passed,
    an enrollment is performed. Else, it means we already know about
    the device to attest, and jump straight to the attestation part.
 */
class ProtocolFragment : Fragment() {
    private var state = State.ENROLLMENT
    private var logger: Logger? = null


    /*
        Hook methods
     */

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        val menuHost = requireActivity()
        menuHost.addMenuProvider(object: MenuProvider {
            override fun onCreateMenu(menu: Menu, menuInflater: MenuInflater) {
                menuInflater.inflate(R.menu.action_bar, menu)
            }
            override fun onMenuItemSelected(menuItem: MenuItem): Boolean {
                return false
            }
            override fun onPrepareMenu(menu: Menu) {
                super.onPrepareMenu(menu)
                menu.findItem(R.id.action_edit).isVisible = false
                menu.findItem(R.id.action_add).isVisible = false
            }
        })
        (activity as MainActivity).showUpButton()
        return inflater.inflate(R.layout.fragment_protocol, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
		val device = requireArguments().getSerializable("device") as Device
        if (device.name.isEmpty()) {
            state = State.ENROLLMENT
            activity?.title = "Enrollment in progress"
        } else {
            state = State.AUTHENTICATION
            activity?.title = "Attestation in progress"
        }
        logger = Logger(activity as MainActivity?, view.findViewById(R.id.logger_text_view), onError = {
            // TODO: Ask the user if he wants to inspect the logs or go back to the menu
        })
    }

    override fun onDestroyView() {
        super.onDestroyView()
        (activity as MainActivity).hideUpButton()
    }

}