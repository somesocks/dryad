#
# Fingerprint primitives.
#
# These intentionally avoid external base32 or BLAKE2 tools so dryad-sh can
# act as a small bootstrap implementation on a plain POSIX system.
#

dryad_hex_nibble () {
    case $1 in
        0 ) printf '0\n' ;;
        1 ) printf '1\n' ;;
        2 ) printf '2\n' ;;
        3 ) printf '3\n' ;;
        4 ) printf '4\n' ;;
        5 ) printf '5\n' ;;
        6 ) printf '6\n' ;;
        7 ) printf '7\n' ;;
        8 ) printf '8\n' ;;
        9 ) printf '9\n' ;;
        a | A ) printf '10\n' ;;
        b | B ) printf '11\n' ;;
        c | C ) printf '12\n' ;;
        d | D ) printf '13\n' ;;
        e | E ) printf '14\n' ;;
        f | F ) printf '15\n' ;;
        * ) dryad_die "invalid hex digit: $1" ;;
    esac
}

dryad_hex_byte_value () {
    dryad_hex_byte=$1
    dryad_hex_byte_hi=${dryad_hex_byte%?}
    dryad_hex_byte_lo=${dryad_hex_byte#?}
    dryad_hex_byte_hi=$(dryad_hex_nibble "$dryad_hex_byte_hi")
    dryad_hex_byte_lo=$(dryad_hex_nibble "$dryad_hex_byte_lo")
    printf '%s\n' $((dryad_hex_byte_hi * 16 + dryad_hex_byte_lo))
}

dryad_base32_char () {
    case $1 in
        0 ) printf 'a' ;; 1 ) printf 'b' ;; 2 ) printf 'c' ;; 3 ) printf 'd' ;;
        4 ) printf 'e' ;; 5 ) printf 'f' ;; 6 ) printf 'g' ;; 7 ) printf 'h' ;;
        8 ) printf 'i' ;; 9 ) printf 'j' ;; 10 ) printf 'k' ;; 11 ) printf 'l' ;;
        12 ) printf 'm' ;; 13 ) printf 'n' ;; 14 ) printf 'o' ;; 15 ) printf 'p' ;;
        16 ) printf 'q' ;; 17 ) printf 'r' ;; 18 ) printf 's' ;; 19 ) printf 't' ;;
        20 ) printf 'u' ;; 21 ) printf 'v' ;; 22 ) printf 'w' ;; 23 ) printf 'x' ;;
        24 ) printf 'y' ;; 25 ) printf 'z' ;; 26 ) printf '2' ;; 27 ) printf '3' ;;
        28 ) printf '4' ;; 29 ) printf '5' ;; 30 ) printf '6' ;; 31 ) printf '7' ;;
        * ) dryad_die "invalid base32 index: $1" ;;
    esac
}

dryad_base32_encode_hex () {
    dryad_base32_hex=$1
    dryad_base32_value=0
    dryad_base32_bits=0

    for dryad_base32_byte_hex in $(printf '%s\n' "$dryad_base32_hex" | sed 's/../& /g'); do
        [ -n "$dryad_base32_byte_hex" ] || continue
        dryad_base32_byte=$(dryad_hex_byte_value "$dryad_base32_byte_hex")
        dryad_base32_value=$((dryad_base32_value * 256 + dryad_base32_byte))
        dryad_base32_bits=$((dryad_base32_bits + 8))
        while [ "$dryad_base32_bits" -ge 5 ]; do
            dryad_base32_shift=$((dryad_base32_bits - 5))
            dryad_base32_index=$(((dryad_base32_value >> dryad_base32_shift) & 31))
            dryad_base32_char "$dryad_base32_index"
            dryad_base32_bits=$dryad_base32_shift
            if [ "$dryad_base32_bits" -gt 0 ]; then
                dryad_base32_value=$((dryad_base32_value & ((1 << dryad_base32_bits) - 1)))
            else
                dryad_base32_value=0
            fi
        done
    done

    if [ "$dryad_base32_bits" -gt 0 ]; then
        dryad_base32_index=$(((dryad_base32_value << (5 - dryad_base32_bits)) & 31))
        dryad_base32_char "$dryad_base32_index"
    fi
    printf '\n'
}

