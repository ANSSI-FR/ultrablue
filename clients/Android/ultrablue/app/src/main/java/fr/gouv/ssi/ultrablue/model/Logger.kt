package fr.gouv.ssi.ultrablue.model

import android.text.Html
import android.widget.TextView
import fr.gouv.ssi.ultrablue.MainActivity

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
        return "[" + "=".repeat(filledLength) + ">" + "&nbsp;".repeat(barLength - filledLength) + "]"
    }
}

class CLog(private val msg: String, val success: Boolean, val fatal: Boolean = true): Log(msg = msg) {
    override fun toString(): String {
        return if (success) {
            "[<font color='green'>ok</font>]"
        } else {
            "[<font color='red'>ko</font>]"
        } + "&nbsp;" + msg
    }
}

open class Log(private val msg: String) {
    override fun toString(): String {
        return "&nbsp;".repeat(5) + msg + "..."
    }
}

class Logger(private var activity: MainActivity?, private var textView: TextView, private var onError: () -> Unit) {
    private var logs = listOf<Log>()

    fun push(log: Log) {
        logs = logs + log
        activity?.runOnUiThread {
            textView.setText(
                Html.fromHtml(this.toString(), Html.FROM_HTML_MODE_COMPACT),
                TextView.BufferType.SPANNABLE
            )
        }
        if (log is CLog && !log.success && log.fatal) {
            onError()
        }
    }

    fun update(log: PLog) {
        if (logs.last() is PLog) {
            logs = logs.dropLast(1) + log
            activity?.runOnUiThread {
                textView.setText(
                    Html.fromHtml(this.toString(), Html.FROM_HTML_MODE_COMPACT),
                    TextView.BufferType.SPANNABLE
                )
            }
        }
    }

    fun last(): Log {
        return logs.last()
    }

    override fun toString(): String {
        return logs.joinToString(separator = "<br/>", transform = {
            it.toString()
        })
    }
}