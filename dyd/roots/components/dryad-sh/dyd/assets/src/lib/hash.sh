#
# Fingerprint primitives.
#
# These intentionally avoid external base32 or BLAKE2 tools so dryad-sh can
# act as a small bootstrap implementation on a plain POSIX system.
#

dryad_base32_char_load () {
    case $1 in
        0 ) dyd_ret0=a ;; 1 ) dyd_ret0=b ;; 2 ) dyd_ret0=c ;; 3 ) dyd_ret0=d ;;
        4 ) dyd_ret0=e ;; 5 ) dyd_ret0=f ;; 6 ) dyd_ret0=g ;; 7 ) dyd_ret0=h ;;
        8 ) dyd_ret0=i ;; 9 ) dyd_ret0=j ;; 10 ) dyd_ret0=k ;; 11 ) dyd_ret0=l ;;
        12 ) dyd_ret0=m ;; 13 ) dyd_ret0=n ;; 14 ) dyd_ret0=o ;; 15 ) dyd_ret0=p ;;
        16 ) dyd_ret0=q ;; 17 ) dyd_ret0=r ;; 18 ) dyd_ret0=s ;; 19 ) dyd_ret0=t ;;
        20 ) dyd_ret0=u ;; 21 ) dyd_ret0=v ;; 22 ) dyd_ret0=w ;; 23 ) dyd_ret0=x ;;
        24 ) dyd_ret0=y ;; 25 ) dyd_ret0=z ;; 26 ) dyd_ret0=2 ;; 27 ) dyd_ret0=3 ;;
        28 ) dyd_ret0=4 ;; 29 ) dyd_ret0=5 ;; 30 ) dyd_ret0=6 ;; 31 ) dyd_ret0=7 ;;
        * ) dryad_die "invalid base32 index: $1" ;;
    esac
}

dryad_base32_char () {
    dryad_base32_char_load "$1"
    printf '%s' "$dyd_ret0"
}

dryad_hash_impl_resolve () {
    dryad_hash_impl_request=${DRYAD_SH_HASH_IMPL:-auto}
    if [ "${dryad_hash_impl_resolved_request+x}" = x ] \
        && [ "$dryad_hash_impl_resolved_request" = "$dryad_hash_impl_request" ]; then
        return 0
    fi

    case $dryad_hash_impl_request in
        auto | shell )
            if dryad_hash_shell_32_supported; then
                dryad_hash_impl=shell-32
            else
                dryad_hash_impl=shell-16
            fi
            ;;
        shell-32 )
            dryad_hash_shell_32_supported \
                || dryad_die "DRYAD_SH_HASH_IMPL=shell-32 requested, but 32-bit shell arithmetic is not supported"
            dryad_hash_impl=shell-32
            ;;
        shell-16 )
            dryad_hash_impl=shell-16
            ;;
        * )
            dryad_die "invalid DRYAD_SH_HASH_IMPL: $dryad_hash_impl_request"
            ;;
    esac

    dryad_hash_impl_resolved_request=$dryad_hash_impl_request
}

dryad_hash_shell_32_supported () {
    if [ "${dryad_hash_shell_32_supported_status+x}" = x ]; then
        return "$dryad_hash_shell_32_supported_status"
    fi

    if (
        dryad_hash_shell_32_probe=$(((4294967295 + 1) & 4294967295)):$(((4294967295 + 4294967295 + 4294967295) >> 32)):$(((2147483648 >> 1) & 4294967295))
        [ "$dryad_hash_shell_32_probe" = 0:2:1073741824 ]
    ) 2>/dev/null; then
        dryad_hash_shell_32_supported_status=0
    else
        dryad_hash_shell_32_supported_status=1
    fi

    return "$dryad_hash_shell_32_supported_status"
}

dryad_blake2b_128_file_hex_shell_load () {
    case $dryad_hash_impl in
        shell-32 ) dryad_blake2b_128_file_hex_shell_32_load "$1" ;;
        shell-16 ) dryad_blake2b_128_file_hex_shell_16_load "$1" ;;
        * ) dryad_die "invalid shell hash implementation: $dryad_hash_impl" ;;
    esac
}

