dryad_main () {
    dryad_scope_rewrite_done=${dryad_scope_rewrite_done:-0}

    while [ "$#" -gt 0 ]; do
        dryad_global_arg=$(dryad_strip_option_quotes "$1")
        case $1 in
            --scope=* )
                dryad_scope_arg=${dryad_global_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* )
                dryad_log_level=${dryad_global_arg#--log-level=}
                shift
                ;;
            --log-level )
                [ "$#" -gt 1 ] || dryad_die "--log-level requires a value"
                dryad_log_level=$2
                shift 2
                ;;
            --log-format=* | --parallel=* )
                shift
                ;;
            --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            --help | -h )
                dryad_usage
                exit 0
                ;;
            --version | version )
                printf 'dryad-sh %s\n' "$DRYAD_SH_VERSION"
                exit 0
                ;;
            --* )
                dryad_die "unsupported global option: $1"
                ;;
            * )
                break
                ;;
        esac
    done

    dryad_resource=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    dryad_find_command_scope "$@"

    if [ "$dryad_scope_rewrite_done" = 0 ]; then
        dryad_scope_rewrite "$dryad_resource" "$@"
    fi

    case $dryad_resource in
        '' )
            dryad_usage
            ;;
        garden )
            dryad_cmd_garden "$@"
            ;;
        roots )
            dryad_cmd_roots "$@"
            ;;
        sprouts )
            dryad_cmd_sprouts "$@"
            ;;
        root )
            dryad_cmd_root "$@"
            ;;
        scopes )
            dryad_cmd_scopes "$@"
            ;;
        scope )
            dryad_cmd_scope "$@"
            ;;
        run )
            dryad_cmd_run "$@"
            ;;
        scripts )
            dryad_cmd_scripts "$@"
            ;;
        script )
            dryad_cmd_script "$@"
            ;;
        * )
            dryad_die "unsupported command resource: $dryad_resource"
            ;;
    esac
}

dryad_find_command_scope () {
    while [ "$#" -gt 0 ]; do
        dryad_find_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_find_arg in
            -- )
                return 0
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_find_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* )
                dryad_log_level=${dryad_find_arg#--log-level=}
                shift
                ;;
            --log-level )
                [ "$#" -gt 1 ] || dryad_die "--log-level requires a value"
                dryad_log_level=$2
                shift 2
                ;;
            * )
                shift
                ;;
        esac
    done
}

dryad_scope_rewrite () {
    dryad_rewrite_resource=$1
    shift

    case $dryad_rewrite_resource in
        '' | garden | run )
            return 0
            ;;
    esac

    for dryad_rewrite_arg in "$@"; do
        case $dryad_rewrite_arg in
            --help | -h )
                return 0
                ;;
        esac
    done

    dryad_rewrite_action=${1:-}
    [ -n "$dryad_rewrite_action" ] || return 0

    dryad_rewrite_scope=$(dryad_scope_resolve)
    if [ -z "$dryad_rewrite_scope" ] || [ "$dryad_rewrite_scope" = none ]; then
        return 0
    fi

    case $dryad_rewrite_resource in
        root )
            case ${1:-} in
                requirement | requirements | secrets | variants )
                    dryad_rewrite_setting=$dryad_rewrite_resource-${1:-}-${2:-}
                    ;;
                * )
                    dryad_rewrite_setting=$dryad_rewrite_resource-$dryad_rewrite_action
                    ;;
            esac
            ;;
        * )
            dryad_rewrite_setting=$dryad_rewrite_resource-$dryad_rewrite_action
            ;;
    esac
    dryad_rewrite_args=$(dryad_scope_setting_get "$dryad_rewrite_scope" "$dryad_rewrite_setting")
    [ -n "$dryad_rewrite_args" ] || return 0

    dryad_scope_rewrite_done=1
    set -f
    dryad_debug "rewriting args to: dryad --scope=none $dryad_rewrite_resource $* $dryad_rewrite_args"
    dryad_scope_rewrite_run "$dryad_rewrite_resource" "$@" $dryad_rewrite_args
    dryad_rewrite_status=$?
    set +f
    exit "$dryad_rewrite_status"
}

dryad_scope_rewrite_run () {
    dryad_rewrite_run_resource=$1
    shift

    set +f
    dryad_main --scope=none "$dryad_rewrite_run_resource" "$@"
}

dryad_main "$@"
