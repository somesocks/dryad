dryad_cmd_scripts () {
    dryad_scripts_action=${1:-}
    dryad_scripts_next=${2:-}
    case $dryad_scripts_action in
        list )
            case $dryad_scripts_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad scripts list [--scope=<name>] [--path] [--oneline=false]
EOF
                    ;;
                * )
                    dryad_scripts_list "$@"
                    ;;
            esac
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad scripts list [--scope=<name>] [--path] [--oneline=false]
EOF
            ;;
        * )
            dryad_die "unsupported scripts action: $dryad_scripts_action"
            ;;
    esac
}

dryad_cmd_script () {
    dryad_script_action=${1:-}
    case $dryad_script_action in
        create | edit )
            shift
            dryad_script_edit "$dryad_script_action" "$@"
            ;;
        get )
            shift
            dryad_script_get "$@"
            ;;
        path )
            shift
            dryad_script_path_command "$@"
            ;;
        run )
            shift
            dryad_cmd_run "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad script create <script>
  dryad script edit <script>
  dryad script get <script>
  dryad script path <script>
  dryad script run <script> [--scope=<name>] -- [args...]
EOF
            ;;
        * )
            dryad_die "unsupported script action: $dryad_script_action"
            ;;
    esac
}

dryad_script_scope_dir () {
    dryad_garden_find_load
    dryad_script_scope_garden=$dyd_ret0
    dryad_scope_resolve_load
    dryad_script_scope_name=$dyd_ret0

    if [ -z "$dryad_script_scope_name" ] || [ "$dryad_script_scope_name" = none ]; then
        dryad_die "no scope set, can't find command"
    fi

    dryad_script_scope_path=$dryad_script_scope_garden/dyd/shed/scopes/$dryad_script_scope_name
    if [ ! -d "$dryad_script_scope_path" ]; then
        dryad_die "scope not found: $dryad_script_scope_name"
    fi

    printf '%s\n' "$dryad_script_scope_path"
}

dryad_script_file_path () {
    dryad_script_file_name=$1
    dryad_script_file_scope_dir=$(dryad_script_scope_dir)
    printf '%s\n' "$dryad_script_file_scope_dir/script-run-$dryad_script_file_name"
}

dryad_script_parse_one_name () {
    dryad_script_parse_action=$1
    shift
    dryad_script_parse_name=

    while [ "$#" -gt 0 ]; do
        dryad_strip_option_quotes_load "$1"
        dryad_script_parse_arg=$dyd_ret0
        case $dryad_script_parse_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad script $dryad_script_parse_action <script>
EOF
                return 2
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_script_parse_arg#--scope=}
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
                dryad_die "unsupported script $dryad_script_parse_action option: $1"
                ;;
            * )
                [ -z "$dryad_script_parse_name" ] ||
                    dryad_die "script $dryad_script_parse_action accepts one script"
                dryad_script_parse_name=$1
                shift
                ;;
        esac
    done

    [ -n "$dryad_script_parse_name" ] ||
        dryad_die "script $dryad_script_parse_action requires a script"

    return 0
}

dryad_script_path_command () {
    dryad_script_parse_one_name path "$@" || {
        dryad_script_parse_status=$?
        [ "$dryad_script_parse_status" = 2 ] && return 0
        return "$dryad_script_parse_status"
    }

    dryad_script_file_path "$dryad_script_parse_name"
}

dryad_script_get () {
    dryad_script_parse_one_name get "$@" || {
        dryad_script_parse_status=$?
        [ "$dryad_script_parse_status" = 2 ] && return 0
        return "$dryad_script_parse_status"
    }

    dryad_script_get_path=$(dryad_script_file_path "$dryad_script_parse_name")
    [ -f "$dryad_script_get_path" ] ||
        dryad_die "script not found: $dryad_script_parse_name"

    cat "$dryad_script_get_path"
    printf '\n'
}

