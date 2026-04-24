dryad_bool_value () {
    case $1 in
        true | 1 )
            printf '1\n'
            ;;
        false | 0 )
            printf '0\n'
            ;;
        * )
            dryad_die "expected boolean value, got: $1"
            ;;
    esac
}
