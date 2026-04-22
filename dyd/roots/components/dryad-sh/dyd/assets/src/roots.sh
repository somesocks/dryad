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

    sed 's/^[[:space:]]*//;s/[[:space:]]*$//' "$dryad_requirement_file"
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

dryad_cmd_root () {
    dryad_root_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_root_action in
        ancestors )
            dryad_cmd_root_ancestors "$@"
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

dryad_roots_variant_descriptors () {
    dryad_roots_variant_root=$1
    dryad_roots_variant_dir=$dryad_roots_variant_root/dyd/variants

    if [ ! -d "$dryad_roots_variant_dir" ]; then
        printf '\n'
        return 0
    fi

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

    if [ -z "$dryad_roots_include" ]; then
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

dryad_roots_owning_abs_path () {
    dryad_roots_owning_abs_input=$1
    case $dryad_roots_owning_abs_input in
        /* )
            printf '%s\n' "$dryad_roots_owning_abs_input"
            ;;
        * )
            printf '%s/%s\n' "$(pwd -P)" "$dryad_roots_owning_abs_input"
            ;;
    esac
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
    dryad_roots_owning_changed=$(dryad_roots_owning_abs_path "$dryad_roots_owning_input_path")
    dryad_roots_owning_root=$(dryad_roots_owning_root_for_path "$dryad_roots_owning_changed" || true)

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
        owning )
            dryad_cmd_roots_owning "$@"
            ;;
        affected | build )
            dryad_roots_next=${1:-}
            case $dryad_roots_next in
                --help | -h )
                    cat <<EOF
Usage:
  dryad roots $dryad_roots_action
EOF
                    ;;
                * )
                    dryad_die "roots $dryad_roots_action is not supported by dryad-sh yet"
                    ;;
            esac
            ;;
        list )
            dryad_roots_include=
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
                    if dryad_roots_entry_matches_includes "$dryad_roots_display" "$dryad_roots_root" "$dryad_roots_descriptor"; then
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
