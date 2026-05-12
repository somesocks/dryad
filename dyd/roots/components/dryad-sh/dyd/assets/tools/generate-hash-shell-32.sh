#!/usr/bin/env sh

set -eu

p=dyd_b2_32

emit_line () {
    printf '%s\n' "$1"
}

emit_word_assign () {
    prefix=$1
    idx=$2
    lo=$3
    hi=$4

    printf '    %s_%s%s_0=%s\n' "$p" "$prefix" "$idx" "$lo"
    printf '    %s_%s%s_1=%s\n' "$p" "$prefix" "$idx" "$hi"
}

emit_copy_word () {
    dst_prefix=$1
    dst=$2
    src_prefix=$3
    src=$4

    printf '    %s_%s%s_0=$%s_%s%s_0\n' "$p" "$dst_prefix" "$dst" "$p" "$src_prefix" "$src"
    printf '    %s_%s%s_1=$%s_%s%s_1\n' "$p" "$dst_prefix" "$dst" "$p" "$src_prefix" "$src"
}

emit_xor_word () {
    dst=$1
    src=$2

    printf '    %s_v%s_0=$((%s_v%s_0 ^ %s_v%s_0))\n' "$p" "$dst" "$p" "$dst" "$p" "$src"
    printf '    %s_v%s_1=$((%s_v%s_1 ^ %s_v%s_1))\n' "$p" "$dst" "$p" "$dst" "$p" "$src"
}

emit_xor_counter () {
    dst=$1

    printf '    %s_v%s_0=$((%s_v%s_0 ^ %s_t0))\n' "$p" "$dst" "$p" "$dst" "$p"
    printf '    %s_v%s_1=$((%s_v%s_1 ^ %s_t1))\n' "$p" "$dst" "$p" "$dst" "$p"
}

emit_xor_mask () {
    dst=$1

    printf '        %s_v%s_0=$((%s_v%s_0 ^ %s_mask))\n' "$p" "$dst" "$p" "$dst" "$p"
    printf '        %s_v%s_1=$((%s_v%s_1 ^ %s_mask))\n' "$p" "$dst" "$p" "$dst" "$p"
}

emit_add3_args () {
    dst=$1
    src=$2
    arg_lo=$3
    arg_hi=$4

    printf '    %s_sum=$((%s_v%s_0 + %s_v%s_0 + ${%s}))\n' "$p" "$p" "$dst" "$p" "$src" "$arg_lo"
    printf '    %s_v%s_0=$((%s_sum & %s_mask))\n' "$p" "$dst" "$p" "$p"
    printf '    %s_carry=$((%s_sum >> 32))\n' "$p" "$p"
    printf '    %s_v%s_1=$(((%s_v%s_1 + %s_v%s_1 + ${%s} + %s_carry) & %s_mask))\n' "$p" "$dst" "$p" "$dst" "$p" "$src" "$arg_hi" "$p" "$p"
}

emit_add2 () {
    dst=$1
    src=$2

    printf '    %s_sum=$((%s_v%s_0 + %s_v%s_0))\n' "$p" "$p" "$dst" "$p" "$src"
    printf '    %s_v%s_0=$((%s_sum & %s_mask))\n' "$p" "$dst" "$p" "$p"
    printf '    %s_carry=$((%s_sum >> 32))\n' "$p" "$p"
    printf '    %s_v%s_1=$(((%s_v%s_1 + %s_v%s_1 + %s_carry) & %s_mask))\n' "$p" "$dst" "$p" "$dst" "$p" "$src" "$p" "$p"
}

