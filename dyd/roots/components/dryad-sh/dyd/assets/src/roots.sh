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
    dryad_root_path_abs=$(dryad_join_path "$(pwd -P)" "$dryad_root_path_input")

    if [ -d "$dryad_root_path_abs" ]; then
        dryad_clean_cd "$dryad_root_path_abs"
        return 0
    fi

    dryad_clean_cd "$(dirname "$dryad_root_path_abs")"
}

dryad_root_path_find () {
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

dryad_cmd_root_path () {
    dryad_root_path_target=

    while [ "$#" -gt 0 ]; do
        dryad_root_path_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_path_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root path [path]
EOF
                return 0
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_path_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root path option: $1"
                ;;
            * )
                [ -z "$dryad_root_path_target" ] ||
                    dryad_die "root path accepts one path"
                dryad_root_path_target=$1
                shift
                ;;
        esac
    done

    dryad_root_path_find "${dryad_root_path_target:-.}"
}

dryad_root_create_resolve_target () {
    dryad_root_create_input=$1
    dryad_root_create_abs=$(dryad_join_path "$(pwd -P)" "$dryad_root_create_input")

    if [ -d "$dryad_root_create_abs" ]; then
        dryad_clean_cd "$dryad_root_create_abs"
        return 0
    fi

    dryad_root_create_parent=$(dirname "$dryad_root_create_abs")
    dryad_root_create_name=$(basename "$dryad_root_create_abs")
    dryad_root_create_parent_abs=$(dryad_clean_cd "$dryad_root_create_parent")
    printf '%s\n' "$dryad_root_create_parent_abs/$dryad_root_create_name"
}

dryad_root_create () {
    dryad_root_create_target=$1
    dryad_root_create_garden=$(dryad_garden_find)
    dryad_root_create_roots=$dryad_root_create_garden/dyd/roots
    dryad_root_create_path=$(dryad_root_create_resolve_target "$dryad_root_create_target")

    case $dryad_root_create_path in
        "$dryad_root_create_roots"/* )
            ;;
        * )
            dryad_die "root destination $dryad_root_create_path must be in roots directory $dryad_root_create_roots"
            ;;
    esac

    if [ -e "$dryad_root_create_path" ]; then
        dryad_die "root destination $dryad_root_create_path already exists"
    fi

    mkdir -p "$dryad_root_create_path/dyd/assets"
    mkdir -p "$dryad_root_create_path/dyd/commands"
    mkdir -p "$dryad_root_create_path/dyd/docs"
    mkdir -p "$dryad_root_create_path/dyd/requirements"
    mkdir -p "$dryad_root_create_path/dyd/secrets"
    mkdir -p "$dryad_root_create_path/dyd/traits"

    printf '%s' root > "$dryad_root_create_path/dyd/type"
    : > "$dryad_root_create_path/dyd/commands/dyd-root-build"
    chmod 775 "$dryad_root_create_path/dyd/commands/dyd-root-build"

    printf '%s\n' "$dryad_root_create_path"
}

dryad_cmd_root_create () {
    dryad_root_create_target=

    while [ "$#" -gt 0 ]; do
        dryad_root_create_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_create_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root create <path>
EOF
                return 0
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_create_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root create option: $1"
                ;;
            * )
                [ -z "$dryad_root_create_target" ] ||
                    dryad_die "root create accepts one path"
                dryad_root_create_target=$1
                shift
                ;;
        esac
    done

    [ -n "$dryad_root_create_target" ] ||
        dryad_die "root create requires a path"

    dryad_root_create "$dryad_root_create_target"
}

dryad_package_path_resolve_start () {
    dryad_package_path_input=${1:-.}
    dryad_package_path_abs=$(dryad_join_path "$(pwd -P)" "$dryad_package_path_input")

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

dryad_root_secrets_path_find () {
    dryad_root_secrets_start=${1:-.}
    dryad_root_secrets_package=$(dryad_package_path_find "$dryad_root_secrets_start")
    printf '%s\n' "$dryad_root_secrets_package/dyd/secrets"
}

dryad_cmd_root_secrets_one_path () {
    dryad_root_secrets_action=$1
    shift
    dryad_root_secrets_target=

    while [ "$#" -gt 0 ]; do
        dryad_root_secrets_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_secrets_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad root secrets $dryad_root_secrets_action [path]
EOF
                return 0
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_secrets_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root secrets $dryad_root_secrets_action option: $1"
                ;;
            * )
                [ -z "$dryad_root_secrets_target" ] ||
                    dryad_die "root secrets $dryad_root_secrets_action accepts one path"
                dryad_root_secrets_target=$1
                shift
                ;;
        esac
    done

    dryad_root_secrets_path=$(dryad_root_secrets_path_find "${dryad_root_secrets_target:-.}")
    case $dryad_root_secrets_action in
        path )
            [ -d "$dryad_root_secrets_path" ] || return 0
            printf '%s\n' "$dryad_root_secrets_path"
            ;;
        list )
            [ -d "$dryad_root_secrets_path" ] || return 0
            find "$dryad_root_secrets_path" -print | sort
            ;;
    esac
}

dryad_cmd_root_secrets () {
    dryad_root_secrets_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_root_secrets_action in
        path | list )
            dryad_cmd_root_secrets_one_path "$dryad_root_secrets_action" "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad root secrets path [path]
  dryad root secrets list [path]
EOF
            ;;
        * )
            dryad_die "unsupported root secrets action: $dryad_root_secrets_action"
            ;;
    esac
}

dryad_relative_path () {
    dryad_relative_from=$(dryad_clean_cd "$1")
    dryad_relative_to=$(dryad_clean_cd "$2")
    dryad_relative_prefix=$dryad_relative_from
    dryad_relative_ups=

    while :; do
        case $dryad_relative_to in
            "$dryad_relative_prefix" )
                printf '%s\n' "${dryad_relative_ups:-.}"
                return 0
                ;;
            "$dryad_relative_prefix"/* )
                dryad_relative_tail=${dryad_relative_to#"$dryad_relative_prefix"/}
                if [ -n "$dryad_relative_ups" ]; then
                    printf '%s/%s\n' "$dryad_relative_ups" "$dryad_relative_tail"
                else
                    printf '%s\n' "$dryad_relative_tail"
                fi
                return 0
                ;;
        esac

        if [ "$dryad_relative_prefix" = / ]; then
            printf '%s\n' "$dryad_relative_to"
            return 0
        fi

        dryad_relative_prefix=$(dirname "$dryad_relative_prefix")
        if [ -n "$dryad_relative_ups" ]; then
            dryad_relative_ups=$dryad_relative_ups/..
        else
            dryad_relative_ups=..
        fi
    done
}

dryad_file_abs_path () {
    dryad_file_abs_input=$1
    dryad_file_abs_dir=$(dirname "$dryad_file_abs_input")
    dryad_file_abs_name=$(basename "$dryad_file_abs_input")
    dryad_file_abs_dir=$(dryad_clean_cd "$dryad_file_abs_dir")
    printf '%s/%s\n' "$dryad_file_abs_dir" "$dryad_file_abs_name"
}

dryad_bool_value () {
    case $1 in
        true | 1 )
            printf '1\n'
            ;;
        false | 0 )
            printf '0\n'
            ;;
        * )
            dryad_die "expected boolean value, got: $1"
            ;;
    esac
}

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

dryad_requirement_name_normalize () {
    dryad_requirement_name_raw=$1
    dryad_requirement_name_alias=$dryad_requirement_name_raw
    dryad_requirement_name_condition=

    case $dryad_requirement_name_raw in
        *~* )
            dryad_requirement_name_alias=${dryad_requirement_name_raw%%~*}
            dryad_requirement_name_condition=${dryad_requirement_name_raw#*~}
            [ -n "$dryad_requirement_name_condition" ] ||
                dryad_die "malformed requirement condition descriptor: $dryad_requirement_name_raw"
            ;;
    esac

    case $dryad_requirement_name_alias in
        '' | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]* )
            dryad_die "malformed requirement name: $dryad_requirement_name_raw"
            ;;
    esac

    if [ -n "$dryad_requirement_name_condition" ]; then
        dryad_requirement_name_condition=$(dryad_fs_descriptor_normalize "$dryad_requirement_name_condition")
        printf '%s~%s\n' "$dryad_requirement_name_alias" "$dryad_requirement_name_condition"
    else
        printf '%s\n' "$dryad_requirement_name_alias"
    fi
}

dryad_root_selected_variant_descriptor () {
    dryad_root_selected_root=$1
    dryad_root_selected_requested=$2

    if [ -n "$dryad_root_selected_requested" ]; then
        dryad_fs_descriptor_normalize "$dryad_root_selected_requested"
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
        printf '%s\n' "$dryad_root_selected_descriptor"
    fi
}

dryad_root_requirements_path_for_variant () {
    dryad_root_requirements_variant_root=$1
    dryad_root_requirements_variant_descriptor=$2

    if [ -n "$dryad_root_requirements_variant_descriptor" ]; then
        printf '%s\n' "$dryad_root_requirements_variant_root/dyd/requirements~$dryad_root_requirements_variant_descriptor"
    else
        printf '%s\n' "$dryad_root_requirements_variant_root/dyd/requirements"
    fi
}

dryad_root_requirements_selected_path () {
    dryad_root_requirements_root=$1
    dryad_root_requirements_variant=$2
    dryad_root_requirements_variant=$(dryad_root_selected_variant_descriptor "$dryad_root_requirements_root" "$dryad_root_requirements_variant")

    if [ -n "$dryad_root_requirements_variant" ]; then
        dryad_root_requirements_match=
        for dryad_root_requirements_candidate in "$dryad_root_requirements_root"/dyd/requirements~*; do
            [ -d "$dryad_root_requirements_candidate" ] || continue
            dryad_root_requirements_selector=${dryad_root_requirements_candidate##*/requirements~}
            if dryad_selector_matches_descriptor "$dryad_root_requirements_selector" "$dryad_root_requirements_variant"; then
                [ -z "$dryad_root_requirements_match" ] ||
                    dryad_die "multiple requirements paths match variant: $dryad_root_requirements_variant"
                dryad_root_requirements_match=$dryad_root_requirements_candidate
            fi
        done

        if [ -n "$dryad_root_requirements_match" ]; then
            printf '%s\n' "$dryad_root_requirements_match"
            return 0
        fi
    fi

    if [ -d "$dryad_root_requirements_root/dyd/requirements" ]; then
        printf '%s\n' "$dryad_root_requirements_root/dyd/requirements"
    fi
}

dryad_root_requirement_parse_target () {
    dryad_root_requirement_target_raw=$1

    case $dryad_root_requirement_target_raw in
        *://* )
            dryad_die "unsupported scheme for root requirement: ${dryad_root_requirement_target_raw%%://*}"
            ;;
        root:* )
            dryad_root_requirement_target_body=${dryad_root_requirement_target_raw#root:}
            ;;
        * )
            dryad_root_requirement_target_body=$dryad_root_requirement_target_raw
            ;;
    esac

    dryad_root_requirement_target_query=
    case $dryad_root_requirement_target_body in
        *\?* )
            dryad_root_requirement_target_path=${dryad_root_requirement_target_body%%\?*}
            dryad_root_requirement_target_query=${dryad_root_requirement_target_body#*\?}
            ;;
        * )
            dryad_root_requirement_target_path=$dryad_root_requirement_target_body
            ;;
    esac

    [ -n "$dryad_root_requirement_target_path" ] ||
        dryad_die "missing root requirement target path"

    printf '%s\n%s\n' "$dryad_root_requirement_target_path" "$dryad_root_requirement_target_query"
}

dryad_cmd_root_requirement_add () {
    dryad_root_requirement_add_target=
    dryad_root_requirement_add_alias=
    dryad_root_requirement_add_variant=

    while [ "$#" -gt 0 ]; do
        dryad_root_requirement_add_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_requirement_add_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root requirement add [--variant=<descriptor>] <path> [alias]
EOF
                return 0
                ;;
            --variant=* )
                dryad_root_requirement_add_variant=${dryad_root_requirement_add_arg#--variant=}
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                dryad_root_requirement_add_variant=$2
                shift 2
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_requirement_add_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root requirement add option: $1"
                ;;
            * )
                if [ -z "$dryad_root_requirement_add_target" ]; then
                    dryad_root_requirement_add_target=$1
                elif [ -z "$dryad_root_requirement_add_alias" ]; then
                    dryad_root_requirement_add_alias=$1
                else
                    dryad_die "root requirement add accepts a target and optional alias"
                fi
                shift
                ;;
        esac
    done

    [ -n "$dryad_root_requirement_add_target" ] ||
        dryad_die "root requirement add requires a target"

    dryad_root_requirement_add_root=$(dryad_root_path_find .)
    dryad_root_requirement_add_variant=$(dryad_root_selected_variant_descriptor "$dryad_root_requirement_add_root" "$dryad_root_requirement_add_variant")
    dryad_root_requirement_add_dir=$(dryad_root_requirements_path_for_variant "$dryad_root_requirement_add_root" "$dryad_root_requirement_add_variant")

    dryad_root_requirement_add_parsed=$(dryad_root_requirement_parse_target "$dryad_root_requirement_add_target")
    dryad_root_requirement_add_dep_path=$(printf '%s\n' "$dryad_root_requirement_add_parsed" | sed -n '1p')
    dryad_root_requirement_add_dep_query=$(printf '%s\n' "$dryad_root_requirement_add_parsed" | sed -n '2p')
    dryad_root_requirement_add_dep_root=$(dryad_root_path_find "$dryad_root_requirement_add_dep_path")

    if [ -z "$dryad_root_requirement_add_alias" ]; then
        dryad_root_requirement_add_alias=$(basename "$dryad_root_requirement_add_dep_root")
    fi
    dryad_root_requirement_add_alias=$(dryad_requirement_name_normalize "$dryad_root_requirement_add_alias")

    mkdir -p "$dryad_root_requirement_add_dir"
    dryad_root_requirement_add_rel=$(dryad_relative_path "$dryad_root_requirement_add_dir" "$dryad_root_requirement_add_dep_root")
    dryad_root_requirement_add_query=$(dryad_url_query_normalize "$dryad_root_requirement_add_dep_query")
    printf 'root:%s%s' "$dryad_root_requirement_add_rel" "$dryad_root_requirement_add_query" > "$dryad_root_requirement_add_dir/$dryad_root_requirement_add_alias"
}

dryad_cmd_root_requirement_remove () {
    dryad_root_requirement_remove_name=
    dryad_root_requirement_remove_variant=

    while [ "$#" -gt 0 ]; do
        dryad_root_requirement_remove_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_requirement_remove_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root requirement remove [--variant=<descriptor>] <name>
EOF
                return 0
                ;;
            --variant=* )
                dryad_root_requirement_remove_variant=${dryad_root_requirement_remove_arg#--variant=}
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                dryad_root_requirement_remove_variant=$2
                shift 2
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_requirement_remove_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root requirement remove option: $1"
                ;;
            * )
                [ -z "$dryad_root_requirement_remove_name" ] ||
                    dryad_die "root requirement remove accepts one requirement name"
                dryad_root_requirement_remove_name=$1
                shift
                ;;
        esac
    done

    [ -n "$dryad_root_requirement_remove_name" ] ||
        dryad_die "root requirement remove requires a requirement name"

    dryad_root_requirement_remove_name=$(dryad_requirement_name_normalize "$dryad_root_requirement_remove_name")
    dryad_root_requirement_remove_root=$(dryad_root_path_find .)
    dryad_root_requirement_remove_variant=$(dryad_root_selected_variant_descriptor "$dryad_root_requirement_remove_root" "$dryad_root_requirement_remove_variant")
    dryad_root_requirement_remove_dir=$(dryad_root_requirements_path_for_variant "$dryad_root_requirement_remove_root" "$dryad_root_requirement_remove_variant")
    dryad_root_requirement_remove_path=$dryad_root_requirement_remove_dir/$dryad_root_requirement_remove_name

    [ -e "$dryad_root_requirement_remove_path" ] ||
        dryad_die "requirement does not exist"

    rm "$dryad_root_requirement_remove_path"
}

dryad_cmd_root_requirement () {
    dryad_root_requirement_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_root_requirement_action in
        add )
            dryad_cmd_root_requirement_add "$@"
            ;;
        remove )
            dryad_cmd_root_requirement_remove "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad root requirement add <path> [alias]
  dryad root requirement remove <name>
EOF
            ;;
        * )
            dryad_die "unsupported root requirement action: $dryad_root_requirement_action"
            ;;
    esac
}

dryad_requirement_target_spec () {
    dryad_requirement_file=$1

    if [ -L "$dryad_requirement_file" ]; then
        dryad_requirement_link=$(readlink "$dryad_requirement_file") ||
            dryad_die "could not read requirement symlink: $dryad_requirement_file"
        case $dryad_requirement_link in
            root:* )
                printf '%s\n' "$dryad_requirement_link"
                ;;
            * )
                dryad_requirement_dir=$(dirname "$dryad_requirement_file")
                dryad_requirement_target=$(dryad_join_path "$dryad_requirement_dir" "$dryad_requirement_link")
                dryad_requirement_root=$(dryad_root_path_find "$dryad_requirement_target")
                dryad_requirement_rel=$(dryad_relative_path "$dryad_requirement_dir" "$dryad_requirement_root")
                printf 'root:%s\n' "$dryad_requirement_rel"
                ;;
        esac
        return 0
    fi

    dryad_requirement_spec=$(sed 's/^[[:space:]]*//;s/[[:space:]]*$//' "$dryad_requirement_file")
    dryad_requirement_size=$(wc -c < "$dryad_requirement_file" | tr -d ' ')
    dryad_requirement_trimmed_size=$(printf '%s' "$dryad_requirement_spec" | wc -c | tr -d ' ')
    if [ "$dryad_requirement_size" != "$dryad_requirement_trimmed_size" ]; then
        dryad_requirement_abs=$(dryad_file_abs_path "$dryad_requirement_file")
        dryad_requirement_garden=$(dryad_garden_find)
        dryad_requirement_display=${dryad_requirement_abs#"$dryad_requirement_garden"/}
        printf '%s\n' "dryad-sh: malformed requirement file path=$dryad_requirement_display expected=\"$dryad_requirement_spec\"" >&2
    fi
    printf '%s\n' "$dryad_requirement_spec"
}

