dryad_garden_find () {
    dryad_garden_dir=$(pwd -P)

    while :; do
        dryad_garden_type=$dryad_garden_dir/dyd/type
        if [ -f "$dryad_garden_type" ]; then
            dryad_garden_sentinel=$(cat "$dryad_garden_type")
            if [ "$dryad_garden_sentinel" = garden ]; then
                dryad_garden_size=$(wc -c < "$dryad_garden_type" | tr -d ' ')
                if [ "$dryad_garden_size" != 6 ]; then
                    printf '%s\n' 'dryad-sh: malformed sentinel file path=dyd/type expected="garden"' >&2
                fi
                printf '%s\n' "$dryad_garden_dir"
                return 0
            fi
        fi

        dryad_garden_parent=$(dirname "$dryad_garden_dir")
        if [ "$dryad_garden_parent" = "$dryad_garden_dir" ]; then
            dryad_die "not inside a dryad garden"
        fi
        dryad_garden_dir=$dryad_garden_parent
    done
}

dryad_garden_create () {
    dryad_garden_create_target=${1:-.}
    dryad_join_path_load "$(pwd -P)" "$dryad_garden_create_target"
    dryad_garden_create_abs=$dyd_ret0
    dryad_garden_create_parent=$(dirname "$dryad_garden_create_abs")
    dryad_garden_create_name=$(basename "$dryad_garden_create_abs")
    dryad_garden_create_parent=$(dryad_clean_cd "$dryad_garden_create_parent")
    dryad_garden_create_base=$dryad_garden_create_parent/$dryad_garden_create_name

    mkdir -p "$dryad_garden_create_base/dyd/heap/files/v2"
    mkdir -p "$dryad_garden_create_base/dyd/heap/secrets/v2"
    mkdir -p "$dryad_garden_create_base/dyd/heap/stems/v2"
    mkdir -p "$dryad_garden_create_base/dyd/heap/sprouts/v2"
    mkdir -p "$dryad_garden_create_base/dyd/heap/derivations/roots/v2"
    mkdir -p "$dryad_garden_create_base/dyd/heap/contexts"
    mkdir -p "$dryad_garden_create_base/dyd/roots"
    mkdir -p "$dryad_garden_create_base/dyd/sprouts"
    mkdir -p "$dryad_garden_create_base/dyd/shed/scopes"
    mkdir -p "$dryad_garden_create_base/dyd/shed/heap/files"
    mkdir -p "$dryad_garden_create_base/dyd/shed/heap/secrets"
    mkdir -p "$dryad_garden_create_base/dyd/shed/heap/stems"
    mkdir -p "$dryad_garden_create_base/dyd/shed/heap/sprouts"
    mkdir -p "$dryad_garden_create_base/dyd/shed/heap/derivations/roots"

    printf '%s' garden > "$dryad_garden_create_base/dyd/type"
    printf '%s' 1 > "$dryad_garden_create_base/dyd/shed/heap/files/depth"
    printf '%s' 1 > "$dryad_garden_create_base/dyd/shed/heap/secrets/depth"
    printf '%s' 1 > "$dryad_garden_create_base/dyd/shed/heap/stems/depth"
    printf '%s' 1 > "$dryad_garden_create_base/dyd/shed/heap/sprouts/depth"
    printf '%s' 1 > "$dryad_garden_create_base/dyd/shed/heap/derivations/roots/depth"
}

dryad_garden_prune () {
    dryad_garden_prune_garden=$(dryad_garden_find)
    dryad_garden_prune_derivations=$dryad_garden_prune_garden/dyd/heap/derivations/roots/v2

    [ -d "$dryad_garden_prune_derivations" ] || return 0

    find "$dryad_garden_prune_derivations" -type f |
        while IFS= read -r dryad_garden_prune_derivation; do
            case $(basename "$dryad_garden_prune_derivation") in
                .tmp-* )
                    rm -f "$dryad_garden_prune_derivation"
                    continue
                    ;;
            esac

            dryad_garden_prune_result=$(cat "$dryad_garden_prune_derivation")
            if [ -z "$dryad_garden_prune_result" ]; then
                rm -f "$dryad_garden_prune_derivation"
                continue
            fi

            dryad_garden_prune_stem=$(dryad_root_build_heap_package_path "$dryad_garden_prune_garden" stems "$dryad_garden_prune_result" 2>/dev/null || true)
            if [ -z "$dryad_garden_prune_stem" ] || [ ! -d "$dryad_garden_prune_stem" ]; then
                rm -f "$dryad_garden_prune_derivation"
            fi
        done
}

