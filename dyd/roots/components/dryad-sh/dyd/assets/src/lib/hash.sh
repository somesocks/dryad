#
# Fingerprint primitives.
#
# These intentionally avoid external base32 or BLAKE2 tools so dryad-sh can
# act as a small bootstrap implementation on a plain POSIX system.
#

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

dryad_blake2b_128_file_hex () {
    dryad_b2_file=$1
    dryad_b2_format=${2:-hex}

    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        [ "$dryad_b2_format" = hex ] || dryad_die "unsupported shell hash format: $dryad_b2_format"
        dryad_blake2b_128_file_hex_shell "$dryad_b2_file"
        return 0
    fi

    dryad_blake2b_128_file_hex_awk "$dryad_b2_file" "$dryad_b2_format"
}

dryad_blake2b_128_file_base32 () {
    dryad_b2_base32_file=$1

    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        dryad_blake2b_128_file_base32_shell "$dryad_b2_base32_file"
        return 0
    fi

    dryad_blake2b_128_file_hex_awk "$dryad_b2_base32_file" base32
}

dryad_blake2b_128_files_table_base32 () {
    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        while IFS='	' read -r dryad_b2_files_table_rel dryad_b2_files_table_path; do
            [ -n "$dryad_b2_files_table_rel" ] || continue
            printf '%s\t' "$dryad_b2_files_table_rel"
            dryad_blake2b_128_file_prefixed_base32_shell "$dryad_b2_files_table_path"
        done
        return 0
    fi

    dryad_blake2b_128_file_hex_awk - files-table
}

dryad_blake2b_128_stream_base32 () {
    if [ "${DRYAD_SH_HASH_IMPL:-awk}" = shell ]; then
        dryad_b2_stream_tmp=${TMPDIR:-/tmp}/dryad-sh-hash-stream.$$
        rm -f "$dryad_b2_stream_tmp"
        cat > "$dryad_b2_stream_tmp"
        dryad_blake2b_128_file_base32_shell "$dryad_b2_stream_tmp"
        rm -f "$dryad_b2_stream_tmp"
        return 0
    fi

    dryad_blake2b_128_file_hex_awk - base32
}

dryad_blake2b_128_file_fingerprint () {
    dryad_b2_fp_file=$1
    printf 'v2-'
    dryad_blake2b_128_file_base32 "$dryad_b2_fp_file"
}
