dryad_sprouts_path () {
    dryad_sprouts_garden=$(dryad_garden_find)
    printf '%s\n' "$dryad_sprouts_garden/dyd/sprouts"
}

dryad_sprouts_candidate_is_sprout () {
    dryad_sprouts_candidate=$1

    if [ -f "$dryad_sprouts_candidate/dyd/type" ]; then
        dryad_sprouts_candidate_type=$(tr -d '[:space:]' < "$dryad_sprouts_candidate/dyd/type")
        case $dryad_sprouts_candidate_type in
            sprout | stem )
                return 0
                ;;
        esac
    fi

    [ -f "$dryad_sprouts_candidate/dyd/fingerprint" ]
}

dryad_sprouts_find_sprouts () {
    dryad_sprouts_find_dir=$1

    [ -d "$dryad_sprouts_find_dir" ] || return 0

    find "$dryad_sprouts_find_dir" \( -type d -o -type l \) -print |
        while IFS= read -r dryad_sprouts_find_candidate; do
            [ "$dryad_sprouts_find_candidate" != "$dryad_sprouts_find_dir" ] || continue
            if dryad_sprouts_candidate_is_sprout "$dryad_sprouts_find_candidate"; then
                printf '%s\n' "$dryad_sprouts_find_candidate"
            fi
        done |
        sort
}

dryad_sprouts_ref_descriptor () {
    dryad_sprouts_ref=$1

    case $dryad_sprouts_ref in
        *~* )
            dryad_fs_descriptor_normalize "${dryad_sprouts_ref#*~}"
            ;;
        * )
            printf '\n'
            ;;
    esac
}