dryad_garden_wipe_path_contents () {
    dryad_garden_wipe_contents_path=$1

    [ -d "$dryad_garden_wipe_contents_path" ] || return 0
    chmod -R u+w "$dryad_garden_wipe_contents_path" 2>/dev/null || true
    find "$dryad_garden_wipe_contents_path" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
}

dryad_garden_wipe () {
    dryad_garden_wipe_garden=$(dryad_garden_find)
    dryad_garden_wipe_heap=$dryad_garden_wipe_garden/dyd/heap

    dryad_sprouts_make_tree_removable "$dryad_garden_wipe_garden/dyd/sprouts"
    rm -rf "$dryad_garden_wipe_garden/dyd/sprouts"

    if [ -L "$dryad_garden_wipe_heap" ]; then
        dryad_garden_wipe_heap_target=$(dryad_clean_cd "$dryad_garden_wipe_heap")
        dryad_garden_wipe_path_contents "$dryad_garden_wipe_heap_target"
    else
        dryad_garden_wipe_path_contents "$dryad_garden_wipe_heap"
    fi
}

dryad_cmd_garden () {
    dryad_garden_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi
    case $dryad_garden_action in
        create )
            dryad_garden_create_target=
            while [ "$#" -gt 0 ]; do
                dryad_strip_option_quotes_load "$1"
                dryad_garden_create_arg=$dyd_ret0
                case $dryad_garden_create_arg in
                    --help | -h )
                        cat <<'EOF'
Usage:
  dryad garden create <path>
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
                        dryad_die "unsupported garden create option: $1"
                        ;;
                    * )
                        [ -z "$dryad_garden_create_target" ] ||
                            dryad_die "garden create accepts one path"
                        dryad_garden_create_target=$1
                        shift
                        ;;
                esac
            done
            dryad_garden_create "${dryad_garden_create_target:-.}"
            ;;
        path )
            dryad_garden_next=${1:-}
            case $dryad_garden_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad garden path
EOF
                    ;;
                * )
                    dryad_garden_find
                    ;;
            esac
            ;;
        pack )
            dryad_garden_next=${1:-}
            case $dryad_garden_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad garden pack
EOF
                    ;;
                * )
                    dryad_die "garden pack is not supported by dryad-sh yet"
                    ;;
            esac
            ;;
        prune )
            dryad_garden_next=${1:-}
            case $dryad_garden_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad garden prune
EOF
                    ;;
                * )
                    while [ "$#" -gt 0 ]; do
                        dryad_strip_option_quotes_load "$1"
                        dryad_garden_prune_arg=$dyd_ret0
                        case $dryad_garden_prune_arg in
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
                                dryad_die "unsupported garden prune option: $1"
                                ;;
                            * )
                                dryad_die "garden prune accepts no arguments"
                                ;;
                        esac
                    done
                    dryad_garden_prune
                    ;;
            esac
            ;;
        wipe )
            dryad_garden_next=${1:-}
            case $dryad_garden_next in
                --help | -h )
                    cat <<'EOF'
Usage:
  dryad garden wipe
EOF
                    ;;
                * )
                    while [ "$#" -gt 0 ]; do
                        dryad_strip_option_quotes_load "$1"
                        dryad_garden_wipe_arg=$dyd_ret0
                        case $dryad_garden_wipe_arg in
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
                                dryad_die "unsupported garden wipe option: $1"
                                ;;
                            * )
                                dryad_die "garden wipe accepts no arguments"
                                ;;
                        esac
                    done
                    dryad_garden_wipe
                    ;;
            esac
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad garden create <path>
  dryad garden path
  dryad garden pack
  dryad garden prune
  dryad garden wipe
EOF
            ;;
        * )
            dryad_die "unsupported garden action: $dryad_garden_action"
            ;;
    esac
}
