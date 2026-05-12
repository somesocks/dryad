dryad_blake2b_128_file_hex_awk () {
    dyd_b2_file=$1
    dyd_b2_format=${2:-hex}

    if [ "$dyd_b2_format" = files-table ]; then
        cat
    elif [ "$dyd_b2_file" = - ]; then
        od -An -v -tu1
    else
        od -An -v -tu1 "$dyd_b2_file"
    fi | awk -v format="$dyd_b2_format" '
        function bxor8(a, b,    key, aa, bb, bit, value, place) {
            key = a * 256 + b
            if (key in xor8_cache) {
                return xor8_cache[key]
            }

            aa = a
            bb = b
            value = 0
            place = 1
            for (bit = 0; bit < 8; bit++) {
                if ((aa % 2) != (bb % 2)) {
                    value += place
                }
                aa = int(aa / 2)
                bb = int(bb / 2)
                place *= 2
            }

            xor8_cache[key] = value
            return value
        }

        function xor16(a, b) {
            return bxor8(a % 256, b % 256) + 256 * bxor8(int(a / 256), int(b / 256))
        }

        function copy_h_to_v(dst, src) {
            v0[dst] = h0[src]
            v1[dst] = h1[src]
            v2[dst] = h2[src]
            v3[dst] = h3[src]
        }

        function copy_iv_to_v(dst, src) {
            v0[dst] = iv0[src]
            v1[dst] = iv1[src]
            v2[dst] = iv2[src]
            v3[dst] = iv3[src]
        }

        function xor_v_v(dst, src) {
            v0[dst] = xor16(v0[dst], v0[src])
            v1[dst] = xor16(v1[dst], v1[src])
            v2[dst] = xor16(v2[dst], v2[src])
            v3[dst] = xor16(v3[dst], v3[src])
        }

        function xor_v_t(dst) {
            v0[dst] = xor16(v0[dst], t[0])
            v1[dst] = xor16(v1[dst], t[1])
            v2[dst] = xor16(v2[dst], t[2])
            v3[dst] = xor16(v3[dst], t[3])
        }

        function xor_v_mask(dst) {
            v0[dst] = xor16(v0[dst], 65535)
            v1[dst] = xor16(v1[dst], 65535)
            v2[dst] = xor16(v2[dst], 65535)
            v3[dst] = xor16(v3[dst], 65535)
        }

        function add3_v_v_m(dst, src, msg,    sum, carry) {
            sum = v0[dst] + v0[src] + m0[msg]
            v0[dst] = sum % 65536
            carry = int(sum / 65536)
            sum = v1[dst] + v1[src] + m1[msg] + carry
            v1[dst] = sum % 65536
            carry = int(sum / 65536)
            sum = v2[dst] + v2[src] + m2[msg] + carry
            v2[dst] = sum % 65536
            carry = int(sum / 65536)
            sum = v3[dst] + v3[src] + m3[msg] + carry
            v3[dst] = sum % 65536
        }

        function add2_v_v(dst, src,    sum, carry) {
            sum = v0[dst] + v0[src]
            v0[dst] = sum % 65536
            carry = int(sum / 65536)
            sum = v1[dst] + v1[src] + carry
            v1[dst] = sum % 65536
            carry = int(sum / 65536)
            sum = v2[dst] + v2[src] + carry
            v2[dst] = sum % 65536
            carry = int(sum / 65536)
            sum = v3[dst] + v3[src] + carry
            v3[dst] = sum % 65536
        }

        function rotr_v(idx, bits,    a0, a1, a2, a3) {
            a0 = v0[idx]
            a1 = v1[idx]
            a2 = v2[idx]
            a3 = v3[idx]
            if (bits == 16) {
                v0[idx] = a1
                v1[idx] = a2
                v2[idx] = a3
                v3[idx] = a0
            } else if (bits == 32) {
                v0[idx] = a2
                v1[idx] = a3
                v2[idx] = a0
                v3[idx] = a1
            } else if (bits == 24) {
                v0[idx] = int(a1 / 256) + 256 * (a2 % 256)
                v1[idx] = int(a2 / 256) + 256 * (a3 % 256)
                v2[idx] = int(a3 / 256) + 256 * (a0 % 256)
                v3[idx] = int(a0 / 256) + 256 * (a1 % 256)
            } else if (bits == 63) {
                v0[idx] = (2 * a0) % 65536 + int(a3 / 32768)
                v1[idx] = (2 * a1) % 65536 + int(a0 / 32768)
                v2[idx] = (2 * a2) % 65536 + int(a1 / 32768)
                v3[idx] = (2 * a3) % 65536 + int(a2 / 32768)
            } else {
                exit 2
            }
        }

        function G(a, b, c, d, x, y) {
            add3_v_v_m(a, b, x)
            xor_v_v(d, a)
            rotr_v(d, 32)
            add2_v_v(c, d)
            xor_v_v(b, c)
            rotr_v(b, 24)
            add3_v_v_m(a, b, y)
            xor_v_v(d, a)
            rotr_v(d, 16)
            add2_v_v(c, d)
            xor_v_v(b, c)
            rotr_v(b, 63)
        }

        function round(spec,    s) {
            split(spec, s, " ")
            G(0, 4, 8, 12, s[1], s[2])
            G(1, 5, 9, 13, s[3], s[4])
            G(2, 6, 10, 14, s[5], s[6])
            G(3, 7, 11, 15, s[7], s[8])
            G(0, 5, 10, 15, s[9], s[10])
            G(1, 6, 11, 12, s[11], s[12])
            G(2, 7, 8, 13, s[13], s[14])
            G(3, 4, 9, 14, s[15], s[16])
        }

        function load_block_word(idx,    base) {
            base = idx * 8
            m0[idx] = byte[base] + 256 * byte[base + 1]
            m1[idx] = byte[base + 2] + 256 * byte[base + 3]
            m2[idx] = byte[base + 4] + 256 * byte[base + 5]
            m3[idx] = byte[base + 6] + 256 * byte[base + 7]
        }

        function add_counter(n,    sum, carry) {
            sum = t[0] + n
            t[0] = sum % 65536
            carry = int(sum / 65536)
            sum = t[1] + carry
            t[1] = sum % 65536
            carry = int(sum / 65536)
            sum = t[2] + carry
            t[2] = sum % 65536
            carry = int(sum / 65536)
            t[3] = (t[3] + carry) % 65536
        }

        function compress(final,    i) {
            for (i = 0; i < 16; i++) {
                load_block_word(i)
            }
            for (i = 0; i < 8; i++) {
                copy_h_to_v(i, i)
                copy_iv_to_v(i + 8, i)
            }
            xor_v_t(12)
            if (final == 1) {
                xor_v_mask(14)
            }

            round("0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15")
            round("14 10 4 8 9 15 13 6 1 12 0 2 11 7 5 3")
            round("11 8 12 0 5 2 15 13 10 14 3 6 7 1 9 4")
            round("7 9 3 1 13 12 11 14 2 6 5 10 4 0 15 8")
            round("9 0 5 7 2 4 10 15 14 1 11 12 6 8 3 13")
            round("2 12 6 10 0 11 8 3 4 13 7 5 15 14 1 9")
            round("12 5 1 15 14 13 4 10 0 7 6 3 9 2 8 11")
            round("13 11 7 14 12 1 3 9 5 0 15 4 8 6 2 10")
            round("6 15 14 9 11 3 0 8 12 2 13 7 1 4 10 5")
            round("10 2 8 4 7 6 1 5 15 11 9 14 3 12 13 0")
            round("0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15")
            round("14 10 4 8 9 15 13 6 1 12 0 2 11 7 5 3")

            for (i = 0; i < 8; i++) {
                h0[i] = xor16(xor16(h0[i], v0[i]), v0[i + 8])
                h1[i] = xor16(xor16(h1[i], v1[i]), v1[i + 8])
                h2[i] = xor16(xor16(h2[i], v2[i]), v2[i + 8])
                h3[i] = xor16(xor16(h3[i], v3[i]), v3[i + 8])
            }
        }

        function hx(b) {
            return substr(hex, int(b / 16) + 1, 1) substr(hex, (b % 16) + 1, 1)
        }

        function word_hex(idx) {
            return hx(h0[idx] % 256) hx(int(h0[idx] / 256)) \
                hx(h1[idx] % 256) hx(int(h1[idx] / 256)) \
                hx(h2[idx] % 256) hx(int(h2[idx] / 256)) \
                hx(h3[idx] % 256) hx(int(h3[idx] / 256))
        }

        function hex_value(c) {
            return index(hex, c) - 1
        }

        function pow2(n,    value) {
            value = 1
            while (n > 0) {
                value *= 2
                n--
            }
            return value
        }

        function base32_char(idx) {
            return substr(base32_alphabet, idx + 1, 1)
        }

        function base32_hex(digest,    out, value, bits, pos, byte, shift, index_value) {
            out = ""
            value = 0
            bits = 0
            for (pos = 1; pos <= length(digest); pos += 2) {
                byte = hex_value(substr(digest, pos, 1)) * 16 + hex_value(substr(digest, pos + 1, 1))
                value = value * 256 + byte
                bits += 8
                while (bits >= 5) {
                    shift = bits - 5
                    index_value = int(value / pow2(shift)) % 32
                    out = out base32_char(index_value)
                    bits = shift
                    if (bits > 0) {
                        value = value % pow2(bits)
                    } else {
                        value = 0
                    }
                }
            }
            if (bits > 0) {
                index_value = (value * pow2(5 - bits)) % 32
                out = out base32_char(index_value)
            }
            return out
        }

        function reset_hash(    i) {
            for (i = 0; i < 8; i++) {
                h0[i] = iv0[i]
                h1[i] = iv1[i]
                h2[i] = iv2[i]
                h3[i] = iv3[i]
            }
            h0[0] = xor16(h0[0], 16)
            h1[0] = xor16(h1[0], 257)

            block_len = 0
            processed = 0
            delete byte
            for (i = 0; i < 4; i++) {
                t[i] = 0
            }
        }

        function hash_byte(b) {
            if (block_len == 128) {
                add_counter(128)
                compress(0)
                processed += 128
                delete byte
                block_len = 0
            }
            byte[block_len] = b + 0
            block_len++
        }

        function digest_hex() {
            add_counter(block_len)
            compress(1)
            return word_hex(0) word_hex(1)
        }

        function shell_quote(s,    q) {
            q = s
            gsub(/\047/, sprintf("%c%c%c%c", 39, 92, 39, 39), q)
            return sprintf("%c%s%c", 39, q, 39)
        }

        function hash_file_base32(path,    cmd, line, fields, fields_len, i, close_status) {
            reset_hash()
            hash_byte(102)
            hash_byte(105)
            hash_byte(108)
            hash_byte(101)
            hash_byte(0)

            cmd = "od -An -v -tu1 " shell_quote(path)
            while ((cmd | getline line) > 0) {
                fields_len = split(line, fields, " ")
                for (i = 1; i <= fields_len; i++) {
                    if (fields[i] != "") {
                        hash_byte(fields[i])
                    }
                }
            }
            close_status = close(cmd)
            if (close_status != 0) {
                exit 3
            }

            return base32_hex(digest_hex())
        }

        BEGIN {
            hex = "0123456789abcdef"
            base32_alphabet = "abcdefghijklmnopqrstuvwxyz234567"
            FS = "\t"

            iv0[0] = 51464; iv1[0] = 62396; iv2[0] = 58983; iv3[0] = 27145
            iv0[1] = 42811; iv1[1] = 33994; iv2[1] = 44677; iv3[1] = 47975
            iv0[2] = 63531; iv1[2] = 65172; iv2[2] = 62322; iv3[2] = 15470
            iv0[3] = 14065; iv1[3] = 24349; iv2[3] = 62778; iv3[3] = 42319
            iv0[4] = 33489; iv1[4] = 44518; iv2[4] = 21119; iv3[4] = 20750
            iv0[5] = 27679; iv1[5] = 11070; iv2[5] = 26764; iv3[5] = 39685
            iv0[6] = 48491; iv1[6] = 64321; iv2[6] = 55723; iv3[6] = 8067
            iv0[7] = 8569;  iv1[7] = 4990;  iv2[7] = 52505; iv3[7] = 23520

            reset_hash()
        }

        {
            if (format == "files-table") {
                if ($1 != "" && $2 != "") {
                    printf "%s\t%s\n", $1, hash_file_base32($2)
                }
                next
            }

            fields_len = split($0, fields, " ")
            for (i = 1; i <= fields_len; i++) {
                if (fields[i] != "") {
                    hash_byte(fields[i])
                }
            }
        }

        END {
            if (format == "files-table") {
                exit 0
            }

            digest = digest_hex()
            if (format == "base32") {
                printf "%s\n", base32_hex(digest)
            } else {
                printf "%s\n", digest
            }
        }
    '
}
