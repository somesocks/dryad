dryad_url_query_join () {
    dryad_url_query_join_sep=$1
    shift
    dryad_url_query_join_value=$1

    [ -n "$dryad_url_query_join_value" ] || return 0

    printf '%s\n' "$dryad_url_query_join_value" |
        tr '&' '\n' |
        sed '/^$/d' |
        sort |
        awk -v sep="$dryad_url_query_join_sep" '
            {
                if (seen) {
                    printf "%s", sep
                }
                printf "%s", $0
                seen = 1
            }
            END {
                if (seen) {
                    printf "\n"
                }
            }
        '
}

dryad_url_query_to_descriptor () {
    dryad_url_query_join '+' "$1"
}

dryad_url_query_normalize () {
    dryad_url_query_normalized=$(dryad_url_query_join '&' "$1")
    if [ -n "$dryad_url_query_normalized" ]; then
        printf '?%s\n' "$dryad_url_query_normalized"
    fi
}

dryad_url_query_warn_if_noncanonical () {
    dryad_url_query_warn_query=$1

    [ -n "$dryad_url_query_warn_query" ] || return 0

    dryad_url_query_warn_normalized=$(dryad_url_query_join '&' "$dryad_url_query_warn_query")
    if [ "$dryad_url_query_warn_query" != "$dryad_url_query_warn_normalized" ]; then
        printf '%s\n' "dryad-sh: variant descriptor dimensions should be sorted alphabetically (ascending): ?$dryad_url_query_warn_query" >&2
    fi
}

dryad_fs_descriptor_normalize () {
    dryad_fs_descriptor_raw=$1

    [ -n "$dryad_fs_descriptor_raw" ] || return 0

    printf '%s\n' "$dryad_fs_descriptor_raw" |
        tr '+' '\n' |
        sed '/^$/d' |
        sort |
        awk '
            {
                if ($0 !~ /^[A-Za-z0-9._-]+=.+$/) {
                    exit 2
                }
                if (seen) {
                    printf "+"
                }
                printf "%s", $0
                seen = 1
            }
            END {
                if (seen) {
                    printf "\n"
                }
            }
        ' || dryad_die "malformed variant descriptor: $dryad_fs_descriptor_raw"
}

dryad_descriptor_value () {
    dryad_descriptor=$1
    dryad_descriptor_dim=$2

    dryad_descriptor_old_ifs=$IFS
    IFS=+
    set -- $dryad_descriptor
    IFS=$dryad_descriptor_old_ifs

    for dryad_descriptor_pair do
        case $dryad_descriptor_pair in
            "$dryad_descriptor_dim="* )
                printf '%s\n' "${dryad_descriptor_pair#*=}"
                return 0
                ;;
        esac
    done

    return 1
}

dryad_option_list_contains () {
    dryad_option_list=$1
    dryad_option_target=$2

    dryad_option_old_ifs=$IFS
    IFS=,
    set -- $dryad_option_list
    IFS=$dryad_option_old_ifs

    for dryad_option_item do
        if [ "$dryad_option_item" = "$dryad_option_target" ]; then
            return 0
        fi
    done

    return 1
}

dryad_selector_matches_descriptor () {
    dryad_selector=$1
    dryad_descriptor=$2

    dryad_selector_old_ifs=$IFS
    IFS=+
    set -- $dryad_selector
    IFS=$dryad_selector_old_ifs

    for dryad_selector_pair do
        dryad_selector_dim=${dryad_selector_pair%%=*}
        dryad_selector_options=${dryad_selector_pair#*=}
        dryad_selector_value=$(dryad_descriptor_value "$dryad_descriptor" "$dryad_selector_dim" || true)

        if [ "$dryad_selector_options" = any ]; then
            [ -n "$dryad_selector_value" ] || return 1
            continue
        fi

        if [ "$dryad_selector_options" = none ]; then
            [ -z "$dryad_selector_value" ] || return 1
            continue
        fi

        [ -n "$dryad_selector_value" ] || return 1
        if ! dryad_option_list_contains "$dryad_selector_options" "$dryad_selector_value"; then
            return 1
        fi
    done

    return 0
}
