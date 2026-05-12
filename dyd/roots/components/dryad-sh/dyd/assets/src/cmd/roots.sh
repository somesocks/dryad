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

    dryad_root_path_find_load "${dryad_root_path_target:-.}"
    printf '%s\n' "$dyd_ret0"
}

dryad_root_create_resolve_target () {
    dryad_root_create_input=$1
    dryad_join_path_load "$(pwd -P)" "$dryad_root_create_input"
    dryad_root_create_abs=$dyd_ret0

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
    dryad_root_selected_variant_descriptor_load "$dryad_root_requirements_root" "$dryad_root_requirements_variant"
    dryad_root_requirements_variant=$dyd_ret0

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

    dryad_root_path_find_load .
    dryad_root_requirement_add_root=$dyd_ret0
    dryad_root_selected_variant_descriptor_load "$dryad_root_requirement_add_root" "$dryad_root_requirement_add_variant"
    dryad_root_requirement_add_variant=$dyd_ret0
    dryad_root_requirement_add_dir=$(dryad_root_requirements_path_for_variant "$dryad_root_requirement_add_root" "$dryad_root_requirement_add_variant")

    dryad_root_requirement_add_parsed=$(dryad_root_requirement_parse_target "$dryad_root_requirement_add_target")
    dryad_root_requirement_add_dep_path=$(printf '%s\n' "$dryad_root_requirement_add_parsed" | sed -n '1p')
    dryad_root_requirement_add_dep_query=$(printf '%s\n' "$dryad_root_requirement_add_parsed" | sed -n '2p')
    dryad_root_path_find_load "$dryad_root_requirement_add_dep_path"
    dryad_root_requirement_add_dep_root=$dyd_ret0

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
    dryad_root_path_find_load .
    dryad_root_requirement_remove_root=$dyd_ret0
    dryad_root_selected_variant_descriptor_load "$dryad_root_requirement_remove_root" "$dryad_root_requirement_remove_variant"
    dryad_root_requirement_remove_variant=$dyd_ret0
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

dryad_requirement_target_spec_load () {
    dryad_requirement_file=$1

    if [ -L "$dryad_requirement_file" ]; then
        dryad_requirement_link=$(readlink "$dryad_requirement_file") ||
            dryad_die "could not read requirement symlink: $dryad_requirement_file"
        case $dryad_requirement_link in
            root:* )
                dyd_ret0=$dryad_requirement_link
                ;;
            * )
                dryad_requirement_dir=$(dirname "$dryad_requirement_file")
                dryad_join_path_load "$dryad_requirement_dir" "$dryad_requirement_link"
                dryad_requirement_target=$dyd_ret0
                dryad_root_path_find_load "$dryad_requirement_target"
                dryad_requirement_root=$dyd_ret0
                dryad_requirement_rel=$(dryad_relative_path "$dryad_requirement_dir" "$dryad_requirement_root")
                dyd_ret0=root:$dryad_requirement_rel
                ;;
        esac
        return 0
    fi

    dryad_requirement_spec=$(sed 's/^[[:space:]]*//;s/[[:space:]]*$//' "$dryad_requirement_file")
    dryad_requirement_size=$(wc -c < "$dryad_requirement_file" | tr -d ' ')
    dryad_requirement_trimmed_size=$(printf '%s' "$dryad_requirement_spec" | wc -c | tr -d ' ')
    if [ "$dryad_requirement_size" != "$dryad_requirement_trimmed_size" ]; then
        dryad_file_abs_path_load "$dryad_requirement_file"
        dryad_requirement_abs=$dyd_ret0
        dryad_requirement_garden=$(dryad_garden_find)
        dryad_requirement_display=${dryad_requirement_abs#"$dryad_requirement_garden"/}
        printf '%s\n' "dryad-sh: malformed requirement file path=$dryad_requirement_display expected=\"$dryad_requirement_spec\"" >&2
    fi
    dyd_ret0=$dryad_requirement_spec
}