emit_rotr () {
    idx=$1
    bits=$2

    printf '    %s_r0=$%s_v%s_0\n' "$p" "$p" "$idx"
    printf '    %s_r1=$%s_v%s_1\n' "$p" "$p" "$idx"
    case $bits in
        16 | 24 )
            printf '    %s_v%s_0=$(((%s_r0 >> %s) | ((%s_r1 << %s) & %s_mask)))\n' "$p" "$idx" "$p" "$bits" "$p" "$((32 - bits))" "$p"
            printf '    %s_v%s_1=$(((%s_r1 >> %s) | ((%s_r0 << %s) & %s_mask)))\n' "$p" "$idx" "$p" "$bits" "$p" "$((32 - bits))" "$p"
            ;;
        32 )
            printf '    %s_v%s_0=$%s_r1\n' "$p" "$idx" "$p"
            printf '    %s_v%s_1=$%s_r0\n' "$p" "$idx" "$p"
            ;;
        63 )
            printf '    %s_v%s_0=$((((%s_r0 << 1) & %s_mask) | (%s_r1 >> 31)))\n' "$p" "$idx" "$p" "$p" "$p"
            printf '    %s_v%s_1=$((((%s_r1 << 1) & %s_mask) | (%s_r0 >> 31)))\n' "$p" "$idx" "$p" "$p" "$p"
            ;;
    esac
}

emit_g_helper () {
    a=$1
    b=$2
    c=$3
    d=$4

    emit_line "${p}_G_${a}_${b}_${c}_${d} () {"
    emit_add3_args "$a" "$b" 1 2
    emit_xor_word "$d" "$a"
    emit_rotr "$d" 32
    emit_add2 "$c" "$d"
    emit_xor_word "$b" "$c"
    emit_rotr "$b" 24
    emit_add3_args "$a" "$b" 3 4
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

    printf '    %s_G_%s_%s_%s_%s "$%s_m%s_0" "$%s_m%s_1" "$%s_m%s_0" "$%s_m%s_1"\n' \
        "$p" "$a" "$b" "$c" "$d" "$p" "$x" "$p" "$x" "$p" "$y" "$p" "$y"
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
    emit_line "${p}_load_block_shell () {"
    emit_line "    ${p}_block_len=\$#"
    word=0
    while [ "$word" -lt 16 ]; do
        lo0=$((word * 8 + 1))
        lo1=$((lo0 + 1))
        lo2=$((lo0 + 2))
        lo3=$((lo0 + 3))
        hi0=$((lo0 + 4))
        hi1=$((lo0 + 5))
        hi2=$((lo0 + 6))
        hi3=$((lo0 + 7))
        printf '    %s_m%s_0=$((%s | (%s << 8) | (%s << 16) | (%s << 24)))\n' "$p" "$word" "\${$lo0:-0}" "\${$lo1:-0}" "\${$lo2:-0}" "\${$lo3:-0}"
        printf '    %s_m%s_1=$((%s | (%s << 8) | (%s << 16) | (%s << 24)))\n' "$p" "$word" "\${$hi0:-0}" "\${$hi1:-0}" "\${$hi2:-0}" "\${$hi3:-0}"
        word=$((word + 1))
    done
    emit_line '}'
    emit_line ''
}

prefixed_expr () {
    mode=$1
    byte=$2

    if [ "$mode" = prefix ]; then
        case $byte in
            1 ) printf 102 ;;
            2 ) printf 105 ;;
            3 ) printf 108 ;;
            4 ) printf 101 ;;
            5 ) printf 0 ;;
            * ) printf '${%s:-0}' "$((byte - 5))" ;;
        esac
    else
        case $byte in
            1 ) printf '$%s_c0' "$p" ;;
            2 ) printf '$%s_c1' "$p" ;;
            3 ) printf '$%s_c2' "$p" ;;
            4 ) printf '$%s_c3' "$p" ;;
            5 ) printf '$%s_c4' "$p" ;;
            * ) printf '${%s:-0}' "$((byte - 5))" ;;
        esac
    fi
}

