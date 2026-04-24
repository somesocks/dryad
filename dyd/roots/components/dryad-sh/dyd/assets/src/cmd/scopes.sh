dryad_scope_arg=

dryad_scopes_path () {
    dryad_scopes_garden=$(dryad_garden_find)
    printf '%s\n' "$dryad_scopes_garden/dyd/shed/scopes"
}

dryad_scope_active () {
    dryad_scope_garden=$(dryad_garden_find)
    dryad_scope_default=$dryad_scope_garden/dyd/shed/scopes/default

    if [ ! -d "$dryad_scope_default" ]; then
        return 1
    fi

    dryad_scope_real=$(cd "$dryad_scope_default" 2>/dev/null && pwd -P)
    basename "$dryad_scope_real"
}

dryad_scope_resolve () {
    if [ -n "$dryad_scope_arg" ]; then
        printf '%s\n' "$dryad_scope_arg"
        return 0
    fi

    dryad_scope_active || true
}

dryad_scope_setting_path () {
    dryad_scope_setting_scope=$1
    dryad_scope_setting_name=$2
    dryad_scope_setting_garden=$(dryad_garden_find)
    printf '%s\n' "$dryad_scope_setting_garden/dyd/shed/scopes/$dryad_scope_setting_scope/$dryad_scope_setting_name"
}

dryad_scope_setting_get () {
    dryad_scope_setting_scope=$1
    dryad_scope_setting_name=$2
    dryad_scope_setting_scope_dir=$(dryad_scopes_path)/$dryad_scope_setting_scope

    if [ ! -d "$dryad_scope_setting_scope_dir" ]; then
        dryad_die "scope not found: $dryad_scope_setting_scope"
    fi

    dryad_scope_setting_file=$dryad_scope_setting_scope_dir/$dryad_scope_setting_name
    if [ ! -f "$dryad_scope_setting_file" ]; then
        return 0
    fi

    cat "$dryad_scope_setting_file"
}

dryad_scope_setting_set () {
    dryad_scope_setting_scope=$1
    dryad_scope_setting_name=$2
    dryad_scope_setting_value=$3
    dryad_scope_setting_scope_dir=$(dryad_scopes_path)/$dryad_scope_setting_scope

    if [ ! -d "$dryad_scope_setting_scope_dir" ]; then
        dryad_die "scope not found: $dryad_scope_setting_scope"
    fi

    printf '%s' "$dryad_scope_setting_value" > "$dryad_scope_setting_scope_dir/$dryad_scope_setting_name"
}

dryad_scope_setting_unset () {
    dryad_scope_setting_scope=$1
    dryad_scope_setting_name=$2
    dryad_scope_setting_scope_dir=$(dryad_scopes_path)/$dryad_scope_setting_scope

    if [ ! -d "$dryad_scope_setting_scope_dir" ]; then
        dryad_die "scope not found: $dryad_scope_setting_scope"
    fi

    dryad_scope_setting_file=$dryad_scope_setting_scope_dir/$dryad_scope_setting_name
    if [ ! -e "$dryad_scope_setting_file" ]; then
        dryad_die "setting not found: $dryad_scope_setting_name"
    fi

    rm "$dryad_scope_setting_file"
}

dryad_cmd_scope_setting () {
    dryad_scope_setting_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_scope_setting_action in
        get | unset )
            dryad_scope_setting_scope=
            dryad_scope_setting_name=
            while [ "$#" -gt 0 ]; do
                dryad_scope_setting_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scope_setting_arg in
                    --help | -h )
                        cat <<EOF
Usage:
  dryad scope setting $dryad_scope_setting_action <scope> <setting>