dryad_requirement_target_url () {
    dryad_requirement_file=$1
    dryad_requirement_dir=$(dirname "$dryad_requirement_file")
    dryad_requirement_target_spec_load "$dryad_requirement_file"
    dryad_requirement_spec=$dyd_ret0

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

    dryad_join_path_load "$dryad_requirement_dir" "$dryad_requirement_target_ref"
    dryad_requirement_target_path=$dyd_ret0
    dryad_root_path_find_load "$dryad_requirement_target_path"
    dryad_requirement_target_root=$dyd_ret0
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

        dryad_file_abs_path_load "$dryad_root_requirement_entry"
        dryad_root_requirement_abs=$dyd_ret0
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

    dryad_root_path_find_load "$dryad_root_requirements_path"
    dryad_root_requirements_root=$dyd_ret0
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

    dryad_root_path_find_load "${dryad_root_variants_target:-.}"
    dryad_root_variants_root=$dyd_ret0
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
    dryad_root_path_find_load "$dryad_root_walk_path"
    dryad_root_walk_root=$dyd_ret0
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
        dryad_root_path_find_load "$dryad_roots_build_stdin_path"
        dryad_roots_build_stdin_root=$dyd_ret0

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
    dryad_roots_build_run_tab=$(printf '\t')

    printf '%s\n' "$dryad_roots_build_run_entries" |
        while IFS= read -r dryad_roots_build_run_entry; do
            case $dryad_roots_build_run_entry in
                *"$dryad_roots_build_run_tab"* )
                    dryad_roots_build_run_entry_root=${dryad_roots_build_run_entry%%"$dryad_roots_build_run_tab"*}
                    ;;
                * )
                    dryad_roots_build_run_entry_root=$dryad_roots_build_run_entry
                    ;;
            esac
            [ -n "$dryad_roots_build_run_entry_root" ] || continue
            printf '%s\n' "$dryad_roots_build_run_entry_root"
        done |
        sort -u > "$dryad_roots_build_run_roots"

    while IFS= read -r dryad_roots_build_run_root; do
        [ -n "$dryad_roots_build_run_root" ] || continue
        dryad_roots_build_run_descriptors=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-descriptors.XXXXXX")
        printf '%s\n' "$dryad_roots_build_run_entries" |
            while IFS= read -r dryad_roots_build_run_entry; do
                case $dryad_roots_build_run_entry in
                    *"$dryad_roots_build_run_tab"* )
                        dryad_roots_build_run_entry_root=${dryad_roots_build_run_entry%%"$dryad_roots_build_run_tab"*}
                        dryad_roots_build_run_entry_descriptor=${dryad_roots_build_run_entry#*"$dryad_roots_build_run_tab"}
                        dryad_roots_build_run_entry_descriptor=${dryad_roots_build_run_entry_descriptor%%"$dryad_roots_build_run_tab"*}
                        ;;
                    * )
                        dryad_roots_build_run_entry_root=$dryad_roots_build_run_entry
                        dryad_roots_build_run_entry_descriptor=
                        ;;
                esac
                if [ "$dryad_roots_build_run_entry_root" = "$dryad_roots_build_run_root" ]; then
                    printf '%s\n' "$dryad_roots_build_run_entry_descriptor"
                fi
            done > "$dryad_roots_build_run_descriptors"
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
    dryad_root_path_find_load "$dryad_root_build_path"
    dryad_root_build_root=$dyd_ret0
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
        dryad_root_path_find_load "$dryad_roots_each_stdin_path"
        dryad_roots_each_stdin_root=$dyd_ret0

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
    dryad_roots_owning_clean_rest=${dryad_roots_owning_clean_input#/}
    dryad_roots_owning_clean_out=

    while [ -n "$dryad_roots_owning_clean_rest" ]; do
        case $dryad_roots_owning_clean_rest in
            */* )
                dryad_roots_owning_clean_part=${dryad_roots_owning_clean_rest%%/*}
                dryad_roots_owning_clean_rest=${dryad_roots_owning_clean_rest#*/}
                ;;
            * )
                dryad_roots_owning_clean_part=$dryad_roots_owning_clean_rest
                dryad_roots_owning_clean_rest=
                ;;
        esac

        case $dryad_roots_owning_clean_part in
            '' | . )
                ;;
            .. )
                case $dryad_roots_owning_clean_out in
                    */* ) dryad_roots_owning_clean_out=${dryad_roots_owning_clean_out%/*} ;;
                    * ) dryad_roots_owning_clean_out= ;;
                esac
                ;;
            * )
                if [ -n "$dryad_roots_owning_clean_out" ]; then
                    dryad_roots_owning_clean_out=$dryad_roots_owning_clean_out/$dryad_roots_owning_clean_part
                else
                    dryad_roots_owning_clean_out=$dryad_roots_owning_clean_part
                fi
                ;;
        esac
    done

    if [ -n "$dryad_roots_owning_clean_out" ]; then
        printf '/%s\n' "$dryad_roots_owning_clean_out"
    else
        printf '/\n'
    fi
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

dryad_roots_owning_selected_path_compute_load () {
    dryad_roots_owning_selected_root=$1
    dryad_roots_owning_selected_descriptor=$2
    dryad_roots_owning_selected_kind=$3
    dyd_ret0=
    dryad_roots_owning_selected_match=
    dryad_roots_owning_selected_count=0

    dryad_roots_owning_selected_plain=$dryad_roots_owning_selected_root/dyd/$dryad_roots_owning_selected_kind
    if [ -d "$dryad_roots_owning_selected_plain" ]; then
        dryad_roots_owning_selected_match=$dryad_roots_owning_selected_plain
        dryad_roots_owning_selected_count=1
    fi

    for dryad_roots_owning_selected_candidate in "$dryad_roots_owning_selected_root"/dyd/"$dryad_roots_owning_selected_kind"~*; do
        [ -d "$dryad_roots_owning_selected_candidate" ] || continue
        dryad_roots_owning_selected_selector=${dryad_roots_owning_selected_candidate##*/$dryad_roots_owning_selected_kind~}
        if dryad_selector_matches_descriptor "$dryad_roots_owning_selected_selector" "$dryad_roots_owning_selected_descriptor"; then
            dryad_roots_owning_selected_count=$((dryad_roots_owning_selected_count + 1))
            [ "$dryad_roots_owning_selected_count" -le 1 ] ||
                dryad_die "multiple matching dyd/$dryad_roots_owning_selected_kind selectors for variant $dryad_roots_owning_selected_descriptor"
            dryad_roots_owning_selected_match=$dryad_roots_owning_selected_candidate
        fi
    done

    dyd_ret0=$dryad_roots_owning_selected_match
}

