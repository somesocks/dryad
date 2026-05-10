#!/usr/bin/env sh

set -eu

emit_line () {
    printf '%s\n' "$1"
}

emit_word_assign () {
    prefix=$1
    idx=$2
    a=$3
    b=$4
    c=$5
    d=$6

    printf '    dryad_b2_%s%s_0=%s\n' "$prefix" "$idx" "$a"
    printf '    dryad_b2_%s%s_1=%s\n' "$prefix" "$idx" "$b"
    printf '    dryad_b2_%s%s_2=%s\n' "$prefix" "$idx" "$c"
    printf '    dryad_b2_%s%s_3=%s\n' "$prefix" "$idx" "$d"
}

emit_copy_word () {
    dst_prefix=$1
    dst=$2
    src_prefix=$3
    src=$4

    for lane in 0 1 2 3; do
        printf '    dryad_b2_%s%s_%s=$dryad_b2_%s%s_%s\n' "$dst_prefix" "$dst" "$lane" "$src_prefix" "$src" "$lane"
    done
}

emit_xor_word () {
    dst=$1
    src=$2

    for lane in 0 1 2 3; do
        printf '    dryad_b2_v%s_%s=$((dryad_b2_v%s_%s ^ dryad_b2_v%s_%s))\n' "$dst" "$lane" "$dst" "$lane" "$src" "$lane"
    done
}

emit_xor_word_m () {
    dst=$1
    src=$2

    for lane in 0 1 2 3; do
        printf '    dryad_b2_v%s_%s=$((dryad_b2_v%s_%s ^ dryad_b2_m%s_%s))\n' "$dst" "$lane" "$dst" "$lane" "$src" "$lane"
    done
}

emit_xor_counter () {
    dst=$1

    for lane in 0 1 2 3; do
        printf '    dryad_b2_v%s_%s=$((dryad_b2_v%s_%s ^ dryad_b2_t%s))\n' "$dst" "$lane" "$dst" "$lane" "$lane"
    done
}

emit_xor_mask () {
    dst=$1

    for lane in 0 1 2 3; do
        printf '        dryad_b2_v%s_%s=$((dryad_b2_v%s_%s ^ 65535))\n' "$dst" "$lane" "$dst" "$lane"
    done
}

emit_add3 () {
    dst=$1
    src=$2
    msg=$3

    printf '    dryad_b2_sum=$((dryad_b2_v%s_0 + dryad_b2_v%s_0 + dryad_b2_m%s_0))\n' "$dst" "$src" "$msg"
    printf '    dryad_b2_v%s_0=$((dryad_b2_sum & 65535))\n' "$dst"
    emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    for lane in 1 2 3; do
        printf '    dryad_b2_sum=$((dryad_b2_v%s_%s + dryad_b2_v%s_%s + dryad_b2_m%s_%s + dryad_b2_carry))\n' "$dst" "$lane" "$src" "$lane" "$msg" "$lane"
        printf '    dryad_b2_v%s_%s=$((dryad_b2_sum & 65535))\n' "$dst" "$lane"
        [ "$lane" = 3 ] || emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    done
}

emit_add3_args () {
    dst=$1
    src=$2
    arg_base=$3

    printf '    dryad_b2_sum=$((dryad_b2_v%s_0 + dryad_b2_v%s_0 + ${%s}))\n' "$dst" "$src" "$arg_base"
    printf '    dryad_b2_v%s_0=$((dryad_b2_sum & 65535))\n' "$dst"
    emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    for lane in 1 2 3; do
        arg=$((arg_base + lane))
        printf '    dryad_b2_sum=$((dryad_b2_v%s_%s + dryad_b2_v%s_%s + ${%s} + dryad_b2_carry))\n' "$dst" "$lane" "$src" "$lane" "$arg"
        printf '    dryad_b2_v%s_%s=$((dryad_b2_sum & 65535))\n' "$dst" "$lane"
        [ "$lane" = 3 ] || emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    done
}