EOF
                        return 0
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
                        dryad_die "unsupported scope setting $dryad_scope_setting_action option: $1"
                        ;;
                    * )
                        if [ -z "$dryad_scope_setting_scope" ]; then
                            dryad_scope_setting_scope=$1
                        elif [ -z "$dryad_scope_setting_name" ]; then
                            dryad_scope_setting_name=$1
                        else
                            dryad_die "scope setting $dryad_scope_setting_action accepts two arguments"
                        fi
                        shift
                        ;;
                esac
            done

            [ -n "$dryad_scope_setting_scope" ] ||
                dryad_die "scope setting $dryad_scope_setting_action requires a scope"
            [ -n "$dryad_scope_setting_name" ] ||
                dryad_die "scope setting $dryad_scope_setting_action requires a setting"

            case $dryad_scope_setting_action in
                get )
                    dryad_scope_setting_get "$dryad_scope_setting_scope" "$dryad_scope_setting_name"
                    ;;
                unset )
                    dryad_scope_setting_unset "$dryad_scope_setting_scope" "$dryad_scope_setting_name"
                    ;;
            esac
            ;;
        set )
            dryad_scope_setting_scope=
            dryad_scope_setting_name=
            dryad_scope_setting_value=
            dryad_scope_setting_has_value=0
            while [ "$#" -gt 0 ]; do
                dryad_scope_setting_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scope_setting_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad scope setting set <scope> <setting> <value>
EOF
                        return 0
                        ;;
                    --log-level=* | --log-format=* | --parallel=* )
                        if [ "$dryad_scope_setting_has_value" = 0 ]; then
                            shift
                        else
                            dryad_die "scope setting set accepts three arguments"
                        fi
                        ;;
                    --log-level | --log-format | --parallel )
                        if [ "$dryad_scope_setting_has_value" = 0 ]; then
                            [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                            shift 2
                        else
                            dryad_die "scope setting set accepts three arguments"
                        fi
                        ;;
                    -- )
                        shift
                        if [ "$#" -gt 0 ]; then
                            [ "$dryad_scope_setting_has_value" = 0 ] ||
                                dryad_die "scope setting set accepts three arguments"
                            dryad_scope_setting_value=$1
                            dryad_scope_setting_has_value=1
                            shift
                        fi
                        break
                        ;;
                    * )
                        if [ -z "$dryad_scope_setting_scope" ]; then
                            dryad_scope_setting_scope=$1
                        elif [ -z "$dryad_scope_setting_name" ]; then
                            dryad_scope_setting_name=$1
                        elif [ "$dryad_scope_setting_has_value" = 0 ]; then
                            dryad_scope_setting_value=$1
                            dryad_scope_setting_has_value=1
                        else
                            dryad_die "scope setting set accepts three arguments"
                        fi
                        shift
                        ;;
                esac
            done

            [ -n "$dryad_scope_setting_scope" ] ||
                dryad_die "scope setting set requires a scope"
            [ -n "$dryad_scope_setting_name" ] ||
                dryad_die "scope setting set requires a setting"
            [ "$dryad_scope_setting_has_value" = 1 ] ||
                dryad_die "scope setting set requires a value"

            dryad_scope_setting_set "$dryad_scope_setting_scope" "$dryad_scope_setting_name" "$dryad_scope_setting_value"
            ;;
        --help | -h | '' )
            cat <<'EOF'
Usage:
  dryad scope setting get <scope> <setting>
  dryad scope setting set <scope> <setting> <value>
  dryad scope setting unset <scope> <setting>
EOF
            ;;
        * )
            dryad_die "unsupported scope setting action: $dryad_scope_setting_action"
            ;;
    esac
}

dryad_scopes_default_get () {
    dryad_scopes_default_alias=$(dryad_scopes_path)/default
    if [ ! -e "$dryad_scopes_default_alias" ]; then
        return 0
    fi

    dryad_scopes_default_target=$(readlink "$dryad_scopes_default_alias") ||
        dryad_die "default scope is not a symlink"
    basename "$dryad_scopes_default_target"
}

dryad_scopes_default_set () {
    dryad_scopes_default_scope=$1
    dryad_scopes_default_dir=$(dryad_scopes_path)
    dryad_scopes_default_scope_dir=$dryad_scopes_default_dir/$dryad_scopes_default_scope

    if [ ! -e "$dryad_scopes_default_scope_dir" ]; then
        dryad_die "scope not found: $dryad_scopes_default_scope"
    fi

    dryad_scopes_default_alias=$dryad_scopes_default_dir/default
    if [ -e "$dryad_scopes_default_alias" ] || [ -L "$dryad_scopes_default_alias" ]; then
        rm "$dryad_scopes_default_alias"
    fi
    ln -s "$dryad_scopes_default_scope" "$dryad_scopes_default_alias"
}

dryad_scopes_default_unset () {
    dryad_scopes_default_alias=$(dryad_scopes_path)/default
    if [ -e "$dryad_scopes_default_alias" ] || [ -L "$dryad_scopes_default_alias" ]; then
        rm "$dryad_scopes_default_alias"
    fi
}

