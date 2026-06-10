package de.arcan.arclib.msgpack

import java.io.ByteArrayOutputStream

/**
 * Minimal MessagePack encoder/decoder for the MyClerk protocol.
 * Supports: maps, arrays, strings, integers, longs, booleans, binary, null.
 *
 * @see https://github.com/msgpack/msgpack/blob/master/spec.md
 */
object MsgPack {

    fun encode(value: Any?): ByteArray {
        val out = ByteArrayOutputStream()
        writeValue(out, value)
        return out.toByteArray()
    }

    /** Hard cap on nesting depth to stop a malicious peer from overflowing the JVM stack
     *  (StackOverflowError is an Error, not caught by callers' catch(Exception)). */
    private const val MAX_DEPTH = 64

    fun decode(buf: ByteArray, off: Int = 0): Pair<Any?, Int> {
        val cur = intArrayOf(off)
        val v = readAny(buf, cur, 0)
        return Pair(v, cur[0] - off)
    }

    private fun writeValue(out: ByteArrayOutputStream, value: Any?) {
        when (value) {
            null -> out.write(0xC0)
            is Boolean -> out.write(if (value) 0xC3 else 0xC2)
            is Int -> writeInt(out, value)
            is Long -> writeLong(out, value)
            is String -> writeString(out, value)
            is ByteArray -> writeBin(out, value)
            is Map<*, *> -> writeMap(out, value)
            is List<*> -> writeArray(out, value)
            is Array<*> -> writeArray(out, value.toList())
            else -> writeString(out, value.toString())
        }
    }

    private fun writeLong(out: ByteArrayOutputStream, value: Long) {
        if (value in Int.MIN_VALUE..Int.MAX_VALUE) { writeInt(out, value.toInt()); return }
        if (value >= 0) {
            out.write(0xCF)
            for (i in 7 downTo 0) out.write(((value ushr (i * 8)) and 0xFF).toInt())
        } else {
            out.write(0xD3)
            for (i in 7 downTo 0) out.write(((value ushr (i * 8)) and 0xFF).toInt())
        }
    }

    private fun writeBin(out: ByteArrayOutputStream, bytes: ByteArray) {
        when {
            bytes.size <= 0xFF -> { out.write(0xC4); out.write(bytes.size) }
            bytes.size <= 0xFFFF -> { out.write(0xC5); writeU16(out, bytes.size) }
            else -> { out.write(0xC6); writeU32(out, bytes.size) }
        }
        out.write(bytes)
    }

    private fun writeInt(out: ByteArrayOutputStream, value: Int) {
        when {
            value in 0..0x7F -> out.write(value)
            value in -32..-1 -> out.write(value and 0xFF)
            value in 0..0xFF -> { out.write(0xCC); out.write(value) }
            value in 0..0xFFFF -> { out.write(0xCD); writeU16(out, value) }
            value >= 0 -> { out.write(0xCE); writeU32(out, value) }
            value in -128..-1 -> { out.write(0xD0); out.write(value and 0xFF) }
            value in -32768..-1 -> { out.write(0xD1); writeU16(out, value and 0xFFFF) }
            else -> { out.write(0xD2); writeU32(out, value) }
        }
    }

    private fun writeString(out: ByteArrayOutputStream, str: String) {
        val bytes = str.toByteArray(Charsets.UTF_8)
        when {
            bytes.size <= 31 -> out.write(0xA0 or bytes.size)
            bytes.size <= 0xFF -> { out.write(0xD9); out.write(bytes.size) }
            bytes.size <= 0xFFFF -> { out.write(0xDA); writeU16(out, bytes.size) }
            else -> { out.write(0xDB); writeU32(out, bytes.size) }
        }
        out.write(bytes)
    }

    private fun writeMap(out: ByteArrayOutputStream, map: Map<*, *>) {
        val size = map.size
        when {
            size <= 15 -> out.write(0x80 or size)
            size <= 0xFFFF -> { out.write(0xDE); writeU16(out, size) }
            else -> { out.write(0xDF); writeU32(out, size) }
        }
        for ((key, value) in map) { writeValue(out, key); writeValue(out, value) }
    }

    private fun writeArray(out: ByteArrayOutputStream, list: List<*>) {
        val size = list.size
        when {
            size <= 15 -> out.write(0x90 or size)
            size <= 0xFFFF -> { out.write(0xDC); writeU16(out, size) }
            else -> { out.write(0xDD); writeU32(out, size) }
        }
        for (item in list) writeValue(out, item)
    }