dryad_requirement_target_url () {
    dryad_requirement_file=$1
    dryad_requirement_dir=$(dirname "$dryad_requirement_file")
    dryad_requirement_spec=$(dryad_requirement_target_spec "$dryad_requirement_file")

    case $dryad_requirement_spec in
        root:* )
            ;;
        * )
            dryad_die "requirement target must use root: scheme: $dryad_requirement_file"
            ;;
    esac

    dryad_requirement_body=${dryad_requirement_spec#root:}
    dryad_requirement_query=
    case $dryad_requirement_body in
        *\?* )
            dryad_requirement_query=${dryad_requirement_body#*\?}
            dryad_requirement_target_ref=${dryad_requirement_body%%\?*}
            ;;
        * )
            dryad_requirement_target_ref=$dryad_requirement_body
            ;;
    esac

    case $dryad_requirement_target_ref in
        /* )
            dryad_die "root requirement target must be relative: $dryad_requirement_file"
            ;;
    esac

    dryad_requirement_target_path=$(dryad_join_path "$dryad_requirement_dir" "$dryad_requirement_target_ref")
    dryad_requirement_target_root=$(dryad_root_path_find "$dryad_requirement_target_path")
    dryad_requirement_target_rel=$(dryad_relative_path "$dryad_requirement_dir" "$dryad_requirement_target_root")
    dryad_requirement_target_query=$(dryad_url_query_normalize "$dryad_requirement_query")
    printf 'root:%s%s\n' "$dryad_requirement_target_rel" "$dryad_requirement_target_query"
}

dryad_root_requirements_list_entries () {
    dryad_root_requirements_dir=$1
    dryad_root_requirements_relative=$2
    dryad_root_requirements_garden=$(dryad_garden_find)

    for dryad_root_requirement_entry in "$dryad_root_requirements_dir"/* "$dryad_root_requirements_dir"/.[!.]* "$dryad_root_requirements_dir"/..?*; do
        [ -e "$dryad_root_requirement_entry" ] || [ -L "$dryad_root_requirement_entry" ] || continue
        [ -f "$dryad_root_requirement_entry" ] || [ -L "$dryad_root_requirement_entry" ] || continue

        dryad_root_requirement_abs=$(dryad_file_abs_path "$dryad_root_requirement_entry")
        if [ "$dryad_root_requirements_relative" = 1 ]; then
            case $dryad_root_requirement_abs in
                "$dryad_root_requirements_garden"/* )
                    dryad_root_requirement_display=${dryad_root_requirement_abs#"$dryad_root_requirements_garden"/}
                    ;;
                * )
                    dryad_root_requirement_display=$dryad_root_requirement_abs
                    ;;
            esac
        else
            dryad_root_requirement_display=$dryad_root_requirement_abs
        fi

        dryad_root_requirement_target=$(dryad_requirement_target_url "$dryad_root_requirement_abs")
        printf '%s -> %s\n' "$dryad_root_requirement_display" "$dryad_root_requirement_target"
    done | sort
}

dryad_cmd_root_requirements_list () {
    dryad_root_requirements_target=
    dryad_root_requirements_variant=
    dryad_root_requirements_relative=1

    while [ "$#" -gt 0 ]; do
        dryad_root_requirements_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_requirements_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root requirements list [root] [--variant=<descriptor>] [--relative=<bool>]
EOF
                return 0
                ;;
            --variant=* )
                dryad_root_requirements_variant=${dryad_root_requirements_arg#--variant=}
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                dryad_root_requirements_variant=$2
                shift 2
                ;;
            --relative=* )
                dryad_root_requirements_relative=$(dryad_bool_value "${dryad_root_requirements_arg#--relative=}")
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_root_requirements_relative=$(dryad_bool_value "$2")
                            shift 2
                            ;;
                        * )
                            dryad_root_requirements_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_root_requirements_relative=1
                    shift
                fi
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_requirements_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root requirements list option: $1"
                ;;
            * )
                [ -z "$dryad_root_requirements_target" ] ||
                    dryad_die "root requirements list accepts one root"
                dryad_root_requirements_target=$1
                shift
                ;;
        esac
    done

    dryad_root_requirements_ref=${dryad_root_requirements_target:-.}
    dryad_root_requirements_path=$dryad_root_requirements_ref
    dryad_root_requirements_query=
    case $dryad_root_requirements_ref in
        *\?* )
            dryad_root_requirements_path=${dryad_root_requirements_ref%%\?*}
            dryad_root_requirements_query=${dryad_root_requirements_ref#*\?}
            [ -z "$dryad_root_requirements_variant" ] ||
                dryad_die "cannot use --variant when root reference already has a selector"
            dryad_root_requirements_variant=$(dryad_url_query_to_descriptor "$dryad_root_requirements_query")
            ;;
    esac

    dryad_root_requirements_root=$(dryad_root_path_find "$dryad_root_requirements_path")
    dryad_root_requirements_dir=$(dryad_root_requirements_selected_path "$dryad_root_requirements_root" "$dryad_root_requirements_variant")
    [ -n "$dryad_root_requirements_dir" ] || return 0

    dryad_root_requirements_list_entries "$dryad_root_requirements_dir" "$dryad_root_requirements_relative"
}

dryad_cmd_root_requirements () {
    dryad_root_requirements_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_root_requirements_action in
        list )
            dryad_cmd_root_requirements_list "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad root requirements list [root]
EOF
            ;;
        * )
            dryad_die "unsupported root requirements action: $dryad_root_requirements_action"
            ;;
    esac
}

dryad_cmd_root_variants_list () {
    dryad_root_variants_target=

    while [ "$#" -gt 0 ]; do
        dryad_root_variants_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_variants_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root variants list [root]
EOF
                return 0
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_variants_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root variants list option: $1"
                ;;
            * )
                [ -z "$dryad_root_variants_target" ] ||
                    dryad_die "root variants list accepts one root"
                dryad_root_variants_target=$1
                shift
                ;;
        esac
    done

    dryad_root_variants_root=$(dryad_root_path_find "${dryad_root_variants_target:-.}")
    dryad_roots_variant_descriptors "$dryad_root_variants_root" | while IFS= read -r dryad_root_variants_descriptor; do
        if [ -n "$dryad_root_variants_descriptor" ]; then
            printf '%s\n' "$dryad_root_variants_descriptor"
        else
            printf '%s\n' default
        fi
    done
}

dryad_cmd_root_variants () {
    dryad_root_variants_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_root_variants_action in
        list )
            dryad_cmd_root_variants_list "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad root variants list [root]
EOF
            ;;
        * )
            dryad_die "unsupported root variants action: $dryad_root_variants_action"
            ;;
    esac
}

dryad_root_ref_parse () {
    dryad_root_ref_raw=${1:-.}
    dryad_root_ref_path=$dryad_root_ref_raw
    dryad_root_ref_selector=

    case $dryad_root_ref_raw in
        *\?* )
            dryad_root_ref_path=${dryad_root_ref_raw%%\?*}
            dryad_root_ref_query=${dryad_root_ref_raw#*\?}
            [ -n "$dryad_root_ref_path" ] ||
                dryad_die "missing root ref path"
            dryad_root_ref_selector=$(dryad_url_query_to_descriptor "$dryad_root_ref_query")
            ;;
        *~* )
            dryad_root_ref_path=${dryad_root_ref_raw%%~*}
            dryad_root_ref_selector=${dryad_root_ref_raw#*~}
            [ -n "$dryad_root_ref_path" ] ||
                dryad_die "missing root ref path"
            dryad_root_ref_selector=$(dryad_fs_descriptor_normalize "$dryad_root_ref_selector")
            ;;
    esac

    printf '%s\n%s\n' "$dryad_root_ref_path" "$dryad_root_ref_selector"
}

dryad_root_variant_selector_matches_descriptor () {
    dryad_root_variant_selector=$1
    dryad_root_variant_descriptor=$2

    [ -n "$dryad_root_variant_selector" ] || return 0

    dryad_root_variant_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_variant_selector
    IFS=$dryad_root_variant_old_ifs

    for dryad_root_variant_pair do
        dryad_root_variant_dim=${dryad_root_variant_pair%%=*}
        dryad_root_variant_want=${dryad_root_variant_pair#*=}
        dryad_root_variant_got=$(dryad_descriptor_value "$dryad_root_variant_descriptor" "$dryad_root_variant_dim" || true)

        case $dryad_root_variant_want in
            inherit )
                dryad_die "inherit is invalid in root variant selectors"
                ;;
            host )
                case $dryad_root_variant_dim in
                    os ) dryad_root_variant_want=$(dryad_host_os) ;;
                    arch ) dryad_root_variant_want=$(dryad_host_arch) ;;
                esac
                ;;
        esac

        if [ "$dryad_root_variant_want" = any ]; then
            [ -n "$dryad_root_variant_got" ] || return 1
            continue
        fi

        if [ "$dryad_root_variant_want" = none ]; then
            [ -z "$dryad_root_variant_got" ] || return 1
            continue
        fi

        [ -n "$dryad_root_variant_got" ] || return 1
        dryad_option_list_contains "$dryad_root_variant_want" "$dryad_root_variant_got" || return 1
    done

    return 0
}

dryad_root_ref_format () {
    dryad_root_ref_format_root=$1
    dryad_root_ref_format_descriptor=$2
    dryad_root_ref_format_relative=$3
    dryad_root_ref_format_garden=$4

    if [ "$dryad_root_ref_format_relative" = 1 ]; then
        dryad_root_ref_format_display=${dryad_root_ref_format_root#"$dryad_root_ref_format_garden"/}
    else
        dryad_root_ref_format_display=$dryad_root_ref_format_root
    fi

    if [ -n "$dryad_root_ref_format_descriptor" ]; then
        printf '%s~%s\n' "$dryad_root_ref_format_display" "$dryad_root_ref_format_descriptor"
    else
        printf '%s\n' "$dryad_root_ref_format_display"
    fi
}

dryad_root_selected_refs () {
    dryad_root_selected_refs_root=$1
    dryad_root_selected_refs_selector=$2
    dryad_root_selected_refs_relative=$3
    dryad_root_selected_refs_garden=$4

    dryad_roots_variant_descriptors "$dryad_root_selected_refs_root" | while IFS= read -r dryad_root_selected_refs_descriptor; do
        if dryad_root_variant_selector_matches_descriptor "$dryad_root_selected_refs_selector" "$dryad_root_selected_refs_descriptor"; then
            dryad_root_ref_format "$dryad_root_selected_refs_root" "$dryad_root_selected_refs_descriptor" "$dryad_root_selected_refs_relative" "$dryad_root_selected_refs_garden"
        fi
    done
}

dryad_cmd_root_graph_walk () {
    dryad_root_walk_kind=$1
    dryad_root_walk_transpose=$2
    shift 2

    dryad_root_walk_ref=
    dryad_root_walk_variant=
    dryad_root_walk_relative=1

    while [ "$#" -gt 0 ]; do
        dryad_root_walk_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_walk_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad root $dryad_root_walk_kind [root_ref] [--variant=<descriptor>] [--relative=<bool>]
EOF
                return 0
                ;;
            --variant=* )
                [ -z "$dryad_root_walk_variant" ] ||
                    dryad_die "root ${dryad_root_walk_kind%?} selector specified in both root_ref and --variant"
                dryad_root_walk_variant=${dryad_root_walk_arg#--variant=}
                dryad_root_walk_variant=$(dryad_fs_descriptor_normalize "$dryad_root_walk_variant")
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                [ -z "$dryad_root_walk_variant" ] ||
                    dryad_die "root ${dryad_root_walk_kind%?} selector specified in both root_ref and --variant"
                dryad_root_walk_variant=$(dryad_fs_descriptor_normalize "$2")
                shift 2
                ;;
            --relative=* )
                dryad_root_walk_relative=$(dryad_bool_value "${dryad_root_walk_arg#--relative=}")
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_root_walk_relative=$(dryad_bool_value "$2")
                            shift 2
                            ;;
                        * )
                            dryad_root_walk_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_root_walk_relative=1
                    shift
                fi
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_walk_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root $dryad_root_walk_kind option: $1"
                ;;
            * )
                [ -z "$dryad_root_walk_ref" ] ||
                    dryad_die "root $dryad_root_walk_kind accepts one root_ref"
                dryad_root_walk_ref=$1
                shift
                ;;
        esac
    done

    dryad_root_walk_parsed=$(dryad_root_ref_parse "${dryad_root_walk_ref:-.}")
    dryad_root_walk_path=$(printf '%s\n' "$dryad_root_walk_parsed" | sed -n '1p')
    dryad_root_walk_ref_selector=$(printf '%s\n' "$dryad_root_walk_parsed" | sed -n '2p')

    if [ -n "$dryad_root_walk_ref_selector" ]; then
        [ -z "$dryad_root_walk_variant" ] ||
            dryad_die "root ${dryad_root_walk_kind%?} selector specified in both root_ref and --variant"
        dryad_root_walk_variant=$dryad_root_walk_ref_selector
    fi

    dryad_root_walk_garden=$(dryad_garden_find)
    dryad_roots_graph_garden=$dryad_root_walk_garden
    dryad_root_walk_root=$(dryad_root_path_find "$dryad_root_walk_path")
    dryad_root_walk_start=$(dryad_root_selected_refs "$dryad_root_walk_root" "$dryad_root_walk_variant" "$dryad_root_walk_relative" "$dryad_root_walk_garden")
    [ -n "$dryad_root_walk_start" ] ||
        dryad_die "resolved root ${dryad_root_walk_kind%?} variants are empty"

    dryad_roots_graph_lines "$dryad_root_walk_relative" "$dryad_root_walk_transpose" |
        awk -F '\t' -v starts="$dryad_root_walk_start" '
            BEGIN {
                start_count = split(starts, start_parts, "\n")
                for (i = 1; i <= start_count; i++) {
                    if (start_parts[i] != "") {
                        start[start_parts[i]] = 1
                        queue[++queue_count] = start_parts[i]
                    }
                }
            }

            $2 != "" {
                edge_count[$1]++
                edges[$1, edge_count[$1]] = $3
            }

            END {
                for (head = 1; head <= queue_count; head++) {
                    node = queue[head]
                    for (i = 1; i <= edge_count[node]; i++) {
                        next_node = edges[node, i]
                        if (!(next_node in seen)) {
                            seen[next_node] = 1
                            queue[++queue_count] = next_node
                        }
                    }
                }

                for (node in seen) {
                    if (!(node in start)) {
                        print node
                    }
                }
            }
        ' |
        sort
}

dryad_cmd_root_ancestors () {
    dryad_cmd_root_graph_walk ancestors 0 "$@"
}

dryad_cmd_root_descendants () {
    dryad_cmd_root_graph_walk descendants 1 "$@"
}

dryad_root_build_log () {
    case $dryad_log_level in
        info | debug | trace )
            printf 'dryad-sh: %s\n' "$*" >&2
            ;;
    esac
}

dryad_root_build_selected_descriptors () {
    dryad_root_build_selected_root=$1
    dryad_root_build_selected_selector=$2

    dryad_roots_variant_descriptors "$dryad_root_build_selected_root" |
        while IFS= read -r dryad_root_build_selected_descriptor; do
            if dryad_root_variant_selector_matches_descriptor "$dryad_root_build_selected_selector" "$dryad_root_build_selected_descriptor"; then
                printf '%s\n' "$dryad_root_build_selected_descriptor"
            fi
        done
}

dryad_root_build_copy_dir_contents () {
    dryad_root_build_copy_src=$1
    dryad_root_build_copy_dst=$2

    [ -d "$dryad_root_build_copy_src" ] || return 0
    mkdir -p "$dryad_root_build_copy_dst"
    cp -R "$dryad_root_build_copy_src/." "$dryad_root_build_copy_dst/"
}

dryad_root_build_materialize_traits () {
    dryad_root_build_traits_root=$1
    dryad_root_build_traits_descriptor=$2
    dryad_root_build_traits_workspace=$3
    dryad_root_build_traits_selected=$(dryad_roots_owning_selected_path "$dryad_root_build_traits_root" "$dryad_root_build_traits_descriptor" traits)
    dryad_root_build_traits_dest=$dryad_root_build_traits_workspace/dyd/traits
    dryad_root_build_traits_garden=$(dryad_garden_find)
    dryad_root_build_traits_rel=${dryad_root_build_traits_root#"$dryad_root_build_traits_garden"/dyd/roots/}

    mkdir -p "$dryad_root_build_traits_dest"
    if [ -n "$dryad_root_build_traits_selected" ]; then
        dryad_root_build_copy_dir_contents "$dryad_root_build_traits_selected" "$dryad_root_build_traits_dest"
    fi

    [ -n "$dryad_root_build_traits_descriptor" ] || return 0

    dryad_root_build_traits_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_traits_descriptor
    IFS=$dryad_root_build_traits_old_ifs

    for dryad_root_build_traits_pair do
        dryad_root_build_traits_dim=${dryad_root_build_traits_pair%%=*}
        dryad_root_build_traits_value=${dryad_root_build_traits_pair#*=}
        dryad_root_build_traits_path=$dryad_root_build_traits_dest/$dryad_root_build_traits_dim
        if [ -f "$dryad_root_build_traits_path" ]; then
            dryad_root_build_traits_existing=$(cat "$dryad_root_build_traits_path")
            if [ "$dryad_root_build_traits_existing" != "$dryad_root_build_traits_value" ]; then
                printf '%s\n' "dryad-sh: variant option overrides existing trait value path=dyd/roots/$dryad_root_build_traits_rel/dyd/traits/$dryad_root_build_traits_dim" >&2
            fi
        fi
        printf '%s' "$dryad_root_build_traits_value" > "$dryad_root_build_traits_dest/$dryad_root_build_traits_dim"
    done
}

dryad_root_build_prepare_source () {
    dryad_root_build_prepare_root=$1
    dryad_root_build_prepare_descriptor=$2
    dryad_root_build_prepare_workspace=$3

    mkdir -p "$dryad_root_build_prepare_workspace/dyd/dependencies"
    dryad_root_build_materialize_traits "$dryad_root_build_prepare_root" "$dryad_root_build_prepare_descriptor" "$dryad_root_build_prepare_workspace"

    for dryad_root_build_prepare_kind in assets commands secrets docs; do
        dryad_root_build_prepare_selected=$(dryad_roots_owning_selected_path "$dryad_root_build_prepare_root" "$dryad_root_build_prepare_descriptor" "$dryad_root_build_prepare_kind")
        if [ -n "$dryad_root_build_prepare_selected" ]; then
            ln -s "$dryad_root_build_prepare_selected" "$dryad_root_build_prepare_workspace/dyd/$dryad_root_build_prepare_kind"
        fi
    done

    dryad_root_build_prepare_requirements=$(dryad_roots_owning_selected_path "$dryad_root_build_prepare_root" "$dryad_root_build_prepare_descriptor" requirements)
    if [ -n "$dryad_root_build_prepare_requirements" ]; then
        ln -s "$dryad_root_build_prepare_requirements" "$dryad_root_build_prepare_workspace/dyd/~requirements"
        ln -s "$dryad_root_build_prepare_requirements" "$dryad_root_build_prepare_workspace/dyd/requirements"
    fi
}

dryad_root_build_target_selector_matches_descriptor () {
    dryad_root_build_target_selector=$1
    dryad_root_build_target_descriptor=$2
    dryad_root_build_target_parent_descriptor=$3

    [ -n "$dryad_root_build_target_selector" ] || return 0

    dryad_root_build_target_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_target_selector
    IFS=$dryad_root_build_target_old_ifs

    for dryad_root_build_target_pair do
        dryad_root_build_target_dim=${dryad_root_build_target_pair%%=*}
        dryad_root_build_target_want=${dryad_root_build_target_pair#*=}
        dryad_root_build_target_got=$(dryad_descriptor_value "$dryad_root_build_target_descriptor" "$dryad_root_build_target_dim" || true)

        case $dryad_root_build_target_want in
            inherit )
                dryad_root_build_target_want=$(dryad_descriptor_value "$dryad_root_build_target_parent_descriptor" "$dryad_root_build_target_dim" || true)
                [ -n "$dryad_root_build_target_want" ] || dryad_root_build_target_want=none
                ;;
            host )
                case $dryad_root_build_target_dim in
                    os ) dryad_root_build_target_want=$(dryad_host_os) ;;
                    arch ) dryad_root_build_target_want=$(dryad_host_arch) ;;
                    * ) ;;
                esac
                ;;
        esac

        if [ "$dryad_root_build_target_want" = any ]; then
            [ -n "$dryad_root_build_target_got" ] || return 1
            continue
        fi

        if [ "$dryad_root_build_target_want" = none ]; then
            [ -z "$dryad_root_build_target_got" ] || return 1
            continue
        fi

        [ -n "$dryad_root_build_target_got" ] || return 1
        dryad_option_list_contains "$dryad_root_build_target_want" "$dryad_root_build_target_got" || return 1
    done

    return 0
}

dryad_root_build_dependency_die () {
    dryad_root_build_dep_die_garden=$1
    dryad_root_build_dep_die_root=$2
    shift 2
    dryad_root_build_dep_die_rel=${dryad_root_build_dep_die_root#"$dryad_root_build_dep_die_garden"/dyd/roots/}
    dryad_die "error resolving root dependencies for dyd/roots/$dryad_root_build_dep_die_rel: $*"
}

dryad_root_build_variant_option_enabled () {
    dryad_root_build_variant_option_root=$1
    dryad_root_build_variant_option_dim=$2
    dryad_root_build_variant_option_name=$3
    dryad_root_build_variant_option_path=$dryad_root_build_variant_option_root/dyd/variants/$dryad_root_build_variant_option_dim/$dryad_root_build_variant_option_name

    [ -f "$dryad_root_build_variant_option_path" ] || return 1
    dryad_root_build_variant_option_value=$(tr -d '[:space:]' < "$dryad_root_build_variant_option_path")
    [ "$dryad_root_build_variant_option_value" = true ]
}

dryad_root_build_validate_target_option () {
    dryad_root_build_validate_option_garden=$1
    dryad_root_build_validate_option_source_root=$2
    dryad_root_build_validate_option_target_root=$3
    dryad_root_build_validate_option_dim=$4
    dryad_root_build_validate_option_name=$5
    dryad_root_build_validate_option_path=$dryad_root_build_validate_option_target_root/dyd/variants/$dryad_root_build_validate_option_dim/$dryad_root_build_validate_option_name

    if [ ! -f "$dryad_root_build_validate_option_path" ]; then
        dryad_root_build_dependency_die "$dryad_root_build_validate_option_garden" "$dryad_root_build_validate_option_source_root" "wrongly-specified requirement variant option: $dryad_root_build_validate_option_dim=$dryad_root_build_validate_option_name"
    fi

    if ! dryad_root_build_variant_option_enabled "$dryad_root_build_validate_option_target_root" "$dryad_root_build_validate_option_dim" "$dryad_root_build_validate_option_name"; then
        dryad_root_build_dependency_die "$dryad_root_build_validate_option_garden" "$dryad_root_build_validate_option_source_root" "disabled requirement variant option: $dryad_root_build_validate_option_dim=$dryad_root_build_validate_option_name"
    fi
}

dryad_root_build_target_selector_value () {
    dryad_root_build_target_selector_value_selector=$1
    dryad_root_build_target_selector_value_dim=$2

    dryad_root_build_target_selector_value_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_target_selector_value_selector
    IFS=$dryad_root_build_target_selector_value_old_ifs

    for dryad_root_build_target_selector_value_pair do
        case $dryad_root_build_target_selector_value_pair in
            "$dryad_root_build_target_selector_value_dim="* )
                printf '%s\n' "${dryad_root_build_target_selector_value_pair#*=}"
                return 0
                ;;
        esac
    done

    return 1
}

dryad_root_build_validate_target_selector () {
    dryad_root_build_validate_garden=$1
    dryad_root_build_validate_source_root=$2
    dryad_root_build_validate_target_root=$3
    dryad_root_build_validate_selector=$4
    dryad_root_build_validate_parent_descriptor=$5
    dryad_root_build_validate_variants=$dryad_root_build_validate_target_root/dyd/variants

    if [ ! -d "$dryad_root_build_validate_variants" ]; then
        if [ -n "$dryad_root_build_validate_selector" ]; then
            dryad_root_build_validate_old_ifs=$IFS
            IFS=+
            set -- $dryad_root_build_validate_selector
            IFS=$dryad_root_build_validate_old_ifs
            for dryad_root_build_validate_pair do
                dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "over-specified requirement variant dimension: ${dryad_root_build_validate_pair%%=*}"
            done
        fi
        return 0
    fi

    dryad_root_build_validate_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_validate_selector
    IFS=$dryad_root_build_validate_old_ifs
    for dryad_root_build_validate_pair do
        dryad_root_build_validate_dim=${dryad_root_build_validate_pair%%=*}
        case $dryad_root_build_validate_dim in
            '' ) continue ;;
            _include | _exclude ) ;;
        esac
        if [ ! -d "$dryad_root_build_validate_variants/$dryad_root_build_validate_dim" ]; then
            dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "over-specified requirement variant dimension: $dryad_root_build_validate_dim"
        fi
    done

    find "$dryad_root_build_validate_variants" -mindepth 1 -maxdepth 1 -type d |
        while IFS= read -r dryad_root_build_validate_dim_path; do
            dryad_root_build_validate_dim=$(basename "$dryad_root_build_validate_dim_path")
            case $dryad_root_build_validate_dim in
                _include | _exclude ) continue ;;
            esac

            dryad_root_build_validate_requested=$(dryad_root_build_target_selector_value "$dryad_root_build_validate_selector" "$dryad_root_build_validate_dim" || true)
            if [ -z "$dryad_root_build_validate_requested" ]; then
                if ! dryad_root_build_variant_option_enabled "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" none; then
                    dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "under-specified requirement variant dimension: $dryad_root_build_validate_dim"
                fi
                continue
            fi

            dryad_root_build_validate_req_old_ifs=$IFS
            IFS=,
            set -- $dryad_root_build_validate_requested
            IFS=$dryad_root_build_validate_req_old_ifs

            for dryad_root_build_validate_option do
                case $dryad_root_build_validate_option in
                    inherit )
                        dryad_root_build_validate_inherited=$(dryad_descriptor_value "$dryad_root_build_validate_parent_descriptor" "$dryad_root_build_validate_dim" || true)
                        [ -n "$dryad_root_build_validate_inherited" ] || dryad_root_build_validate_inherited=none
                        dryad_root_build_validate_target_option "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_inherited"
                        ;;
                    any )
                        dryad_root_build_validate_any=0
                        for dryad_root_build_validate_option_path in "$dryad_root_build_validate_dim_path"/*; do
                            [ -f "$dryad_root_build_validate_option_path" ] || continue
                            dryad_root_build_validate_option_name=$(basename "$dryad_root_build_validate_option_path")
                            [ "$dryad_root_build_validate_option_name" != none ] || continue
                            if dryad_root_build_variant_option_enabled "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_option_name"; then
                                dryad_root_build_validate_any=1
                                break
                            fi
                        done
                        [ "$dryad_root_build_validate_any" = 1 ] ||
                            dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "no enabled variant options for any resolution: $dryad_root_build_validate_dim"
                        ;;
                    host )
                        case $dryad_root_build_validate_dim in
                            os ) dryad_root_build_validate_host=$(dryad_host_os) ;;
                            arch ) dryad_root_build_validate_host=$(dryad_host_arch) ;;
                            * )
                                dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "host option is only supported for variant dimensions os/arch: $dryad_root_build_validate_dim"
                                ;;
                        esac
                        dryad_root_build_validate_target_option "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_host"
                        ;;
                    * )
                        dryad_root_build_validate_target_option "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_option"
                        ;;
                esac
            done
        done
}

dryad_root_build_validate_requirement_condition () {
    dryad_root_build_validate_condition_garden=$1
    dryad_root_build_validate_condition_root=$2
    dryad_root_build_validate_condition_name=$3
    dryad_root_build_validate_condition=$4

    case $dryad_root_build_validate_condition_name in
        *~* )
            [ -n "$dryad_root_build_validate_condition" ] ||
                dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
            ;;
        * )
            return 0
            ;;
    esac

    dryad_root_build_validate_condition_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_validate_condition
    IFS=$dryad_root_build_validate_condition_old_ifs

    for dryad_root_build_validate_condition_pair do
        case $dryad_root_build_validate_condition_pair in
            [ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*=* )
                ;;
            * )
                dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
                ;;
        esac

        dryad_root_build_validate_condition_dim=${dryad_root_build_validate_condition_pair%%=*}
        dryad_root_build_validate_condition_options=${dryad_root_build_validate_condition_pair#*=}
        case $dryad_root_build_validate_condition_dim in
            *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]* | '' )
                dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
                ;;
        esac

        dryad_root_build_validate_condition_opt_old_ifs=$IFS
        IFS=,
        set -- $dryad_root_build_validate_condition_options
        IFS=$dryad_root_build_validate_condition_opt_old_ifs

        for dryad_root_build_validate_condition_option do
            case $dryad_root_build_validate_condition_option in
                host )
                    case $dryad_root_build_validate_condition_dim in
                        os | arch ) ;;
                        * )
                            dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "host option is only supported for variant dimensions os/arch: $dryad_root_build_validate_condition_dim"
                            ;;
                    esac
                    ;;
                '' )
                    dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
                    ;;
            esac
        done
    done
}

dryad_root_build_ref_path () {
    dryad_root_build_ref=$1
    case $dryad_root_build_ref in
        *~* )
            printf '%s\n' "${dryad_root_build_ref%%~*}"
            ;;
        * )
            printf '%s\n' "$dryad_root_build_ref"
            ;;
    esac
}

dryad_root_build_ref_descriptor () {
    dryad_root_build_ref=$1
    case $dryad_root_build_ref in
        *~* )
            printf '%s\n' "${dryad_root_build_ref#*~}"
            ;;
    esac
}

dryad_root_build_descriptor_needs_suffix () {
    dryad_root_build_suffix_selector=$1
    dryad_root_build_suffix_count=$2

    [ "$dryad_root_build_suffix_count" -gt 1 ] && return 0
    case $dryad_root_build_suffix_selector in
        *'=any'* ) return 0 ;;
        *','* ) return 0 ;;
    esac
    return 1
}

dryad_root_build_prepare_dependencies () {
    dryad_root_build_deps_garden=$1
    dryad_root_build_deps_root=$2
    dryad_root_build_deps_descriptor=$3
    dryad_root_build_deps_workspace=$4
    dryad_root_build_deps_req_dir=$dryad_root_build_deps_workspace/dyd/~requirements

    [ -d "$dryad_root_build_deps_req_dir" ] || return 0

    for dryad_root_build_deps_file in "$dryad_root_build_deps_req_dir"/* "$dryad_root_build_deps_req_dir"/.[!.]* "$dryad_root_build_deps_req_dir"/..?*; do
        [ -f "$dryad_root_build_deps_file" ] || [ -L "$dryad_root_build_deps_file" ] || continue

        dryad_root_build_deps_name=$(basename "$dryad_root_build_deps_file")
        dryad_root_build_deps_alias=${dryad_root_build_deps_name%%~*}
        dryad_root_build_deps_condition=
        case $dryad_root_build_deps_name in
            *~* )
                dryad_root_build_deps_condition=${dryad_root_build_deps_name#*~}
                ;;
        esac

        dryad_root_build_validate_requirement_condition "$dryad_root_build_deps_garden" "$dryad_root_build_deps_root" "$dryad_root_build_deps_name" "$dryad_root_build_deps_condition"
        if ! dryad_condition_matches_descriptor "$dryad_root_build_deps_condition" "$dryad_root_build_deps_descriptor"; then
            continue
        fi

        dryad_root_build_deps_spec=$(dryad_requirement_target_spec "$dryad_root_build_deps_file")
        case $dryad_root_build_deps_spec in
            root:* )
                ;;
            * )
                dryad_die "requirement target must use root: scheme: $dryad_root_build_deps_file"
                ;;
        esac

        dryad_root_build_deps_body=${dryad_root_build_deps_spec#root:}
        dryad_root_build_deps_query=
        case $dryad_root_build_deps_body in
            *\?* )
                dryad_root_build_deps_target_rel=${dryad_root_build_deps_body%%\?*}
                dryad_root_build_deps_query=${dryad_root_build_deps_body#*\?}
                ;;
            * )
                dryad_root_build_deps_target_rel=$dryad_root_build_deps_body
                ;;
        esac

        dryad_url_query_warn_if_noncanonical "$dryad_root_build_deps_query"
        dryad_root_build_deps_selector=$(dryad_url_query_to_descriptor "$dryad_root_build_deps_query")
        dryad_root_build_deps_file_dir=$(dryad_clean_cd "$(dirname "$dryad_root_build_deps_file")")
        dryad_root_build_deps_target_root=$(dryad_root_path_find "$(dryad_join_path "$dryad_root_build_deps_file_dir" "$dryad_root_build_deps_target_rel")")
        dryad_root_build_validate_target_selector "$dryad_root_build_deps_garden" "$dryad_root_build_deps_root" "$dryad_root_build_deps_target_root" "$dryad_root_build_deps_selector" "$dryad_root_build_deps_descriptor"
        dryad_root_build_deps_matches_file=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-deps.XXXXXX")
        dryad_root_build_selected_descriptors "$dryad_root_build_deps_target_root" "" |
            while IFS= read -r dryad_root_build_deps_candidate; do
                if dryad_root_build_target_selector_matches_descriptor "$dryad_root_build_deps_selector" "$dryad_root_build_deps_candidate" "$dryad_root_build_deps_descriptor"; then
                    printf '%s\n' "$dryad_root_build_deps_candidate"
                fi
            done > "$dryad_root_build_deps_matches_file"

        dryad_root_build_deps_count=$(wc -l < "$dryad_root_build_deps_matches_file" | tr -d ' ')
        [ "$dryad_root_build_deps_count" -gt 0 ] ||
            dryad_die "root requirement target has no matching variants: $dryad_root_build_deps_file"

        while IFS= read -r dryad_root_build_deps_target_descriptor; do
            dryad_root_build_deps_fingerprint=$(dryad_root_build_stem "$dryad_root_build_deps_garden" "$dryad_root_build_deps_target_root" "$dryad_root_build_deps_target_descriptor")
            dryad_root_build_deps_dep_name=$dryad_root_build_deps_alias
            if dryad_root_build_descriptor_needs_suffix "$dryad_root_build_deps_selector" "$dryad_root_build_deps_count" &&
                [ -n "$dryad_root_build_deps_target_descriptor" ]; then
                dryad_root_build_deps_dep_name=$dryad_root_build_deps_alias~$dryad_root_build_deps_target_descriptor
            fi
            ln -s "$(dryad_root_build_heap_package_path "$dryad_root_build_deps_garden" stems "$dryad_root_build_deps_fingerprint")" "$dryad_root_build_deps_workspace/dyd/dependencies/$dryad_root_build_deps_dep_name"
        done < "$dryad_root_build_deps_matches_file"
        rm -f "$dryad_root_build_deps_matches_file"
    done
}

dryad_root_build_prepare_path () {
    dryad_root_build_path_workspace=$1
    dryad_root_build_path_dir=$dryad_root_build_path_workspace/dyd/path

    rm -rf "$dryad_root_build_path_dir"
    mkdir -p "$dryad_root_build_path_dir"
    [ -d "$dryad_root_build_path_workspace/dyd/dependencies" ] || return 0

    for dryad_root_build_path_dep in "$dryad_root_build_path_workspace"/dyd/dependencies/*; do
        [ -e "$dryad_root_build_path_dep" ] || [ -L "$dryad_root_build_path_dep" ] || continue
        dryad_root_build_path_dep_name=$(basename "$dryad_root_build_path_dep")
        [ -d "$dryad_root_build_path_dep/dyd/commands" ] || continue
        for dryad_root_build_path_command in "$dryad_root_build_path_dep"/dyd/commands/*; do
            [ -f "$dryad_root_build_path_command" ] || continue
            dryad_root_build_path_command_name=$(basename "$dryad_root_build_path_command")
            case $dryad_root_build_path_command_name in
                dyd-stem-run | default )
                    dryad_root_build_path_stub_name=$dryad_root_build_path_dep_name
                    ;;
                * )
                    dryad_root_build_path_stub_name=$dryad_root_build_path_dep_name--$dryad_root_build_path_command_name
                    ;;
            esac
            {
                printf '%s\n' '#!/bin/sh'
                printf '%s\n' 'set -eu'
                printf '%s\n' 'export DYD_STEM="$(dirname "$0")/../dependencies/'"$dryad_root_build_path_dep_name"'"'
                printf '%s\n' 'export PATH="$DYD_STEM/dyd/path:$PATH"'
                printf '%s\n' 'exec "$DYD_STEM/dyd/commands/'"$dryad_root_build_path_command_name"'" "$@"'
            } > "$dryad_root_build_path_dir/$dryad_root_build_path_stub_name"
            chmod 755 "$dryad_root_build_path_dir/$dryad_root_build_path_stub_name"
        done
    done
}

dryad_root_build_prepare_built_requirements () {
    dryad_root_build_reqs_stem=$1
    dryad_root_build_reqs_dir=$dryad_root_build_reqs_stem/dyd/requirements
    dryad_root_build_reqs_deps=$dryad_root_build_reqs_stem/dyd/dependencies

    rm -rf "$dryad_root_build_reqs_dir"
    mkdir -p "$dryad_root_build_reqs_dir"
    [ -d "$dryad_root_build_reqs_deps" ] || return 0

    for dryad_root_build_reqs_dep in "$dryad_root_build_reqs_deps"/*; do
        [ -e "$dryad_root_build_reqs_dep" ] || [ -L "$dryad_root_build_reqs_dep" ] || continue
        dryad_root_build_reqs_name=$(basename "$dryad_root_build_reqs_dep")
        [ -f "$dryad_root_build_reqs_dep/dyd/fingerprint" ] ||
            dryad_die "dependency missing fingerprint: $dryad_root_build_reqs_dep"
        cp "$dryad_root_build_reqs_dep/dyd/fingerprint" "$dryad_root_build_reqs_dir/$dryad_root_build_reqs_name"
    done
}

dryad_root_build_prepare_built_dependencies () {
    dryad_root_build_built_deps_source=$1
    dryad_root_build_built_deps_stem=$2
    dryad_root_build_built_deps_source_dir=$dryad_root_build_built_deps_source/dyd/dependencies
    dryad_root_build_built_deps_dest_dir=$dryad_root_build_built_deps_stem/dyd/dependencies

    [ -d "$dryad_root_build_built_deps_source_dir" ] || return 0
    mkdir -p "$dryad_root_build_built_deps_dest_dir"

    for dryad_root_build_built_deps_source_link in "$dryad_root_build_built_deps_source_dir"/*; do
        [ -e "$dryad_root_build_built_deps_source_link" ] || [ -L "$dryad_root_build_built_deps_source_link" ] || continue
        dryad_root_build_built_deps_name=$(basename "$dryad_root_build_built_deps_source_link")
        dryad_root_build_built_deps_dest_link=$dryad_root_build_built_deps_dest_dir/$dryad_root_build_built_deps_name
        [ ! -e "$dryad_root_build_built_deps_dest_link" ] && [ ! -L "$dryad_root_build_built_deps_dest_link" ] || continue

        dryad_root_build_built_deps_target=$(dryad_clean_cd "$dryad_root_build_built_deps_source_link")
        [ -f "$dryad_root_build_built_deps_target/dyd/fingerprint" ] ||
            dryad_die "dependency missing fingerprint: $dryad_root_build_built_deps_source_link"
        ln -s "$dryad_root_build_built_deps_target" "$dryad_root_build_built_deps_dest_link"
    done
}

dryad_root_build_init_stem () {
    dryad_root_build_init_stem_path=$1

    mkdir -p "$dryad_root_build_init_stem_path/dyd/assets"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/commands"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/dependencies"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/docs"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/requirements"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/traits"
}


dryad_root_build_file_hash () {
    dryad_root_build_hash_file=$1
    dryad_root_build_hash_tmp=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-file-hash.XXXXXX")
    {
        printf 'file\000'
        cat "$dryad_root_build_hash_file"
    } > "$dryad_root_build_hash_tmp"
    dryad_root_build_hash_hex=$(dryad_blake2b_128_file_hex "$dryad_root_build_hash_tmp")
    rm -f "$dryad_root_build_hash_tmp"
    dryad_base32_encode_hex "$dryad_root_build_hash_hex"
}

dryad_root_build_link_hash () {
    dryad_root_build_hash_link=$1
    dryad_root_build_hash_link_target=$(readlink "$dryad_root_build_hash_link")
    dryad_root_build_hash_tmp=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-link-hash.XXXXXX")
    {
        printf 'link\000'
        printf '%s' "$dryad_root_build_hash_link_target"
    } > "$dryad_root_build_hash_tmp"
    dryad_root_build_hash_hex=$(dryad_blake2b_128_file_hex "$dryad_root_build_hash_tmp")
    rm -f "$dryad_root_build_hash_tmp"
    dryad_base32_encode_hex "$dryad_root_build_hash_hex"
}

dryad_root_build_fingerprint () {
    dryad_root_build_fingerprint_path=$1
    dryad_root_build_fingerprint_payload=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint.XXXXXX")
    dryad_root_build_fingerprint_table=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint-table.XXXXXX")

    (
        cd "$dryad_root_build_fingerprint_path" || exit 1
        find . -mindepth 1 -print | sort | while IFS= read -r dryad_root_build_fingerprint_entry; do
            dryad_root_build_fingerprint_rel=${dryad_root_build_fingerprint_entry#./}
            case $dryad_root_build_fingerprint_rel in
                dyd/path/* | dyd/assets/* | dyd/secrets/* | dyd/commands/* | dyd/docs/* | dyd/type | dyd/variants/* | dyd/traits/* | dyd/requirements/* )
                    ;;
                * )
                    continue
                    ;;
            esac
            if [ -L "$dryad_root_build_fingerprint_entry" ]; then
                printf '%s ./%s\n' "$(dryad_root_build_link_hash "$dryad_root_build_fingerprint_entry")" "$dryad_root_build_fingerprint_rel"
            elif [ -f "$dryad_root_build_fingerprint_entry" ]; then
                printf '%s ./%s\n' "$(dryad_root_build_file_hash "$dryad_root_build_fingerprint_entry")" "$dryad_root_build_fingerprint_rel"
            fi
        done
    ) > "$dryad_root_build_fingerprint_table"

    printf 'stem\000' > "$dryad_root_build_fingerprint_payload"
    dryad_root_build_fingerprint_first=1
    while IFS= read -r dryad_root_build_fingerprint_line; do
        if [ "$dryad_root_build_fingerprint_first" = 1 ]; then
            dryad_root_build_fingerprint_first=0
        else
            printf '\000' >> "$dryad_root_build_fingerprint_payload"
        fi
        printf '%s' "$dryad_root_build_fingerprint_line" >> "$dryad_root_build_fingerprint_payload"
    done < "$dryad_root_build_fingerprint_table"

    dryad_blake2b_128_file_fingerprint "$dryad_root_build_fingerprint_payload"
    rm -f "$dryad_root_build_fingerprint_payload" "$dryad_root_build_fingerprint_table"
}

dryad_root_build_heap_package_path () {
    dryad_root_build_heap_fingerprint_path "$1" "$2" "$3"
}

dryad_root_build_heap_depth () {
    dryad_root_build_heap_depth_garden=$1
    dryad_root_build_heap_depth_kind=$2
    dryad_root_build_heap_depth_file=$dryad_root_build_heap_depth_garden/dyd/shed/heap/$dryad_root_build_heap_depth_kind/depth

    if [ ! -f "$dryad_root_build_heap_depth_file" ]; then
        printf '1\n'
        return 0
    fi

    dryad_root_build_heap_depth_value=$(tr -d '[:space:]' < "$dryad_root_build_heap_depth_file")
    case $dryad_root_build_heap_depth_value in
        '' | *[!0-9]* )
            dryad_die "invalid shed heap depth in $dryad_root_build_heap_depth_file"
            ;;
    esac
    printf '%s\n' "$dryad_root_build_heap_depth_value"
}

dryad_root_build_heap_fingerprint_path () {
    dryad_root_build_heap_garden=$1
    dryad_root_build_heap_kind=$2
    dryad_root_build_heap_fingerprint=$3
    dryad_root_build_heap_encoded=${dryad_root_build_heap_fingerprint#v2-}

    case $dryad_root_build_heap_encoded in
        "$dryad_root_build_heap_fingerprint" )
            dryad_die "invalid heap fingerprint: $dryad_root_build_heap_fingerprint"
            ;;
    esac

    dryad_root_build_heap_depth_value=$(dryad_root_build_heap_depth "$dryad_root_build_heap_garden" "$dryad_root_build_heap_kind")
    dryad_root_build_heap_remaining=$dryad_root_build_heap_encoded
    dryad_root_build_heap_path=$dryad_root_build_heap_garden/dyd/heap/$dryad_root_build_heap_kind/v2
    dryad_root_build_heap_i=0
    while [ "$dryad_root_build_heap_i" -lt "$dryad_root_build_heap_depth_value" ]; do
        dryad_root_build_heap_remaining_len=$(printf '%s' "$dryad_root_build_heap_remaining" | wc -c | tr -d ' ')
        [ "$dryad_root_build_heap_remaining_len" -gt 2 ] ||
            dryad_die "invalid shed heap depth $dryad_root_build_heap_depth_value for fingerprint $dryad_root_build_heap_fingerprint"
        dryad_root_build_heap_segment=$(printf '%s\n' "$dryad_root_build_heap_remaining" | sed 's/^\(..\).*/\1/')
        dryad_root_build_heap_remaining=$(printf '%s\n' "$dryad_root_build_heap_remaining" | sed 's/^..//')
        dryad_root_build_heap_path=$dryad_root_build_heap_path/$dryad_root_build_heap_segment
        dryad_root_build_heap_i=$((dryad_root_build_heap_i + 1))
    done
    printf '%s/%s\n' "$dryad_root_build_heap_path" "$dryad_root_build_heap_remaining"
}