dryad_scope_create () {
    dryad_scope_create_name=$1
    dryad_scope_create_path=$(dryad_scopes_path)/$dryad_scope_create_name
    mkdir -p "$dryad_scope_create_path"
    printf '%s\n' "$dryad_scope_create_path"
}

dryad_scope_delete () {
    dryad_scope_delete_name=$1
    dryad_scope_delete_path=$(dryad_scopes_path)/$dryad_scope_delete_name
    rm -rf "$dryad_scope_delete_path"
}

dryad_scope_one_name_command () {
    dryad_scope_one_action=$1
    shift
    dryad_scope_one_name=

    while [ "$#" -gt 0 ]; do
        dryad_scope_one_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_scope_one_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad scope $dryad_scope_one_action <scope>
EOF
                return 0
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
                dryad_die "unsupported scope $dryad_scope_one_action option: $1"
                ;;
            * )
                [ -z "$dryad_scope_one_name" ] ||
                    dryad_die "scope $dryad_scope_one_action accepts one scope"
                dryad_scope_one_name=$1
                shift
                ;;
        esac
    done

    [ -n "$dryad_scope_one_name" ] ||
        dryad_die "scope $dryad_scope_one_action requires a scope"

    case $dryad_scope_one_action in
        create )
            dryad_scope_create "$dryad_scope_one_name"
            ;;
        delete )
            dryad_scope_delete "$dryad_scope_one_name"
            ;;
    esac
}

dryad_cmd_scopes_default () {
    dryad_scopes_default_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_scopes_default_action in
        get )
            while [ "$#" -gt 0 ]; do
                dryad_scopes_default_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scopes_default_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad scopes default get
EOF
                        return 0
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
                        dryad_die "unsupported scopes default get argument: $1"
                        ;;
                esac
            done
            dryad_scopes_default_get
            ;;
        set )
            dryad_scopes_default_scope=
            while [ "$#" -gt 0 ]; do
                dryad_scopes_default_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scopes_default_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad scopes default set <scope>
EOF
                        return 0
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
                        dryad_die "unsupported scopes default set option: $1"
                        ;;
                    * )
                        [ -z "$dryad_scopes_default_scope" ] ||
                            dryad_die "scopes default set accepts one scope"
                        dryad_scopes_default_scope=$1
                        shift
                        ;;
                esac
            done
            [ -n "$dryad_scopes_default_scope" ] ||
                dryad_die "scopes default set requires a scope"
            dryad_scopes_default_set "$dryad_scopes_default_scope"
            ;;
        unset )
            while [ "$#" -gt 0 ]; do
                dryad_scopes_default_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scopes_default_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad scopes default unset
EOF
                        return 0
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
                        dryad_die "unsupported scopes default unset argument: $1"
                        ;;
                esac
            done
            dryad_scopes_default_unset
            ;;
        --help | -h | '' )
            cat <<'EOF'
Usage:
  dryad scopes default get
  dryad scopes default set <scope>
  dryad scopes default unset
EOF
            ;;
        * )
            dryad_die "unsupported scopes default action: $dryad_scopes_default_action"
            ;;
    esac
}

dryad_cmd_scopes () {
    dryad_scopes_action=${1:-}
    dryad_scopes_next=${2:-}
    dryad_scopes_next_next=${3:-}
    case $dryad_scopes_action in
        default )
            shift
            dryad_cmd_scopes_default "$@"
            ;;
        path )
            case $dryad_scopes_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad scopes path
  dryad scopes list
EOF
                    ;;
                * )
                    dryad_scopes_path
                    ;;
            esac
            ;;
        list )
            case $dryad_scopes_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad scopes path
  dryad scopes list