dryad_b2_init_word () {
    dryad_b2_init_word_name=$1
    dryad_b2_init_word_0=$2
    dryad_b2_init_word_1=$3
    dryad_b2_init_word_2=$4
    dryad_b2_init_word_3=$5

    eval "${dryad_b2_init_word_name}_0=\$dryad_b2_init_word_0"
    eval "${dryad_b2_init_word_name}_1=\$dryad_b2_init_word_1"
    eval "${dryad_b2_init_word_name}_2=\$dryad_b2_init_word_2"
    eval "${dryad_b2_init_word_name}_3=\$dryad_b2_init_word_3"
}

dryad_b2_copy_word () {
    dryad_b2_copy_dst=$1
    dryad_b2_copy_src=$2
    eval "${dryad_b2_copy_dst}_0=\${${dryad_b2_copy_src}_0}"
    eval "${dryad_b2_copy_dst}_1=\${${dryad_b2_copy_src}_1}"
    eval "${dryad_b2_copy_dst}_2=\${${dryad_b2_copy_src}_2}"
    eval "${dryad_b2_copy_dst}_3=\${${dryad_b2_copy_src}_3}"
}

dryad_b2_word_xor_to () {
    dryad_b2_xor_dst=$1
    dryad_b2_xor_src=$2
    eval "${dryad_b2_xor_dst}_0=\$((\${${dryad_b2_xor_dst}_0} ^ \${${dryad_b2_xor_src}_0}))"
    eval "${dryad_b2_xor_dst}_1=\$((\${${dryad_b2_xor_dst}_1} ^ \${${dryad_b2_xor_src}_1}))"
    eval "${dryad_b2_xor_dst}_2=\$((\${${dryad_b2_xor_dst}_2} ^ \${${dryad_b2_xor_src}_2}))"
    eval "${dryad_b2_xor_dst}_3=\$((\${${dryad_b2_xor_dst}_3} ^ \${${dryad_b2_xor_src}_3}))"
}

dryad_b2_word_xor_mask_to () {
    dryad_b2_xor_mask_dst=$1
    eval "${dryad_b2_xor_mask_dst}_0=\$((\${${dryad_b2_xor_mask_dst}_0} ^ 65535))"
    eval "${dryad_b2_xor_mask_dst}_1=\$((\${${dryad_b2_xor_mask_dst}_1} ^ 65535))"
    eval "${dryad_b2_xor_mask_dst}_2=\$((\${${dryad_b2_xor_mask_dst}_2} ^ 65535))"
    eval "${dryad_b2_xor_mask_dst}_3=\$((\${${dryad_b2_xor_mask_dst}_3} ^ 65535))"
}

dryad_b2_word_add3_to () {
    dryad_b2_add_dst=$1
    dryad_b2_add_src=$2
    dryad_b2_add_msg=$3

    eval "dryad_b2_add_a0=\${${dryad_b2_add_dst}_0}; dryad_b2_add_a1=\${${dryad_b2_add_dst}_1}; dryad_b2_add_a2=\${${dryad_b2_add_dst}_2}; dryad_b2_add_a3=\${${dryad_b2_add_dst}_3}"
    eval "dryad_b2_add_b0=\${${dryad_b2_add_src}_0}; dryad_b2_add_b1=\${${dryad_b2_add_src}_1}; dryad_b2_add_b2=\${${dryad_b2_add_src}_2}; dryad_b2_add_b3=\${${dryad_b2_add_src}_3}"
    eval "dryad_b2_add_c0=\${${dryad_b2_add_msg}_0}; dryad_b2_add_c1=\${${dryad_b2_add_msg}_1}; dryad_b2_add_c2=\${${dryad_b2_add_msg}_2}; dryad_b2_add_c3=\${${dryad_b2_add_msg}_3}"

    dryad_b2_add_sum=$((dryad_b2_add_a0 + dryad_b2_add_b0 + dryad_b2_add_c0))
    dryad_b2_add_r0=$((dryad_b2_add_sum & 65535))
    dryad_b2_add_carry=$((dryad_b2_add_sum >> 16))
    dryad_b2_add_sum=$((dryad_b2_add_a1 + dryad_b2_add_b1 + dryad_b2_add_c1 + dryad_b2_add_carry))
    dryad_b2_add_r1=$((dryad_b2_add_sum & 65535))
    dryad_b2_add_carry=$((dryad_b2_add_sum >> 16))
    dryad_b2_add_sum=$((dryad_b2_add_a2 + dryad_b2_add_b2 + dryad_b2_add_c2 + dryad_b2_add_carry))
    dryad_b2_add_r2=$((dryad_b2_add_sum & 65535))
    dryad_b2_add_carry=$((dryad_b2_add_sum >> 16))
    dryad_b2_add_sum=$((dryad_b2_add_a3 + dryad_b2_add_b3 + dryad_b2_add_c3 + dryad_b2_add_carry))
    dryad_b2_add_r3=$((dryad_b2_add_sum & 65535))

    dryad_b2_init_word "$dryad_b2_add_dst" "$dryad_b2_add_r0" "$dryad_b2_add_r1" "$dryad_b2_add_r2" "$dryad_b2_add_r3"
}