dryad_roots_owning_selected_path_load () {
    dryad_roots_owning_selected_root=$1
    dryad_roots_owning_selected_descriptor=$2
    dryad_roots_owning_selected_kind=$3

    if dryad_memo_get_line_load roots-owning-selected-path "$dryad_roots_owning_selected_root" "$dryad_roots_owning_selected_descriptor" "$dryad_roots_owning_selected_kind"; then
        dryad_roots_owning_selected_value=$dyd_ret0
        dryad_profile_count memo.hit.roots-owning-selected-path
        dyd_ret0=$dryad_roots_owning_selected_value
        return 0
    fi

    dryad_profile_count memo.miss.roots-owning-selected-path
    dryad_profile_count call.roots-owning-selected-path.uncached
    dryad_roots_owning_selected_path_compute_load "$dryad_roots_owning_selected_root" "$dryad_roots_owning_selected_descriptor" "$dryad_roots_owning_selected_kind" ||
        return $?
    dryad_roots_owning_selected_value=$dyd_ret0
    dryad_memo_put_value roots-owning-selected-path "$dryad_roots_owning_selected_value" "$dryad_roots_owning_selected_root" "$dryad_roots_owning_selected_descriptor" "$dryad_roots_owning_selected_kind"
    dyd_ret0=$dryad_roots_owning_selected_value
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
        if dryad_descriptor_value_load "$dryad_condition_descriptor" "$dryad_condition_dim"; then
            dryad_condition_value=$dyd_ret0
        else
            dryad_condition_value=
        fi

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
        dryad_roots_owning_selected_path_load "$dryad_roots_owning_variant_root" "$dryad_roots_owning_variant_descriptor" "$dryad_roots_owning_variant_kind"
        dryad_roots_owning_variant_selected=$dyd_ret0
        if dryad_roots_owning_path_within "$dryad_roots_owning_variant_selected" "$dryad_roots_owning_variant_changed"; then
            return 0
        fi
    done

    dryad_roots_owning_selected_path_load "$dryad_roots_owning_variant_root" "$dryad_roots_owning_variant_descriptor" requirements
    dryad_roots_owning_variant_requirements=$dyd_ret0
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
                dryad_requirement_target_spec_load "$dryad_roots_graph_record_req_file"
                dryad_roots_graph_record_req_spec=$dyd_ret0
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

dryad_roots_graph_json_escape () {
    printf '%s\n' "$1" | sed 's/\\/\\\\/g;s/"/\\"/g'
}

dryad_roots_graph_print_json_flush () {
    [ -n "$dryad_roots_graph_print_current" ] || return 0

    if [ "$dryad_roots_graph_print_printed" = 1 ]; then
        printf ','
    fi
    printf '"%s":' "$(dryad_roots_graph_json_escape "$dryad_roots_graph_print_current")"

    dryad_roots_graph_print_edge_printed=0
    printf '{'
    while IFS= read -r dryad_roots_graph_print_edge; do
        [ -n "$dryad_roots_graph_print_edge" ] || continue
        dryad_roots_graph_print_edge_label=${dryad_roots_graph_print_edge%%"$dryad_roots_graph_print_tab"*}
        dryad_roots_graph_print_edge_target=${dryad_roots_graph_print_edge#*"$dryad_roots_graph_print_tab"}
        if [ "$dryad_roots_graph_print_edge_printed" = 1 ]; then
            printf ','
        fi
        printf '"%s":"%s"' \
            "$(dryad_roots_graph_json_escape "$dryad_roots_graph_print_edge_label")" \
            "$(dryad_roots_graph_json_escape "$dryad_roots_graph_print_edge_target")"
        dryad_roots_graph_print_edge_printed=1
    done <<EOF
$dryad_roots_graph_print_edges
EOF
    printf '}'
    dryad_roots_graph_print_printed=1
}

dryad_roots_graph_print_yaml_flush () {
    [ -n "$dryad_roots_graph_print_current" ] || return 0

    if [ -z "$dryad_roots_graph_print_edges" ]; then
        printf '%s: {}\n' "$dryad_roots_graph_print_current"
        return 0
    fi

    printf '%s:\n' "$dryad_roots_graph_print_current"
    while IFS= read -r dryad_roots_graph_print_edge; do
        [ -n "$dryad_roots_graph_print_edge" ] || continue
        dryad_roots_graph_print_edge_label=${dryad_roots_graph_print_edge%%"$dryad_roots_graph_print_tab"*}
        dryad_roots_graph_print_edge_target=${dryad_roots_graph_print_edge#*"$dryad_roots_graph_print_tab"}
        printf '  %s: %s\n' "$dryad_roots_graph_print_edge_label" "$dryad_roots_graph_print_edge_target"
    done <<EOF
$dryad_roots_graph_print_edges
EOF
}

dryad_roots_graph_print_json_compact () {
    dryad_roots_graph_print_tab=$(printf '\t')
    printf '{'
    sort -t "$dryad_roots_graph_print_tab" -k1,1 -k2,2 -k3,3 |
        {
            dryad_roots_graph_print_current=
            dryad_roots_graph_print_edges=
            dryad_roots_graph_print_printed=0
            while IFS= read -r dryad_roots_graph_print_line; do
                case $dryad_roots_graph_print_line in
                    *"$dryad_roots_graph_print_tab"* )
                        dryad_roots_graph_print_node=${dryad_roots_graph_print_line%%"$dryad_roots_graph_print_tab"*}
                        dryad_roots_graph_print_rest=${dryad_roots_graph_print_line#*"$dryad_roots_graph_print_tab"}
                        case $dryad_roots_graph_print_rest in
                            *"$dryad_roots_graph_print_tab"* )
                                dryad_roots_graph_print_label=${dryad_roots_graph_print_rest%%"$dryad_roots_graph_print_tab"*}
                                dryad_roots_graph_print_target=${dryad_roots_graph_print_rest#*"$dryad_roots_graph_print_tab"}
                                dryad_roots_graph_print_target=${dryad_roots_graph_print_target%%"$dryad_roots_graph_print_tab"*}
                                ;;
                            * )
                                dryad_roots_graph_print_label=$dryad_roots_graph_print_rest
                                dryad_roots_graph_print_target=
                                ;;
                        esac
                        ;;
                    * )
                        dryad_roots_graph_print_node=$dryad_roots_graph_print_line
                        dryad_roots_graph_print_label=
                        dryad_roots_graph_print_target=
                        ;;
                esac

                if [ "$dryad_roots_graph_print_node" != "$dryad_roots_graph_print_current" ]; then
                    dryad_roots_graph_print_json_flush
                    dryad_roots_graph_print_current=$dryad_roots_graph_print_node
                    dryad_roots_graph_print_edges=
                fi
                if [ -n "$dryad_roots_graph_print_label" ]; then
                    dryad_roots_graph_print_edges="${dryad_roots_graph_print_edges}${dryad_roots_graph_print_label}	${dryad_roots_graph_print_target}
"
                fi
            done
            dryad_roots_graph_print_json_flush
        }
    printf '}\n'
}

dryad_roots_graph_print_yaml () {
    dryad_roots_graph_print_tab=$(printf '\t')
    sort -t "$dryad_roots_graph_print_tab" -k1,1 -k2,2 -k3,3 |
        {
            dryad_roots_graph_print_current=
            dryad_roots_graph_print_edges=
            while IFS= read -r dryad_roots_graph_print_line; do
                case $dryad_roots_graph_print_line in
                    *"$dryad_roots_graph_print_tab"* )
                        dryad_roots_graph_print_node=${dryad_roots_graph_print_line%%"$dryad_roots_graph_print_tab"*}
                        dryad_roots_graph_print_rest=${dryad_roots_graph_print_line#*"$dryad_roots_graph_print_tab"}
                        case $dryad_roots_graph_print_rest in
                            *"$dryad_roots_graph_print_tab"* )
                                dryad_roots_graph_print_label=${dryad_roots_graph_print_rest%%"$dryad_roots_graph_print_tab"*}
                                dryad_roots_graph_print_target=${dryad_roots_graph_print_rest#*"$dryad_roots_graph_print_tab"}
                                dryad_roots_graph_print_target=${dryad_roots_graph_print_target%%"$dryad_roots_graph_print_tab"*}
                                ;;
                            * )
                                dryad_roots_graph_print_label=$dryad_roots_graph_print_rest
                                dryad_roots_graph_print_target=
                                ;;
                        esac
                        ;;
                    * )
                        dryad_roots_graph_print_node=$dryad_roots_graph_print_line
                        dryad_roots_graph_print_label=
                        dryad_roots_graph_print_target=
                        ;;
                esac

                if [ "$dryad_roots_graph_print_node" != "$dryad_roots_graph_print_current" ]; then
                    dryad_roots_graph_print_yaml_flush
                    dryad_roots_graph_print_current=$dryad_roots_graph_print_node
                    dryad_roots_graph_print_edges=
                fi
                if [ -n "$dryad_roots_graph_print_label" ]; then
                    dryad_roots_graph_print_edges="${dryad_roots_graph_print_edges}${dryad_roots_graph_print_label}	${dryad_roots_graph_print_target}
"
                fi
            done
            dryad_roots_graph_print_yaml_flush
        }
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