dryad_root_build_file_fingerprint () {
    dryad_root_build_file_fp_file=$1
    dryad_root_build_file_fp_tmp=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-file-fingerprint.XXXXXX")
    {
        printf 'file\000'
        cat "$dryad_root_build_file_fp_file"
    } > "$dryad_root_build_file_fp_tmp"
    dryad_root_build_file_fp_hex=$(dryad_blake2b_128_file_hex "$dryad_root_build_file_fp_tmp")
    rm -f "$dryad_root_build_file_fp_tmp"
    printf 'v2-%s\n' "$(dryad_base32_encode_hex "$dryad_root_build_file_fp_hex")"
}

dryad_root_build_heap_add_file () {
    dryad_root_build_add_file_garden=$1
    dryad_root_build_add_file_kind=$2
    dryad_root_build_add_file_src=$3
    dryad_root_build_add_file_fingerprint=$(dryad_root_build_file_fingerprint "$dryad_root_build_add_file_src")
    dryad_root_build_add_file_dest=$(dryad_root_build_heap_fingerprint_path "$dryad_root_build_add_file_garden" "$dryad_root_build_add_file_kind" "$dryad_root_build_add_file_fingerprint")

    if [ -f "$dryad_root_build_add_file_dest" ]; then
        touch "$dryad_root_build_add_file_dest" 2>/dev/null || true
        printf '%s\n' "$dryad_root_build_add_file_fingerprint"
        return 0
    fi

    mkdir -p "$(dirname "$dryad_root_build_add_file_dest")"
    dryad_root_build_add_file_tmp=$(dirname "$dryad_root_build_add_file_dest")/.tmp-$(basename "$dryad_root_build_add_file_dest").$$
    rm -f "$dryad_root_build_add_file_tmp"
    cp "$dryad_root_build_add_file_src" "$dryad_root_build_add_file_tmp"
    chmod 511 "$dryad_root_build_add_file_tmp"
    if ! ln "$dryad_root_build_add_file_tmp" "$dryad_root_build_add_file_dest" 2>/dev/null; then
        if [ ! -f "$dryad_root_build_add_file_dest" ]; then
            rm -f "$dryad_root_build_add_file_tmp"
            dryad_die "could not publish heap file: $dryad_root_build_add_file_dest"
        fi
    fi
    rm -f "$dryad_root_build_add_file_tmp"
    printf '%s\n' "$dryad_root_build_add_file_fingerprint"
}