dryad_b2_word_add2_to () {
    dryad_b2_add2_zero=0
    dryad_b2_init_word dryad_b2_add2_zero_word 0 0 0 0
    dryad_b2_word_add3_to "$1" "$2" dryad_b2_add2_zero_word
}

dryad_b2_word_rotr_to () {
    dryad_b2_rotr_dst=$1
    dryad_b2_rotr_bits=$2
    eval "dryad_b2_rotr_0=\${${dryad_b2_rotr_dst}_0}; dryad_b2_rotr_1=\${${dryad_b2_rotr_dst}_1}; dryad_b2_rotr_2=\${${dryad_b2_rotr_dst}_2}; dryad_b2_rotr_3=\${${dryad_b2_rotr_dst}_3}"

    case $dryad_b2_rotr_bits in
        16 )
            dryad_b2_init_word "$dryad_b2_rotr_dst" "$dryad_b2_rotr_1" "$dryad_b2_rotr_2" "$dryad_b2_rotr_3" "$dryad_b2_rotr_0"
            ;;
        32 )
            dryad_b2_init_word "$dryad_b2_rotr_dst" "$dryad_b2_rotr_2" "$dryad_b2_rotr_3" "$dryad_b2_rotr_0" "$dryad_b2_rotr_1"
            ;;
        24 )
            dryad_b2_rotr_a=$dryad_b2_rotr_1
            dryad_b2_rotr_b=$dryad_b2_rotr_2
            dryad_b2_rotr_c=$dryad_b2_rotr_3
            dryad_b2_rotr_d=$dryad_b2_rotr_0
            dryad_b2_init_word "$dryad_b2_rotr_dst" \
                $(((dryad_b2_rotr_a >> 8) | ((dryad_b2_rotr_b & 255) << 8))) \
                $(((dryad_b2_rotr_b >> 8) | ((dryad_b2_rotr_c & 255) << 8))) \
                $(((dryad_b2_rotr_c >> 8) | ((dryad_b2_rotr_d & 255) << 8))) \
                $(((dryad_b2_rotr_d >> 8) | ((dryad_b2_rotr_a & 255) << 8)))
            ;;
        63 )
            dryad_b2_init_word "$dryad_b2_rotr_dst" \
                $((((dryad_b2_rotr_0 << 1) & 65535) | (dryad_b2_rotr_3 >> 15))) \
                $((((dryad_b2_rotr_1 << 1) & 65535) | (dryad_b2_rotr_0 >> 15))) \
                $((((dryad_b2_rotr_2 << 1) & 65535) | (dryad_b2_rotr_1 >> 15))) \
                $((((dryad_b2_rotr_3 << 1) & 65535) | (dryad_b2_rotr_2 >> 15)))
            ;;
        * )
            dryad_die "unsupported BLAKE2b rotation: $dryad_b2_rotr_bits"
            ;;
    esac
}

