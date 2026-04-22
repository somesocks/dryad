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
        prune | run | wipe )
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