dryad_root_build_publish_should_include () {
    case $1 in
        dyd | dyd/path | dyd/path/* | dyd/assets | dyd/assets/* | dyd/secrets | dyd/secrets/* | dyd/commands | dyd/commands/* | dyd/docs | dyd/docs/* | dyd/type | dyd/fingerprint | dyd/requirements | dyd/requirements/* | dyd/variants | dyd/variants/* | dyd/dependencies | dyd/traits | dyd/traits/* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_root_build_link_is_internal () {
    dryad_root_build_link_base=$1
    dryad_root_build_link_rel=$2
    dryad_root_build_link_target=$3
    dryad_root_build_link_base_abs=$(dryad_clean_cd "$dryad_root_build_link_base")

    case $dryad_root_build_link_target in
        /* )
            dryad_root_build_link_target_abs=$dryad_root_build_link_target
            ;;
        * )
            dryad_root_build_link_target_abs=$dryad_root_build_link_base_abs/$(dirname "$dryad_root_build_link_rel")/$dryad_root_build_link_target
            ;;
    esac

    dryad_root_build_link_target_dir=$(dirname "$dryad_root_build_link_target_abs")
    if [ -d "$dryad_root_build_link_target_dir" ]; then
        dryad_root_build_link_target_abs=$(dryad_file_abs_path "$dryad_root_build_link_target_abs")
    fi

    case $dryad_root_build_link_target_abs in
        "$dryad_root_build_link_base_abs" | "$dryad_root_build_link_base_abs"/* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_root_build_publish_tree () {
    dryad_root_build_publish_tree_garden=$1
    dryad_root_build_publish_tree_kind=$2
    dryad_root_build_publish_tree_src=$3
    dryad_root_build_publish_tree_tmp=$4

    (
        cd "$dryad_root_build_publish_tree_src" || exit 1
        find . -print | sort | while IFS= read -r dryad_root_build_publish_tree_entry; do
            dryad_root_build_publish_tree_rel=${dryad_root_build_publish_tree_entry#./}
            [ "$dryad_root_build_publish_tree_rel" != . ] || continue
            dryad_root_build_publish_should_include "$dryad_root_build_publish_tree_rel" || continue

            dryad_root_build_publish_tree_dest=$dryad_root_build_publish_tree_tmp/$dryad_root_build_publish_tree_rel
            if [ -L "$dryad_root_build_publish_tree_entry" ]; then
                dryad_root_build_publish_tree_target=$(readlink "$dryad_root_build_publish_tree_entry")
                if dryad_root_build_link_is_internal "$dryad_root_build_publish_tree_src" "$dryad_root_build_publish_tree_rel" "$dryad_root_build_publish_tree_target"; then
                    mkdir -p "$(dirname "$dryad_root_build_publish_tree_dest")"
                    ln -s "$dryad_root_build_publish_tree_target" "$dryad_root_build_publish_tree_dest"
                fi
            elif [ -d "$dryad_root_build_publish_tree_entry" ]; then
                mkdir -p "$dryad_root_build_publish_tree_dest"
            elif [ -f "$dryad_root_build_publish_tree_entry" ]; then
                dryad_root_build_publish_tree_file_kind=files
                case $dryad_root_build_publish_tree_kind:$dryad_root_build_publish_tree_rel in
                    stem:dyd/secrets/* )
                        dryad_root_build_publish_tree_file_kind=secrets
                        ;;
                esac
                dryad_root_build_publish_tree_file_fp=$(dryad_root_build_heap_add_file "$dryad_root_build_publish_tree_garden" "$dryad_root_build_publish_tree_file_kind" "$dryad_root_build_publish_tree_entry")
                dryad_root_build_publish_tree_file_heap=$(dryad_root_build_heap_fingerprint_path "$dryad_root_build_publish_tree_garden" "$dryad_root_build_publish_tree_file_kind" "$dryad_root_build_publish_tree_file_fp")
                mkdir -p "$(dirname "$dryad_root_build_publish_tree_dest")"
                ln "$dryad_root_build_publish_tree_file_heap" "$dryad_root_build_publish_tree_dest"
            fi
        done
    )
}

dryad_root_build_publish_dependency_links () {
    dryad_root_build_publish_deps_garden=$1
    dryad_root_build_publish_deps_src=$2
    dryad_root_build_publish_deps_tmp=$3
    dryad_root_build_publish_deps_src_dir=$dryad_root_build_publish_deps_src/dyd/dependencies
    dryad_root_build_publish_deps_tmp_dir=$dryad_root_build_publish_deps_tmp/dyd/dependencies

    mkdir -p "$dryad_root_build_publish_deps_tmp_dir"
    [ -d "$dryad_root_build_publish_deps_src_dir" ] || return 0

    for dryad_root_build_publish_deps_link in "$dryad_root_build_publish_deps_src_dir"/*; do
        [ -e "$dryad_root_build_publish_deps_link" ] || [ -L "$dryad_root_build_publish_deps_link" ] || continue
        dryad_root_build_publish_deps_name=$(basename "$dryad_root_build_publish_deps_link")
        dryad_root_build_publish_deps_target=$(dryad_clean_cd "$dryad_root_build_publish_deps_link")
        [ -f "$dryad_root_build_publish_deps_target/dyd/fingerprint" ] ||
            dryad_die "dependency missing fingerprint: $dryad_root_build_publish_deps_link"
        dryad_root_build_publish_deps_fingerprint=$(cat "$dryad_root_build_publish_deps_target/dyd/fingerprint")
        dryad_root_build_publish_deps_heap_path=$(dryad_root_build_heap_package_path "$dryad_root_build_publish_deps_garden" stems "$dryad_root_build_publish_deps_fingerprint")
        dryad_root_build_publish_deps_rel=$(dryad_relative_path "$dryad_root_build_publish_deps_tmp_dir" "$dryad_root_build_publish_deps_heap_path")
        ln -s "$dryad_root_build_publish_deps_rel" "$dryad_root_build_publish_deps_tmp_dir/$dryad_root_build_publish_deps_name"
    done
}

dryad_root_build_publish_dir () {
    dryad_root_build_publish_garden=$1
    dryad_root_build_publish_kind=$2
    dryad_root_build_publish_src=$3
    dryad_root_build_publish_dest=$4

    if [ -e "$dryad_root_build_publish_dest" ] || [ -L "$dryad_root_build_publish_dest" ]; then
        return 0
    fi

    mkdir -p "$(dirname "$dryad_root_build_publish_dest")"
    dryad_root_build_publish_tmp=$(dirname "$dryad_root_build_publish_dest")/.tmp-$(basename "$dryad_root_build_publish_dest").$$
    rm -rf "$dryad_root_build_publish_tmp"
    mkdir -p "$dryad_root_build_publish_tmp"
    dryad_root_build_publish_tree "$dryad_root_build_publish_garden" "$dryad_root_build_publish_kind" "$dryad_root_build_publish_src" "$dryad_root_build_publish_tmp"
    dryad_root_build_publish_dependency_links "$dryad_root_build_publish_garden" "$dryad_root_build_publish_src" "$dryad_root_build_publish_tmp"
    if ! mv "$dryad_root_build_publish_tmp" "$dryad_root_build_publish_dest" 2>/dev/null; then
        rm -rf "$dryad_root_build_publish_tmp"
        [ -e "$dryad_root_build_publish_dest" ] || return 1
    fi
    find "$dryad_root_build_publish_dest" -type d -exec chmod 511 {} \;
}

dryad_root_build_ensure_sprout_parent () {
    dryad_root_build_parent_garden=$1
    dryad_root_build_parent_rel=$2
    dryad_root_build_parent_dir=$(dirname "$dryad_root_build_parent_rel")
    dryad_root_build_parent_current=$(dryad_sprouts_ensure_dir "$dryad_root_build_parent_garden")

    if [ "$dryad_root_build_parent_dir" = . ]; then
        printf '%s\n' "$dryad_root_build_parent_current"
        return 0
    fi

    dryad_root_build_parent_old_ifs=$IFS
    IFS=/
    set -- $dryad_root_build_parent_dir
    IFS=$dryad_root_build_parent_old_ifs

    for dryad_root_build_parent_segment do
        [ -n "$dryad_root_build_parent_segment" ] || continue
        dryad_root_build_parent_next=$dryad_root_build_parent_current/$dryad_root_build_parent_segment
        if [ -e "$dryad_root_build_parent_next" ] && [ ! -d "$dryad_root_build_parent_next" ]; then
            dryad_die "sprout parent path exists and is not a directory: $dryad_root_build_parent_next"
        fi
        if [ ! -d "$dryad_root_build_parent_next" ]; then
            dryad_root_build_parent_restore=0
            if ! dryad_path_has_owner_write "$dryad_root_build_parent_current"; then
                dryad_root_build_parent_restore=1
                chmod u+w "$dryad_root_build_parent_current"
            fi
            mkdir "$dryad_root_build_parent_next"
            chmod 551 "$dryad_root_build_parent_next"
            if [ "$dryad_root_build_parent_restore" = 1 ]; then
                chmod u-w "$dryad_root_build_parent_current"
            fi
        fi
        dryad_root_build_parent_current=$dryad_root_build_parent_next
    done

    printf '%s\n' "$dryad_root_build_parent_current"
}

dryad_root_build_run_command () {
    dryad_root_build_run_garden=$1
    dryad_root_build_run_rel=$2
    dryad_root_build_run_source=$3
    dryad_root_build_run_dest=$4
    dryad_root_build_run_join_stdout=$5
    dryad_root_build_run_join_stderr=$6
    dryad_root_build_run_log_stdout=$7
    dryad_root_build_run_log_stderr=$8

    dryad_root_build_run_command=$dryad_root_build_run_source/dyd/commands/dyd-root-build
    [ -f "$dryad_root_build_run_command" ] ||
        dryad_die "missing root build command: $dryad_root_build_run_command"

    dryad_root_build_run_stdout=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-build.stdout.XXXXXX")
    dryad_root_build_run_stderr=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-build.stderr.XXXXXX")
    dryad_root_build_run_path=$dryad_root_build_run_source/dyd/commands:$dryad_root_build_run_source/dyd/path:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    case $(dryad_host_os) in
        darwin )
            dryad_root_build_run_path=$dryad_root_build_run_path:/opt/homebrew/bin:/opt/homebrew/sbin
            ;;
    esac

    set +e
    (
        cd "$dryad_root_build_run_source" || exit 1
        PATH=$dryad_root_build_run_path \
        DYD_STEM=$dryad_root_build_run_source \
        DYD_BUILD=$dryad_root_build_run_dest \
        DYD_GARDEN=$dryad_root_build_run_garden \
        DYD_OS=$(dryad_host_os) \
        DYD_ARCH=$(dryad_host_arch) \
        DYD_LOG_LEVEL=$dryad_log_level \
        "$dryad_root_build_run_command" "$dryad_root_build_run_dest"
    ) > "$dryad_root_build_run_stdout" 2> "$dryad_root_build_run_stderr"
    dryad_root_build_run_status=$?
    set -e

    if [ "$dryad_root_build_run_join_stdout" = 1 ]; then
        cat "$dryad_root_build_run_stdout" >&2
    fi
    if [ "$dryad_root_build_run_join_stderr" = 1 ]; then
        cat "$dryad_root_build_run_stderr" >&2
    fi
    if [ -n "$dryad_root_build_run_log_stdout" ]; then
        mkdir -p "$dryad_root_build_run_log_stdout"
        cp "$dryad_root_build_run_stdout" "$dryad_root_build_run_log_stdout/dyd-root-build--$(dryad_sprouts_run_sanitize_segment "$dryad_root_build_run_rel").out"
    fi
    if [ -n "$dryad_root_build_run_log_stderr" ]; then
        mkdir -p "$dryad_root_build_run_log_stderr"
        cp "$dryad_root_build_run_stderr" "$dryad_root_build_run_log_stderr/dyd-root-build--$(dryad_sprouts_run_sanitize_segment "$dryad_root_build_run_rel").err"
    fi

    rm -f "$dryad_root_build_run_stdout" "$dryad_root_build_run_stderr"
    return "$dryad_root_build_run_status"
}

dryad_root_build_stem () {
    dryad_root_build_stem_garden=$1
    dryad_root_build_stem_root=$2
    dryad_root_build_stem_descriptor=$3

    dryad_root_build_stem_rel=${dryad_root_build_stem_root#"$dryad_root_build_stem_garden"/dyd/roots/}
    dryad_root_build_stem_workspace=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-root.XXXXXX")
    dryad_root_build_stem_dest=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-stem.XXXXXX")

    dryad_root_build_log "root build - verifying root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"

    dryad_root_build_prepare_source "$dryad_root_build_stem_root" "$dryad_root_build_stem_descriptor" "$dryad_root_build_stem_workspace"
    dryad_root_build_prepare_dependencies "$dryad_root_build_stem_garden" "$dryad_root_build_stem_root" "$dryad_root_build_stem_descriptor" "$dryad_root_build_stem_workspace"
    dryad_root_build_prepare_built_requirements "$dryad_root_build_stem_workspace"
    dryad_root_build_prepare_path "$dryad_root_build_stem_workspace"

    dryad_root_build_init_stem "$dryad_root_build_stem_dest"
    dryad_root_build_log "root build - building root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"
    dryad_root_build_run_command \
        "$dryad_root_build_stem_garden" \
        "$dryad_root_build_stem_rel" \
        "$dryad_root_build_stem_workspace" \
        "$dryad_root_build_stem_dest" \
        "${dryad_root_build_join_stdout:-0}" \
        "${dryad_root_build_join_stderr:-0}" \
        "${dryad_root_build_log_stdout:-}" \
        "${dryad_root_build_log_stderr:-}" ||
        dryad_die "error executing root to build stem: dyd/roots/$dryad_root_build_stem_rel"

    dryad_root_build_prepare_built_dependencies "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_dest"
    dryad_root_build_prepare_built_requirements "$dryad_root_build_stem_dest"
    dryad_root_build_prepare_path "$dryad_root_build_stem_dest"
    printf '%s' stem > "$dryad_root_build_stem_dest/dyd/type"
    dryad_root_build_stem_fingerprint=$(dryad_root_build_fingerprint "$dryad_root_build_stem_dest")
    printf '%s' "$dryad_root_build_stem_fingerprint" > "$dryad_root_build_stem_dest/dyd/fingerprint"
    dryad_root_build_publish_dir "$dryad_root_build_stem_garden" stem "$dryad_root_build_stem_dest" "$(dryad_root_build_heap_package_path "$dryad_root_build_stem_garden" stems "$dryad_root_build_stem_fingerprint")"

    rm -rf "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_dest"
    dryad_root_build_log "root build - done building root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"
    printf '%s\n' "$dryad_root_build_stem_fingerprint"
}

dryad_root_build_materialize_sprout () {
    dryad_root_build_sprout_garden=$1
    dryad_root_build_sprout_root=$2
    dryad_root_build_sprout_descriptors_file=$3
    dryad_root_build_sprout_tmp=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-sprout.XXXXXX")
    dryad_root_build_sprout_rel=${dryad_root_build_sprout_root#"$dryad_root_build_sprout_garden"/dyd/roots/}

    mkdir -p "$dryad_root_build_sprout_tmp/dyd/dependencies"
    mkdir -p "$dryad_root_build_sprout_tmp/dyd/requirements"
    mkdir -p "$dryad_root_build_sprout_tmp/dyd/traits"

    if [ -d "$dryad_root_build_sprout_root/dyd/traits" ]; then
        dryad_root_build_copy_dir_contents "$dryad_root_build_sprout_root/dyd/traits" "$dryad_root_build_sprout_tmp/dyd/traits"
    fi

    while IFS= read -r dryad_root_build_sprout_descriptor; do
        dryad_root_build_sprout_stem_fingerprint=$(dryad_root_build_stem "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_descriptor")
        dryad_root_build_sprout_dep_name=stem
        if [ -n "$dryad_root_build_sprout_descriptor" ]; then
            dryad_root_build_sprout_dep_name=stem~$dryad_root_build_sprout_descriptor
        fi
        ln -s "$(dryad_root_build_heap_package_path "$dryad_root_build_sprout_garden" stems "$dryad_root_build_sprout_stem_fingerprint")" "$dryad_root_build_sprout_tmp/dyd/dependencies/$dryad_root_build_sprout_dep_name"
        printf '%s' "$dryad_root_build_sprout_stem_fingerprint" > "$dryad_root_build_sprout_tmp/dyd/requirements/$dryad_root_build_sprout_dep_name"
    done < "$dryad_root_build_sprout_descriptors_file"

    printf '%s' sprout > "$dryad_root_build_sprout_tmp/dyd/type"
    dryad_root_build_sprout_fingerprint=$(dryad_root_build_fingerprint "$dryad_root_build_sprout_tmp")
    printf '%s' "$dryad_root_build_sprout_fingerprint" > "$dryad_root_build_sprout_tmp/dyd/fingerprint"
    dryad_root_build_sprout_heap_path=$(dryad_root_build_heap_package_path "$dryad_root_build_sprout_garden" sprouts "$dryad_root_build_sprout_fingerprint")
    dryad_root_build_publish_dir "$dryad_root_build_sprout_garden" sprout "$dryad_root_build_sprout_tmp" "$dryad_root_build_sprout_heap_path"

    dryad_root_build_sprout_link_parent=$(dryad_root_build_ensure_sprout_parent "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_rel")
    dryad_root_build_sprout_link=$dryad_root_build_sprout_link_parent/$(basename "$dryad_root_build_sprout_rel")
    dryad_root_build_sprout_restore_parent=0
    if [ -d "$dryad_root_build_sprout_link_parent" ] &&
        ! dryad_path_has_owner_write "$dryad_root_build_sprout_link_parent"; then
        dryad_root_build_sprout_restore_parent=1
        chmod u+w "$dryad_root_build_sprout_link_parent"
    fi
    if [ -e "$dryad_root_build_sprout_link" ] || [ -L "$dryad_root_build_sprout_link" ]; then
        dryad_sprouts_make_tree_removable "$dryad_root_build_sprout_link"
        rm -rf "$dryad_root_build_sprout_link"
    fi
    dryad_root_build_sprout_ln_status=0
    dryad_root_build_sprout_target=$(dryad_relative_path "$dryad_root_build_sprout_link_parent" "$dryad_root_build_sprout_heap_path")
    ln -s "$dryad_root_build_sprout_target" "$dryad_root_build_sprout_link" ||
        dryad_root_build_sprout_ln_status=$?
    if [ "$dryad_root_build_sprout_restore_parent" = 1 ]; then
        chmod u-w "$dryad_root_build_sprout_link_parent"
    fi
    [ "$dryad_root_build_sprout_ln_status" = 0 ] ||
        return "$dryad_root_build_sprout_ln_status"

    rm -rf "$dryad_root_build_sprout_tmp"
    dryad_root_build_log "root build - done verifying root path=dyd/roots/$dryad_root_build_sprout_rel"
    printf '%s\n' "$dryad_root_build_sprout_fingerprint"
}

dryad_root_build_sprout () {
    dryad_root_build_sprout_root=$1
    dryad_root_build_sprout_selector=$2
    dryad_root_build_sprout_garden=$3
    dryad_root_build_sprout_descriptors=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-sprout-descriptors.XXXXXX")
    dryad_root_build_selected_descriptors "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_selector" > "$dryad_root_build_sprout_descriptors"

    if [ ! -s "$dryad_root_build_sprout_descriptors" ] &&
        [ -d "$dryad_root_build_sprout_root/dyd/variants" ]; then
        rm -f "$dryad_root_build_sprout_descriptors"
        dryad_die "resolved root build variants are filtered by variants/_include and variants/_exclude"
    fi

    dryad_root_build_materialize_sprout "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_descriptors"
    rm -f "$dryad_root_build_sprout_descriptors"
}

dryad_roots_build_entry () {
    dryad_roots_build_entry_root=$1
    dryad_roots_build_entry_descriptor=$2
    dryad_roots_build_entry_garden=$3
    dryad_roots_build_entry_display=${dryad_roots_build_entry_root#"$dryad_roots_build_entry_garden"/}

    if [ -n "$dryad_roots_build_entry_descriptor" ]; then
        dryad_roots_build_entry_ref=$dryad_roots_build_entry_display~$dryad_roots_build_entry_descriptor
    else
        dryad_roots_build_entry_ref=$dryad_roots_build_entry_display
    fi

    if dryad_root_variant_selector_matches_descriptor "$dryad_roots_build_variant" "$dryad_roots_build_entry_descriptor" &&
        dryad_roots_entry_matches_filters "$dryad_roots_build_entry_ref" "$dryad_roots_build_entry_root" "$dryad_roots_build_entry_descriptor"; then
        printf '%s\t%s\n' "$dryad_roots_build_entry_root" "$dryad_roots_build_entry_descriptor"
    fi
}

dryad_roots_build_entries_all () {
    dryad_roots_build_entries_garden=$1
    dryad_roots_build_entries_dir=$dryad_roots_build_entries_garden/dyd/roots

    [ -d "$dryad_roots_build_entries_dir" ] || return 0

    dryad_roots_find_roots "$dryad_roots_build_entries_dir" | while IFS= read -r dryad_roots_build_entry_root; do
        dryad_roots_variant_descriptors "$dryad_roots_build_entry_root" | while IFS= read -r dryad_roots_build_entry_descriptor; do
            dryad_roots_build_entry "$dryad_roots_build_entry_root" "$dryad_roots_build_entry_descriptor" "$dryad_roots_build_entries_garden"
        done
    done
}

dryad_roots_build_entries_from_stdin () {
    dryad_roots_build_entries_garden=$1

    while IFS= read -r dryad_roots_build_stdin_ref; do
        [ -n "$dryad_roots_build_stdin_ref" ] || continue
        dryad_roots_build_stdin_parsed=$(dryad_root_ref_parse "$dryad_roots_build_stdin_ref")
        dryad_roots_build_stdin_path=$(printf '%s\n' "$dryad_roots_build_stdin_parsed" | sed -n '1p')
        dryad_roots_build_stdin_selector=$(printf '%s\n' "$dryad_roots_build_stdin_parsed" | sed -n '2p')
        dryad_roots_build_stdin_root=$(dryad_root_path_find "$dryad_roots_build_stdin_path")

        dryad_roots_variant_descriptors "$dryad_roots_build_stdin_root" | while IFS= read -r dryad_roots_build_stdin_descriptor; do
            if dryad_root_variant_selector_matches_descriptor "$dryad_roots_build_stdin_selector" "$dryad_roots_build_stdin_descriptor"; then
                dryad_roots_build_entry "$dryad_roots_build_stdin_root" "$dryad_roots_build_stdin_descriptor" "$dryad_roots_build_entries_garden"
            fi
        done
    done
}

dryad_roots_build_run_entries () {
    dryad_roots_build_run_entries=$1
    dryad_roots_build_run_garden=$2
    dryad_roots_build_run_roots=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-roots-build.XXXXXX")

    printf '%s\n' "$dryad_roots_build_run_entries" |
        awk -F '\t' '$1 != "" { print $1 }' |
        sort -u > "$dryad_roots_build_run_roots"

    while IFS= read -r dryad_roots_build_run_root; do
        [ -n "$dryad_roots_build_run_root" ] || continue
        dryad_roots_build_run_descriptors=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-descriptors.XXXXXX")
        printf '%s\n' "$dryad_roots_build_run_entries" |
            awk -F '\t' -v root="$dryad_roots_build_run_root" '$1 == root { print $2 }' > "$dryad_roots_build_run_descriptors"
        dryad_root_build_materialize_sprout "$dryad_roots_build_run_garden" "$dryad_roots_build_run_root" "$dryad_roots_build_run_descriptors" >/dev/null
        rm -f "$dryad_roots_build_run_descriptors"
    done < "$dryad_roots_build_run_roots"

    rm -f "$dryad_roots_build_run_roots"
}

dryad_cmd_roots_build () {
    dryad_roots_include=
    dryad_roots_exclude=
    dryad_roots_build_from_stdin=0
    dryad_roots_build_variant=
    dryad_root_build_join_stdout=0
    dryad_root_build_join_stderr=0
    dryad_root_build_log_stdout=
    dryad_root_build_log_stderr=

    while [ "$#" -gt 0 ]; do
        dryad_roots_build_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_roots_build_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad roots build [--include=<filter>] [--exclude=<filter>] [--from-stdin] [--variant=<descriptor>]
EOF
                return 0
                ;;
            --include=* )
                dryad_roots_include="${dryad_roots_include}
${dryad_roots_build_arg#--include=}"
                shift
                ;;
            --include )
                [ "$#" -gt 1 ] || dryad_die "--include requires a value"
                dryad_roots_include="${dryad_roots_include}
$2"
                shift 2
                ;;
            --exclude=* )
                dryad_roots_exclude="${dryad_roots_exclude}
${dryad_roots_build_arg#--exclude=}"
                shift
                ;;
            --exclude )
                [ "$#" -gt 1 ] || dryad_die "--exclude requires a value"
                dryad_roots_exclude="${dryad_roots_exclude}
$2"
                shift 2
                ;;
            --from-stdin )
                dryad_roots_build_from_stdin=1
                shift
                ;;
            --variant=* )
                dryad_roots_build_variant=$(dryad_fs_descriptor_normalize "${dryad_roots_build_arg#--variant=}")
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                dryad_roots_build_variant=$(dryad_fs_descriptor_normalize "$2")
                shift 2
                ;;
            --join-stdout=* )
                dryad_root_build_join_stdout=$(dryad_bool_value "${dryad_roots_build_arg#--join-stdout=}")
                shift
                ;;
            --join-stdout )
                dryad_root_build_join_stdout=1
                shift
                ;;
            --join-stderr=* )
                dryad_root_build_join_stderr=$(dryad_bool_value "${dryad_roots_build_arg#--join-stderr=}")
                shift
                ;;
            --join-stderr )
                dryad_root_build_join_stderr=1
                shift
                ;;
            --log-stdout=* )
                dryad_root_build_log_stdout=${dryad_roots_build_arg#--log-stdout=}
                dryad_root_build_join_stdout=0
                shift
                ;;
            --log-stdout )
                [ "$#" -gt 1 ] || dryad_die "--log-stdout requires a value"
                dryad_root_build_log_stdout=$2
                dryad_root_build_join_stdout=0
                shift 2
                ;;
            --log-stderr=* )
                dryad_root_build_log_stderr=${dryad_roots_build_arg#--log-stderr=}
                dryad_root_build_join_stderr=0
                shift
                ;;
            --log-stderr )
                [ "$#" -gt 1 ] || dryad_die "--log-stderr requires a value"
                dryad_root_build_log_stderr=$2
                dryad_root_build_join_stderr=0
                shift 2
                ;;
            --scope=* | --log-level=* | --log-format=* | --parallel=* | --path=* )
                shift
                ;;
            --scope | --log-level | --log-format | --parallel | --path )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported roots build option: $1"
                ;;
            * )
                dryad_die "unsupported roots build argument: $1"
                ;;
        esac
    done

    dryad_roots_build_garden=$(dryad_garden_find)
    dryad_sprouts_prune "$dryad_roots_build_garden"

    if [ "$dryad_roots_build_from_stdin" = 1 ]; then
        dryad_roots_build_entries=$(dryad_roots_build_entries_from_stdin "$dryad_roots_build_garden")
    else
        dryad_roots_build_entries=$(dryad_roots_build_entries_all "$dryad_roots_build_garden")
    fi

    [ -n "$dryad_roots_build_entries" ] || return 0
    dryad_roots_build_run_entries "$dryad_roots_build_entries" "$dryad_roots_build_garden"
}

dryad_cmd_root_build () {
    dryad_root_build_ref=
    dryad_root_build_variant=
    dryad_root_build_join_stdout=0
    dryad_root_build_join_stderr=0
    dryad_root_build_log_stdout=
    dryad_root_build_log_stderr=

    while [ "$#" -gt 0 ]; do
        dryad_root_build_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_root_build_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad root build [root_ref] [--variant=<descriptor>]
EOF
                return 0
                ;;
            --variant=* )
                [ -z "$dryad_root_build_variant" ] ||
                    dryad_die "root build selector specified in both root_ref and --variant"
                dryad_root_build_variant=$(dryad_fs_descriptor_normalize "${dryad_root_build_arg#--variant=}")
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                [ -z "$dryad_root_build_variant" ] ||
                    dryad_die "root build selector specified in both root_ref and --variant"
                dryad_root_build_variant=$(dryad_fs_descriptor_normalize "$2")
                shift 2
                ;;
            --join-stdout=* )
                dryad_root_build_join_stdout=$(dryad_bool_value "${dryad_root_build_arg#--join-stdout=}")
                shift
                ;;
            --join-stdout )
                dryad_root_build_join_stdout=1
                shift
                ;;
            --join-stderr=* )
                dryad_root_build_join_stderr=$(dryad_bool_value "${dryad_root_build_arg#--join-stderr=}")
                shift
                ;;
            --join-stderr )
                dryad_root_build_join_stderr=1
                shift
                ;;
            --log-stdout=* )
                dryad_root_build_log_stdout=${dryad_root_build_arg#--log-stdout=}
                dryad_root_build_join_stdout=0
                shift
                ;;
            --log-stdout )
                [ "$#" -gt 1 ] || dryad_die "--log-stdout requires a value"
                dryad_root_build_log_stdout=$2
                dryad_root_build_join_stdout=0
                shift 2
                ;;
            --log-stderr=* )
                dryad_root_build_log_stderr=${dryad_root_build_arg#--log-stderr=}
                dryad_root_build_join_stderr=0
                shift
                ;;
            --log-stderr )
                [ "$#" -gt 1 ] || dryad_die "--log-stderr requires a value"
                dryad_root_build_log_stderr=$2
                dryad_root_build_join_stderr=0
                shift 2
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_root_build_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported root build option: $1"
                ;;
            * )
                [ -z "$dryad_root_build_ref" ] ||
                    dryad_die "root build accepts one root_ref"
                dryad_root_build_ref=$1
                shift
                ;;
        esac
    done

    dryad_root_build_parsed=$(dryad_root_ref_parse "${dryad_root_build_ref:-.}")
    dryad_root_build_path=$(printf '%s\n' "$dryad_root_build_parsed" | sed -n '1p')
    dryad_root_build_ref_selector=$(printf '%s\n' "$dryad_root_build_parsed" | sed -n '2p')
    if [ -n "$dryad_root_build_ref_selector" ]; then
        [ -z "$dryad_root_build_variant" ] ||
            dryad_die "root build selector specified in both root_ref and --variant"
        dryad_root_build_variant=$dryad_root_build_ref_selector
    fi

    dryad_root_build_garden=$(dryad_garden_find)
    dryad_root_build_root=$(dryad_root_path_find "$dryad_root_build_path")
    dryad_root_build_sprout "$dryad_root_build_root" "$dryad_root_build_variant" "$dryad_root_build_garden"
}

dryad_cmd_root () {
    dryad_root_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_root_action in
        ancestors )
            dryad_cmd_root_ancestors "$@"
            ;;
        build )
            dryad_cmd_root_build "$@"
            ;;
        descendants )
            dryad_cmd_root_descendants "$@"
            ;;
        create )
            dryad_cmd_root_create "$@"
            ;;
        path )
            dryad_cmd_root_path "$@"
            ;;
        requirement )
            dryad_cmd_root_requirement "$@"
            ;;
        requirements )
            dryad_cmd_root_requirements "$@"
            ;;
        secrets )
            dryad_cmd_root_secrets "$@"
            ;;
        variants )
            dryad_cmd_root_variants "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad root ancestors [path]
  dryad root build [path]
  dryad root create <path>
  dryad root descendants [path]
  dryad root path [path]
  dryad root requirement add <path> [alias]
  dryad root requirement remove <name>
  dryad root requirements list [path]
  dryad root secrets path [path]
  dryad root secrets list [path]
  dryad root variants list [path]
EOF
            ;;
        * )
            dryad_die "unsupported root action: $dryad_root_action"
            ;;
    esac
}

dryad_roots_has_root_ancestor () {
    dryad_roots_candidate_parent=$(dirname "$1")

    while [ "$dryad_roots_candidate_parent" != "$dryad_roots_dir" ]; do
        if [ -f "$dryad_roots_candidate_parent/dyd/type" ] &&
            [ "$(cat "$dryad_roots_candidate_parent/dyd/type")" = root ]; then
            return 0
        fi

        dryad_roots_next_parent=$(dirname "$dryad_roots_candidate_parent")
        if [ "$dryad_roots_next_parent" = "$dryad_roots_candidate_parent" ]; then
            return 1
        fi
        dryad_roots_candidate_parent=$dryad_roots_next_parent
    done

    return 1
}

dryad_roots_find_roots () {
    dryad_roots_find_dir=$1

    [ -d "$dryad_roots_find_dir" ] || return 0

    find "$dryad_roots_find_dir" -type d -name dyd -prune -o -type d -print |
        while IFS= read -r dryad_roots_find_candidate; do
            dryad_roots_find_type=$dryad_roots_find_candidate/dyd/type
            [ -f "$dryad_roots_find_type" ] || continue
            if [ "$(cat "$dryad_roots_find_type")" = root ]; then
                printf '%s\n' "$dryad_roots_find_candidate"
            fi
        done |
        sort
}

dryad_roots_validate_filter_rules () {
    dryad_roots_filter_root=$1
    dryad_roots_filter_dir=$dryad_roots_filter_root/dyd/variants

    for dryad_roots_filter_kind in _include _exclude; do
        dryad_roots_filter_kind_dir=$dryad_roots_filter_dir/$dryad_roots_filter_kind
        [ -d "$dryad_roots_filter_kind_dir" ] || continue
        case $dryad_roots_filter_kind in
            _include ) dryad_roots_filter_label=included ;;
            _exclude ) dryad_roots_filter_label=excluded ;;
        esac

        for dryad_roots_filter_file in "$dryad_roots_filter_kind_dir"/*; do
            [ -f "$dryad_roots_filter_file" ] || continue
            dryad_roots_filter_enabled=$(tr -d '[:space:]' < "$dryad_roots_filter_file")
            [ "$dryad_roots_filter_enabled" = true ] || continue
            dryad_roots_filter_descriptor=$(basename "$dryad_roots_filter_file")

            dryad_roots_filter_old_ifs=$IFS
            IFS=+
            set -- $dryad_roots_filter_descriptor
            IFS=$dryad_roots_filter_old_ifs

            for dryad_roots_filter_pair do
                dryad_roots_filter_options=${dryad_roots_filter_pair#*=}
                dryad_roots_filter_opt_old_ifs=$IFS
                IFS=,
                set -- $dryad_roots_filter_options
                IFS=$dryad_roots_filter_opt_old_ifs
                for dryad_roots_filter_option do
                    case $dryad_roots_filter_option in
                        inherit | host )
                            dryad_die "$dryad_roots_filter_option option is not supported for $dryad_roots_filter_label variant selectors"
                            ;;
                    esac
                done
            done
        done
    done
}

dryad_roots_variant_descriptors () {
    dryad_roots_variant_root=$1
    dryad_roots_variant_dir=$dryad_roots_variant_root/dyd/variants

    if [ ! -d "$dryad_roots_variant_dir" ]; then
        printf '\n'
        return 0
    fi

    dryad_roots_validate_filter_rules "$dryad_roots_variant_root"

    find "$dryad_roots_variant_dir" -mindepth 2 -maxdepth 2 -type f |
        sed "s|^$dryad_roots_variant_dir/||" |
        sort |
        awk -F/ -v base="$dryad_roots_variant_dir" '
            function trim(value) {
                sub(/^[[:space:]]+/, "", value)
                sub(/[[:space:]]+$/, "", value)
                return value
            }

            function read_enabled(rel,    raw, path) {
                path = base "/" rel
                if ((getline raw < path) < 0) {
                    return 0
                }
                close(path)
                raw = trim(raw)
                return raw == "true"
            }

            function append_option(dim, option) {
                if (!(dim in dim_seen)) {
                    dims[++dim_count] = dim
                    dim_seen[dim] = 1
                }
                options[dim] = options[dim] "\034" option
            }

            function descriptor_value(descriptor, dim,    parts, count, i, pair) {
                count = split(descriptor, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    if (pair ~ "^" dim "=") {
                        sub("^[^=]+=", "", pair)
                        return pair
                    }
                }
                return ""
            }

            function option_list_contains(list, target,    parts, count, i) {
                count = split(list, parts, ",")
                for (i = 1; i <= count; i++) {
                    if (parts[i] == target) {
                        return 1
                    }
                }
                return 0
            }

            function rule_matches(descriptor, rule,    parts, count, i, pair, dim, want, got) {
                if (rule == "") {
                    return 1
                }

                count = split(rule, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    dim = pair
                    sub("=.*$", "", dim)
                    want = pair
                    sub("^[^=]+=", "", want)
                    got = descriptor_value(descriptor, dim)

                    if (want == "none") {
                        if (got != "") {
                            return 0
                        }
                    } else if (want == "any") {
                        if (got == "") {
                            return 0
                        }
                    } else if (!option_list_contains(want, got)) {
                        return 0
                    }
                }
                return 1
            }

            function included(descriptor,    i) {
                if (include_count == 0) {
                    return 1
                }
                for (i = 1; i <= include_count; i++) {
                    if (rule_matches(descriptor, include_rules[i])) {
                        return 1
                    }
                }
                return 0
            }

            function excluded(descriptor,    i) {
                for (i = 1; i <= exclude_count; i++) {
                    if (rule_matches(descriptor, exclude_rules[i])) {
                        return 1
                    }
                }
                return 0
            }

            function emit_variant(descriptor) {
                if (!included(descriptor) || excluded(descriptor)) {
                    return
                }
                if (!(descriptor in emitted)) {
                    emitted[descriptor] = 1
                    output[++output_count] = descriptor
                }
            }

            function walk(dim_index, prefix,    dim, parts, count, i, option, next_prefix) {
                if (dim_index > dim_count) {
                    emit_variant(prefix)
                    return
                }

                dim = dims[dim_index]
                count = split(options[dim], parts, "\034")
                for (i = 2; i <= count; i++) {
                    option = parts[i]
                    if (option == "none") {
                        next_prefix = prefix
                    } else if (prefix == "") {
                        next_prefix = dim "=" option
                    } else {
                        next_prefix = prefix "+" dim "=" option
                    }
                    walk(dim_index + 1, next_prefix)
                }
            }

            {
                rel = $0
                if ($1 == "_include") {
                    if (read_enabled(rel)) {
                        include_rules[++include_count] = $2
                    }
                    next
                }
                if ($1 == "_exclude") {
                    if (read_enabled(rel)) {
                        exclude_rules[++exclude_count] = $2
                    }
                    next
                }
                if (read_enabled(rel)) {
                    append_option($1, $2)
                }
            }

            END {
                if (dim_count == 0) {
                    emit_variant("")
                } else {
                    walk(1, "")
                }

                for (i = 1; i <= output_count; i++) {
                    print output[i]
                }
            }
        '
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

dryad_roots_trait_matches () {
    dryad_roots_trait_root=$1
    dryad_roots_trait_descriptor=$2
    dryad_roots_trait_name=$3
    dryad_roots_trait_expected=$4

    for dryad_roots_trait_dir in "$dryad_roots_trait_root"/dyd/traits~*; do
        [ -d "$dryad_roots_trait_dir" ] || continue
        dryad_roots_trait_selector=${dryad_roots_trait_dir##*traits~}
        if ! dryad_selector_matches_descriptor "$dryad_roots_trait_selector" "$dryad_roots_trait_descriptor"; then
            continue
        fi

        dryad_roots_trait_file=$dryad_roots_trait_dir/$dryad_roots_trait_name
        [ -f "$dryad_roots_trait_file" ] || continue
        if [ "$(cat "$dryad_roots_trait_file")" = "$dryad_roots_trait_expected" ]; then
            return 0
        fi
    done

    return 1
}

dryad_roots_entry_matches_include () {
    dryad_roots_match_display=$1
    dryad_roots_match_root=$2
    dryad_roots_match_descriptor=$3
    dryad_roots_match_include=$4

    case $dryad_roots_match_include in
        "~"*".txt="* )
            dryad_roots_match_trait=${dryad_roots_match_include#\~}
            dryad_roots_match_trait_name=${dryad_roots_match_trait%%=*}
            dryad_roots_match_trait_value=${dryad_roots_match_trait#*=}
            dryad_roots_trait_matches "$dryad_roots_match_root" "$dryad_roots_match_descriptor" "$dryad_roots_match_trait_name" "$dryad_roots_match_trait_value"
            return $?
            ;;
        "**~"* )
            [ "$dryad_roots_match_descriptor" = "${dryad_roots_match_include#**\~}" ]
            return $?
            ;;
        "~"* )
            dryad_roots_match_selector=${dryad_roots_match_include#\~}
            dryad_selector_matches_descriptor "$dryad_roots_match_selector" "$dryad_roots_match_descriptor"
            return $?
            ;;
        * )
            case $dryad_roots_match_display in
                $dryad_roots_match_include ) return 0 ;;
                * ) return 1 ;;
            esac
            ;;
    esac
}

dryad_roots_entry_matches_includes () {
    dryad_roots_match_display=$1
    dryad_roots_match_root=$2
    dryad_roots_match_descriptor=$3

    if [ -z "${dryad_roots_include:-}" ]; then
        return 0
    fi

    while IFS= read -r dryad_roots_match_include; do
        [ -n "$dryad_roots_match_include" ] || continue
        if dryad_roots_entry_matches_include "$dryad_roots_match_display" "$dryad_roots_match_root" "$dryad_roots_match_descriptor" "$dryad_roots_match_include"; then
            return 0
        fi
    done <<EOF
$dryad_roots_include
EOF

    return 1
}

dryad_roots_entry_matches_excludes () {
    dryad_roots_match_display=$1
    dryad_roots_match_root=$2
    dryad_roots_match_descriptor=$3

    if [ -z "${dryad_roots_exclude:-}" ]; then
        return 1
    fi

    while IFS= read -r dryad_roots_match_exclude; do
        [ -n "$dryad_roots_match_exclude" ] || continue
        if dryad_roots_entry_matches_include "$dryad_roots_match_display" "$dryad_roots_match_root" "$dryad_roots_match_descriptor" "$dryad_roots_match_exclude"; then
            return 0
        fi
    done <<EOF
$dryad_roots_exclude
EOF

    return 1
}

dryad_roots_entry_matches_filters () {
    dryad_roots_match_display=$1
    dryad_roots_match_root=$2
    dryad_roots_match_descriptor=$3

    dryad_roots_entry_matches_includes "$dryad_roots_match_display" "$dryad_roots_match_root" "$dryad_roots_match_descriptor" || return 1
    ! dryad_roots_entry_matches_excludes "$dryad_roots_match_display" "$dryad_roots_match_root" "$dryad_roots_match_descriptor"
}

dryad_roots_print_display () {
    dryad_roots_print_display_value=$1
    if [ "$dryad_roots_to_sprouts" = 1 ]; then
        case $dryad_roots_print_display_value in
            dyd/roots/* )
                dryad_roots_print_display_value=dyd/sprouts/${dryad_roots_print_display_value#dyd/roots/}
                ;;
        esac
    fi
    printf '%s\n' "$dryad_roots_print_display_value"
}

dryad_roots_list_from_stdin () {
    while IFS= read -r dryad_roots_stdin_ref; do
        [ -n "$dryad_roots_stdin_ref" ] || continue
        dryad_roots_stdin_path=${dryad_roots_stdin_ref%%\?*}
        dryad_roots_stdin_query=
        case $dryad_roots_stdin_ref in
            *\?* )
                dryad_roots_stdin_query=${dryad_roots_stdin_ref#*\?}
                ;;
        esac
        dryad_roots_stdin_descriptor=$(printf '%s\n' "$dryad_roots_stdin_query" | tr '&' '+')
        if [ -n "$dryad_roots_stdin_descriptor" ]; then
            dryad_roots_print_display "$dryad_roots_stdin_path~$dryad_roots_stdin_descriptor"
        else
            dryad_roots_print_display "$dryad_roots_stdin_path"
        fi
    done
}

dryad_roots_each_print_entry () {
    dryad_roots_each_entry_root=$1
    dryad_roots_each_entry_descriptor=$2
    dryad_roots_each_entry_garden=$3
    dryad_roots_each_entry_display=${dryad_roots_each_entry_root#"$dryad_roots_each_entry_garden"/}

    if [ -n "$dryad_roots_each_entry_descriptor" ]; then
        dryad_roots_each_entry_ref=$dryad_roots_each_entry_display~$dryad_roots_each_entry_descriptor
    else
        dryad_roots_each_entry_ref=$dryad_roots_each_entry_display
    fi

    if dryad_roots_entry_matches_filters "$dryad_roots_each_entry_ref" "$dryad_roots_each_entry_root" "$dryad_roots_each_entry_descriptor"; then
        printf '%s\t%s\n' "$dryad_roots_each_entry_root" "$dryad_roots_each_entry_descriptor"
    fi
}

dryad_roots_each_entries_all () {
    dryad_roots_each_entries_garden=$1
    dryad_roots_each_entries_dir=$dryad_roots_each_entries_garden/dyd/roots

    [ -d "$dryad_roots_each_entries_dir" ] || return 0

    dryad_roots_find_roots "$dryad_roots_each_entries_dir" | while IFS= read -r dryad_roots_each_entry_root; do
        dryad_roots_variant_descriptors "$dryad_roots_each_entry_root" | while IFS= read -r dryad_roots_each_entry_descriptor; do
            dryad_roots_each_print_entry "$dryad_roots_each_entry_root" "$dryad_roots_each_entry_descriptor" "$dryad_roots_each_entries_garden"
        done
    done
}

dryad_roots_each_entries_from_stdin () {
    dryad_roots_each_entries_garden=$1

    while IFS= read -r dryad_roots_each_stdin_ref; do
        [ -n "$dryad_roots_each_stdin_ref" ] || continue
        dryad_roots_each_stdin_parsed=$(dryad_root_ref_parse "$dryad_roots_each_stdin_ref")
        dryad_roots_each_stdin_path=$(printf '%s\n' "$dryad_roots_each_stdin_parsed" | sed -n '1p')
        dryad_roots_each_stdin_selector=$(printf '%s\n' "$dryad_roots_each_stdin_parsed" | sed -n '2p')
        dryad_roots_each_stdin_root=$(dryad_root_path_find "$dryad_roots_each_stdin_path")

        dryad_roots_variant_descriptors "$dryad_roots_each_stdin_root" | while IFS= read -r dryad_roots_each_stdin_descriptor; do
            if dryad_root_variant_selector_matches_descriptor "$dryad_roots_each_stdin_selector" "$dryad_roots_each_stdin_descriptor"; then
                dryad_roots_each_print_entry "$dryad_roots_each_stdin_root" "$dryad_roots_each_stdin_descriptor" "$dryad_roots_each_entries_garden"
            fi
        done
    done
}

dryad_roots_each_run_entry () {
    dryad_roots_each_run_root=$1
    dryad_roots_each_run_descriptor=$2

    if [ -n "$dryad_roots_each_run_descriptor" ]; then
        dryad_roots_each_run_ref=$dryad_roots_each_run_root~$dryad_roots_each_run_descriptor
    else
        dryad_roots_each_run_ref=$dryad_roots_each_run_root
    fi

    if [ "$dryad_roots_each_join_stdout" = 1 ] && [ "$dryad_roots_each_join_stderr" = 1 ]; then
        (
            cd "$dryad_roots_each_run_root" || exit 1
            DYD_ROOT=$dryad_roots_each_run_root \
                DYD_ROOT_REF=$dryad_roots_each_run_ref \
                DYD_VARIANT=$dryad_roots_each_run_descriptor \
                "$dryad_roots_each_shell" -c "$dryad_roots_each_command"
        )
    elif [ "$dryad_roots_each_join_stdout" = 1 ]; then
        (
            cd "$dryad_roots_each_run_root" || exit 1
            DYD_ROOT=$dryad_roots_each_run_root \
                DYD_ROOT_REF=$dryad_roots_each_run_ref \
                DYD_VARIANT=$dryad_roots_each_run_descriptor \
                "$dryad_roots_each_shell" -c "$dryad_roots_each_command" 2>/dev/null
        )
    elif [ "$dryad_roots_each_join_stderr" = 1 ]; then
        (
            cd "$dryad_roots_each_run_root" || exit 1
            DYD_ROOT=$dryad_roots_each_run_root \
                DYD_ROOT_REF=$dryad_roots_each_run_ref \
                DYD_VARIANT=$dryad_roots_each_run_descriptor \
                "$dryad_roots_each_shell" -c "$dryad_roots_each_command" >/dev/null
        )
    else
        (
            cd "$dryad_roots_each_run_root" || exit 1
            DYD_ROOT=$dryad_roots_each_run_root \
                DYD_ROOT_REF=$dryad_roots_each_run_ref \
                DYD_VARIANT=$dryad_roots_each_run_descriptor \
                "$dryad_roots_each_shell" -c "$dryad_roots_each_command" >/dev/null 2>/dev/null
        )
    fi
}

dryad_cmd_roots_each () {
    dryad_roots_include=
    dryad_roots_exclude=
    dryad_roots_each_from_stdin=0
    dryad_roots_each_ignore_errors=0
    dryad_roots_each_join_stdout=0
    dryad_roots_each_join_stderr=0
    dryad_roots_each_shell=${SHELL:-/bin/sh}
    dryad_roots_each_command=

    while [ "$#" -gt 0 ]; do
        dryad_roots_each_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_roots_each_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad roots each [--include=<filter>] [--exclude=<filter>] [--from-stdin] [--shell=<shell>] -- <command>
EOF
                return 0
                ;;
            --include=* )
                dryad_roots_include="${dryad_roots_include}
${dryad_roots_each_arg#--include=}"
                shift
                ;;
            --include )
                [ "$#" -gt 1 ] || dryad_die "--include requires a value"
                dryad_roots_include="${dryad_roots_include}
$2"
                shift 2
                ;;
            --exclude=* )
                dryad_roots_exclude="${dryad_roots_exclude}
${dryad_roots_each_arg#--exclude=}"
                shift
                ;;
            --exclude )
                [ "$#" -gt 1 ] || dryad_die "--exclude requires a value"
                dryad_roots_exclude="${dryad_roots_exclude}
$2"
                shift 2
                ;;
            --from-stdin )
                dryad_roots_each_from_stdin=1
                shift
                ;;
            --ignore-errors )
                dryad_roots_each_ignore_errors=1
                shift
                ;;
            --join-stdout )
                dryad_roots_each_join_stdout=1
                shift
                ;;
            --join-stderr )
                dryad_roots_each_join_stderr=1
                shift
                ;;
            --shell=* )
                dryad_roots_each_shell=${dryad_roots_each_arg#--shell=}
                shift
                ;;
            --shell )
                [ "$#" -gt 1 ] || dryad_die "--shell requires a value"
                dryad_roots_each_shell=$2
                shift 2
                ;;
            --scope=* | --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --scope | --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported roots each option: $1"
                ;;
            * )
                break
                ;;
        esac
    done

    dryad_roots_each_command=$*
    dryad_roots_each_garden=$(dryad_garden_find)

    if [ "$dryad_roots_each_from_stdin" = 1 ]; then
        dryad_roots_each_entries=$(dryad_roots_each_entries_from_stdin "$dryad_roots_each_garden" | sort -u)
    else
        dryad_roots_each_entries=$(dryad_roots_each_entries_all "$dryad_roots_each_garden" | sort -u)
    fi

    [ -n "$dryad_roots_each_entries" ] || return 0

    printf '%s\n' "$dryad_roots_each_entries" | while IFS="$(printf '\t')" read -r dryad_roots_each_root dryad_roots_each_descriptor; do
        [ -n "$dryad_roots_each_root" ] || continue
        if ! dryad_roots_each_run_entry "$dryad_roots_each_root" "$dryad_roots_each_descriptor"; then
            [ "$dryad_roots_each_ignore_errors" = 1 ] || return 1
        fi
    done
}

dryad_roots_owning_clean_abs_path () {
    dryad_roots_owning_clean_input=$1
    awk -v path="$dryad_roots_owning_clean_input" '
        BEGIN {
            count = split(path, parts, "/")
            depth = 0
            for (i = 1; i <= count; i++) {
                part = parts[i]
                if (part == "" || part == ".") {
                    continue
                }
                if (part == "..") {
                    if (depth > 0) {
                        depth--
                    }
                    continue
                }
                stack[++depth] = part
            }

            if (depth == 0) {
                print "/"
                exit
            }

            out = ""
            for (i = 1; i <= depth; i++) {
                out = out "/" stack[i]
            }
            print out
        }
    '
}

dryad_roots_owning_abs_lexical_path () {
    dryad_roots_owning_abs_input=$1
    case $dryad_roots_owning_abs_input in
        /* )
            dryad_roots_owning_abs_raw=$dryad_roots_owning_abs_input
            ;;
        * )
            dryad_roots_owning_abs_raw=$(pwd -P)/$dryad_roots_owning_abs_input
            ;;
    esac
    dryad_roots_owning_clean_abs_path "$dryad_roots_owning_abs_raw"
}

dryad_roots_owning_dependency_correction () {
    dryad_roots_owning_correction_path=$1
    dryad_roots_owning_correction_parent=$(dirname "$dryad_roots_owning_correction_path")
    dryad_roots_owning_correction_parent_name=$(basename "$dryad_roots_owning_correction_parent")
    dryad_roots_owning_correction_dyd=$(dirname "$dryad_roots_owning_correction_parent")
    dryad_roots_owning_correction_dyd_name=$(basename "$dryad_roots_owning_correction_dyd")

    case $dryad_roots_owning_correction_dyd_name/$dryad_roots_owning_correction_parent_name in
        dyd/requirements | dyd/requirements~* )
            dirname "$dryad_roots_owning_correction_dyd"
            ;;
        * )
            printf '%s\n' "$dryad_roots_owning_correction_path"
            ;;
    esac
}

dryad_roots_owning_eval_existing_prefix () {
    dryad_roots_owning_eval_path=$1

    if [ -d "$dryad_roots_owning_eval_path" ]; then
        dryad_clean_cd "$dryad_roots_owning_eval_path"
        return 0
    fi

    dryad_roots_owning_eval_dir=$(dirname "$dryad_roots_owning_eval_path")
    dryad_roots_owning_eval_suffix=$(basename "$dryad_roots_owning_eval_path")

    while [ ! -d "$dryad_roots_owning_eval_dir" ] &&
        [ "$dryad_roots_owning_eval_dir" != / ]; do
        dryad_roots_owning_eval_suffix=$(basename "$dryad_roots_owning_eval_dir")/$dryad_roots_owning_eval_suffix
        dryad_roots_owning_eval_dir=$(dirname "$dryad_roots_owning_eval_dir")
    done

    if [ -d "$dryad_roots_owning_eval_dir" ]; then
        dryad_roots_owning_eval_real_dir=$(dryad_clean_cd "$dryad_roots_owning_eval_dir")
        case $dryad_roots_owning_eval_real_dir in
            / )
                printf '/%s\n' "$dryad_roots_owning_eval_suffix"
                ;;
            * )
                printf '%s/%s\n' "$dryad_roots_owning_eval_real_dir" "$dryad_roots_owning_eval_suffix"
                ;;
        esac
        return 0
    fi

    printf '%s\n' "$dryad_roots_owning_eval_path"
}

dryad_roots_owning_paths_for_input () {
    dryad_roots_owning_paths_input=$1
    dryad_roots_owning_paths_raw=$(dryad_roots_owning_abs_lexical_path "$dryad_roots_owning_paths_input")
    dryad_roots_owning_paths_corrected=$(dryad_roots_owning_dependency_correction "$dryad_roots_owning_paths_raw")
    dryad_roots_owning_paths_owning=$(dryad_roots_owning_eval_existing_prefix "$dryad_roots_owning_paths_corrected")
    dryad_roots_owning_paths_changed=$dryad_roots_owning_paths_owning

    if [ "$dryad_roots_owning_paths_corrected" != "$dryad_roots_owning_paths_raw" ]; then
        dryad_roots_owning_paths_rel=${dryad_roots_owning_paths_raw#"$dryad_roots_owning_paths_corrected"}
        dryad_roots_owning_paths_rel=${dryad_roots_owning_paths_rel#/}
        if [ -n "$dryad_roots_owning_paths_rel" ]; then
            dryad_roots_owning_paths_changed=$dryad_roots_owning_paths_owning/$dryad_roots_owning_paths_rel
        fi
    fi

    printf '%s\n%s\n' "$dryad_roots_owning_paths_owning" "$dryad_roots_owning_paths_changed"
}

dryad_roots_owning_root_for_path () {
    dryad_roots_owning_root_path=$1
    if [ -d "$dryad_roots_owning_root_path" ]; then
        dryad_roots_owning_root_candidate=$dryad_roots_owning_root_path
    else
        dryad_roots_owning_root_candidate=$(dirname "$dryad_roots_owning_root_path")
    fi

    while :; do
        if [ -f "$dryad_roots_owning_root_candidate/dyd/type" ] &&
            [ "$(cat "$dryad_roots_owning_root_candidate/dyd/type")" = root ]; then
            printf '%s\n' "$dryad_roots_owning_root_candidate"
            return 0
        fi

        dryad_roots_owning_root_next=$(dirname "$dryad_roots_owning_root_candidate")
        [ "$dryad_roots_owning_root_next" != "$dryad_roots_owning_root_candidate" ] || return 1
        dryad_roots_owning_root_candidate=$dryad_roots_owning_root_next
    done
}

dryad_roots_owning_path_within () {
    dryad_roots_owning_within_base=$1
    dryad_roots_owning_within_path=$2

    [ -n "$dryad_roots_owning_within_base" ] || return 1

    case $dryad_roots_owning_within_path in
        "$dryad_roots_owning_within_base" | "$dryad_roots_owning_within_base"/* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_roots_owning_selected_path () {
    dryad_roots_owning_selected_root=$1
    dryad_roots_owning_selected_descriptor=$2
    dryad_roots_owning_selected_kind=$3
    dryad_roots_owning_selected_match=

    for dryad_roots_owning_selected_candidate in "$dryad_roots_owning_selected_root"/dyd/"$dryad_roots_owning_selected_kind"~*; do
        [ -d "$dryad_roots_owning_selected_candidate" ] || continue
        dryad_roots_owning_selected_selector=${dryad_roots_owning_selected_candidate##*/$dryad_roots_owning_selected_kind~}
        if dryad_selector_matches_descriptor "$dryad_roots_owning_selected_selector" "$dryad_roots_owning_selected_descriptor"; then
            [ -z "$dryad_roots_owning_selected_match" ] ||
                dryad_die "multiple $dryad_roots_owning_selected_kind paths match variant: $dryad_roots_owning_selected_descriptor"
            dryad_roots_owning_selected_match=$dryad_roots_owning_selected_candidate
        fi
    done

    if [ -n "$dryad_roots_owning_selected_match" ]; then
        printf '%s\n' "$dryad_roots_owning_selected_match"
        return 0
    fi

    if [ -d "$dryad_roots_owning_selected_root/dyd/$dryad_roots_owning_selected_kind" ]; then
        printf '%s\n' "$dryad_roots_owning_selected_root/dyd/$dryad_roots_owning_selected_kind"
    fi
}