dryad_b2_G () {
    dryad_b2_G_a=$1
    dryad_b2_G_b=$2
    dryad_b2_G_c=$3
    dryad_b2_G_d=$4
    dryad_b2_G_x=$5
    dryad_b2_G_y=$6

    dryad_b2_word_add3_to "dryad_b2_v$dryad_b2_G_a" "dryad_b2_v$dryad_b2_G_b" "dryad_b2_m$dryad_b2_G_x"
    dryad_b2_word_xor_to "dryad_b2_v$dryad_b2_G_d" "dryad_b2_v$dryad_b2_G_a"
    dryad_b2_word_rotr_to "dryad_b2_v$dryad_b2_G_d" 32
    dryad_b2_word_add2_to "dryad_b2_v$dryad_b2_G_c" "dryad_b2_v$dryad_b2_G_d"
    dryad_b2_word_xor_to "dryad_b2_v$dryad_b2_G_b" "dryad_b2_v$dryad_b2_G_c"
    dryad_b2_word_rotr_to "dryad_b2_v$dryad_b2_G_b" 24
    dryad_b2_word_add3_to "dryad_b2_v$dryad_b2_G_a" "dryad_b2_v$dryad_b2_G_b" "dryad_b2_m$dryad_b2_G_y"
    dryad_b2_word_xor_to "dryad_b2_v$dryad_b2_G_d" "dryad_b2_v$dryad_b2_G_a"
    dryad_b2_word_rotr_to "dryad_b2_v$dryad_b2_G_d" 16
    dryad_b2_word_add2_to "dryad_b2_v$dryad_b2_G_c" "dryad_b2_v$dryad_b2_G_d"
    dryad_b2_word_xor_to "dryad_b2_v$dryad_b2_G_b" "dryad_b2_v$dryad_b2_G_c"
    dryad_b2_word_rotr_to "dryad_b2_v$dryad_b2_G_b" 63
}

dryad_b2_round () {
    set -- $1
    dryad_b2_G 0 4 8 12 "$1" "$2"
    dryad_b2_G 1 5 9 13 "$3" "$4"
    dryad_b2_G 2 6 10 14 "$5" "$6"
    dryad_b2_G 3 7 11 15 "$7" "$8"
    dryad_b2_G 0 5 10 15 "$9" "${10}"
    dryad_b2_G 1 6 11 12 "${11}" "${12}"
    dryad_b2_G 2 7 8 13 "${13}" "${14}"
    dryad_b2_G 3 4 9 14 "${15}" "${16}"
}

dryad_b2_load_block_word () {
    dryad_b2_load_idx=$1
    dryad_b2_load_base=$((dryad_b2_load_idx * 8))
    dryad_b2_load_byte_i=$dryad_b2_load_base
    while [ "$dryad_b2_load_byte_i" -lt $((dryad_b2_load_base + 8)) ]; do
        eval "dryad_b2_load_b$((dryad_b2_load_byte_i - dryad_b2_load_base))=\${dryad_b2_block_$dryad_b2_load_byte_i:-0}"
        dryad_b2_load_byte_i=$((dryad_b2_load_byte_i + 1))
    done
    eval "dryad_b2_m${dryad_b2_load_idx}_0=\$((dryad_b2_load_b0 | (dryad_b2_load_b1 << 8)))"
    eval "dryad_b2_m${dryad_b2_load_idx}_1=\$((dryad_b2_load_b2 | (dryad_b2_load_b3 << 8)))"
    eval "dryad_b2_m${dryad_b2_load_idx}_2=\$((dryad_b2_load_b4 | (dryad_b2_load_b5 << 8)))"
    eval "dryad_b2_m${dryad_b2_load_idx}_3=\$((dryad_b2_load_b6 | (dryad_b2_load_b7 << 8)))"
}

dryad_b2_add_counter () {
    dryad_b2_counter_add=$1
    dryad_b2_counter_sum=$((dryad_b2_t0 + dryad_b2_counter_add))
    dryad_b2_t0=$((dryad_b2_counter_sum & 65535))
    dryad_b2_counter_carry=$((dryad_b2_counter_sum >> 16))
    dryad_b2_counter_sum=$((dryad_b2_t1 + dryad_b2_counter_carry))
    dryad_b2_t1=$((dryad_b2_counter_sum & 65535))
    dryad_b2_counter_carry=$((dryad_b2_counter_sum >> 16))
    dryad_b2_counter_sum=$((dryad_b2_t2 + dryad_b2_counter_carry))
    dryad_b2_t2=$((dryad_b2_counter_sum & 65535))
    dryad_b2_counter_carry=$((dryad_b2_counter_sum >> 16))
    dryad_b2_t3=$(((dryad_b2_t3 + dryad_b2_counter_carry) & 65535))
}