dryad_sprouts_print_ref () {
    dryad_sprouts_print_path=$1
    dryad_sprouts_print_garden=$2

    if [ "$dryad_sprouts_relative" = 1 ]; then
        case $dryad_sprouts_print_path in
            "$dryad_sprouts_print_garden"/* )
                printf '%s\n' "${dryad_sprouts_print_path#"$dryad_sprouts_print_garden"/}"
                ;;
            * )
                printf '%s\n' "$dryad_sprouts_print_path"
                ;;
        esac
    else
        printf '%s\n' "$dryad_sprouts_print_path"
    fi
}

dryad_sprouts_entry_matches_filters () {
    dryad_sprouts_match_display=$1
    dryad_sprouts_match_path=$2
    dryad_sprouts_match_descriptor=$3

    dryad_roots_entry_matches_filters "$dryad_sprouts_match_display" "$dryad_sprouts_match_path" "$dryad_sprouts_match_descriptor"
}

dryad_sprouts_list_from_stdin () {
    dryad_sprouts_list_garden=$1

    while IFS= read -r dryad_sprouts_stdin_ref; do
        [ -n "$dryad_sprouts_stdin_ref" ] || continue
        dryad_sprouts_stdin_path=${dryad_sprouts_stdin_ref%%\?*}
        dryad_sprouts_stdin_query=
        case $dryad_sprouts_stdin_ref in
            *\?* )
                dryad_sprouts_stdin_query=${dryad_sprouts_stdin_ref#*\?}
                ;;
        esac
        dryad_sprouts_stdin_descriptor=$(printf '%s\n' "$dryad_sprouts_stdin_query" | tr '&' '+')
        if [ -n "$dryad_sprouts_stdin_descriptor" ]; then
            dryad_sprouts_stdin_display=$dryad_sprouts_stdin_path~$dryad_sprouts_stdin_descriptor
        else
            dryad_sprouts_stdin_display=$dryad_sprouts_stdin_path
        fi
        dryad_sprouts_stdin_descriptor=$(dryad_sprouts_ref_descriptor "$dryad_sprouts_stdin_display")
        if dryad_sprouts_entry_matches_filters "$dryad_sprouts_stdin_display" "$dryad_sprouts_stdin_path" "$dryad_sprouts_stdin_descriptor"; then
            case $dryad_sprouts_stdin_display in
                /* )
                    dryad_sprouts_print_ref "$dryad_sprouts_stdin_display" "$dryad_sprouts_list_garden"
                    ;;
                * )
                    printf '%s\n' "$dryad_sprouts_stdin_display"
                    ;;
            esac
        fi
    done
}

dryad_path_has_owner_write () {
    dryad_path_has_owner_write_path=$1
    dryad_path_has_owner_write_mode=$(ls -ld "$dryad_path_has_owner_write_path" 2>/dev/null | sed 's/ .*//') ||
        return 1

    case $dryad_path_has_owner_write_mode in
        ??w* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_path_chmod_parent_owner_write () {
    dryad_path_chmod_parent_path=$1
    dryad_path_chmod_parent_parent=$(dirname "$dryad_path_chmod_parent_path")
    if [ -d "$dryad_path_chmod_parent_parent" ] &&
        ! dryad_path_has_owner_write "$dryad_path_chmod_parent_parent"; then
        chmod u+w "$dryad_path_chmod_parent_parent"
    fi
}

dryad_sprouts_ensure_dir () {
    dryad_sprouts_ensure_garden=$1
    dryad_sprouts_ensure_dir=$dryad_sprouts_ensure_garden/dyd/sprouts

    if [ -e "$dryad_sprouts_ensure_dir" ] && [ ! -d "$dryad_sprouts_ensure_dir" ]; then
        dryad_die "sprouts path exists and is not a directory: $dryad_sprouts_ensure_dir"
    fi

    if [ ! -d "$dryad_sprouts_ensure_dir" ]; then
        dryad_path_chmod_parent_owner_write "$dryad_sprouts_ensure_dir"
        mkdir -p "$dryad_sprouts_ensure_dir"
        chmod 551 "$dryad_sprouts_ensure_dir"
    fi

    printf '%s\n' "$dryad_sprouts_ensure_dir"
}

dryad_sprouts_make_tree_removable () {
    dryad_sprouts_make_tree_path=$1

    [ -e "$dryad_sprouts_make_tree_path" ] || [ -L "$dryad_sprouts_make_tree_path" ] || return 0

    dryad_path_chmod_parent_owner_write "$dryad_sprouts_make_tree_path"

    if [ -d "$dryad_sprouts_make_tree_path" ] && [ ! -L "$dryad_sprouts_make_tree_path" ]; then
        find "$dryad_sprouts_make_tree_path" -type d -exec chmod u+w {} \;
    fi
}

dryad_sprouts_remove_children () {
    dryad_sprouts_remove_children_dir=$1

    [ -d "$dryad_sprouts_remove_children_dir" ] || return 0

    dryad_sprouts_remove_restore_owner_write=0
    if ! dryad_path_has_owner_write "$dryad_sprouts_remove_children_dir"; then
        dryad_sprouts_remove_restore_owner_write=1
    fi

    dryad_sprouts_make_tree_removable "$dryad_sprouts_remove_children_dir"

    find "$dryad_sprouts_remove_children_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} \;

    if [ "$dryad_sprouts_remove_restore_owner_write" = 1 ]; then
        chmod u-w "$dryad_sprouts_remove_children_dir"
    fi
}

dryad_sprouts_prune () {
    dryad_sprouts_prune_garden=$1
    dryad_sprouts_prune_dir=$(dryad_sprouts_ensure_dir "$dryad_sprouts_prune_garden")
    dryad_sprouts_prune_roots=$dryad_sprouts_prune_garden/dyd/roots

    find "$dryad_sprouts_prune_dir" -depth -mindepth 1 -print |
        while IFS= read -r dryad_sprouts_prune_entry; do
            dryad_sprouts_prune_rel=${dryad_sprouts_prune_entry#"$dryad_sprouts_prune_dir"/}
            dryad_sprouts_prune_root_equiv=$dryad_sprouts_prune_roots/$dryad_sprouts_prune_rel

            if [ ! -e "$dryad_sprouts_prune_root_equiv" ] && [ ! -L "$dryad_sprouts_prune_root_equiv" ]; then
                dryad_sprouts_prune_parent=$(dirname "$dryad_sprouts_prune_entry")
                dryad_sprouts_prune_restore_parent=0
                if [ -d "$dryad_sprouts_prune_parent" ] &&
                    ! dryad_path_has_owner_write "$dryad_sprouts_prune_parent"; then
                    dryad_sprouts_prune_restore_parent=1
                fi

                dryad_sprouts_make_tree_removable "$dryad_sprouts_prune_entry"
                dryad_sprouts_prune_remove_status=0
                rm -rf "$dryad_sprouts_prune_entry" ||
                    dryad_sprouts_prune_remove_status=$?

                if [ "$dryad_sprouts_prune_restore_parent" = 1 ] &&
                    [ -d "$dryad_sprouts_prune_parent" ]; then
                    chmod u-w "$dryad_sprouts_prune_parent"
                fi

                [ "$dryad_sprouts_prune_remove_status" = 0 ] ||
                    return "$dryad_sprouts_prune_remove_status"
            fi
        done
}

dryad_cmd_sprouts_no_arg_action () {
    dryad_sprouts_no_arg_action=$1
    shift

    while [ "$#" -gt 0 ]; do
        dryad_sprouts_no_arg_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_no_arg_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad sprouts $dryad_sprouts_no_arg_action
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
            --* )
                dryad_die "unsupported sprouts $dryad_sprouts_no_arg_action option: $1"
                ;;
            * )
                dryad_die "unsupported sprouts $dryad_sprouts_no_arg_action argument: $1"
                ;;
        esac
    done

    dryad_sprouts_no_arg_garden=$(dryad_garden_find)
    case $dryad_sprouts_no_arg_action in
        wipe )
            dryad_sprouts_no_arg_dir=$(dryad_sprouts_ensure_dir "$dryad_sprouts_no_arg_garden")
            dryad_sprouts_remove_children "$dryad_sprouts_no_arg_dir"
            ;;
        prune )
            dryad_sprouts_prune "$dryad_sprouts_no_arg_garden"
            ;;
        * )
            dryad_die "unsupported sprouts action: $dryad_sprouts_no_arg_action"
            ;;
    esac
}

