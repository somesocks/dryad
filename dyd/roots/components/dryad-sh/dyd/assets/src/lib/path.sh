dryad_relative_path () {
    dryad_relative_path_load "$1" "$2"
    printf '%s\n' "$dyd_ret0"
}

dryad_relative_path_load () {
    dryad_clean_cd_load "$1"
    dryad_relative_from=$dyd_ret0
    dryad_clean_cd_load "$2"
    dryad_relative_to=$dyd_ret0
    dryad_relative_prefix=$dryad_relative_from
    dryad_relative_ups=

    while :; do
        case $dryad_relative_to in
            "$dryad_relative_prefix" )
                dyd_ret0=${dryad_relative_ups:-.}
                return 0
                ;;
            "$dryad_relative_prefix"/* )
                dryad_relative_tail=${dryad_relative_to#"$dryad_relative_prefix"/}
                if [ -n "$dryad_relative_ups" ]; then
                    dyd_ret0=$dryad_relative_ups/$dryad_relative_tail
                else
                    dyd_ret0=$dryad_relative_tail
                fi
                return 0
                ;;
        esac

        if [ "$dryad_relative_prefix" = / ]; then
            dyd_ret0=$dryad_relative_to
            return 0
        fi

        dryad_relative_prefix=$(dirname "$dryad_relative_prefix")
        if [ -n "$dryad_relative_ups" ]; then
            dryad_relative_ups=$dryad_relative_ups/..
        else
            dryad_relative_ups=..
        fi
    done
}

dryad_relative_path_literal_load () {
    dryad_relative_from=${1%/}
    dryad_relative_to=${2%/}
    dryad_relative_prefix=$dryad_relative_from
    dryad_relative_ups=

    while :; do
        case $dryad_relative_to in
            "$dryad_relative_prefix" )
                dyd_ret0=${dryad_relative_ups:-.}
                return 0
                ;;
            "$dryad_relative_prefix"/* )
                dryad_relative_tail=${dryad_relative_to#"$dryad_relative_prefix"/}
                if [ -n "$dryad_relative_ups" ]; then
                    dyd_ret0=$dryad_relative_ups/$dryad_relative_tail
                else
                    dyd_ret0=$dryad_relative_tail
                fi
                return 0
                ;;
        esac

        if [ "$dryad_relative_prefix" = / ]; then
            dyd_ret0=$dryad_relative_to
            return 0
        fi

        dryad_relative_prefix=$(dirname "$dryad_relative_prefix")
        if [ -n "$dryad_relative_ups" ]; then
            dryad_relative_ups=$dryad_relative_ups/..
        else
            dryad_relative_ups=..
        fi
    done
}

dryad_file_abs_path_load () {
    dryad_file_abs_input=$1
    dryad_file_abs_dir=$(dirname "$dryad_file_abs_input")
    dryad_file_abs_name=$(basename "$dryad_file_abs_input")
    dryad_clean_cd_load "$dryad_file_abs_dir"
    dryad_file_abs_dir=$dyd_ret0
    dyd_ret0=$dryad_file_abs_dir/$dryad_file_abs_name
}
