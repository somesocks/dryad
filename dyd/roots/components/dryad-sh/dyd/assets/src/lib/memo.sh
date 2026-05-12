dryad_memo_init () {
    if [ -n "${DRYAD_SH_MEMO_DIR:-}" ]; then
        mkdir -p "$DRYAD_SH_MEMO_DIR"
        export DRYAD_SH_MEMO_DIR
        return 0
    fi

    DRYAD_SH_MEMO_DIR=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-memo.XXXXXX")
    export DRYAD_SH_MEMO_DIR
}

dryad_memo_cleanup () {
    if [ -n "${DRYAD_SH_MEMO_DIR:-}" ]; then
        rm -rf "$DRYAD_SH_MEMO_DIR"
        unset DRYAD_SH_MEMO_DIR
    fi
}

dryad_memo_escape_arg_load () {
    dryad_memo_escape_in=$1

    if [ -z "$dryad_memo_escape_in" ]; then
        dyd_ret0='~e'
        return 0
    fi

    dryad_memo_escape_out=
    while [ -n "$dryad_memo_escape_in" ]; do
        dryad_memo_escape_ch=${dryad_memo_escape_in%"${dryad_memo_escape_in#?}"}
        dryad_memo_escape_in=${dryad_memo_escape_in#?}
        case $dryad_memo_escape_ch in
            '~' )
                dryad_memo_escape_out=${dryad_memo_escape_out}~~
                ;;
            '/' )
                dryad_memo_escape_out=${dryad_memo_escape_out}~s
                ;;
            '^' )
                dryad_memo_escape_out=${dryad_memo_escape_out}~c
                ;;
            * )
                dryad_memo_escape_out=$dryad_memo_escape_out$dryad_memo_escape_ch
                ;;
        esac
    done

    dyd_ret0=$dryad_memo_escape_out
}

dryad_memo_path_load () {
    [ -n "${DRYAD_SH_MEMO_DIR:-}" ] || dryad_die "memo dir is not initialized"
    dryad_memo_path_group=$1
    shift

    dryad_memo_path_dir=$DRYAD_SH_MEMO_DIR/$dryad_memo_path_group
    dryad_memo_path_key=
    dryad_memo_path_sep=
    if [ "$#" -eq 0 ]; then
        dryad_memo_path_key=~k
    else
        for dryad_memo_path_arg do
            dryad_memo_escape_arg_load "$dryad_memo_path_arg"
            dryad_memo_path_escaped=$dyd_ret0
            dryad_memo_path_key=$dryad_memo_path_key$dryad_memo_path_sep$dryad_memo_path_escaped
            dryad_memo_path_sep='^'
        done
    fi
    dyd_ret0=$dryad_memo_path_dir/$dryad_memo_path_key
}

dryad_memo_get_load () {
    dryad_memo_init
    dryad_memo_path_load "$@"
    dryad_memo_get_path=$dyd_ret0
    dyd_ret0=
    [ -f "$dryad_memo_get_path" ] || return 1

    dryad_memo_get_value=
    dryad_memo_get_sep=
    while IFS= read -r dryad_memo_get_line || [ -n "$dryad_memo_get_line" ]; do
        dryad_memo_get_value=$dryad_memo_get_value$dryad_memo_get_sep$dryad_memo_get_line
        dryad_memo_get_sep='
'
    done < "$dryad_memo_get_path"
    dyd_ret0=$dryad_memo_get_value
}

dryad_memo_get_line_load () {
    dryad_memo_init
    dryad_memo_path_load "$@"
    dryad_memo_get_line_path=$dyd_ret0
    [ -f "$dryad_memo_get_line_path" ] || return 1

    dryad_memo_get_line_value=
    IFS= read -r dryad_memo_get_line_value < "$dryad_memo_get_line_path" || [ -n "$dryad_memo_get_line_value" ] || dryad_memo_get_line_value=
    dyd_ret0=$dryad_memo_get_line_value
}

dryad_memo_put () {
    dryad_memo_init
    dryad_memo_path_load "$@"
    dryad_memo_put_path=$dyd_ret0
    dryad_memo_put_tmp=$dryad_memo_put_path.tmp.$$

    mkdir -p "${dryad_memo_put_path%/*}"
    rm -f "$dryad_memo_put_tmp"
    cat > "$dryad_memo_put_tmp"

    if [ -f "$dryad_memo_put_path" ]; then
        rm -f "$dryad_memo_put_tmp"
        return 0
    fi

    if ! mv "$dryad_memo_put_tmp" "$dryad_memo_put_path" 2>/dev/null; then
        rm -f "$dryad_memo_put_tmp"
        [ -f "$dryad_memo_put_path" ] || return 1
    fi
}

dryad_memo_put_value () {
    dryad_memo_put_value_group=$1
    shift
    dryad_memo_put_value_value=$1
    shift

    dryad_memo_init
    dryad_memo_path_load "$dryad_memo_put_value_group" "$@"
    dryad_memo_put_value_path=$dyd_ret0
    dryad_memo_put_value_tmp=$dryad_memo_put_value_path.tmp.$$

    mkdir -p "${dryad_memo_put_value_path%/*}"
    rm -f "$dryad_memo_put_value_tmp"
    printf '%s' "$dryad_memo_put_value_value" > "$dryad_memo_put_value_tmp"

    if [ -f "$dryad_memo_put_value_path" ]; then
        rm -f "$dryad_memo_put_value_tmp"
        return 0
    fi

    if ! mv "$dryad_memo_put_value_tmp" "$dryad_memo_put_value_path" 2>/dev/null; then
        rm -f "$dryad_memo_put_value_tmp"
        [ -f "$dryad_memo_put_value_path" ] || return 1
    fi
}