EOF
                    ;;
                * )
                    dryad_scopes_oneline=1
                    while [ "$#" -gt 0 ]; do
                        dryad_scopes_arg=$(dryad_strip_option_quotes "$1")
                        case $dryad_scopes_arg in
                            list )
                                shift
                                ;;
                            --oneline=false | --oneline=0 )
                                dryad_scopes_oneline=0
                                shift
                                ;;
                            --oneline=true | --oneline=1 | --oneline )
                                dryad_scopes_oneline=1
                                shift
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
                                dryad_die "unsupported scopes list argument: $1"
                                ;;
                        esac
                    done

                    dryad_scopes_dir=$(dryad_scopes_path)
                    if [ ! -d "$dryad_scopes_dir" ]; then
                        return 0
                    fi
                    for dryad_scopes_scope in "$dryad_scopes_dir"/*; do
                        [ -d "$dryad_scopes_scope" ] || continue
                        dryad_scopes_name=$(basename "$dryad_scopes_scope")
                        [ "$dryad_scopes_name" = default ] && continue
                        if [ "$dryad_scopes_oneline" = 1 ] &&
                            [ -f "$dryad_scopes_scope/.oneline" ]; then
                            dryad_scopes_desc=$(cat "$dryad_scopes_scope/.oneline")
                            if [ -n "$dryad_scopes_desc" ]; then
                                printf '%s - %s\n' "$dryad_scopes_name" "$dryad_scopes_desc"
                                continue
                            fi
                        fi
                        printf '%s\n' "$dryad_scopes_name"
                    done | sort
                    ;;
            esac
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad scopes default get
  dryad scopes default set <scope>
  dryad scopes default unset
  dryad scopes path
  dryad scopes list
EOF
            ;;
        * )
            dryad_die "unsupported scopes action: $dryad_scopes_action"
            ;;
    esac
}

dryad_cmd_scope () {
    dryad_scope_action=${1:-}
    dryad_scope_next=${2:-}
    dryad_scope_next_next=${3:-}
    case $dryad_scope_action in
        active )
            dryad_scope_active_oneline=1
            while [ "$#" -gt 1 ]; do
                shift
                dryad_scope_active_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scope_active_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad scope active [--oneline=false]
EOF
                        return 0
                        ;;
                    --oneline=false | --oneline=0 )
                        dryad_scope_active_oneline=0
                        ;;
                    --oneline=true | --oneline=1 | --oneline )
                        dryad_scope_active_oneline=1
                        ;;
                    --log-level=* | --log-format=* | --parallel=* )
                        ;;
                    --log-level | --log-format | --parallel )
                        [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                        shift
                        ;;
                    -- )
                        break
                        ;;
                    * )
                        dryad_die "unsupported scope active argument: $1"
                        ;;
                esac
            done

            dryad_scope_active_name=$(dryad_scope_active || true)
            [ -n "$dryad_scope_active_name" ] || return 0

            if [ "$dryad_scope_active_oneline" = 1 ]; then
                dryad_scope_active_oneline_file=$(dryad_scopes_path)/$dryad_scope_active_name/.oneline
                if [ -f "$dryad_scope_active_oneline_file" ]; then
                    dryad_scope_active_description=$(cat "$dryad_scope_active_oneline_file")
                    if [ -n "$dryad_scope_active_description" ]; then
                        printf '%s - %s\n' "$dryad_scope_active_name" "$dryad_scope_active_description"
                        return 0
                    fi
                fi
            fi

            printf '%s\n' "$dryad_scope_active_name"
            ;;
        create | delete )
            shift
            dryad_scope_one_name_command "$dryad_scope_action" "$@"
            ;;
        use )
            shift
            dryad_scope_use_scope=
            while [ "$#" -gt 0 ]; do
                dryad_scope_use_arg=$(dryad_strip_option_quotes "$1")
                case $dryad_scope_use_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad scope use <scope>
EOF
                        return 0
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
                        dryad_die "unsupported scope use option: $1"
                        ;;
                    * )
                        [ -z "$dryad_scope_use_scope" ] ||
                            dryad_die "scope use accepts one scope"
                        dryad_scope_use_scope=$1
                        shift
                        ;;
                esac
            done

            [ -n "$dryad_scope_use_scope" ] ||
                dryad_die "scope use requires a scope"

            if [ "$dryad_scope_use_scope" = none ]; then
                dryad_scopes_default_unset
            else
                dryad_scopes_default_set "$dryad_scope_use_scope"
            fi
            ;;
        setting )
            shift
            dryad_cmd_scope_setting "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad scope active [--oneline=false]
  dryad scope create <scope>
  dryad scope delete <scope>
  dryad scope setting get
  dryad scope setting set
  dryad scope setting unset
  dryad scope use <scope>
EOF
            ;;
        * )
            dryad_die "unsupported scope action: $dryad_scope_action"
            ;;
    esac
}