emit_add2 () {
    dst=$1
    src=$2

    printf '    dryad_b2_sum=$((dryad_b2_v%s_0 + dryad_b2_v%s_0))\n' "$dst" "$src"
    printf '    dryad_b2_v%s_0=$((dryad_b2_sum & 65535))\n' "$dst"
    emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    for lane in 1 2 3; do
        printf '    dryad_b2_sum=$((dryad_b2_v%s_%s + dryad_b2_v%s_%s + dryad_b2_carry))\n' "$dst" "$lane" "$src" "$lane"
        printf '    dryad_b2_v%s_%s=$((dryad_b2_sum & 65535))\n' "$dst" "$lane"
        [ "$lane" = 3 ] || emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    done
}

emit_rotr () {
    idx=$1
    bits=$2

    for lane in 0 1 2 3; do
        printf '    dryad_b2_r%s=$dryad_b2_v%s_%s\n' "$lane" "$idx" "$lane"
    done
    case $bits in
        16 )
            printf '    dryad_b2_v%s_0=$dryad_b2_r1\n' "$idx"
            printf '    dryad_b2_v%s_1=$dryad_b2_r2\n' "$idx"
            printf '    dryad_b2_v%s_2=$dryad_b2_r3\n' "$idx"
            printf '    dryad_b2_v%s_3=$dryad_b2_r0\n' "$idx"
            ;;
        32 )
            printf '    dryad_b2_v%s_0=$dryad_b2_r2\n' "$idx"
            printf '    dryad_b2_v%s_1=$dryad_b2_r3\n' "$idx"
            printf '    dryad_b2_v%s_2=$dryad_b2_r0\n' "$idx"
            printf '    dryad_b2_v%s_3=$dryad_b2_r1\n' "$idx"
            ;;
        24 )
            printf '    dryad_b2_v%s_0=$(((dryad_b2_r1 >> 8) | ((dryad_b2_r2 & 255) << 8)))\n' "$idx"
            printf '    dryad_b2_v%s_1=$(((dryad_b2_r2 >> 8) | ((dryad_b2_r3 & 255) << 8)))\n' "$idx"
            printf '    dryad_b2_v%s_2=$(((dryad_b2_r3 >> 8) | ((dryad_b2_r0 & 255) << 8)))\n' "$idx"
            printf '    dryad_b2_v%s_3=$(((dryad_b2_r0 >> 8) | ((dryad_b2_r1 & 255) << 8)))\n' "$idx"
            ;;
        63 )
            printf '    dryad_b2_v%s_0=$((((dryad_b2_r0 << 1) & 65535) | (dryad_b2_r3 >> 15)))\n' "$idx"
            printf '    dryad_b2_v%s_1=$((((dryad_b2_r1 << 1) & 65535) | (dryad_b2_r0 >> 15)))\n' "$idx"
            printf '    dryad_b2_v%s_2=$((((dryad_b2_r2 << 1) & 65535) | (dryad_b2_r1 >> 15)))\n' "$idx"
            printf '    dryad_b2_v%s_3=$((((dryad_b2_r3 << 1) & 65535) | (dryad_b2_r2 >> 15)))\n' "$idx"
            ;;
    esac
}

emit_g () {
    a=$1
    b=$2
    c=$3
    d=$4
    x=$5
    y=$6

    emit_add3 "$a" "$b" "$x"
    emit_xor_word "$d" "$a"
    emit_rotr "$d" 32
    emit_add2 "$c" "$d"
    emit_xor_word "$b" "$c"
    emit_rotr "$b" 24
    emit_add3 "$a" "$b" "$y"
    emit_xor_word "$d" "$a"
    emit_rotr "$d" 16
    emit_add2 "$c" "$d"
    emit_xor_word "$b" "$c"
    emit_rotr "$b" 63
}

emit_round () {
    set -- $1
    emit_g 0 4 8 12 "$1" "$2"
    emit_g 1 5 9 13 "$3" "$4"
    emit_g 2 6 10 14 "$5" "$6"
    emit_g 3 7 11 15 "$7" "$8"
    emit_g 0 5 10 15 "$9" "${10}"
    emit_g 1 6 11 12 "${11}" "${12}"
    emit_g 2 7 8 13 "${13}" "${14}"
    emit_g 3 4 9 14 "${15}" "${16}"
}