dryad_b2_compress () {
    dryad_b2_compress_final=$1
    dryad_b2_compress_load_i=0
    while [ "$dryad_b2_compress_load_i" -lt 16 ]; do
        dryad_b2_load_block_word "$dryad_b2_compress_load_i"
        dryad_b2_compress_load_i=$((dryad_b2_compress_load_i + 1))
    done

    dryad_b2_copy_word dryad_b2_v0 dryad_b2_h0
    dryad_b2_copy_word dryad_b2_v1 dryad_b2_h1
    dryad_b2_copy_word dryad_b2_v2 dryad_b2_h2
    dryad_b2_copy_word dryad_b2_v3 dryad_b2_h3
    dryad_b2_copy_word dryad_b2_v4 dryad_b2_h4
    dryad_b2_copy_word dryad_b2_v5 dryad_b2_h5
    dryad_b2_copy_word dryad_b2_v6 dryad_b2_h6
    dryad_b2_copy_word dryad_b2_v7 dryad_b2_h7
    dryad_b2_copy_word dryad_b2_v8 dryad_b2_iv0
    dryad_b2_copy_word dryad_b2_v9 dryad_b2_iv1
    dryad_b2_copy_word dryad_b2_v10 dryad_b2_iv2
    dryad_b2_copy_word dryad_b2_v11 dryad_b2_iv3
    dryad_b2_copy_word dryad_b2_v12 dryad_b2_iv4
    dryad_b2_copy_word dryad_b2_v13 dryad_b2_iv5
    dryad_b2_copy_word dryad_b2_v14 dryad_b2_iv6
    dryad_b2_copy_word dryad_b2_v15 dryad_b2_iv7

    dryad_b2_init_word dryad_b2_tword "$dryad_b2_t0" "$dryad_b2_t1" "$dryad_b2_t2" "$dryad_b2_t3"
    dryad_b2_word_xor_to dryad_b2_v12 dryad_b2_tword
    if [ "$dryad_b2_compress_final" = 1 ]; then
        dryad_b2_word_xor_mask_to dryad_b2_v14
    fi

    dryad_b2_round "0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15"
    dryad_b2_round "14 10 4 8 9 15 13 6 1 12 0 2 11 7 5 3"
    dryad_b2_round "11 8 12 0 5 2 15 13 10 14 3 6 7 1 9 4"
    dryad_b2_round "7 9 3 1 13 12 11 14 2 6 5 10 4 0 15 8"
    dryad_b2_round "9 0 5 7 2 4 10 15 14 1 11 12 6 8 3 13"
    dryad_b2_round "2 12 6 10 0 11 8 3 4 13 7 5 15 14 1 9"
    dryad_b2_round "12 5 1 15 14 13 4 10 0 7 6 3 9 2 8 11"
    dryad_b2_round "13 11 7 14 12 1 3 9 5 0 15 4 8 6 2 10"
    dryad_b2_round "6 15 14 9 11 3 0 8 12 2 13 7 1 4 10 5"
    dryad_b2_round "10 2 8 4 7 6 1 5 15 11 9 14 3 12 13 0"
    dryad_b2_round "0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15"
    dryad_b2_round "14 10 4 8 9 15 13 6 1 12 0 2 11 7 5 3"

    dryad_b2_h_i=0
    while [ "$dryad_b2_h_i" -lt 8 ]; do
        dryad_b2_word_xor_to "dryad_b2_h$dryad_b2_h_i" "dryad_b2_v$dryad_b2_h_i"
        dryad_b2_word_xor_to "dryad_b2_h$dryad_b2_h_i" "dryad_b2_v$((dryad_b2_h_i + 8))"
        dryad_b2_h_i=$((dryad_b2_h_i + 1))
    done
}

