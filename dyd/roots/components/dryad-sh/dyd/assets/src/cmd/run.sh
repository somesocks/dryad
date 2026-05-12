dryad_cmd_run () {
    dryad_run_command=${1:-}
    case $dryad_run_command in
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad run <script> [--scope=<name>] -- [args...]
EOF
            return 0
            ;;
    esac
    [ -n "$dryad_run_command" ] || dryad_die "run requires a script name"
    shift

    while [ "$#" -gt 0 ]; do
        dryad_strip_option_quotes_load "$1"
        dryad_run_arg=$dyd_ret0
        case $dryad_run_arg in
            -- )
                shift
                break
                ;;
            --scope=* )
                dryad_scope_arg=${dryad_run_arg#--scope=}
                shift
                ;;
            --scope )
                [ "$#" -gt 1 ] || dryad_die "--scope requires a value"
                dryad_scope_arg=$2
                shift 2
                ;;
            --log-level=* )
                dryad_log_level=${dryad_run_arg#--log-level=}
                shift
                ;;
            --log-level )
                [ "$#" -gt 1 ] || dryad_die "--log-level requires a value"
                dryad_log_level=$2
                shift 2
                ;;
            --log-format=* | --parallel=* | --inherit )
                shift
                ;;
            --log-format | --parallel )
                [ "$#" -gt 1 ] || dryad_die "$1 requires a value"
                shift 2
                ;;
            --* )
                dryad_die "unsupported run option: $1"
                ;;
            * )
                break
                ;;
        esac
    done

    dryad_garden_find_load
    dryad_run_garden=$dyd_ret0
    dryad_scope_resolve_load
    dryad_run_scope=$dyd_ret0

    if [ -z "$dryad_run_scope" ] || [ "$dryad_run_scope" = none ]; then
        dryad_die "no scope set, can't find command"
    fi

    dryad_run_scope_dir=$dryad_run_garden/dyd/shed/scopes/$dryad_run_scope
    if [ ! -d "$dryad_run_scope_dir" ]; then
        dryad_die "scope not found: $dryad_run_scope"
    fi

    dryad_run_script=$dryad_run_scope_dir/script-run-$dryad_run_command
    if [ ! -f "$dryad_run_script" ]; then
        dryad_die "script not found: $dryad_run_command"
    fi

    dryad_host_os_load
    dryad_run_host_os=$dyd_ret0
    dryad_host_arch_load
    dryad_run_host_arch=$dyd_ret0

    DYD_SCOPE=$dryad_run_scope \
    DYD_GARDEN=$dryad_run_garden \
    DYD_OS=$dryad_run_host_os \
    DYD_ARCH=$dryad_run_host_arch \
    DYD_LOG_LEVEL=$dryad_log_level \
    "$dryad_run_script" "$@"
}