dryad_script_edit () {
    dryad_script_edit_action=$1
    shift
    dryad_script_edit_name=
    dryad_script_edit_editor=

    while [ "$#" -gt 0 ]; do
        dryad_strip_option_quotes_load "$1"
        dryad_script_edit_arg=$dyd_ret0
        case $dryad_script_edit_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad script $dryad_script_edit_action <script>
EOF
                return 0
                ;;
            --editor=* )
                dryad_script_edit_editor=${dryad_script_edit_arg#--editor=}
                shift
                ;;
            --editor )
                [ "$#" -gt 1 ] || dryad_die "--editor requires a value"
                dryad_script_edit_editor=$2
                shift 2
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_script_edit_arg#--scope=}
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
                dryad_die "unsupported script $dryad_script_edit_action option: $1"
                ;;
            * )
                [ -z "$dryad_script_edit_name" ] ||
                    dryad_die "script $dryad_script_edit_action accepts one script"
                dryad_script_edit_name=$1
                shift
                ;;
        esac
    done

    [ -n "$dryad_script_edit_name" ] ||
        dryad_die "script $dryad_script_edit_action requires a script"

    if [ -z "$dryad_script_edit_editor" ]; then
        dryad_script_edit_editor=${EDITOR:-}
    fi
    if [ -z "$dryad_script_edit_editor" ]; then
        dryad_script_edit_editor=${VISUAL:-}
    fi
    [ -n "$dryad_script_edit_editor" ] ||
        dryad_die "no editor found"

    dryad_script_edit_path=$(dryad_script_file_path "$dryad_script_edit_name")
    : >> "$dryad_script_edit_path"
    chmod 755 "$dryad_script_edit_path"
    "$dryad_script_edit_editor" "$dryad_script_edit_path"
}

dryad_scripts_list () {
    dryad_scripts_show_path=0
    dryad_scripts_oneline=1

    while [ "$#" -gt 0 ]; do
        dryad_strip_option_quotes_load "$1"
        dryad_scripts_arg=$dyd_ret0
        case $dryad_scripts_arg in
            list )
                shift
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_scripts_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --path=true | --path=1 | --path )
                dryad_scripts_show_path=1
                shift
                ;;
            --path=false | --path=0 )
                dryad_scripts_show_path=0
                shift
                ;;
            --oneline=false | --oneline=0 )
                dryad_scripts_oneline=0
                shift
                ;;
            --oneline=true | --oneline=1 | --oneline )
                dryad_scripts_oneline=1
                shift
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
            * )
                dryad_die "unsupported scripts list argument: $1"
                ;;
        esac
    done

    dryad_garden_find_load
    dryad_scripts_garden=$dyd_ret0
    dryad_scope_resolve_load
    dryad_scripts_scope=$dyd_ret0

    if [ -z "$dryad_scripts_scope" ] || [ "$dryad_scripts_scope" = none ]; then
        dryad_die "no scope set, can't find command"
    fi

    dryad_scripts_scope_dir=$dryad_scripts_garden/dyd/shed/scopes/$dryad_scripts_scope
    if [ ! -d "$dryad_scripts_scope_dir" ]; then
        dryad_die "scope not found: $dryad_scripts_scope"
    fi

    find "$dryad_scripts_scope_dir" -type f -name 'script-run-*' | while IFS= read -r dryad_scripts_script; do
        case $dryad_scripts_script in
            *.oneline ) continue ;;
        esac

        if [ "$dryad_scripts_show_path" = 1 ]; then
            printf '%s\n' "$dryad_scripts_script"
            continue
        fi

        dryad_scripts_name=$(basename "$dryad_scripts_script")
        dryad_scripts_name=${dryad_scripts_name#script-run-}
        dryad_scripts_display="dryad run $dryad_scripts_name"

        if [ "$dryad_scripts_oneline" = 1 ] &&
            [ -f "$dryad_scripts_script.oneline" ]; then
            dryad_scripts_desc=$(cat "$dryad_scripts_script.oneline")
            if [ -n "$dryad_scripts_desc" ]; then
                dryad_scripts_display="$dryad_scripts_display - $dryad_scripts_desc"
            fi
        fi

        printf '%s\n' "$dryad_scripts_display"
    done | sort
}