emit_load_prefixed_block () {
    name=$1
    mode=$2

    emit_line "$name () {"
    emit_line '    if [ "$#" -gt 122 ]; then'
    emit_line "        ${p}_block_len=128"
    emit_line '    else'
    emit_line "        ${p}_block_len=\$((5 + \$#))"
    emit_line '    fi'
    emit_line '    if [ "$#" -gt 123 ]; then'
    emit_line "        ${p}_carry_len=\$((\$# - 123))"
    emit_line '    else'
    emit_line "        ${p}_carry_len=0"
    emit_line '    fi'

    byte=1
    word=0
    while [ "$word" -lt 16 ]; do
        a=$(prefixed_expr "$mode" "$byte")
        b=$(prefixed_expr "$mode" "$((byte + 1))")
        c=$(prefixed_expr "$mode" "$((byte + 2))")
        d=$(prefixed_expr "$mode" "$((byte + 3))")
        e=$(prefixed_expr "$mode" "$((byte + 4))")
        f=$(prefixed_expr "$mode" "$((byte + 5))")
        g=$(prefixed_expr "$mode" "$((byte + 6))")
        h=$(prefixed_expr "$mode" "$((byte + 7))")
        printf '    %s_m%s_0=$((%s | (%s << 8) | (%s << 16) | (%s << 24)))\n' "$p" "$word" "$a" "$b" "$c" "$d"
        printf '    %s_m%s_1=$((%s | (%s << 8) | (%s << 16) | (%s << 24)))\n' "$p" "$word" "$e" "$f" "$g" "$h"
        byte=$((byte + 8))
        word=$((word + 1))
    done
    printf '    %s_c0=${124:-0}\n' "$p"
    printf '    %s_c1=${125:-0}\n' "$p"
    printf '    %s_c2=${126:-0}\n' "$p"
    printf '    %s_c3=${127:-0}\n' "$p"
    printf '    %s_c4=${128:-0}\n' "$p"
    emit_line '}'
    emit_line ''
}

emit_load_final_carry_block () {
    emit_line "${p}_load_final_carry_block_shell () {"
    printf '    %s_block_len=$%s_carry_len\n' "$p" "$p"
    printf '    %s_m0_0=$(($%s_c0 | ($%s_c1 << 8) | ($%s_c2 << 16) | ($%s_c3 << 24)))\n' "$p" "$p" "$p" "$p" "$p"
    printf '    %s_m0_1=$%s_c4\n' "$p" "$p"
    word=1
    while [ "$word" -lt 16 ]; do
        printf '    %s_m%s_0=0\n' "$p" "$word"
        printf '    %s_m%s_1=0\n' "$p" "$word"
        word=$((word + 1))
    done
    emit_line '}'
    emit_line ''
}

emit_reset () {
    emit_line "${p}_reset_shell () {"
    emit_line "    ${p}_mask=4294967295"
    emit_word_assign iv 0 4089235720 1779033703
    emit_word_assign iv 1 2227873595 3144134277
    emit_word_assign iv 2 4271175723 1013904242
    emit_word_assign iv 3 1595750129 2773480762
    emit_word_assign iv 4 2917565137 1359893119
    emit_word_assign iv 5 725511199 2600822924
    emit_word_assign iv 6 4215389547 528734635
    emit_word_assign iv 7 327033209 1541459225
    for idx in 0 1 2 3 4 5 6 7; do
        emit_copy_word h "$idx" iv "$idx"
    done
    printf '    %s_h0_0=$((%s_h0_0 ^ 16842768))\n' "$p" "$p"
    printf '    %s_t0=0\n' "$p"
    printf '    %s_t1=0\n' "$p"
    emit_line '}'
    emit_line ''
}

emit_add_counter () {
    emit_line "${p}_add_counter_shell () {"
    printf '    %s_sum=$((%s_t0 + $1))\n' "$p" "$p"
    printf '    %s_t0=$((%s_sum & %s_mask))\n' "$p" "$p" "$p"
    printf '    %s_t1=$(((%s_t1 + (%s_sum >> 32)) & %s_mask))\n' "$p" "$p" "$p" "$p"
    emit_line '}'
    emit_line ''
}

