package fr.gouv.ssi.ultrablue.model

import android.text.Html
import android.view.View
import android.widget.ScrollView
import android.widget.TextView
import fr.gouv.ssi.ultrablue.MainActivity

/*
    In this file, we implement four classes, allowing to display different kind
    of logs to the user, through a TextView.
 */

// The PLog class is used to represent the progress of something,
// with a progress bar of this style: "[======>           ] 32/88".
class PLog(private val total: Int, private var barLength: Int = 25): Log(msg = "") {
    private var state: Int = 0

    fun updateProgress(state: Int) {
        this.state = if (state < total) {
            state
        } else {
            total
        }
    }

    override fun toString(): String {
        val percent = state.toFloat() / total.toFloat()
        val filledLength = (barLength * percent).toInt()
        return "[" + "=".repeat(filledLength) + ">" + "&nbsp;".repeat(barLength - filledLength) + "] " + "$state/$total"
    }
}

// The CLog class instantiates a log for a completed action, that can be
// successful of not. Takes the form of: "[ok] Lorem ipsum dolor sit amet".
class CLog(private val msg: String, val success: Boolean, val fatal: Boolean = true): Log(msg = msg) {
    override fun toString(): String {
        return if (success) {
            "[<font color='green'>ok</font>]"
        } else {
            "[<font color='red'>ko</font>]"
        } + "&nbsp;" + msg
    }
}

// The Log class instantiates a log for a pending action.
// Takes the form of: "Lorem ipsum dolor sit amet..."
open class Log(private val msg: String) {
    override fun toString(): String {
        return "&nbsp;".repeat(5) + msg + "..."
    }
}

/*
    The logger keeps an internal list of logs, which can be any of the three types
    of logs. When updated, either by pushing a new log, or updating a progress log,
    It updates the TextView content on the UI thread, so the user can see it
    immediately.
    When an unsuccessful CLog is pushed, the onError callback is called.
 */
class Logger(private var activity: MainActivity?, private var textView: TextView, private var scrollView: ScrollView? = null, private var onError: () -> Unit) {
    private var logs = listOf<Log>()

    fun push(log: Log) {
        logs = logs + log
        updateUI()
        if (log is CLog && !log.success && log.fatal) {
            onError()
        }
    }

    fun update(progress: Int) {
        if (logs.last() is PLog) {
            (logs.last() as PLog).updateProgress(progress)
            updateUI()
        }
    }

    fun reset() {
        this.logs = listOf()
        updateUI()
    }

    /*
        This is where we update the user interface so that the
        user see the logger updates.
     */
    private fun updateUI() {
        activity?.runOnUiThread {
            textView.setText(
                Html.fromHtml(this.toString(), Html.FROM_HTML_MODE_COMPACT),
                TextView.BufferType.SPANNABLE
            )
            scrollView?.let {
                it.post { it.fullScroll(View.FOCUS_DOWN) }
            }
        }
    }

    fun setOnErrorHandler(handler: () -> Unit) {
        this.onError = handler
    }

    override fun toString(): String {
        return logs.joinToString(separator = "<br/>", transform = {
            it.toString()
        })
    }
}