emit_g_helper () {
    a=$1
    b=$2
    c=$3
    d=$4

    emit_line "dryad_b2_G_${a}_${b}_${c}_${d} () {"
    emit_add3_args "$a" "$b" 1
    emit_xor_word "$d" "$a"
    emit_rotr "$d" 32
    emit_add2 "$c" "$d"
    emit_xor_word "$b" "$c"
    emit_rotr "$b" 24
    emit_add3_args "$a" "$b" 5
    emit_xor_word "$d" "$a"
    emit_rotr "$d" 16
    emit_add2 "$c" "$d"
    emit_xor_word "$b" "$c"
    emit_rotr "$b" 63
    emit_line '}'
    emit_line ''
}

emit_g_helpers () {
    emit_g_helper 0 4 8 12
    emit_g_helper 1 5 9 13
    emit_g_helper 2 6 10 14
    emit_g_helper 3 7 11 15
    emit_g_helper 0 5 10 15
    emit_g_helper 1 6 11 12
    emit_g_helper 2 7 8 13
    emit_g_helper 3 4 9 14
}

emit_g_call () {
    a=$1
    b=$2
    c=$3
    d=$4
    x=$5
    y=$6

    printf '    dryad_b2_G_%s_%s_%s_%s "$dryad_b2_m%s_0" "$dryad_b2_m%s_1" "$dryad_b2_m%s_2" "$dryad_b2_m%s_3" "$dryad_b2_m%s_0" "$dryad_b2_m%s_1" "$dryad_b2_m%s_2" "$dryad_b2_m%s_3"\n' \
        "$a" "$b" "$c" "$d" "$x" "$x" "$x" "$x" "$y" "$y" "$y" "$y"
}

emit_round_calls () {
    set -- $1
    emit_g_call 0 4 8 12 "$1" "$2"
    emit_g_call 1 5 9 13 "$3" "$4"
    emit_g_call 2 6 10 14 "$5" "$6"
    emit_g_call 3 7 11 15 "$7" "$8"
    emit_g_call 0 5 10 15 "$9" "${10}"
    emit_g_call 1 6 11 12 "${11}" "${12}"
    emit_g_call 2 7 8 13 "${13}" "${14}"
    emit_g_call 3 4 9 14 "${15}" "${16}"
}

emit_load_block () {
    emit_line 'dryad_b2_load_block_shell () {'
    emit_line '    dryad_b2_block_len=$#'
    word=0
    while [ "$word" -lt 16 ]; do
        lane=0
        while [ "$lane" -lt 4 ]; do
            byte_a=$((word * 8 + lane * 2 + 1))
            byte_b=$((byte_a + 1))
            printf '    dryad_b2_m%s_%s=$((%s | (%s << 8)))\n' "$word" "$lane" "\${$byte_a:-0}" "\${$byte_b:-0}"
            lane=$((lane + 1))
        done
        word=$((word + 1))
    done
    emit_line '}'
    emit_line ''
}

