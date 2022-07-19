package fr.gouv.ssi.ultrablue

import android.text.Html
import android.widget.TextView

class PLog(private val total: Int, private var barLength: Int = 25): Log(msg = "") {
    var state: Int = 0

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

class CLog(private val msg: String, private val success: Boolean): Log(msg = msg) {
    override fun toString(): String {
        return if (success) {
            "[<font color='green'>ok</font>]"
        } else {
            "[<font color='green'>ko</font>]"
        } + "&nbsp;" + msg
    }
}

open class Log(private val msg: String) {
    override fun toString(): String {
        return "&nbsp;".repeat(5) + msg
    }
}

class Logger(private var textView: TextView) {
    private var logs = listOf<Log>()

    fun push(log: Log) {
        logs = logs + log
        textView.setText(
            Html.fromHtml(this.toString(), Html.FROM_HTML_MODE_COMPACT),
            TextView.BufferType.SPANNABLE
        )
    }

    fun update(log: PLog) {
        if (logs.last() is PLog) {
            logs = logs.dropLast(1) + log
            textView.setText(
                Html.fromHtml(this.toString(), Html.FROM_HTML_MODE_COMPACT),
                TextView.BufferType.SPANNABLE
            )
        }
    }

    override fun toString(): String {
        return logs.joinToString(separator = "<br/>", transform = {
            it.toString()
        })
    }
}