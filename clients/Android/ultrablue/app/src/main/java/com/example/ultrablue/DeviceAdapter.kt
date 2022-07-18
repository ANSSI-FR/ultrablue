package com.example.ultrablue

import android.annotation.SuppressLint
import android.view.*
import android.widget.ImageButton
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.example.ultrablue.database.Device

/*
    Allows to dispatch clicks on specific ViewHolder CardView, and handle them
    from the fragment that contains the recyclerview (DeviceListFragment).
 */
interface ItemClickListener {

    // Which section of the CardView has been clicked.
    enum class Target {
        ATTESTATION_BUTTON, TRASH_BUTTON, CARD_VIEW
    }

    fun onClick(id: Target, device: Device)
}

/*
    This Adapter manages a list of DeviceViewCards (defined in the res/layout folder).
    It derives from a RecyclerView Adapter for optimisation reasons.
 */
class DeviceAdapter(private val itemClickListener: ItemClickListener) : RecyclerView.Adapter<DeviceAdapter.ViewHolder>() {
    // The list of registered devices to display.
    private var deviceList = emptyList<Device>()

    class ViewHolder(ItemView: View) : RecyclerView.ViewHolder(ItemView) {
        var nameTextView: TextView = itemView.findViewById(R.id.device_name)
        var addrTextView: TextView = itemView.findViewById(R.id.device_addr)
        val attestationButton: ImageButton = itemView.findViewById(R.id.attestation_button)
        var trashButton: ImageButton = itemView.findViewById(R.id.trash_button)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.device_card_view, parent, false)
        return ViewHolder(view)
    }

    // Instantiate a specific DeviceViewCard.
    @SuppressLint("ClickableViewAccessibility")
    override fun onBindViewHolder(holder: ViewHolder, position: Int) {
        val device = deviceList[position]

        holder.nameTextView.text = device.name
        holder.addrTextView.text = device.addr

        holder.itemView.setOnClickListener {
            itemClickListener.onClick(ItemClickListener.Target.CARD_VIEW, device)
        }
        // Add a visual effect on each DeviceCardView tap, canceled on release.
        holder.itemView.setOnTouchListener {view, motionEvent ->
            val elevationDelta = 10
            when (motionEvent.action) {
                MotionEvent.ACTION_DOWN -> view.elevation -= elevationDelta
                MotionEvent.ACTION_UP -> {
                    view.elevation += elevationDelta
                    view.performClick()
                }
                MotionEvent.ACTION_CANCEL -> view.elevation += elevationDelta
            }
            true
        }

        holder.attestationButton.setOnClickListener {
            itemClickListener.onClick(ItemClickListener.Target.ATTESTATION_BUTTON, device)
        }
        // Change the button color on tap, and revert it on release.
        holder.attestationButton.setOnTouchListener {view, motionEvent ->
            when (motionEvent.action) {
                MotionEvent.ACTION_DOWN -> view.setBackgroundResource(R.drawable.ic_play_solid_selected)
                MotionEvent.ACTION_UP -> {
                    view.setBackgroundResource(R.drawable.ic_play_solid)
                    view.performClick()
                }
                MotionEvent.ACTION_CANCEL -> view.setBackgroundResource(R.drawable.ic_play_solid)
            }
            true
        }

        holder.trashButton.setOnClickListener {
            itemClickListener.onClick(ItemClickListener.Target.TRASH_BUTTON, device)
        }
        // Change the button color on tap, and revert it on release.
        holder.trashButton.setOnTouchListener {view, motionEvent ->
            when (motionEvent.action) {
                MotionEvent.ACTION_DOWN -> view.setBackgroundResource(R.drawable.ic_trash_solid_selected)
                MotionEvent.ACTION_UP -> {
                    view.setBackgroundResource(R.drawable.ic_trash_solid)
                    view.performClick()
                }
                MotionEvent.ACTION_CANCEL -> view.setBackgroundResource(R.drawable.ic_trash_solid)
            }
            true
        }
    }

    override fun getItemCount(): Int {
        return deviceList.size
    }

    @SuppressLint("NotifyDataSetChanged")
    fun setRegisteredDevices(devices: List<Device>) {
        this.deviceList = devices
        notifyDataSetChanged()
    }
}