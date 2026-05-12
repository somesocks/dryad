dryad_bool_value_load () {
    case $1 in
        true | 1 )
            dyd_ret0=1
            ;;
        false | 0 )
            dyd_ret0=0
            ;;
        * )
            dryad_die "expected boolean value, got: $1"
            ;;
    esac
}