dryad_b2_word_hex_le () {
    dryad_b2_hex_word=$1
    eval "dryad_b2_hex_0=\${${dryad_b2_hex_word}_0}; dryad_b2_hex_1=\${${dryad_b2_hex_word}_1}; dryad_b2_hex_2=\${${dryad_b2_hex_word}_2}; dryad_b2_hex_3=\${${dryad_b2_hex_word}_3}"
    printf '%02x%02x%02x%02x%02x%02x%02x%02x' \
        $((dryad_b2_hex_0 & 255)) $((dryad_b2_hex_0 >> 8)) \
        $((dryad_b2_hex_1 & 255)) $((dryad_b2_hex_1 >> 8)) \
        $((dryad_b2_hex_2 & 255)) $((dryad_b2_hex_2 >> 8)) \
        $((dryad_b2_hex_3 & 255)) $((dryad_b2_hex_3 >> 8))
}

dryad_blake2b_128_file_hex () {
    dryad_b2_file=$1
    dryad_b2_total=$(wc -c < "$dryad_b2_file" | tr -d ' ')

    dryad_b2_init_word dryad_b2_iv0 51464 62396 58983 27145
    dryad_b2_init_word dryad_b2_iv1 42811 33994 44677 47975
    dryad_b2_init_word dryad_b2_iv2 63531 65172 62322 15470
    dryad_b2_init_word dryad_b2_iv3 14065 24349 62778 42319
    dryad_b2_init_word dryad_b2_iv4 33489 44518 21119 20750
    dryad_b2_init_word dryad_b2_iv5 27679 11070 26764 39685
    dryad_b2_init_word dryad_b2_iv6 48491 64321 55723 8067
    dryad_b2_init_word dryad_b2_iv7 8569 4990 52505 23520

    dryad_b2_copy_word dryad_b2_h0 dryad_b2_iv0
    dryad_b2_copy_word dryad_b2_h1 dryad_b2_iv1
    dryad_b2_copy_word dryad_b2_h2 dryad_b2_iv2
    dryad_b2_copy_word dryad_b2_h3 dryad_b2_iv3
    dryad_b2_copy_word dryad_b2_h4 dryad_b2_iv4
    dryad_b2_copy_word dryad_b2_h5 dryad_b2_iv5
    dryad_b2_copy_word dryad_b2_h6 dryad_b2_iv6
    dryad_b2_copy_word dryad_b2_h7 dryad_b2_iv7
    dryad_b2_init_word dryad_b2_param 16 257 0 0
    dryad_b2_word_xor_to dryad_b2_h0 dryad_b2_param

    dryad_b2_t0=0
    dryad_b2_t1=0
    dryad_b2_t2=0
    dryad_b2_t3=0
    dryad_b2_block_len=0
    dryad_b2_processed=0

    for dryad_b2_byte in $(od -An -v -tu1 "$dryad_b2_file"); do
        eval "dryad_b2_block_$dryad_b2_block_len=\$dryad_b2_byte"
        dryad_b2_block_len=$((dryad_b2_block_len + 1))
        if [ "$dryad_b2_block_len" -eq 128 ] &&
            [ $((dryad_b2_processed + 128)) -lt "$dryad_b2_total" ]; then
            dryad_b2_add_counter 128
            dryad_b2_compress 0
            dryad_b2_processed=$((dryad_b2_processed + 128))
            dryad_b2_block_i=0
            while [ "$dryad_b2_block_i" -lt 128 ]; do
                eval "unset dryad_b2_block_$dryad_b2_block_i"
                dryad_b2_block_i=$((dryad_b2_block_i + 1))
            done
            dryad_b2_block_len=0
        fi
    done

    dryad_b2_add_counter "$dryad_b2_block_len"
    dryad_b2_compress 1
    {
        dryad_b2_word_hex_le dryad_b2_h0
        dryad_b2_word_hex_le dryad_b2_h1
    }
    printf '\n'
}

dryad_blake2b_128_file_fingerprint () {
    dryad_b2_fp_file=$1
    dryad_b2_fp_hex=$(dryad_blake2b_128_file_hex "$dryad_b2_fp_file")
    printf 'v2-%s\n' "$(dryad_base32_encode_hex "$dryad_b2_fp_hex")"
}