    private fun readAny(buf: ByteArray, cur: IntArray, depth: Int): Any? {
        if (depth > MAX_DEPTH) throw IllegalArgumentException("MsgPack: nesting too deep (> $MAX_DEPTH)")
        val b = buf[cur[0]++].toInt() and 0xFF
        return when {
            b <= 0x7F -> b
            b >= 0xE0 -> b - 256
            b in 0xA0..0xBF -> readStr(buf, cur, b and 0x1F)
            b in 0x90..0x9F -> readArr(buf, cur, b and 0x0F, depth)
            b in 0x80..0x8F -> readMap(buf, cur, b and 0x0F, depth)
            b == 0xC0 -> null
            b == 0xC2 -> false
            b == 0xC3 -> true
            b == 0xC4 -> readBin(buf, cur, buf[cur[0]++].toInt() and 0xFF)
            b == 0xC5 -> { val n = readU16(buf, cur); readBin(buf, cur, n) }
            b == 0xC6 -> { val n = readU32(buf, cur); readBin(buf, cur, n) }
            b == 0xCC -> buf[cur[0]++].toInt() and 0xFF
            b == 0xCD -> readU16(buf, cur)
            b == 0xCE -> readU32(buf, cur)
            b == 0xCF -> readU64(buf, cur)
            b == 0xD0 -> buf[cur[0]++].toInt()
            b == 0xD1 -> { val n = readU16(buf, cur); if (n >= 0x8000) n - 0x10000 else n }
            b == 0xD2 -> readU32(buf, cur)
            b == 0xD3 -> readU64(buf, cur)
            b == 0xD9 -> readStr(buf, cur, buf[cur[0]++].toInt() and 0xFF)
            b == 0xDA -> { val n = readU16(buf, cur); readStr(buf, cur, n) }
            b == 0xDB -> { val n = readU32(buf, cur); readStr(buf, cur, n) }
            b == 0xDC -> { val n = readU16(buf, cur); readArr(buf, cur, n, depth) }
            b == 0xDD -> { val n = readU32(buf, cur); readArr(buf, cur, n, depth) }
            b == 0xDE -> { val n = readU16(buf, cur); readMap(buf, cur, n, depth) }
            b == 0xDF -> { val n = readU32(buf, cur); readMap(buf, cur, n, depth) }
            else -> throw IllegalArgumentException("MsgPack: unsupported tag 0x${b.toString(16)}")
        }
    }

    /** A length/count read off the wire can never exceed the bytes still in the buffer: every
     *  element/byte it claims needs at least one source byte. This single guard defeats both the
     *  pre-allocation OOM (e.g. bin32/array32 declaring 2 GB from a 5-byte frame) and runaway loops. */
    private fun checkLen(buf: ByteArray, cur: IntArray, n: Int) {
        val remaining = buf.size - cur[0]
        if (n < 0 || n > remaining) {
            throw IllegalArgumentException("MsgPack: length $n exceeds remaining $remaining bytes")
        }
    }

    private fun readStr(buf: ByteArray, cur: IntArray, n: Int): String {
        checkLen(buf, cur, n)
        val s = String(buf, cur[0], n, Charsets.UTF_8); cur[0] += n; return s
    }
    private fun readBin(buf: ByteArray, cur: IntArray, n: Int): ByteArray {
        checkLen(buf, cur, n)
        val out = ByteArray(n); System.arraycopy(buf, cur[0], out, 0, n); cur[0] += n; return out
    }
    private fun readArr(buf: ByteArray, cur: IntArray, n: Int, depth: Int): List<Any?> {
        checkLen(buf, cur, n)
        val list = ArrayList<Any?>(minOf(n, 1024))
        repeat(n) { list.add(readAny(buf, cur, depth + 1)) }
        return list
    }
    private fun readMap(buf: ByteArray, cur: IntArray, n: Int, depth: Int): Map<Any?, Any?> {
        checkLen(buf, cur, n)
        val m = LinkedHashMap<Any?, Any?>(); repeat(n) { m[readAny(buf, cur, depth + 1)] = readAny(buf, cur, depth + 1) }; return m
    }
    private fun readU16(buf: ByteArray, cur: IntArray): Int {
        val v = ((buf[cur[0]].toInt() and 0xFF) shl 8) or (buf[cur[0] + 1].toInt() and 0xFF); cur[0] += 2; return v
    }
    private fun readU32(buf: ByteArray, cur: IntArray): Int {
        val v = ((buf[cur[0]].toInt() and 0xFF) shl 24) or ((buf[cur[0]+1].toInt() and 0xFF) shl 16) or
                ((buf[cur[0]+2].toInt() and 0xFF) shl 8) or (buf[cur[0]+3].toInt() and 0xFF); cur[0] += 4; return v
    }
    private fun readU64(buf: ByteArray, cur: IntArray): Long {
        var v = 0L; for (i in 0 until 8) v = (v shl 8) or (buf[cur[0]+i].toLong() and 0xFF); cur[0] += 8; return v
    }
    private fun writeU16(out: ByteArrayOutputStream, v: Int) { out.write((v shr 8) and 0xFF); out.write(v and 0xFF) }
    private fun writeU32(out: ByteArrayOutputStream, v: Int) {
        out.write((v shr 24) and 0xFF); out.write((v shr 16) and 0xFF); out.write((v shr 8) and 0xFF); out.write(v and 0xFF)
    }
}