emit_load_prefixed_block () {
    name=$1
    prefix_exprs=$2

    emit_line "$name () {"
    emit_line '    if [ "$#" -gt 122 ]; then'
    emit_line '        dryad_b2_block_len=128'
    emit_line '    else'
    emit_line '        dryad_b2_block_len=$((5 + $#))'
    emit_line '    fi'
    emit_line '    if [ "$#" -gt 123 ]; then'
    emit_line '        dryad_b2_carry_len=$(($# - 123))'
    emit_line '    else'
    emit_line '        dryad_b2_carry_len=0'
    emit_line '    fi'
    byte=1
    word=0
    while [ "$word" -lt 16 ]; do
        lane=0
        while [ "$lane" -lt 4 ]; do
            a=$byte
            b=$((byte + 1))
            if [ "$prefix_exprs" = prefix ]; then
                case $a in
                    1 ) expr_a=102 ;;
                    2 ) expr_a=105 ;;
                    3 ) expr_a=108 ;;
                    4 ) expr_a=101 ;;
                    5 ) expr_a=0 ;;
                    * ) expr_a="\${$((a - 5)):-0}" ;;
                esac
                case $b in
                    1 ) expr_b=102 ;;
                    2 ) expr_b=105 ;;
                    3 ) expr_b=108 ;;
                    4 ) expr_b=101 ;;
                    5 ) expr_b=0 ;;
                    * ) expr_b="\${$((b - 5)):-0}" ;;
                esac
            else
                case $a in
                    1 ) expr_a='$dryad_b2_c0' ;;
                    2 ) expr_a='$dryad_b2_c1' ;;
                    3 ) expr_a='$dryad_b2_c2' ;;
                    4 ) expr_a='$dryad_b2_c3' ;;
                    5 ) expr_a='$dryad_b2_c4' ;;
                    * ) expr_a="\${$((a - 5)):-0}" ;;
                esac
                case $b in
                    1 ) expr_b='$dryad_b2_c0' ;;
                    2 ) expr_b='$dryad_b2_c1' ;;
                    3 ) expr_b='$dryad_b2_c2' ;;
                    4 ) expr_b='$dryad_b2_c3' ;;
                    5 ) expr_b='$dryad_b2_c4' ;;
                    * ) expr_b="\${$((b - 5)):-0}" ;;
                esac
            fi
            printf '    dryad_b2_m%s_%s=$((%s | (%s << 8)))\n' "$word" "$lane" "$expr_a" "$expr_b"
            byte=$((byte + 2))
            lane=$((lane + 1))
        done
        word=$((word + 1))
    done
    emit_line '    dryad_b2_c0=${124:-0}'
    emit_line '    dryad_b2_c1=${125:-0}'
    emit_line '    dryad_b2_c2=${126:-0}'
    emit_line '    dryad_b2_c3=${127:-0}'
    emit_line '    dryad_b2_c4=${128:-0}'
    emit_line '}'
    emit_line ''
}

emit_load_final_carry_block () {
    emit_line 'dryad_b2_load_final_carry_block_shell () {'
    emit_line '    dryad_b2_block_len=$dryad_b2_carry_len'
    emit_line '    dryad_b2_m0_0=$((dryad_b2_c0 | (dryad_b2_c1 << 8)))'
    emit_line '    dryad_b2_m0_1=$((dryad_b2_c2 | (dryad_b2_c3 << 8)))'
    emit_line '    dryad_b2_m0_2=$dryad_b2_c4'
    emit_line '    dryad_b2_m0_3=0'
    word=1
    while [ "$word" -lt 16 ]; do
        lane=0
        while [ "$lane" -lt 4 ]; do
            printf '    dryad_b2_m%s_%s=0\n' "$word" "$lane"
            lane=$((lane + 1))
        done
        word=$((word + 1))
    done
    emit_line '}'
    emit_line ''
}

emit_reset () {
    emit_line 'dryad_b2_reset_shell () {'
    emit_word_assign iv 0 51464 62396 58983 27145
    emit_word_assign iv 1 42811 33994 44677 47975
    emit_word_assign iv 2 63531 65172 62322 15470
    emit_word_assign iv 3 14065 24349 62778 42319
    emit_word_assign iv 4 33489 44518 21119 20750
    emit_word_assign iv 5 27679 11070 26764 39685
    emit_word_assign iv 6 48491 64321 55723 8067
    emit_word_assign iv 7 8569 4990 52505 23520
    for idx in 0 1 2 3 4 5 6 7; do
        emit_copy_word h "$idx" iv "$idx"
    done
    emit_line '    dryad_b2_h0_0=$((dryad_b2_h0_0 ^ 16))'
    emit_line '    dryad_b2_h0_1=$((dryad_b2_h0_1 ^ 257))'
    emit_line '    dryad_b2_t0=0'
    emit_line '    dryad_b2_t1=0'
    emit_line '    dryad_b2_t2=0'
    emit_line '    dryad_b2_t3=0'
    emit_line '}'
    emit_line ''
}

