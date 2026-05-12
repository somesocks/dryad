dryad_roots_path () {
    dryad_roots_garden=$(dryad_garden_find)
    printf '%s\n' "$dryad_roots_garden/dyd/roots"
}

dryad_root_sentinel_is_root () {
    dryad_root_sentinel_path=$1
    dryad_root_sentinel_garden=$2
    dryad_root_sentinel_value=$(tr -d '[:space:]' < "$dryad_root_sentinel_path")

    if [ "$dryad_root_sentinel_value" != root ]; then
        return 1
    fi

    dryad_root_sentinel_size=$(wc -c < "$dryad_root_sentinel_path" | tr -d ' ')
    if [ "$dryad_root_sentinel_size" != 4 ]; then
        dryad_root_sentinel_rel=${dryad_root_sentinel_path#"$dryad_root_sentinel_garden"/}
        printf '%s\n' "dryad-sh: malformed sentinel file path=$dryad_root_sentinel_rel expected=\"root\"" >&2
    fi

    return 0
}

dryad_root_path_resolve_start () {
    dryad_root_path_input=${1:-.}
    dryad_join_path_load "$(pwd -P)" "$dryad_root_path_input"
    dryad_root_path_abs=$dyd_ret0

    if [ -d "$dryad_root_path_abs" ]; then
        dryad_clean_cd "$dryad_root_path_abs"
        return 0
    fi

    dryad_clean_cd "$(dirname "$dryad_root_path_abs")"
}

dryad_root_path_find_uncached () {
    dryad_root_path_start=${1:-.}
    dryad_root_path_garden=$(dryad_garden_find)
    dryad_root_path_dir=$(dryad_root_path_resolve_start "$dryad_root_path_start")

    case $dryad_root_path_dir in
        "$dryad_root_path_garden"/dyd/roots | "$dryad_root_path_garden"/dyd/roots/* )
            ;;
        * )
            dryad_die "not inside a dryad root"
            ;;
    esac

    while :; do
        dryad_root_type=$dryad_root_path_dir/dyd/type
        if [ -f "$dryad_root_type" ] &&
            dryad_root_sentinel_is_root "$dryad_root_type" "$dryad_root_path_garden"; then
            printf '%s\n' "$dryad_root_path_dir"
            return 0
        fi

        if [ "$dryad_root_path_dir" = "$dryad_root_path_garden/dyd/roots" ] ||
            [ "$dryad_root_path_dir" = "$dryad_root_path_garden" ]; then
            dryad_die "not inside a dryad root"
        fi

        dryad_root_path_parent=$(dirname "$dryad_root_path_dir")
        if [ "$dryad_root_path_parent" = "$dryad_root_path_dir" ]; then
            dryad_die "not inside a dryad root"
        fi
        dryad_root_path_dir=$dryad_root_path_parent
    done
}

dryad_root_path_find_load () {
    dryad_root_path_start=${1:-.}
    dryad_root_path_key_pwd=$(pwd -P)

    if dryad_memo_get_line_load root-path-find "$dryad_root_path_key_pwd" "$dryad_root_path_start"; then
        dryad_profile_count memo.hit.root-path-find
        dryad_root_path_result=$dyd_ret0
        dyd_ret0=$dryad_root_path_result
        return 0
    fi

    dryad_profile_count memo.miss.root-path-find
    dryad_root_path_result=$(dryad_root_path_find_uncached "$dryad_root_path_start") || return $?
    dryad_memo_put_value root-path-find "$dryad_root_path_result" "$dryad_root_path_key_pwd" "$dryad_root_path_start"
    dyd_ret0=$dryad_root_path_result
}

dryad_package_path_resolve_start () {
    dryad_package_path_input=${1:-.}
    dryad_join_path_load "$(pwd -P)" "$dryad_package_path_input"
    dryad_package_path_abs=$dyd_ret0

    if [ -d "$dryad_package_path_abs" ]; then
        dryad_clean_cd "$dryad_package_path_abs"
        return 0
    fi

    dryad_clean_cd "$(dirname "$dryad_package_path_abs")"
}

dryad_package_path_find () {
    dryad_package_path_start=${1:-.}
    dryad_package_path_dir=$(dryad_package_path_resolve_start "$dryad_package_path_start")

    while :; do
        if [ -d "$dryad_package_path_dir/dyd" ]; then
            if [ -f "$dryad_package_path_dir/dyd/type" ]; then
                dryad_package_type=$(tr -d '[:space:]' < "$dryad_package_path_dir/dyd/type")
                if [ "$dryad_package_type" = sprout ] &&
                    [ -e "$dryad_package_path_dir/dyd/dependencies/stem" ]; then
                    dryad_clean_cd "$dryad_package_path_dir/dyd/dependencies/stem"
                    return 0
                fi
            fi
            printf '%s\n' "$dryad_package_path_dir"
            return 0
        fi

        dryad_package_path_parent=$(dirname "$dryad_package_path_dir")
        if [ "$dryad_package_path_parent" = "$dryad_package_path_dir" ]; then
            dryad_die "dyd package path not found"
        fi
        dryad_package_path_dir=$dryad_package_path_parent
    done
}

dryad_root_selected_variant_descriptor_load () {
    dryad_root_selected_root=$1
    dryad_root_selected_requested=$2
    dyd_ret0=

    if [ -n "$dryad_root_selected_requested" ]; then
        dyd_ret0=$(dryad_fs_descriptor_normalize "$dryad_root_selected_requested")
        return 0
    fi

    if [ ! -d "$dryad_root_selected_root/dyd/variants" ]; then
        return 0
    fi

    dryad_root_selected_descriptors=$(dryad_roots_variant_descriptors "$dryad_root_selected_root")
    dryad_root_selected_count=0
    dryad_root_selected_descriptor=
    dryad_root_selected_old_ifs=$IFS
    IFS='
'
    for dryad_root_selected_item in $dryad_root_selected_descriptors; do
        [ -n "$dryad_root_selected_item" ] || continue
        dryad_root_selected_count=$((dryad_root_selected_count + 1))
        dryad_root_selected_descriptor=$dryad_root_selected_item
    done
    IFS=$dryad_root_selected_old_ifs

    if [ "$dryad_root_selected_count" -gt 1 ]; then
        dryad_die "under-specified root variant selector"
    fi

    if [ "$dryad_root_selected_count" -eq 1 ]; then
        dyd_ret0=$dryad_root_selected_descriptor
    fi
}
