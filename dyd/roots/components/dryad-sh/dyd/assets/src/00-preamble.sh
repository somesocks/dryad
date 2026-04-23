set -eu

DRYAD_SH_VERSION=${DRYAD_SH_VERSION:-0.0.0}
dryad_log_level=${DYD_LOG_LEVEL:-info}

dryad_die () {
    printf 'dryad-sh: error: %s\n' "$*" >&2
    exit 1
}

dryad_debug () {
    case $dryad_log_level in
        debug | trace )
            printf 'dryad-sh: %s\n' "$*" >&2
            ;;
    esac
}

dryad_profile_enabled () {
    case ${DRYAD_SH_PROFILE:-} in
        '' | 0 | false | no )
            return 1
            ;;
        * )
            return 0
            ;;
    esac
}

dryad_profile_init () {
    if [ -n "${DRYAD_SH_PROFILE_FILE:-}" ]; then
        export DRYAD_SH_PROFILE_FILE
        return 0
    fi

    dryad_profile_enabled || return 1

    case $DRYAD_SH_PROFILE in
        1 )
            DRYAD_SH_PROFILE_FILE=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-profile.XXXXXX")
            ;;
        * )
            DRYAD_SH_PROFILE_FILE=$DRYAD_SH_PROFILE
            mkdir -p "$(dirname "$DRYAD_SH_PROFILE_FILE")"
            : > "$DRYAD_SH_PROFILE_FILE"
            ;;
    esac

    export DRYAD_SH_PROFILE_FILE
}

dryad_profile_count () {
    dryad_profile_name=$1
    dryad_profile_delta=${2:-1}

    if [ -z "${DRYAD_SH_PROFILE_FILE:-}" ]; then
        dryad_profile_init || return 0
    fi

    printf 'count\t%s\t%s\n' "$dryad_profile_name" "$dryad_profile_delta" >> "$DRYAD_SH_PROFILE_FILE"
}

dryad_profile_report () {
    [ -n "${DRYAD_SH_PROFILE_FILE:-}" ] || return 0
    [ -f "$DRYAD_SH_PROFILE_FILE" ] || return 0

    printf 'dryad-sh: profile file=%s\n' "$DRYAD_SH_PROFILE_FILE" >&2
    awk -F '\t' '
        $1 == "count" {
            counts[$2] += $3
        }

        END {
            for (key in counts) {
                printf "dryad-sh: profile count %s=%s\n", key, counts[key]
            }
        }
    ' "$DRYAD_SH_PROFILE_FILE" >&2
}

dryad_usage () {
    cat <<'EOF'
dryad-sh - bootstrap Dryad implementation

Usage:
  dryad [--scope=<name>|--scope <name>] <resource> <action> [args...]
  dryad --help
  dryad --version

Supported commands:
  garden path
  root ancestors [path]
  root build [path]
  root create <path>
  root descendants [path]
  root path
  root requirement add <path> [alias]
  root requirement remove <name>
  root requirements list [path]
  root secrets path [path]
  root secrets list [path]
  root variants list [path]
  roots affected
  roots build
  roots each
  roots graph
  roots owning
  roots path
  roots list
  sprouts path
  sprouts list
  sprouts prune
  sprouts run
  sprouts wipe
  sprout run <sprout_ref>
  scopes path
  scopes list
  scope active
  run <script> -- [args...]
EOF
}

dryad_path_is_abs () {
    case $1 in
        /* ) return 0 ;;
        * ) return 1 ;;
    esac
}

dryad_join_path () {
    case $2 in
        '' ) printf '%s\n' "$1" ;;
        /* ) printf '%s\n' "$2" ;;
        * ) printf '%s/%s\n' "$1" "$2" ;;
    esac
}

dryad_clean_cd () {
    cd "$1" 2>/dev/null || dryad_die "could not enter directory: $1"
    pwd -P
}

dryad_strip_option_quotes () {
    case $1 in
        *=\'*\' )
            dryad_strip_name=${1%%=*}
            dryad_strip_value=${1#*=}
            dryad_strip_value=${dryad_strip_value#\'}
            dryad_strip_value=${dryad_strip_value%\'}
            printf '%s=%s\n' "$dryad_strip_name" "$dryad_strip_value"
            ;;
        *=* )
            dryad_strip_name=${1%%=*}
            dryad_strip_value=${1#*=}
            dryad_strip_value=${dryad_strip_value#\"}
            dryad_strip_value=${dryad_strip_value%\"}
            printf '%s=%s\n' "$dryad_strip_name" "$dryad_strip_value"
            ;;
        * )
            printf '%s\n' "$1"
            ;;
    esac
}

dryad_host_os () {
    dryad_host_os_name=$(uname -s 2>/dev/null || printf unknown)
    case $dryad_host_os_name in
        Darwin ) printf 'darwin\n' ;;
        Linux ) printf 'linux\n' ;;
        * ) printf '%s\n' "$dryad_host_os_name" | tr '[:upper:]' '[:lower:]' ;;
    esac
}

dryad_host_arch () {
    dryad_host_arch_name=$(uname -m 2>/dev/null || printf unknown)
    case $dryad_host_arch_name in
        x86_64 | amd64 ) printf 'amd64\n' ;;
        aarch64 | arm64 ) printf 'arm64\n' ;;
        * ) printf '%s\n' "$dryad_host_arch_name" ;;
    esac
}