emit_add_counter () {
    emit_line 'dryad_b2_add_counter_shell () {'
    emit_line '    dryad_b2_counter_add=$1'
    emit_line '    dryad_b2_sum=$((dryad_b2_t0 + dryad_b2_counter_add))'
    emit_line '    dryad_b2_t0=$((dryad_b2_sum & 65535))'
    emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    emit_line '    dryad_b2_sum=$((dryad_b2_t1 + dryad_b2_carry))'
    emit_line '    dryad_b2_t1=$((dryad_b2_sum & 65535))'
    emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    emit_line '    dryad_b2_sum=$((dryad_b2_t2 + dryad_b2_carry))'
    emit_line '    dryad_b2_t2=$((dryad_b2_sum & 65535))'
    emit_line '    dryad_b2_carry=$((dryad_b2_sum >> 16))'
    emit_line '    dryad_b2_t3=$(((dryad_b2_t3 + dryad_b2_carry) & 65535))'
    emit_line '}'
    emit_line ''
}

emit_compress () {
    emit_line 'dryad_b2_compress_shell () {'
    emit_line '    dryad_b2_final=$1'
    for idx in 0 1 2 3 4 5 6 7; do
        emit_copy_word v "$idx" h "$idx"
    done
    for idx in 0 1 2 3 4 5 6 7; do
        emit_copy_word v "$((idx + 8))" iv "$idx"
    done
    emit_xor_counter 12
    emit_line '    if [ "$dryad_b2_final" = 1 ]; then'
    emit_xor_mask 14
    emit_line '    fi'
    emit_round_calls '0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15'
    emit_round_calls '14 10 4 8 9 15 13 6 1 12 0 2 11 7 5 3'
    emit_round_calls '11 8 12 0 5 2 15 13 10 14 3 6 7 1 9 4'
    emit_round_calls '7 9 3 1 13 12 11 14 2 6 5 10 4 0 15 8'
    emit_round_calls '9 0 5 7 2 4 10 15 14 1 11 12 6 8 3 13'
    emit_round_calls '2 12 6 10 0 11 8 3 4 13 7 5 15 14 1 9'
    emit_round_calls '12 5 1 15 14 13 4 10 0 7 6 3 9 2 8 11'
    emit_round_calls '13 11 7 14 12 1 3 9 5 0 15 4 8 6 2 10'
    emit_round_calls '6 15 14 9 11 3 0 8 12 2 13 7 1 4 10 5'
    emit_round_calls '10 2 8 4 7 6 1 5 15 11 9 14 3 12 13 0'
    emit_round_calls '0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15'
    emit_round_calls '14 10 4 8 9 15 13 6 1 12 0 2 11 7 5 3'
    for idx in 0 1 2 3 4 5 6 7; do
        for lane in 0 1 2 3; do
            printf '    dryad_b2_h%s_%s=$((dryad_b2_h%s_%s ^ dryad_b2_v%s_%s ^ dryad_b2_v%s_%s))\n' "$idx" "$lane" "$idx" "$lane" "$idx" "$lane" "$((idx + 8))" "$lane"
        done
    done
    emit_line '}'
    emit_line ''
}