dryad_blake2b_128_file_base32_shell_load () {
    case $dryad_hash_impl in
        shell-32 ) dryad_blake2b_128_file_base32_shell_32_load "$1" ;;
        shell-16 ) dryad_blake2b_128_file_base32_shell_16_load "$1" ;;
        * ) dryad_die "invalid shell hash implementation: $dryad_hash_impl" ;;
    esac
}

dryad_blake2b_128_file_prefixed_base32_shell_load () {
    case $dryad_hash_impl in
        shell-32 ) dryad_blake2b_128_file_prefixed_base32_shell_32_load "$1" ;;
        shell-16 ) dryad_blake2b_128_file_prefixed_base32_shell_16_load "$1" ;;
        * ) dryad_die "invalid shell hash implementation: $dryad_hash_impl" ;;
    esac
}

dryad_blake2b_128_file_hex () {
    dryad_blake2b_128_file_hex_load "$@"
    printf '%s\n' "$dyd_ret0"
}

dryad_blake2b_128_file_hex_load () {
    dyd_b2_file=$1
    dyd_b2_format=${2:-hex}

    dryad_hash_impl_resolve
    case $dryad_hash_impl in
    shell-32 | shell-16 )
        [ "$dyd_b2_format" = hex ] || dryad_die "unsupported shell hash format: $dyd_b2_format"
        dryad_blake2b_128_file_hex_shell_load "$dyd_b2_file"
        return $?
        ;;
    esac
}

dryad_blake2b_128_file_base32 () {
    dryad_blake2b_128_file_base32_load "$1"
    printf '%s\n' "$dyd_ret0"
}

dryad_blake2b_128_file_base32_load () {
    dyd_b2_base32_file=$1

    dryad_hash_impl_resolve
    case $dryad_hash_impl in
    shell-32 | shell-16 )
        dryad_blake2b_128_file_base32_shell_load "$dyd_b2_base32_file"
        return $?
        ;;
    esac
}

dryad_blake2b_128_file_prefixed_base32_load () {
    dyd_b2_prefixed_file=$1

    dryad_hash_impl_resolve
    case $dryad_hash_impl in
    shell-32 | shell-16 )
        dryad_blake2b_128_file_prefixed_base32_shell_load "$dyd_b2_prefixed_file"
        return $?
        ;;
    esac
}

dryad_blake2b_128_files_table_base32 () {
    dryad_hash_impl_resolve
    case $dryad_hash_impl in
    shell-32 | shell-16 )
        while IFS='	' read -r dyd_b2_files_table_rel dyd_b2_files_table_path; do
            [ -n "$dyd_b2_files_table_rel" ] || continue
            dryad_blake2b_128_file_prefixed_base32_shell_load "$dyd_b2_files_table_path"
            printf '%s\t%s\n' "$dyd_b2_files_table_rel" "$dyd_ret0"
        done
        return 0
        ;;
    esac
}

dryad_blake2b_128_stream_base32 () {
    dryad_hash_impl_resolve
    case $dryad_hash_impl in
    shell-32 | shell-16 )
        dyd_b2_stream_tmp=${TMPDIR:-/tmp}/dryad-sh-hash-stream.$$
        rm -f "$dyd_b2_stream_tmp"
        cat > "$dyd_b2_stream_tmp"
        dryad_blake2b_128_file_base32_shell_load "$dyd_b2_stream_tmp"
        printf '%s\n' "$dyd_ret0"
        rm -f "$dyd_b2_stream_tmp"
        return 0
        ;;
    esac
}

dryad_blake2b_128_file_fingerprint () {
    dryad_blake2b_128_file_fingerprint_load "$1"
    printf '%s\n' "$dyd_ret0"
}

dryad_blake2b_128_file_fingerprint_load () {
    dyd_b2_fp_file=$1
    dryad_blake2b_128_file_base32_load "$dyd_b2_fp_file"
    dyd_ret0=v2-$dyd_ret0
}
