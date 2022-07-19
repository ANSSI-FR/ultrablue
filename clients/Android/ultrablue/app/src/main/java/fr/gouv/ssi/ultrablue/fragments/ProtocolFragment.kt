package fr.gouv.ssi.ultrablue.fragments

import android.os.Bundle
import android.view.*
import androidx.fragment.app.Fragment
import fr.gouv.ssi.ultrablue.*
import fr.gouv.ssi.ultrablue.database.Device

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
        Member methods
     */

    private fun setFragmentTitle(state: State) {
        (requireActivity() as MainActivity).supportActionBar?.title = when (state) {
            State.AUTHENTICATION -> "Attestation in progress"
            State.ENROLLMENT -> "Enrollment in progress"
        }
    }

    /*
        Hook methods
     */

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        return inflater.inflate(R.layout.fragment_protocol, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        val device = requireArguments().getSerializable("device") as Device?
        val address = requireArguments().getSerializable("address") as String?

        state = if (device != null) { State.AUTHENTICATION } else { State.ENROLLMENT }
        logger = Logger(view.findViewById(R.id.logger_text_view))
        setFragmentTitle(state)

    }

    override fun onCreateOptionsMenu(menu: Menu, inflater: MenuInflater) {
        inflater.inflate(R.menu.action_bar, menu)
    }

    override fun onPrepareOptionsMenu(menu: Menu) {
        menu.findItem(R.id.action_add).isVisible = false
        menu.findItem(R.id.action_edit).isVisible = false
    }
}