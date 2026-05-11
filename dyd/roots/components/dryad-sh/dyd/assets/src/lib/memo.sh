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

dryad_memo_escape_arg () {
    dryad_memo_escape_in=$1

    if [ -z "$dryad_memo_escape_in" ]; then
        printf '~e\n'
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

    printf '%s\n' "$dryad_memo_escape_out"
}

dryad_memo_path () {
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
            dryad_memo_path_escaped=$(dryad_memo_escape_arg "$dryad_memo_path_arg")
            dryad_memo_path_key=$dryad_memo_path_key$dryad_memo_path_sep$dryad_memo_path_escaped
            dryad_memo_path_sep='^'
        done
    fi
    printf '%s/%s\n' "$dryad_memo_path_dir" "$dryad_memo_path_key"
}

dryad_memo_get () {
    dryad_memo_init
    dryad_memo_get_path=$(dryad_memo_path "$@")
    [ -f "$dryad_memo_get_path" ] || return 1
    cat "$dryad_memo_get_path"
}

dryad_memo_get_line_into () {
    dryad_memo_get_line_var=$1
    shift

    dryad_memo_init
    dryad_memo_get_line_path=$(dryad_memo_path "$@")
    [ -f "$dryad_memo_get_line_path" ] || return 1

    dryad_memo_get_line_value=
    IFS= read -r dryad_memo_get_line_value < "$dryad_memo_get_line_path" || [ -n "$dryad_memo_get_line_value" ] || dryad_memo_get_line_value=
    eval "$dryad_memo_get_line_var=\$dryad_memo_get_line_value"
}

dryad_memo_put () {
    dryad_memo_init
    dryad_memo_put_path=$(dryad_memo_path "$@")
    dryad_memo_put_tmp=$dryad_memo_put_path.tmp.$$

    mkdir -p "$(dirname "$dryad_memo_put_path")"
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
    dryad_memo_put_value_path=$(dryad_memo_path "$dryad_memo_put_value_group" "$@")
    dryad_memo_put_value_tmp=$dryad_memo_put_value_path.tmp.$$

    mkdir -p "$(dirname "$dryad_memo_put_value_path")"
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