dryad_roots_owning_is_selectable_family () {
    dryad_roots_owning_family_rel=$1

    case $dryad_roots_owning_family_rel in
        dyd/* )
            dryad_roots_owning_family_name=${dryad_roots_owning_family_rel#dyd/}
            dryad_roots_owning_family_name=${dryad_roots_owning_family_name%%/*}
            dryad_roots_owning_family_base=${dryad_roots_owning_family_name%%~*}
            case $dryad_roots_owning_family_base in
                assets | commands | traits | secrets | docs | requirements )
                    return 0
                    ;;
            esac
            ;;
    esac

    return 1
}

dryad_roots_owning_affects_all_variants () {
    dryad_roots_owning_all_rel=$1

    case $dryad_roots_owning_all_rel in
        . | dyd/type | dyd/variants | dyd/variants/* )
            return 0
            ;;
    esac

    if dryad_roots_owning_is_selectable_family "$dryad_roots_owning_all_rel"; then
        return 1
    fi

    return 0
}

dryad_condition_matches_descriptor () {
    dryad_condition=$1
    dryad_condition_descriptor=$2

    [ -n "$dryad_condition" ] || return 0

    dryad_condition_old_ifs=$IFS
    IFS=+
    set -- $dryad_condition
    IFS=$dryad_condition_old_ifs

    for dryad_condition_pair do
        dryad_condition_dim=${dryad_condition_pair%%=*}
        dryad_condition_options=${dryad_condition_pair#*=}
        dryad_condition_value=$(dryad_descriptor_value "$dryad_condition_descriptor" "$dryad_condition_dim" || true)

        case $dryad_condition_options in
            inherit | any )
                continue
                ;;
            host )
                case $dryad_condition_dim in
                    os )
                        dryad_condition_options=$(dryad_host_os)
                        ;;
                    arch )
                        dryad_condition_options=$(dryad_host_arch)
                        ;;
                esac
                ;;
            none )
                [ -z "$dryad_condition_value" ] || return 1
                continue
                ;;
        esac

        [ -n "$dryad_condition_value" ] || return 1
        if ! dryad_option_list_contains "$dryad_condition_options" "$dryad_condition_value"; then
            return 1
        fi
    done

    return 0
}

dryad_roots_owning_requirements_path_matches_variant () {
    dryad_roots_owning_req_path=$1
    dryad_roots_owning_req_changed=$2
    dryad_roots_owning_req_descriptor=$3

    dryad_roots_owning_path_within "$dryad_roots_owning_req_path" "$dryad_roots_owning_req_changed" || return 1

    if [ "$dryad_roots_owning_req_changed" = "$dryad_roots_owning_req_path" ]; then
        return 0
    fi

    dryad_roots_owning_req_rel=${dryad_roots_owning_req_changed#"$dryad_roots_owning_req_path"/}
    case $dryad_roots_owning_req_rel in
        */* )
            return 0
            ;;
    esac

    case $dryad_roots_owning_req_rel in
        *~* )
            dryad_roots_owning_req_condition=${dryad_roots_owning_req_rel#*~}
            dryad_condition_matches_descriptor "$dryad_roots_owning_req_condition" "$dryad_roots_owning_req_descriptor"
            return $?
            ;;
        * )
            return 0
            ;;
    esac
}