emit_outputs () {
    emit_line 'dryad_b2_output_hex_shell () {'
    printf "    printf '%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x\\n'"
    for word in 0 1; do
        for lane in 0 1 2 3; do
            printf ' $((dryad_b2_h%s_%s & 255)) $((dryad_b2_h%s_%s >> 8))' "$word" "$lane" "$word" "$lane"
        done
    done
    printf '\n'
    emit_line '}'
    emit_line ''
    emit_line 'dryad_b2_output_base32_shell () {'
    byte=0
    for word in 0 1; do
        for lane in 0 1 2 3; do
            printf '    dryad_b2_d%s=$((dryad_b2_h%s_%s & 255))\n' "$byte" "$word" "$lane"
            byte=$((byte + 1))
            printf '    dryad_b2_d%s=$((dryad_b2_h%s_%s >> 8))\n' "$byte" "$word" "$lane"
            byte=$((byte + 1))
        done
    done
    group=0
    while [ "$group" -lt 3 ]; do
        b=$((group * 5))
        printf '    dryad_base32_char $((dryad_b2_d%s >> 3))\n' "$b"
        printf '    dryad_base32_char $((((dryad_b2_d%s & 7) << 2) | (dryad_b2_d%s >> 6)))\n' "$b" "$((b + 1))"
        printf '    dryad_base32_char $(((dryad_b2_d%s >> 1) & 31))\n' "$((b + 1))"
        printf '    dryad_base32_char $((((dryad_b2_d%s & 1) << 4) | (dryad_b2_d%s >> 4)))\n' "$((b + 1))" "$((b + 2))"
        printf '    dryad_base32_char $((((dryad_b2_d%s & 15) << 1) | (dryad_b2_d%s >> 7)))\n' "$((b + 2))" "$((b + 3))"
        printf '    dryad_base32_char $(((dryad_b2_d%s >> 2) & 31))\n' "$((b + 3))"
        printf '    dryad_base32_char $((((dryad_b2_d%s & 3) << 3) | (dryad_b2_d%s >> 5)))\n' "$((b + 3))" "$((b + 4))"
        printf '    dryad_base32_char $((dryad_b2_d%s & 31))\n' "$((b + 4))"
        group=$((group + 1))
    done
    emit_line '    dryad_base32_char $((dryad_b2_d15 >> 3))'
    emit_line '    dryad_base32_char $(((dryad_b2_d15 & 7) << 2))'
    emit_line '    printf '\''\n'\'''
    emit_line '}'
    emit_line ''
}