emit_compress () {
    emit_line "${p}_compress_shell () {"
    printf '    %s_final=$1\n' "$p"
    for idx in 0 1 2 3 4 5 6 7; do
        emit_copy_word v "$idx" h "$idx"
    done
    for idx in 0 1 2 3 4 5 6 7; do
        emit_copy_word v "$((idx + 8))" iv "$idx"
    done
    emit_xor_counter 12
    printf '    if [ "$%s_final" = 1 ]; then\n' "$p"
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
        printf '    %s_h%s_0=$((%s_h%s_0 ^ %s_v%s_0 ^ %s_v%s_0))\n' "$p" "$idx" "$p" "$idx" "$p" "$idx" "$p" "$((idx + 8))"
        printf '    %s_h%s_1=$((%s_h%s_1 ^ %s_v%s_1 ^ %s_v%s_1))\n' "$p" "$idx" "$p" "$idx" "$p" "$idx" "$p" "$((idx + 8))"
    done
    emit_line '}'
    emit_line ''
}

emit_outputs () {
    emit_line "${p}_hex_nibble_shell_load () {"
    emit_line '    case $1 in'
    emit_line '        0 ) dyd_ret0=0 ;; 1 ) dyd_ret0=1 ;; 2 ) dyd_ret0=2 ;; 3 ) dyd_ret0=3 ;;'
    emit_line '        4 ) dyd_ret0=4 ;; 5 ) dyd_ret0=5 ;; 6 ) dyd_ret0=6 ;; 7 ) dyd_ret0=7 ;;'
    emit_line '        8 ) dyd_ret0=8 ;; 9 ) dyd_ret0=9 ;; 10 ) dyd_ret0=a ;; 11 ) dyd_ret0=b ;;'
    emit_line '        12 ) dyd_ret0=c ;; 13 ) dyd_ret0=d ;; 14 ) dyd_ret0=e ;; 15 ) dyd_ret0=f ;;'
    emit_line '        * ) dryad_die "invalid hex nibble: $1" ;;'
    emit_line '    esac'
    emit_line '}'
    emit_line ''
    emit_line "${p}_hex_byte_shell_append () {"
    printf '    %s_hex_byte=$1\n' "$p"
    printf '    %s_hex_nibble_shell_load $(($%s_hex_byte >> 4))\n' "$p" "$p"
    printf '    %s_hex_output=$%s_hex_output$dyd_ret0\n' "$p" "$p"
    printf '    %s_hex_nibble_shell_load $(($%s_hex_byte & 15))\n' "$p" "$p"
    printf '    %s_hex_output=$%s_hex_output$dyd_ret0\n' "$p" "$p"
    emit_line '}'
    emit_line ''
    emit_line "${p}_output_hex_shell_load () {"
    printf '    %s_hex_output=\n' "$p"
    for word in 0 1; do
        printf '    %s_hex_byte_shell_append $((%s_h%s_0 & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $(((%s_h%s_0 >> 8) & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $(((%s_h%s_0 >> 16) & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $(((%s_h%s_0 >> 24) & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $((%s_h%s_1 & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $(((%s_h%s_1 >> 8) & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $(((%s_h%s_1 >> 16) & 255))\n' "$p" "$p" "$word"
        printf '    %s_hex_byte_shell_append $(((%s_h%s_1 >> 24) & 255))\n' "$p" "$p" "$word"
    done
    printf '    dyd_ret0=$%s_hex_output\n' "$p"
    emit_line '}'
    emit_line ''
    emit_line "${p}_base32_char_shell_append () {"
    emit_line '    dryad_base32_char_load "$1"'
    printf '    %s_base32_output=$%s_base32_output$dyd_ret0\n' "$p" "$p"
    emit_line '}'
    emit_line ''
    emit_line "${p}_output_base32_shell_load () {"
    byte=0
    for word in 0 1; do
        printf '    %s_d%s=$((%s_h%s_0 & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$(((%s_h%s_0 >> 8) & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$(((%s_h%s_0 >> 16) & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$(((%s_h%s_0 >> 24) & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$((%s_h%s_1 & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$(((%s_h%s_1 >> 8) & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$(((%s_h%s_1 >> 16) & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
        printf '    %s_d%s=$(((%s_h%s_1 >> 24) & 255))\n' "$p" "$byte" "$p" "$word"; byte=$((byte + 1))
    done
    printf '    %s_base32_output=\n' "$p"
    printf '    %s_base32_char_shell_append $((%s_d0 >> 3))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d0 & 7) << 2) | (%s_d1 >> 6)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d1 >> 1) & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d1 & 1) << 4) | (%s_d2 >> 4)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d2 & 15) << 1) | (%s_d3 >> 7)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d3 >> 2) & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d3 & 3) << 3) | (%s_d4 >> 5)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $((%s_d4 & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((%s_d5 >> 3))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d5 & 7) << 2) | (%s_d6 >> 6)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d6 >> 1) & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d6 & 1) << 4) | (%s_d7 >> 4)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d7 & 15) << 1) | (%s_d8 >> 7)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d8 >> 2) & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d8 & 3) << 3) | (%s_d9 >> 5)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $((%s_d9 & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((%s_d10 >> 3))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d10 & 7) << 2) | (%s_d11 >> 6)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d11 >> 1) & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d11 & 1) << 4) | (%s_d12 >> 4)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d12 & 15) << 1) | (%s_d13 >> 7)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d13 >> 2) & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((((%s_d13 & 3) << 3) | (%s_d14 >> 5)))\n' "$p" "$p" "$p"
    printf '    %s_base32_char_shell_append $((%s_d14 & 31))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $((%s_d15 >> 3))\n' "$p" "$p"
    printf '    %s_base32_char_shell_append $(((%s_d15 & 7) << 2))\n' "$p" "$p"
    printf '    dyd_ret0=$%s_base32_output\n' "$p"
    emit_line '}'
    emit_line ''
}