dryad_roots_owning_variant_matches_path () {
    dryad_roots_owning_variant_root=$1
    dryad_roots_owning_variant_descriptor=$2
    dryad_roots_owning_variant_changed=$3

    for dryad_roots_owning_variant_kind in assets commands traits secrets docs; do
        dryad_roots_owning_variant_selected=$(dryad_roots_owning_selected_path "$dryad_roots_owning_variant_root" "$dryad_roots_owning_variant_descriptor" "$dryad_roots_owning_variant_kind")
        if dryad_roots_owning_path_within "$dryad_roots_owning_variant_selected" "$dryad_roots_owning_variant_changed"; then
            return 0
        fi
    done

    dryad_roots_owning_variant_requirements=$(dryad_roots_owning_selected_path "$dryad_roots_owning_variant_root" "$dryad_roots_owning_variant_descriptor" requirements)
    dryad_roots_owning_requirements_path_matches_variant \
        "$dryad_roots_owning_variant_requirements" \
        "$dryad_roots_owning_variant_changed" \
        "$dryad_roots_owning_variant_descriptor"
}

dryad_roots_owning_print_ref () {
    dryad_roots_owning_ref_root=$1
    dryad_roots_owning_ref_descriptor=$2

    if [ "$dryad_roots_owning_relative" = 1 ]; then
        case $dryad_roots_owning_ref_root in
            "$dryad_roots_owning_garden"/* )
                dryad_roots_owning_ref_display=${dryad_roots_owning_ref_root#"$dryad_roots_owning_garden"/}
                ;;
            * )
                dryad_roots_owning_ref_display=$dryad_roots_owning_ref_root
                ;;
        esac
    else
        dryad_roots_owning_ref_display=$dryad_roots_owning_ref_root
    fi

    if [ -n "$dryad_roots_owning_ref_descriptor" ]; then
        printf '%s~%s\n' "$dryad_roots_owning_ref_display" "$dryad_roots_owning_ref_descriptor"
    else
        printf '%s\n' "$dryad_roots_owning_ref_display"
    fi
}

dryad_roots_owning_refs_for_path () {
    dryad_roots_owning_input_path=$1
    dryad_roots_owning_paths=$(dryad_roots_owning_paths_for_input "$dryad_roots_owning_input_path")
    dryad_roots_owning_path=$(printf '%s\n' "$dryad_roots_owning_paths" | sed -n '1p')
    dryad_roots_owning_changed=$(printf '%s\n' "$dryad_roots_owning_paths" | sed -n '2p')
    dryad_roots_owning_root=$(dryad_roots_owning_root_for_path "$dryad_roots_owning_path" || true)

    [ -n "$dryad_roots_owning_root" ] || return 0

    if [ "$dryad_roots_owning_changed" = "$dryad_roots_owning_root" ]; then
        dryad_roots_owning_rel=.
    else
        dryad_roots_owning_rel=${dryad_roots_owning_changed#"$dryad_roots_owning_root"/}
    fi

    if dryad_roots_owning_affects_all_variants "$dryad_roots_owning_rel"; then
        dryad_roots_variant_descriptors "$dryad_roots_owning_root" | while IFS= read -r dryad_roots_owning_descriptor; do
            dryad_roots_owning_print_ref "$dryad_roots_owning_root" "$dryad_roots_owning_descriptor"
        done
        return 0
    fi

    dryad_roots_variant_descriptors "$dryad_roots_owning_root" | while IFS= read -r dryad_roots_owning_descriptor; do
        if dryad_roots_owning_variant_matches_path "$dryad_roots_owning_root" "$dryad_roots_owning_descriptor" "$dryad_roots_owning_changed"; then
            dryad_roots_owning_print_ref "$dryad_roots_owning_root" "$dryad_roots_owning_descriptor"
        fi
    done
}

dryad_cmd_roots_owning () {
    dryad_roots_owning_relative=1

    while [ "$#" -gt 0 ]; do
        dryad_roots_owning_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_roots_owning_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad roots owning [--relative=<bool>]
EOF
                return 0
                ;;
            --relative=* )
                dryad_roots_owning_relative=$(dryad_bool_value "${dryad_roots_owning_arg#--relative=}")
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_roots_owning_relative=$(dryad_bool_value "$2")
                            shift 2
                            ;;
                        * )
                            dryad_roots_owning_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_roots_owning_relative=1
                    shift
                fi
                ;;
            --scope=* | --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --scope | --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported roots owning option: $1"
                ;;
            * )
                dryad_die "unsupported roots owning argument: $1"
                ;;
        esac
    done

    dryad_roots_owning_garden=$(dryad_garden_find)

    while IFS= read -r dryad_roots_owning_path; do
        [ -n "$dryad_roots_owning_path" ] || continue
        dryad_roots_owning_refs_for_path "$dryad_roots_owning_path"
    done | sort -u
}

dryad_cmd_roots_affected () {
    dryad_roots_affected_relative=1

    while [ "$#" -gt 0 ]; do
        dryad_roots_affected_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_roots_affected_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad roots affected [--relative=<bool>]
EOF
                return 0
                ;;
            --relative=* )
                dryad_roots_affected_relative=$(dryad_bool_value "${dryad_roots_affected_arg#--relative=}")
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_roots_affected_relative=$(dryad_bool_value "$2")
                            shift 2
                            ;;
                        * )
                            dryad_roots_affected_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_roots_affected_relative=1
                    shift
                fi
                ;;
            --scope=* | --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --scope | --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported roots affected option: $1"
                ;;
            * )
                dryad_die "unsupported roots affected argument: $1"
                ;;
        esac
    done

    dryad_roots_affected_garden=$(dryad_garden_find)
    dryad_roots_owning_garden=$dryad_roots_affected_garden
    dryad_roots_owning_relative=$dryad_roots_affected_relative

    dryad_roots_affected_start_nodes=$(
        while IFS= read -r dryad_roots_affected_path; do
            [ -n "$dryad_roots_affected_path" ] || continue
            dryad_roots_owning_refs_for_path "$dryad_roots_affected_path"
        done | sort -u
    )

    [ -n "$dryad_roots_affected_start_nodes" ] || return 0

    dryad_roots_graph_garden=$dryad_roots_affected_garden
    dryad_roots_graph_lines "$dryad_roots_affected_relative" 1 |
        awk -F '\t' -v starts="$dryad_roots_affected_start_nodes" '
            BEGIN {
                start_count = split(starts, start_parts, "\n")
                for (i = 1; i <= start_count; i++) {
                    if (start_parts[i] != "") {
                        seen[start_parts[i]] = 1
                        queue[++queue_count] = start_parts[i]
                    }
                }
            }

            $2 != "" {
                edge_count[$1]++
                edges[$1, edge_count[$1]] = $3
            }

            END {
                for (head = 1; head <= queue_count; head++) {
                    node = queue[head]
                    for (i = 1; i <= edge_count[node]; i++) {
                        next_node = edges[node, i]
                        if (!(next_node in seen)) {
                            seen[next_node] = 1
                            queue[++queue_count] = next_node
                        }
                    }
                }

                for (node in seen) {
                    print node
                }
            }
        ' |
        sort
}

dryad_roots_graph_all_roots () {
    dryad_roots_graph_dir=$(dryad_roots_path)
    dryad_roots_find_roots "$dryad_roots_graph_dir"
}

dryad_roots_graph_records () {
    dryad_roots_graph_all_roots | while IFS= read -r dryad_roots_graph_record_root; do
        printf 'R\t%s\n' "$dryad_roots_graph_record_root"

        if [ -d "$dryad_roots_graph_record_root/dyd/variants" ]; then
            find "$dryad_roots_graph_record_root/dyd/variants" -mindepth 2 -maxdepth 2 -type f |
                sort |
                while IFS= read -r dryad_roots_graph_record_variant; do
                    dryad_roots_graph_record_variant_rel=${dryad_roots_graph_record_variant#"$dryad_roots_graph_record_root"/dyd/variants/}
                    dryad_roots_graph_record_variant_value=$(sed 's/^[[:space:]]*//;s/[[:space:]]*$//' "$dryad_roots_graph_record_variant")
                    printf 'V\t%s\t%s\t%s\n' "$dryad_roots_graph_record_root" "$dryad_roots_graph_record_variant_rel" "$dryad_roots_graph_record_variant_value"
                done
        fi

        for dryad_roots_graph_record_req_dir in "$dryad_roots_graph_record_root"/dyd/requirements "$dryad_roots_graph_record_root"/dyd/requirements~*; do
            [ -d "$dryad_roots_graph_record_req_dir" ] || continue
            for dryad_roots_graph_record_req_file in "$dryad_roots_graph_record_req_dir"/* "$dryad_roots_graph_record_req_dir"/.[!.]* "$dryad_roots_graph_record_req_dir"/..?*; do
                [ -f "$dryad_roots_graph_record_req_file" ] || [ -L "$dryad_roots_graph_record_req_file" ] || continue
                dryad_roots_graph_record_req_rel=${dryad_roots_graph_record_req_file#"$dryad_roots_graph_record_root"/dyd/}
                dryad_roots_graph_record_req_spec=$(dryad_requirement_target_spec "$dryad_roots_graph_record_req_file")
                printf 'Q\t%s\t%s\t%s\n' "$dryad_roots_graph_record_root" "$dryad_roots_graph_record_req_rel" "$dryad_roots_graph_record_req_spec"
            done
        done
    done
}

dryad_roots_graph_lines () {
    dryad_roots_graph_relative=$1
    dryad_roots_graph_transpose=$2
    dryad_roots_graph_host_os=$(dryad_host_os)
    dryad_roots_graph_host_arch=$(dryad_host_arch)

    dryad_roots_graph_records |
        awk -F '\t' \
            -v garden="$dryad_roots_graph_garden" \
            -v relative="$dryad_roots_graph_relative" \
            -v transpose="$dryad_roots_graph_transpose" \
            -v host_os="$dryad_roots_graph_host_os" \
            -v host_arch="$dryad_roots_graph_host_arch" '
            function trim(value) {
                sub(/^[[:space:]]+/, "", value)
                sub(/[[:space:]]+$/, "", value)
                return value
            }

            function add_dim(root, dim) {
                if (!((root, dim) in dim_seen)) {
                    dims[root, ++dim_count[root]] = dim
                    dim_seen[root, dim] = 1
                }
            }

            function add_option(root, dim, option) {
                add_dim(root, dim)
                if (!((root, dim, option) in option_seen)) {
                    options[root, dim] = options[root, dim] "\034" option
                    option_seen[root, dim, option] = 1
                }
            }

            function descriptor_value(descriptor, dim,    parts, count, i, pair) {
                count = split(descriptor, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    if (pair ~ "^" dim "=") {
                        sub("^[^=]+=", "", pair)
                        return pair
                    }
                }
                return ""
            }

            function option_list_contains(list, target,    parts, count, i) {
                count = split(list, parts, ",")
                for (i = 1; i <= count; i++) {
                    if (parts[i] == target) {
                        return 1
                    }
                }
                return 0
            }

            function selector_matches(descriptor, selector, mode,    parts, count, i, pair, dim, want, got) {
                if (selector == "") {
                    return 1
                }
                count = split(selector, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    dim = pair
                    sub("=.*$", "", dim)
                    want = pair
                    sub("^[^=]+=", "", want)
                    got = descriptor_value(descriptor, dim)

                    if (want == "inherit" && mode == "condition") {
                        continue
                    }
                    if (want == "host") {
                        if (dim == "os") {
                            want = host_os
                        } else if (dim == "arch") {
                            want = host_arch
                        }
                    }
                    if (want == "any") {
                        if (mode == "condition") {
                            continue
                        }
                        if (got == "") {
                            return 0
                        }
                        continue
                    }
                    if (want == "none") {
                        if (got != "") {
                            return 0
                        }
                        continue
                    }
                    if (got == "" || !option_list_contains(want, got)) {
                        return 0
                    }
                }
                return 1
            }

            function target_selector_matches(selector, descriptor, parent_descriptor,    parts, count, i, pair, dim, want, got) {
                if (selector == "") {
                    return 1
                }
                count = split(selector, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    dim = pair
                    sub("=.*$", "", dim)
                    want = pair
                    sub("^[^=]+=", "", want)
                    got = descriptor_value(descriptor, dim)

                    if (want == "inherit") {
                        want = descriptor_value(parent_descriptor, dim)
                        if (want == "") {
                            want = "none"
                        }
                    } else if (want == "host") {
                        if (dim == "os") {
                            want = host_os
                        } else if (dim == "arch") {
                            want = host_arch
                        }
                    }

                    if (want == "any") {
                        if (got == "") {
                            return 0
                        }
                        continue
                    }
                    if (want == "none") {
                        if (got != "") {
                            return 0
                        }
                        continue
                    }
                    if (got == "" || !option_list_contains(want, got)) {
                        return 0
                    }
                }
                return 1
            }

            function include_matches(root, descriptor,    i) {
                if (include_count[root] == 0) {
                    return 1
                }
                for (i = 1; i <= include_count[root]; i++) {
                    if (selector_matches(descriptor, include_rule[root, i], "path")) {
                        return 1
                    }
                }
                return 0
            }

            function exclude_matches(root, descriptor,    i) {
                for (i = 1; i <= exclude_count[root]; i++) {
                    if (selector_matches(descriptor, exclude_rule[root, i], "path")) {
                        return 1
                    }
                }
                return 0
            }

            function emit_descriptor(root, descriptor) {
                if (!include_matches(root, descriptor) || exclude_matches(root, descriptor)) {
                    return
                }
                if (!((root, descriptor) in descriptor_seen)) {
                    descriptor_seen[root, descriptor] = 1
                    descriptors[root, ++descriptor_count[root]] = descriptor
                }
            }

            function walk_descriptors(root, dim_index, prefix,    dim, parts, count, i, option, next_prefix) {
                if (dim_index > dim_count[root]) {
                    emit_descriptor(root, prefix)
                    return
                }

                dim = dims[root, dim_index]
                count = split(options[root, dim], parts, "\034")
                for (i = 2; i <= count; i++) {
                    option = parts[i]
                    if (option == "none") {
                        next_prefix = prefix
                    } else if (prefix == "") {
                        next_prefix = dim "=" option
                    } else {
                        next_prefix = prefix "+" dim "=" option
                    }
                    walk_descriptors(root, dim_index + 1, next_prefix)
                }
            }

            function build_descriptors(    i, root) {
                for (i = 1; i <= root_count; i++) {
                    root = roots[i]
                    if (dim_count[root] == 0) {
                        emit_descriptor(root, "")
                    } else {
                        walk_descriptors(root, 1, "")
                    }
                }
            }

            function dirname(path) {
                sub("/[^/]*$", "", path)
                return path
            }

            function basename(path) {
                sub("^.*/", "", path)
                return path
            }

            function normalize_path(path,    parts, count, i, part, stack, depth, out) {
                count = split(path, parts, "/")
                depth = 0
                for (i = 1; i <= count; i++) {
                    part = parts[i]
                    if (part == "" || part == ".") {
                        continue
                    }
                    if (part == "..") {
                        if (depth > 0) {
                            depth--
                        }
                        continue
                    }
                    stack[++depth] = part
                }
                out = "/"
                for (i = 1; i <= depth; i++) {
                    if (i > 1) {
                        out = out "/"
                    }
                    out = out stack[i]
                }
                return out
            }

            function url_query_to_descriptor(query) {
                gsub(/&/, "+", query)
                return query
            }

            function ref_for(root, descriptor,    display) {
                if (relative == 1 && index(root, garden "/") == 1) {
                    display = substr(root, length(garden) + 2)
                } else {
                    display = root
                }
                if (descriptor != "") {
                    return display "~" descriptor
                }
                return display
            }

            function emit_node(ref) {
                print ref "\t\t"
            }

            function emit_edge(from_ref, req_name, to_ref) {
                if (transpose == 1) {
                    print to_ref "\t" req_name "\t" from_ref
                } else {
                    print from_ref "\t" req_name "\t" to_ref
                }
            }

            function emit_graph(    i, j, q, root, descriptor, from_ref, req_dir, req_selector, req_name, req_alias, edge_name, req_condition, spec, body, query, target_path, req_abs_dir, target_abs, target_root, target_selector, k, target_descriptor, to_ref, match_count, m) {
                for (i = 1; i <= root_count; i++) {
                    root = roots[i]
                    for (j = 1; j <= descriptor_count[root]; j++) {
                        descriptor = descriptors[root, j]
                        from_ref = ref_for(root, descriptor)
                        emit_node(from_ref)

                        for (q = 1; q <= req_count; q++) {
                            if (req_root[q] != root) {
                                continue
                            }

                            req_dir = req_rel[q]
                            sub("/[^/]*$", "", req_dir)
                            req_selector = ""
                            if (req_dir ~ /^requirements~/) {
                                req_selector = req_dir
                                sub(/^requirements~/, "", req_selector)
                                if (!selector_matches(descriptor, req_selector, "path")) {
                                    continue
                                }
                            }

                            req_name = basename(req_rel[q])
                            req_condition = ""
                            if (req_name ~ /~/) {
                                req_condition = req_name
                                sub(/^[^~]*~/, "", req_condition)
                                if (!selector_matches(descriptor, req_condition, "condition")) {
                                    continue
                                }
                            }

                            spec = req_spec[q]
                            if (spec !~ /^root:/) {
                                printf "dryad-sh: error: requirement target must use root: scheme: %s\n", spec > "/dev/stderr"
                                exit 2
                            }
                            body = spec
                            sub(/^root:/, "", body)
                            query = ""
                            target_path = body
                            if (body ~ /\?/) {
                                query = body
                                sub(/^[^?]*\?/, "", query)
                                target_path = body
                                sub(/\?.*$/, "", target_path)
                            }
                            target_selector = url_query_to_descriptor(query)
                            req_abs_dir = root "/dyd/" req_dir
                            target_abs = normalize_path(req_abs_dir "/" target_path)
                            target_root = root_by_path[target_abs]
                            if (target_root == "") {
                                printf "dryad-sh: error: root requirement target not found: %s\n", target_abs > "/dev/stderr"
                                exit 2
                            }

                            match_count = 0
                            for (k = 1; k <= descriptor_count[target_root]; k++) {
                                target_descriptor = descriptors[target_root, k]
                                if (target_selector_matches(target_selector, target_descriptor, descriptor)) {
                                    match_descriptor[++match_count] = target_descriptor
                                    match_ref[match_count] = ref_for(target_root, target_descriptor)
                                }
                            }

                            req_alias = req_name
                            sub(/~.*$/, "", req_alias)
                            for (m = 1; m <= match_count; m++) {
                                edge_name = req_name
                                if (match_count > 1 && match_descriptor[m] != "") {
                                    edge_name = req_alias "~" match_descriptor[m]
                                }
                                emit_edge(from_ref, edge_name, match_ref[m])
                            }
                        }
                    }
                }
            }

            $1 == "R" {
                root = $2
                if (!(root in root_seen)) {
                    roots[++root_count] = root
                    root_seen[root] = 1
                    root_by_path[root] = root
                }
                next
            }

            $1 == "V" {
                root = $2
                rel = $3
                value = trim($4)
                if (value != "true") {
                    next
                }
                split(rel, rel_parts, "/")
                if (rel_parts[1] == "_include") {
                    include_rule[root, ++include_count[root]] = rel_parts[2]
                } else if (rel_parts[1] == "_exclude") {
                    exclude_rule[root, ++exclude_count[root]] = rel_parts[2]
                } else {
                    add_option(root, rel_parts[1], rel_parts[2])
                }
                next
            }

            $1 == "Q" {
                req_root[++req_count] = $2
                req_rel[req_count] = $3
                req_spec[req_count] = trim($4)
                next
            }

            END {
                build_descriptors()
                emit_graph()
            }
        '
}