emit_hash_functions () {
    emit_line 'dryad_b2_hash_file_shell () {'
    emit_line '    dryad_b2_file=$1'
    emit_line '    dryad_b2_format=${2:-hex}'
    emit_line '    dryad_b2_od_tmp=${TMPDIR:-/tmp}/dryad-sh-hash-od.$$'
    emit_line '    rm -f "$dryad_b2_od_tmp"'
    emit_line '    if ! od -An -v -tu1 -w128 "$dryad_b2_file" > "$dryad_b2_od_tmp"; then'
    emit_line '        rm -f "$dryad_b2_od_tmp"'
    emit_line '        return 1'
    emit_line '    fi'
    emit_line '    dryad_b2_reset_shell'
    emit_line '    dryad_b2_have_block=0'
    emit_line '    dryad_b2_pending_line='
    emit_line '    while IFS= read -r dryad_b2_line; do'
    emit_line '        [ -n "$dryad_b2_line" ] || continue'
    emit_line '        if [ "$dryad_b2_have_block" = 1 ]; then'
    emit_line '            set -- $dryad_b2_pending_line'
    emit_line '            dryad_b2_load_block_shell "$@"'
    emit_line '            dryad_b2_add_counter_shell "$dryad_b2_block_len"'
    emit_line '            dryad_b2_compress_shell 0'
    emit_line '        fi'
    emit_line '        dryad_b2_pending_line=$dryad_b2_line'
    emit_line '        dryad_b2_have_block=1'
    emit_line '    done < "$dryad_b2_od_tmp"'
    emit_line '    rm -f "$dryad_b2_od_tmp"'
    emit_line '    if [ "$dryad_b2_have_block" = 1 ]; then'
    emit_line '        set -- $dryad_b2_pending_line'
    emit_line '        dryad_b2_load_block_shell "$@"'
    emit_line '    else'
    emit_line '        dryad_b2_load_block_shell'
    emit_line '    fi'
    emit_line '    dryad_b2_add_counter_shell "$dryad_b2_block_len"'
    emit_line '    dryad_b2_compress_shell 1'
    emit_line '    case $dryad_b2_format in'
    emit_line '        base32 ) dryad_b2_output_base32_shell ;;'
    emit_line '        hex ) dryad_b2_output_hex_shell ;;'
    emit_line '        * ) dryad_die "unsupported shell hash format: $dryad_b2_format" ;;'
    emit_line '    esac'
    emit_line '}'
    emit_line ''
    emit_line 'dryad_blake2b_128_file_hex_shell () {'
    emit_line '    dryad_b2_hash_file_shell "$1" hex'
    emit_line '}'
    emit_line ''
    emit_line 'dryad_blake2b_128_file_base32_shell () {'
    emit_line '    dryad_b2_hash_file_shell "$1" base32'
    emit_line '}'
    emit_line ''
    emit_line 'dryad_blake2b_128_file_prefixed_base32_shell () {'
    emit_line '    dryad_b2_prefixed_file=$1'
    emit_line '    dryad_b2_od_tmp=${TMPDIR:-/tmp}/dryad-sh-hash-od.$$'
    emit_line '    rm -f "$dryad_b2_od_tmp"'
    emit_line '    if ! od -An -v -tu1 -w128 "$dryad_b2_prefixed_file" > "$dryad_b2_od_tmp"; then'
    emit_line '        rm -f "$dryad_b2_od_tmp"'
    emit_line '        return 1'
    emit_line '    fi'
    emit_line '    dryad_b2_reset_shell'
    emit_line '    dryad_b2_have_block=0'
    emit_line '    dryad_b2_loaded_block=0'
    emit_line '    dryad_b2_pending_line='
    emit_line '    while IFS= read -r dryad_b2_line; do'
    emit_line '        [ -n "$dryad_b2_line" ] || continue'
    emit_line '        if [ "$dryad_b2_have_block" = 1 ]; then'
    emit_line '            set -- $dryad_b2_pending_line'
    emit_line '            if [ "$dryad_b2_loaded_block" = 0 ]; then'
    emit_line '                dryad_b2_load_initial_prefixed_block_shell "$@"'
    emit_line '            else'
    emit_line '                dryad_b2_load_carried_prefixed_block_shell "$@"'
    emit_line '            fi'
    emit_line '            dryad_b2_add_counter_shell "$dryad_b2_block_len"'
    emit_line '            dryad_b2_compress_shell 0'
    emit_line '            dryad_b2_loaded_block=1'
    emit_line '        fi'
    emit_line '        dryad_b2_pending_line=$dryad_b2_line'
    emit_line '        dryad_b2_have_block=1'
    emit_line '    done < "$dryad_b2_od_tmp"'
    emit_line '    rm -f "$dryad_b2_od_tmp"'
    emit_line '    if [ "$dryad_b2_have_block" = 1 ]; then'
    emit_line '        set -- $dryad_b2_pending_line'
    emit_line '        if [ "$dryad_b2_loaded_block" = 0 ]; then'
    emit_line '            dryad_b2_load_initial_prefixed_block_shell "$@"'
    emit_line '        else'
    emit_line '            dryad_b2_load_carried_prefixed_block_shell "$@"'
    emit_line '        fi'
    emit_line '    else'
    emit_line '        dryad_b2_load_initial_prefixed_block_shell'
    emit_line '    fi'
    emit_line '    if [ "$dryad_b2_carry_len" -gt 0 ]; then'
    emit_line '        dryad_b2_add_counter_shell "$dryad_b2_block_len"'
    emit_line '        dryad_b2_compress_shell 0'
    emit_line '        dryad_b2_load_final_carry_block_shell'
    emit_line '    fi'
    emit_line '    dryad_b2_add_counter_shell "$dryad_b2_block_len"'
    emit_line '    dryad_b2_compress_shell 1'
    emit_line '    dryad_b2_output_base32_shell'
    emit_line '}'
}

emit_line '# Generated by dyd/assets/tools/generate-hash-shell.sh. Do not edit by hand.'
emit_line ''
emit_load_block
emit_load_prefixed_block dryad_b2_load_initial_prefixed_block_shell prefix
emit_load_prefixed_block dryad_b2_load_carried_prefixed_block_shell carry
emit_load_final_carry_block
emit_reset
emit_add_counter
emit_g_helpers
emit_compress
emit_outputs
emit_hash_functions