emit_hash_functions () {
    emit_line "${p}_hash_file_shell () {"
    printf '    %s_file=$1\n' "$p"
    printf '    %s_format=${2:-hex}\n' "$p"
    printf '    %s_od_tmp=${TMPDIR:-/tmp}/dryad-sh-hash-32-od.$$\n' "$p"
    printf '    rm -f "$%s_od_tmp"\n' "$p"
    printf '    if ! od -An -v -tu1 -w128 "$%s_file" > "$%s_od_tmp"; then\n' "$p" "$p"
    printf '        rm -f "$%s_od_tmp"\n' "$p"
    emit_line '        return 1'
    emit_line '    fi'
    printf '    %s_reset_shell\n' "$p"
    printf '    %s_have_block=0\n' "$p"
    printf '    %s_pending_line=\n' "$p"
    printf '    while IFS= read -r %s_line; do\n' "$p"
    printf '        [ -n "$%s_line" ] || continue\n' "$p"
    printf '        if [ "$%s_have_block" = 1 ]; then\n' "$p"
    printf '            set -- $%s_pending_line\n' "$p"
    printf '            %s_load_block_shell "$@"\n' "$p"
    printf '            %s_add_counter_shell "$%s_block_len"\n' "$p" "$p"
    printf '            %s_compress_shell 0\n' "$p"
    emit_line '        fi'
    printf '        %s_pending_line=$%s_line\n' "$p" "$p"
    printf '        %s_have_block=1\n' "$p"
    printf '    done < "$%s_od_tmp"\n' "$p"
    printf '    rm -f "$%s_od_tmp"\n' "$p"
    printf '    if [ "$%s_have_block" = 1 ]; then\n' "$p"
    printf '        set -- $%s_pending_line\n' "$p"
    printf '        %s_load_block_shell "$@"\n' "$p"
    emit_line '    else'
    printf '        %s_load_block_shell\n' "$p"
    emit_line '    fi'
    printf '    %s_add_counter_shell "$%s_block_len"\n' "$p" "$p"
    printf '    %s_compress_shell 1\n' "$p"
    printf '    case $%s_format in\n' "$p"
    printf '        base32 ) %s_output_base32_shell_load ;;\n' "$p"
    printf '        hex ) %s_output_hex_shell_load ;;\n' "$p"
    printf '        * ) dryad_die "unsupported 32-bit shell hash format: $%s_format" ;;\n' "$p"
    emit_line '    esac'
    emit_line '}'
    emit_line ''
    emit_line 'dryad_blake2b_128_file_hex_shell_32_load () {'
    printf '    %s_hash_file_shell "$1" hex\n' "$p"
    emit_line '}'
    emit_line ''
    emit_line 'dryad_blake2b_128_file_base32_shell_32_load () {'
    printf '    %s_hash_file_shell "$1" base32\n' "$p"
    emit_line '}'
    emit_line ''
    emit_line 'dryad_blake2b_128_file_prefixed_base32_shell_32_load () {'
    printf '    %s_prefixed_file=$1\n' "$p"
    printf '    %s_od_tmp=${TMPDIR:-/tmp}/dryad-sh-hash-32-od.$$\n' "$p"
    printf '    rm -f "$%s_od_tmp"\n' "$p"
    printf '    if ! od -An -v -tu1 -w128 "$%s_prefixed_file" > "$%s_od_tmp"; then\n' "$p" "$p"
    printf '        rm -f "$%s_od_tmp"\n' "$p"
    emit_line '        return 1'
    emit_line '    fi'
    printf '    %s_reset_shell\n' "$p"
    printf '    %s_have_block=0\n' "$p"
    printf '    %s_loaded_block=0\n' "$p"
    printf '    %s_pending_line=\n' "$p"
    printf '    while IFS= read -r %s_line; do\n' "$p"
    printf '        [ -n "$%s_line" ] || continue\n' "$p"
    printf '        if [ "$%s_have_block" = 1 ]; then\n' "$p"
    printf '            set -- $%s_pending_line\n' "$p"
    printf '            if [ "$%s_loaded_block" = 0 ]; then\n' "$p"
    printf '                %s_load_initial_prefixed_block_shell "$@"\n' "$p"
    emit_line '            else'
    printf '                %s_load_carried_prefixed_block_shell "$@"\n' "$p"
    emit_line '            fi'
    printf '            %s_add_counter_shell "$%s_block_len"\n' "$p" "$p"
    printf '            %s_compress_shell 0\n' "$p"
    printf '            %s_loaded_block=1\n' "$p"
    emit_line '        fi'
    printf '        %s_pending_line=$%s_line\n' "$p" "$p"
    printf '        %s_have_block=1\n' "$p"
    printf '    done < "$%s_od_tmp"\n' "$p"
    printf '    rm -f "$%s_od_tmp"\n' "$p"
    printf '    if [ "$%s_have_block" = 1 ]; then\n' "$p"
    printf '        set -- $%s_pending_line\n' "$p"
    printf '        if [ "$%s_loaded_block" = 0 ]; then\n' "$p"
    printf '            %s_load_initial_prefixed_block_shell "$@"\n' "$p"
    emit_line '        else'
    printf '            %s_load_carried_prefixed_block_shell "$@"\n' "$p"
    emit_line '        fi'
    emit_line '    else'
    printf '        %s_load_initial_prefixed_block_shell\n' "$p"
    emit_line '    fi'
    printf '    if [ "$%s_carry_len" -gt 0 ]; then\n' "$p"
    printf '        %s_add_counter_shell "$%s_block_len"\n' "$p" "$p"
    printf '        %s_compress_shell 0\n' "$p"
    printf '        %s_load_final_carry_block_shell\n' "$p"
    emit_line '    fi'
    printf '    %s_add_counter_shell "$%s_block_len"\n' "$p" "$p"
    printf '    %s_compress_shell 1\n' "$p"
    printf '    %s_output_base32_shell_load\n' "$p"
    emit_line '}'
}

emit_line '# Generated by dyd/assets/tools/generate-hash-shell-32.sh. Do not edit by hand.'
emit_line ''
emit_load_block
emit_load_prefixed_block ${p}_load_initial_prefixed_block_shell prefix
emit_load_prefixed_block ${p}_load_carried_prefixed_block_shell carry
emit_load_final_carry_block
emit_reset
emit_add_counter
emit_g_helpers
emit_compress
emit_outputs
emit_hash_functions