dryad_roots_graph_print_json_compact () {
    sort -t "$(printf '\t')" -k1,1 -k2,2 -k3,3 |
        awk -F '\t' '
            function json_escape(value) {
                gsub(/\\/, "\\\\", value)
                gsub(/"/, "\\\"", value)
                return value
            }

            function flush_node(    edge_count, i, pair_count, pair) {
                if (current == "") {
                    return
                }

                if (printed_node) {
                    printf ","
                }
                printf "\"%s\":", json_escape(current)

                edge_count = split(edges, edge_parts, "\034")
                if (edge_count < 2) {
                    printf "{}"
                } else {
                    printf "{"
                    for (i = 2; i <= edge_count; i++) {
                        pair_count = split(edge_parts[i], pair, "\035")
                        if (i > 2) {
                            printf ","
                        }
                        printf "\"%s\":\"%s\"", json_escape(pair[1]), json_escape(pair[2])
                    }
                    printf "}"
                }

                printed_node = 1
            }

            BEGIN {
                printf "{"
            }

            {
                if ($1 != current) {
                    flush_node()
                    current = $1
                    edges = ""
                }
                if ($2 != "") {
                    edges = edges "\034" $2 "\035" $3
                }
            }

            END {
                flush_node()
                printf "}\n"
            }
        '
}

