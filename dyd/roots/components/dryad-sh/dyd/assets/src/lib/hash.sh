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

dryad_blake2b_128_file_hex_shell () {
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

dryad_blake2b_128_file_hex () {
    dryad_b2_file=$1
    dryad_b2_format=${2:-hex}

    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        dryad_blake2b_128_file_hex_shell "$dryad_b2_file"
        return 0
    fi

    if [ "$dryad_b2_format" = files-table ]; then
        cat
    elif [ "$dryad_b2_file" = - ]; then
        od -An -v -tu1
    else
        od -An -v -tu1 "$dryad_b2_file"
    fi | awk -v format="$dryad_b2_format" '
        function bxor8(a, b,    key, aa, bb, bit, value, place) {
            key = a SUBSEP b
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

        function hash_file_base32(path,    cmd, line, fields, i, close_status) {
            reset_hash()
            hash_byte(102)
            hash_byte(105)
            hash_byte(108)
            hash_byte(101)
            hash_byte(0)

            cmd = "od -An -v -tu1 " shell_quote(path)
            while ((cmd | getline line) > 0) {
                split(line, fields, " ")
                for (i = 1; i <= length(fields); i++) {
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

            split($0, fields, " ")
            for (i = 1; i <= length(fields); i++) {
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

dryad_blake2b_128_file_base32 () {
    dryad_b2_base32_file=$1

    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        dryad_b2_base32_hex=$(dryad_blake2b_128_file_hex_shell "$dryad_b2_base32_file")
        dryad_base32_encode_hex "$dryad_b2_base32_hex"
        return 0
    fi

    dryad_blake2b_128_file_hex "$dryad_b2_base32_file" base32
}

dryad_blake2b_128_files_table_base32 () {
    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        dryad_b2_files_table_sep=$(printf '\t')
        while IFS=$dryad_b2_files_table_sep read -r dryad_b2_files_table_rel dryad_b2_files_table_path; do
            [ -n "$dryad_b2_files_table_rel" ] || continue
            dryad_b2_files_table_hash=$(
                {
                    printf 'file\000'
                    cat "$dryad_b2_files_table_path"
                } | dryad_blake2b_128_stream_base32
            )
            printf '%s\t%s\n' "$dryad_b2_files_table_rel" "$dryad_b2_files_table_hash"
        done
        return 0
    fi

    dryad_blake2b_128_file_hex - files-table
}

dryad_blake2b_128_stream_base32 () {
    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        dryad_b2_stream_tmp=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-hash-stream.XXXXXX")
        cat > "$dryad_b2_stream_tmp"
        dryad_b2_stream_hex=$(dryad_blake2b_128_file_hex_shell "$dryad_b2_stream_tmp")
        rm -f "$dryad_b2_stream_tmp"
        dryad_base32_encode_hex "$dryad_b2_stream_hex"
        return 0
    fi

    dryad_blake2b_128_file_hex - base32
}

dryad_memo_init () {
    if [ -n "${DRYAD_SH_MEMO_DIR:-}" ]; then
        mkdir -p "$DRYAD_SH_MEMO_DIR"
        export DRYAD_SH_MEMO_DIR
        return 0
    fi

    DRYAD_SH_MEMO_DIR=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-memo.XXXXXX")
    export DRYAD_SH_MEMO_DIR
}

dryad_memo_cleanup () {
    if [ -n "${DRYAD_SH_MEMO_DIR:-}" ]; then
        rm -rf "$DRYAD_SH_MEMO_DIR"
        unset DRYAD_SH_MEMO_DIR
    fi
}

dryad_memo_escape_arg () {
    dryad_memo_escape_in=$1

    if [ -z "$dryad_memo_escape_in" ]; then
        printf '~e\n'
        return 0
    fi

    dryad_memo_escape_out=
    while [ -n "$dryad_memo_escape_in" ]; do
        dryad_memo_escape_ch=${dryad_memo_escape_in%"${dryad_memo_escape_in#?}"}
        dryad_memo_escape_in=${dryad_memo_escape_in#?}
        case $dryad_memo_escape_ch in
            '~' )
                dryad_memo_escape_out=${dryad_memo_escape_out}~~
                ;;
            '/' )
                dryad_memo_escape_out=${dryad_memo_escape_out}~s
                ;;
            '^' )
                dryad_memo_escape_out=${dryad_memo_escape_out}~c
                ;;
            * )
                dryad_memo_escape_out=$dryad_memo_escape_out$dryad_memo_escape_ch
                ;;
        esac
    done

    printf '%s\n' "$dryad_memo_escape_out"
}

dryad_memo_path () {
    [ -n "${DRYAD_SH_MEMO_DIR:-}" ] || dryad_die "memo dir is not initialized"
    dryad_memo_path_group=$1
    shift

    dryad_memo_path_dir=$DRYAD_SH_MEMO_DIR/$dryad_memo_path_group
    dryad_memo_path_key=
    dryad_memo_path_sep=
    if [ "$#" -eq 0 ]; then
        dryad_memo_path_key=~k
    else
        for dryad_memo_path_arg do
            dryad_memo_path_escaped=$(dryad_memo_escape_arg "$dryad_memo_path_arg")
            dryad_memo_path_key=$dryad_memo_path_key$dryad_memo_path_sep$dryad_memo_path_escaped
            dryad_memo_path_sep='^'
        done
    fi
    printf '%s/%s\n' "$dryad_memo_path_dir" "$dryad_memo_path_key"
}

dryad_memo_get () {
    dryad_memo_init
    dryad_memo_get_path=$(dryad_memo_path "$@")
    [ -f "$dryad_memo_get_path" ] || return 1
    cat "$dryad_memo_get_path"
}

dryad_memo_get_line_into () {
    dryad_memo_get_line_var=$1
    shift

    dryad_memo_init
    dryad_memo_get_line_path=$(dryad_memo_path "$@")
    [ -f "$dryad_memo_get_line_path" ] || return 1

    dryad_memo_get_line_value=
    IFS= read -r dryad_memo_get_line_value < "$dryad_memo_get_line_path" || dryad_memo_get_line_value=
    eval "$dryad_memo_get_line_var=\$dryad_memo_get_line_value"
}

dryad_memo_put () {
    dryad_memo_init
    dryad_memo_put_path=$(dryad_memo_path "$@")
    dryad_memo_put_tmp=$dryad_memo_put_path.tmp.$$

    mkdir -p "$(dirname "$dryad_memo_put_path")"
    rm -f "$dryad_memo_put_tmp"
    cat > "$dryad_memo_put_tmp"

    if [ -f "$dryad_memo_put_path" ]; then
        rm -f "$dryad_memo_put_tmp"
        return 0
    fi

    if ! mv "$dryad_memo_put_tmp" "$dryad_memo_put_path" 2>/dev/null; then
        rm -f "$dryad_memo_put_tmp"
        [ -f "$dryad_memo_put_path" ] || return 1
    fi
}

dryad_memo_put_value () {
    dryad_memo_put_value_group=$1
    shift
    dryad_memo_put_value_value=$1
    shift

    dryad_memo_init
    dryad_memo_put_value_path=$(dryad_memo_path "$dryad_memo_put_value_group" "$@")
    dryad_memo_put_value_tmp=$dryad_memo_put_value_path.tmp.$$

    mkdir -p "$(dirname "$dryad_memo_put_value_path")"
    rm -f "$dryad_memo_put_value_tmp"
    printf '%s' "$dryad_memo_put_value_value" > "$dryad_memo_put_value_tmp"

    if [ -f "$dryad_memo_put_value_path" ]; then
        rm -f "$dryad_memo_put_value_tmp"
        return 0
    fi

    if ! mv "$dryad_memo_put_value_tmp" "$dryad_memo_put_value_path" 2>/dev/null; then
        rm -f "$dryad_memo_put_value_tmp"
        [ -f "$dryad_memo_put_value_path" ] || return 1
    fi
}

dryad_blake2b_128_file_fingerprint () {
    dryad_b2_fp_file=$1
    printf 'v2-%s\n' "$(dryad_blake2b_128_file_base32 "$dryad_b2_fp_file")"
}