dryad_cmd_sprouts_path () {
    while [ "$#" -gt 0 ]; do
        dryad_sprouts_path_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_path_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad sprouts path
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
            --* )
                dryad_die "unsupported sprouts path option: $1"
                ;;
            * )
                dryad_die "unsupported sprouts path argument: $1"
                ;;
        esac
    done

    dryad_sprouts_path
}

dryad_cmd_sprouts_list () {
    dryad_roots_include=
    dryad_roots_exclude=
    dryad_sprouts_from_stdin=0
    dryad_sprouts_relative=1

    while [ "$#" -gt 0 ]; do
        dryad_sprouts_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad sprouts list [--relative=<bool>] [--include=<filter>] [--exclude=<filter>] [--from-stdin]
EOF
                return 0
                ;;
            --include=* )
                dryad_roots_include="${dryad_roots_include}
${dryad_sprouts_arg#--include=}"
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
${dryad_sprouts_arg#--exclude=}"
                shift
                ;;
            --exclude )
                [ "$#" -gt 1 ] || dryad_die "--exclude requires a value"
                dryad_roots_exclude="${dryad_roots_exclude}
$2"
                shift 2
                ;;
            --from-stdin )
                dryad_sprouts_from_stdin=1
                shift
                ;;
            --relative=* )
                dryad_sprouts_relative=$(dryad_bool_value "${dryad_sprouts_arg#--relative=}")
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_sprouts_relative=$(dryad_bool_value "$2")
                            shift 2
                            ;;
                        * )
                            dryad_sprouts_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_sprouts_relative=1
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
                dryad_die "unsupported sprouts list option: $1"
                ;;
            * )
                dryad_die "unsupported sprouts list argument: $1"
                ;;
        esac
    done

    dryad_sprouts_garden=$(dryad_garden_find)

    if [ "$dryad_sprouts_from_stdin" = 1 ]; then
        dryad_sprouts_list_from_stdin "$dryad_sprouts_garden" | sort
        return 0
    fi

    dryad_sprouts_dir=$dryad_sprouts_garden/dyd/sprouts
    dryad_sprouts_find_sprouts "$dryad_sprouts_dir" | while IFS= read -r dryad_sprouts_path_entry; do
        dryad_sprouts_display=${dryad_sprouts_path_entry#"$dryad_sprouts_garden"/}
        dryad_sprouts_descriptor=$(dryad_sprouts_ref_descriptor "$dryad_sprouts_display")
        if dryad_sprouts_entry_matches_filters "$dryad_sprouts_display" "$dryad_sprouts_path_entry" "$dryad_sprouts_descriptor"; then
            dryad_sprouts_print_ref "$dryad_sprouts_path_entry" "$dryad_sprouts_garden"
        fi
    done | sort
}

dryad_cmd_sprouts () {
    dryad_sprouts_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_sprouts_action in
        path )
            dryad_cmd_sprouts_path "$@"
            ;;
        list )
            dryad_cmd_sprouts_list "$@"
            ;;
        prune | wipe )
            dryad_cmd_sprouts_no_arg_action "$dryad_sprouts_action" "$@"
            ;;
        run )
            dryad_sprouts_next=${1:-}
            case $dryad_sprouts_next in
                --help | -h )
                    cat <<EOF
Usage:
  dryad sprouts $dryad_sprouts_action
EOF
                    ;;
                * )
                    dryad_die "sprouts $dryad_sprouts_action is not supported by dryad-sh yet"
                    ;;
            esac
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad sprouts list
  dryad sprouts path
  dryad sprouts prune
  dryad sprouts run
  dryad sprouts wipe
EOF
            ;;
        * )
            dryad_die "unsupported sprouts action: $dryad_sprouts_action"
            ;;
    esac
}