dryad_roots_graph_print_yaml () {
    sort -t "$(printf '\t')" -k1,1 -k2,2 -k3,3 |
        awk -F '\t' '
            function flush_node(    edge_count, i, pair_count, pair) {
                if (current == "") {
                    return
                }

                edge_count = split(edges, edge_parts, "\034")
                if (edge_count < 2) {
                    printf "%s: {}\n", current
                } else {
                    printf "%s:\n", current
                    for (i = 2; i <= edge_count; i++) {
                        pair_count = split(edge_parts[i], pair, "\035")
                        printf "  %s: %s\n", pair[1], pair[2]
                    }
                }
            }

            {
                if ($1 != current) {
                    flush_node()
                    current = $1
                    edges = ""
                }
                if ($2 != "") {
                    edges = edges "\034" $2 "\035" $3
                }
            }

            END {
                flush_node()
            }
        '
}

dryad_cmd_roots_graph () {
    dryad_roots_graph_relative=1
    dryad_roots_graph_transpose=0
    dryad_roots_graph_format=yaml

    while [ "$#" -gt 0 ]; do
        dryad_roots_graph_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_roots_graph_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad roots graph [--format=yaml|json|json-compact] [--transpose] [--relative=<bool>]
EOF
                return 0
                ;;
            --format=* )
                dryad_roots_graph_format=${dryad_roots_graph_arg#--format=}
                shift
                ;;
            --format )
                [ "$#" -gt 1 ] || dryad_die "--format requires a value"
                dryad_roots_graph_format=$2
                shift 2
                ;;
            --transpose )
                dryad_roots_graph_transpose=1
                shift
                ;;
            --relative=* )
                dryad_roots_graph_relative=$(dryad_bool_value "${dryad_roots_graph_arg#--relative=}")
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_roots_graph_relative=$(dryad_bool_value "$2")
                            shift 2
                            ;;
                        * )
                            dryad_roots_graph_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_roots_graph_relative=1
                    shift
                fi
                ;;
            --scope=* | --log-level=* | --log-format=* | --parallel=* )
                shift
                ;;
            --scope | --log-level | --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            -- )
                shift
                break
                ;;
            --* )
                dryad_die "unsupported roots graph option: $1"
                ;;
            * )
                dryad_die "unsupported roots graph argument: $1"
                ;;
        esac
    done

    dryad_roots_graph_garden=$(dryad_garden_find)

    case $dryad_roots_graph_format in
        json | JSON | json-compact | JSON-COMPACT )
            dryad_roots_graph_lines "$dryad_roots_graph_relative" "$dryad_roots_graph_transpose" |
                dryad_roots_graph_print_json_compact
            ;;
        yaml | YAML )
            dryad_roots_graph_lines "$dryad_roots_graph_relative" "$dryad_roots_graph_transpose" |
                dryad_roots_graph_print_yaml
            ;;
        * )
            dryad_die "unrecognized output format: $dryad_roots_graph_format"
            ;;
    esac
}

dryad_cmd_roots () {
    dryad_roots_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi
    case $dryad_roots_action in
        path )
            dryad_roots_next=${1:-}
            case $dryad_roots_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad roots path
  dryad roots list
EOF
                    ;;
                * )
                    dryad_roots_path
                    ;;
            esac
            ;;
        graph )
            dryad_cmd_roots_graph "$@"
            ;;
        each )
            dryad_cmd_roots_each "$@"
            ;;
        owning )
            dryad_cmd_roots_owning "$@"
            ;;
        affected )
            dryad_cmd_roots_affected "$@"
            ;;
        build )
            dryad_cmd_roots_build "$@"
            ;;
        list )
            dryad_roots_include=
            dryad_roots_exclude=
            dryad_roots_from_stdin=0
            dryad_roots_to_sprouts=0
            while [ "$#" -gt 0 ]; do
                dryad_roots_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_roots_arg in
                    --include=* )
                        dryad_roots_include="${dryad_roots_include}
${dryad_roots_arg#--include=}"
                        shift
                        ;;
                    --include )
                        [ "$#" -gt 1 ] || dryad_die "--include requires a value"
                        dryad_roots_include="${dryad_roots_include}
$2"
                        shift 2
                        ;;
                    --exclude=* )
                        dryad_roots_exclude="${dryad_roots_exclude}
${dryad_roots_arg#--exclude=}"
                        shift
                        ;;
                    --exclude )
                        [ "$#" -gt 1 ] || dryad_die "--exclude requires a value"
                        dryad_roots_exclude="${dryad_roots_exclude}
$2"
                        shift 2
                        ;;
                    --from-stdin )
                        dryad_roots_from_stdin=1
                        shift
                        ;;
                    --to-sprouts )
                        dryad_roots_to_sprouts=1
                        shift
                        ;;
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad roots list
EOF
                        return 0
                        ;;
                    --scope=* | --log-level=* | --log-format=* | --parallel=* )
                        shift
                        ;;
                    --scope | --log-level | --log-format | --parallel )
                        [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                        shift 2
                        ;;
                    -- )
                        shift
                        break
                        ;;
                    * )
                        dryad_die "unsupported roots list argument: $1"
                        ;;
                esac
            done

            if [ "$dryad_roots_from_stdin" = 1 ]; then
                dryad_roots_list_from_stdin
                return 0
            fi

            dryad_roots_dir=$(dryad_roots_path)
            if [ ! -d "$dryad_roots_dir" ]; then
                return 0
            fi
            dryad_roots_find_roots "$dryad_roots_dir" | while IFS= read -r dryad_roots_root; do
                dryad_roots_rel=${dryad_roots_root#"$dryad_roots_dir"/}
                dryad_roots_variant_descriptors "$dryad_roots_root" | while IFS= read -r dryad_roots_descriptor; do
                    dryad_roots_display=dyd/roots/$dryad_roots_rel
                    if [ -n "$dryad_roots_descriptor" ]; then
                        dryad_roots_display=$dryad_roots_display~$dryad_roots_descriptor
                    fi
                    if dryad_roots_entry_matches_filters "$dryad_roots_display" "$dryad_roots_root" "$dryad_roots_descriptor"; then
                        dryad_roots_print_display "$dryad_roots_display"
                    fi
                done
            done | sort
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad roots affected
  dryad roots build
  dryad roots each
  dryad roots graph
  dryad roots path
  dryad roots list
  dryad roots owning
EOF
            ;;
        * )
            dryad_die "unsupported roots action: $dryad_roots_action"
            ;;
    esac
}
