dryad_sprouts_path () {
    dryad_sprouts_garden=$(dryad_garden_find)
    printf '%s\n' "$dryad_sprouts_garden/dyd/sprouts"
}

dryad_sprouts_candidate_is_sprout () {
    dryad_sprouts_candidate=$1

    if [ -f "$dryad_sprouts_candidate/dyd/type" ]; then
        dryad_sprouts_candidate_type=$(tr -d '[:space:]' < "$dryad_sprouts_candidate/dyd/type")
        case $dryad_sprouts_candidate_type in
            sprout | stem )
                return 0
                ;;
        esac
    fi

    [ -f "$dryad_sprouts_candidate/dyd/fingerprint" ]
}

dryad_sprouts_find_sprouts () {
    dryad_sprouts_find_dir=$1

    [ -d "$dryad_sprouts_find_dir" ] || return 0

    find "$dryad_sprouts_find_dir" \( -type d -o -type l \) -print |
        while IFS= read -r dryad_sprouts_find_candidate; do
            [ "$dryad_sprouts_find_candidate" != "$dryad_sprouts_find_dir" ] || continue
            if dryad_sprouts_candidate_is_sprout "$dryad_sprouts_find_candidate"; then
                printf '%s\n' "$dryad_sprouts_find_candidate"
            fi
        done |
        sort
}

dryad_sprouts_ref_descriptor_load () {
    dryad_sprouts_ref=$1
    dyd_ret0=

    case $dryad_sprouts_ref in
        *~* )
            dryad_fs_descriptor_normalize_load "${dryad_sprouts_ref#*~}"
            ;;
    esac
}

dryad_sprouts_print_ref () {
    dryad_sprouts_print_path=$1
    dryad_sprouts_print_garden=$2

    if [ "$dryad_sprouts_relative" = 1 ]; then
        case $dryad_sprouts_print_path in
            "$dryad_sprouts_print_garden"/* )
                printf '%s\n' "${dryad_sprouts_print_path#"$dryad_sprouts_print_garden"/}"
                ;;
            * )
                printf '%s\n' "$dryad_sprouts_print_path"
                ;;
        esac
    else
        printf '%s\n' "$dryad_sprouts_print_path"
    fi
}

dryad_sprouts_entry_matches_filters () {
    dryad_sprouts_match_display=$1
    dryad_sprouts_match_path=$2
    dryad_sprouts_match_descriptor=$3

    dryad_roots_entry_matches_filters "$dryad_sprouts_match_display" "$dryad_sprouts_match_path" "$dryad_sprouts_match_descriptor"
}

dryad_sprouts_list_from_stdin () {
    dryad_sprouts_list_garden=$1

    while IFS= read -r dryad_sprouts_stdin_ref; do
        [ -n "$dryad_sprouts_stdin_ref" ] || continue
        dryad_sprouts_stdin_path=${dryad_sprouts_stdin_ref%%\?*}
        dryad_sprouts_stdin_query=
        case $dryad_sprouts_stdin_ref in
            *\?* )
                dryad_sprouts_stdin_query=${dryad_sprouts_stdin_ref#*\?}
                ;;
        esac
        dryad_sprouts_stdin_descriptor=$(printf '%s\n' "$dryad_sprouts_stdin_query" | tr '&' '+')
        if [ -n "$dryad_sprouts_stdin_descriptor" ]; then
            dryad_sprouts_stdin_display=$dryad_sprouts_stdin_path~$dryad_sprouts_stdin_descriptor
        else
            dryad_sprouts_stdin_display=$dryad_sprouts_stdin_path
        fi
        dryad_sprouts_ref_descriptor_load "$dryad_sprouts_stdin_display"
        dryad_sprouts_stdin_descriptor=$dyd_ret0
        if dryad_sprouts_entry_matches_filters "$dryad_sprouts_stdin_display" "$dryad_sprouts_stdin_path" "$dryad_sprouts_stdin_descriptor"; then
            case $dryad_sprouts_stdin_display in
                /* )
                    dryad_sprouts_print_ref "$dryad_sprouts_stdin_display" "$dryad_sprouts_list_garden"
                    ;;
                * )
                    printf '%s\n' "$dryad_sprouts_stdin_display"
                    ;;
            esac
        fi
    done
}

dryad_sprouts_run_sanitize_segment_load () {
    dyd_ret0=$(printf '%s\n' "$1" | sed 's#[<>:"/\\|?*]#-#g')
}

dryad_sprouts_run_log () {
    case $dryad_log_level in
        disabled | none | off )
            ;;
        * )
            printf 'dryad-sh: %s\n' "$*" >&2
            ;;
    esac
}

dryad_sprouts_run_ref_parse () {
    dryad_sprouts_run_ref_raw=$1
    dryad_sprouts_run_ref_path=$dryad_sprouts_run_ref_raw
    dryad_sprouts_run_ref_selector=

    case $dryad_sprouts_run_ref_raw in
        *\?* )
            dryad_sprouts_run_ref_path=${dryad_sprouts_run_ref_raw%%\?*}
            dryad_sprouts_run_ref_query=${dryad_sprouts_run_ref_raw#*\?}
            [ -n "$dryad_sprouts_run_ref_path" ] ||
                dryad_die "missing sprouts run sprout_ref path"
            dryad_url_query_to_descriptor_load "$dryad_sprouts_run_ref_query"
            dryad_sprouts_run_ref_selector=$dyd_ret0
            ;;
        *~* )
            dryad_sprouts_run_ref_path=${dryad_sprouts_run_ref_raw%%~*}
            dryad_sprouts_run_ref_selector=${dryad_sprouts_run_ref_raw#*~}
            [ -n "$dryad_sprouts_run_ref_path" ] ||
                dryad_die "missing sprouts run sprout_ref path"
            dryad_fs_descriptor_normalize_load "$dryad_sprouts_run_ref_selector"
            dryad_sprouts_run_ref_selector=$dyd_ret0
            ;;
    esac
}

dryad_sprouts_run_resolve_path_load () {
    dryad_sprouts_run_resolve_garden=$1
    dryad_sprouts_run_resolve_ref=$2

    case $dryad_sprouts_run_resolve_ref in
        /* )
            dryad_sprouts_run_resolve_path=$dryad_sprouts_run_resolve_ref
            ;;
        * )
            dryad_join_path_load "$dryad_sprouts_run_resolve_garden" "$dryad_sprouts_run_resolve_ref"
            dryad_sprouts_run_resolve_path=$dyd_ret0
            ;;
    esac

    if ! dryad_sprouts_candidate_is_sprout "$dryad_sprouts_run_resolve_path"; then
        dryad_die "sprout not found: $dryad_sprouts_run_resolve_ref"
    fi

    dryad_sprouts_run_resolve_dir=$(dirname "$dryad_sprouts_run_resolve_path")
    dryad_sprouts_run_resolve_base=$(basename "$dryad_sprouts_run_resolve_path")
    dryad_sprouts_run_resolve_dir=$(dryad_clean_cd "$dryad_sprouts_run_resolve_dir")
    dyd_ret0=$dryad_sprouts_run_resolve_dir/$dryad_sprouts_run_resolve_base
}

dryad_sprouts_run_stem_dependency_descriptor () {
    dryad_sprouts_run_stem_dependency_name=$1

    case $dryad_sprouts_run_stem_dependency_name in
        stem )
            printf '\n'
            ;;
        stem~* )
            dryad_sprouts_run_stem_dependency_raw=${dryad_sprouts_run_stem_dependency_name#stem~}
            dryad_fs_descriptor_normalize_load "$dryad_sprouts_run_stem_dependency_raw"
            dryad_sprouts_run_stem_dependency_normalized=$dyd_ret0
            if [ "$dryad_sprouts_run_stem_dependency_normalized" != "$dryad_sprouts_run_stem_dependency_raw" ]; then
                dryad_die "non-canonical sprout stem dependency descriptor: $dryad_sprouts_run_stem_dependency_name"
            fi
            printf '%s\n' "$dryad_sprouts_run_stem_dependency_normalized"
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_sprouts_run_stem_path () {
    dryad_sprouts_run_stem_input=$1

    if [ ! -d "$dryad_sprouts_run_stem_input" ]; then
        printf '%s\n' "dryad-sh: error: dyd stem path not found: $dryad_sprouts_run_stem_input" >&2
        return 1
    fi

    dryad_sprouts_run_stem_working=$(dryad_clean_cd "$dryad_sprouts_run_stem_input")
    while :; do
        if [ -d "$dryad_sprouts_run_stem_working/dyd" ]; then
            if [ -f "$dryad_sprouts_run_stem_working/dyd/type" ]; then
                dryad_sprouts_run_stem_type=$(tr -d '[:space:]' < "$dryad_sprouts_run_stem_working/dyd/type")
                if [ "$dryad_sprouts_run_stem_type" = sprout ] &&
                    [ -e "$dryad_sprouts_run_stem_working/dyd/dependencies/stem" ]; then
                    dryad_clean_cd "$dryad_sprouts_run_stem_working/dyd/dependencies/stem"
                    return 0
                fi
            fi

            printf '%s\n' "$dryad_sprouts_run_stem_working"
            return 0
        fi

        dryad_sprouts_run_stem_parent=$(dirname "$dryad_sprouts_run_stem_working")
        if [ "$dryad_sprouts_run_stem_parent" = "$dryad_sprouts_run_stem_working" ]; then
            printf '%s\n' "dryad-sh: error: dyd stem path not found: $dryad_sprouts_run_stem_input" >&2
            return 1
        fi
        dryad_sprouts_run_stem_working=$dryad_sprouts_run_stem_parent
    done
}

dryad_sprouts_run_available_stems () {
    dryad_sprouts_run_available_sprout=$1
    dryad_sprouts_run_available_deps=$dryad_sprouts_run_available_sprout/dyd/dependencies

    if [ ! -d "$dryad_sprouts_run_available_deps" ]; then
        printf '%s\n' "dryad-sh: error: sprout has no stem dependencies: $dryad_sprouts_run_available_sprout" >&2
        return 1
    fi

    find "$dryad_sprouts_run_available_deps" -mindepth 1 -maxdepth 1 \( -type d -o -type l \) -print |
        while IFS= read -r dryad_sprouts_run_available_dep; do
            dryad_sprouts_run_available_name=$(basename "$dryad_sprouts_run_available_dep")
            if ! dryad_sprouts_run_available_descriptor=$(dryad_sprouts_run_stem_dependency_descriptor "$dryad_sprouts_run_available_name"); then
                continue
            fi
            dryad_sprouts_run_available_stem=$(dryad_sprouts_run_stem_path "$dryad_sprouts_run_available_dep") ||
                exit 1
            printf '%s|%s\n' "$dryad_sprouts_run_available_descriptor" "$dryad_sprouts_run_available_stem"
        done |
        sort
}

dryad_sprouts_run_selector_matches_any_stem () {
    dryad_sprouts_run_selector_match_selector=$1
    dryad_sprouts_run_selector_match_stems=$2

    while IFS='|' read -r dryad_sprouts_run_selector_match_descriptor dryad_sprouts_run_selector_match_path; do
        [ -n "$dryad_sprouts_run_selector_match_path" ] || continue
        if dryad_root_variant_selector_matches_descriptor "$dryad_sprouts_run_selector_match_selector" "$dryad_sprouts_run_selector_match_descriptor"; then
            return 0
        fi
    done <<EOF
$dryad_sprouts_run_selector_match_stems
EOF

    return 1
}

dryad_sprouts_run_stems_have_dimension () {
    dryad_sprouts_run_dimension_stems=$1
    dryad_sprouts_run_dimension_name=$2

    while IFS='|' read -r dryad_sprouts_run_dimension_descriptor dryad_sprouts_run_dimension_path; do
        [ -n "$dryad_sprouts_run_dimension_path" ] || continue
        if dryad_descriptor_value_load "$dryad_sprouts_run_dimension_descriptor" "$dryad_sprouts_run_dimension_name"; then
            return 0
        fi
    done <<EOF
$dryad_sprouts_run_dimension_stems
EOF

    return 1
}

dryad_sprouts_run_selector_normalize () {
    dryad_sprouts_run_normalize_selector=$1

    [ -n "$dryad_sprouts_run_normalize_selector" ] || return 0

    dryad_sprouts_run_normalize_output=
    dryad_sprouts_run_normalize_old_ifs=$IFS
    IFS=+
    set -- $dryad_sprouts_run_normalize_selector
    IFS=$dryad_sprouts_run_normalize_old_ifs

    for dryad_sprouts_run_normalize_pair do
        dryad_sprouts_run_normalize_dim=${dryad_sprouts_run_normalize_pair%%=*}
        dryad_sprouts_run_normalize_options=${dryad_sprouts_run_normalize_pair#*=}
        dryad_sprouts_run_normalize_options_output=

        dryad_sprouts_run_normalize_options_old_ifs=$IFS
        IFS=,
        set -- $dryad_sprouts_run_normalize_options
        IFS=$dryad_sprouts_run_normalize_options_old_ifs

        for dryad_sprouts_run_normalize_option do
            case $dryad_sprouts_run_normalize_option in
                inherit )
                    dryad_die "inherit option is not supported for sprout run variant selectors: $dryad_sprouts_run_normalize_dim"
                    ;;
                host )
                    case $dryad_sprouts_run_normalize_dim in
                        os ) dryad_host_os_load; dryad_sprouts_run_normalize_option=$dyd_ret0 ;;
                        arch ) dryad_host_arch_load; dryad_sprouts_run_normalize_option=$dyd_ret0 ;;
                        * ) dryad_die "host option is only supported for variant dimensions os/arch: $dryad_sprouts_run_normalize_dim" ;;
                    esac
                    ;;
            esac

            case ,$dryad_sprouts_run_normalize_options_output, in
                *,"$dryad_sprouts_run_normalize_option",* )
                    ;;
                * )
                    if [ -n "$dryad_sprouts_run_normalize_options_output" ]; then
                        dryad_sprouts_run_normalize_options_output=$dryad_sprouts_run_normalize_options_output,$dryad_sprouts_run_normalize_option
                    else
                        dryad_sprouts_run_normalize_options_output=$dryad_sprouts_run_normalize_option
                    fi
                    ;;
            esac
        done

        if [ -n "$dryad_sprouts_run_normalize_output" ]; then
            dryad_sprouts_run_normalize_output=$dryad_sprouts_run_normalize_output+$dryad_sprouts_run_normalize_dim=$dryad_sprouts_run_normalize_options_output
        else
            dryad_sprouts_run_normalize_output=$dryad_sprouts_run_normalize_dim=$dryad_sprouts_run_normalize_options_output
        fi
    done

    dryad_fs_descriptor_normalize_load "$dryad_sprouts_run_normalize_output"
    [ -z "$dyd_ret0" ] || printf '%s\n' "$dyd_ret0"
}

dryad_sprouts_run_selector_option_exists () {
    dryad_sprouts_run_option_stems=$1
    dryad_sprouts_run_option_dim=$2
    dryad_sprouts_run_option_want=$3

    while IFS='|' read -r dryad_sprouts_run_option_descriptor dryad_sprouts_run_option_path; do
        [ -n "$dryad_sprouts_run_option_path" ] || continue
        if dryad_descriptor_value_load "$dryad_sprouts_run_option_descriptor" "$dryad_sprouts_run_option_dim"; then
            dryad_sprouts_run_option_got=$dyd_ret0
        else
            dryad_sprouts_run_option_got=
        fi

        case $dryad_sprouts_run_option_want in
            any )
                [ -n "$dryad_sprouts_run_option_got" ] && return 0
                ;;
            none )
                [ -z "$dryad_sprouts_run_option_got" ] && return 0
                ;;
            * )
                [ "$dryad_sprouts_run_option_got" = "$dryad_sprouts_run_option_want" ] && return 0
                ;;
        esac
    done <<EOF
$dryad_sprouts_run_option_stems
EOF

    return 1
}

dryad_sprouts_run_selector_validate () {
    dryad_sprouts_run_validate_stems=$1
    dryad_sprouts_run_validate_selector=$2

    [ -n "$dryad_sprouts_run_validate_selector" ] || return 0

    dryad_sprouts_run_validate_old_ifs=$IFS
    IFS=+
    set -- $dryad_sprouts_run_validate_selector
    IFS=$dryad_sprouts_run_validate_old_ifs

    for dryad_sprouts_run_validate_pair do
        dryad_sprouts_run_validate_dim=${dryad_sprouts_run_validate_pair%%=*}
        dryad_sprouts_run_validate_options=${dryad_sprouts_run_validate_pair#*=}

        if ! dryad_sprouts_run_stems_have_dimension "$dryad_sprouts_run_validate_stems" "$dryad_sprouts_run_validate_dim"; then
            printf '%s\n' "dryad-sh: error: over-specified sprout run variant dimension: $dryad_sprouts_run_validate_dim" >&2
            return 1
        fi

        dryad_sprouts_run_validate_options_old_ifs=$IFS
        IFS=,
        set -- $dryad_sprouts_run_validate_options
        IFS=$dryad_sprouts_run_validate_options_old_ifs

        for dryad_sprouts_run_validate_option do
            if ! dryad_sprouts_run_selector_option_exists "$dryad_sprouts_run_validate_stems" "$dryad_sprouts_run_validate_dim" "$dryad_sprouts_run_validate_option"; then
                printf '%s\n' "dryad-sh: error: wrongly-specified sprout run variant option: $dryad_sprouts_run_validate_dim=$dryad_sprouts_run_validate_option" >&2
                return 1
            fi
        done
    done

    return 0
}

dryad_sprouts_run_options_match_value () {
    dryad_sprouts_run_options_match_options=$1
    dryad_sprouts_run_options_match_value=${2:-}
    dryad_sprouts_run_options_match_exists=$3

    dryad_sprouts_run_options_match_old_ifs=$IFS
    IFS=,
    set -- $dryad_sprouts_run_options_match_options
    IFS=$dryad_sprouts_run_options_match_old_ifs

    for dryad_sprouts_run_options_match_option do
        case $dryad_sprouts_run_options_match_option in
            any )
                [ "$dryad_sprouts_run_options_match_exists" = 1 ] && return 0
                ;;
            none )
                [ "$dryad_sprouts_run_options_match_exists" = 0 ] && return 0
                ;;
            * )
                [ "$dryad_sprouts_run_options_match_exists" = 1 ] &&
                    [ "$dryad_sprouts_run_options_match_value" = "$dryad_sprouts_run_options_match_option" ] &&
                    return 0
                ;;
        esac
    done

    return 1
}

dryad_sprouts_run_stems_match_dimension_options () {
    dryad_sprouts_run_dimension_option_stems=$1
    dryad_sprouts_run_dimension_option_dim=$2
    dryad_sprouts_run_dimension_option_options=$3

    while IFS='|' read -r dryad_sprouts_run_dimension_option_descriptor dryad_sprouts_run_dimension_option_path; do
        [ -n "$dryad_sprouts_run_dimension_option_path" ] || continue
        if dryad_descriptor_value_load "$dryad_sprouts_run_dimension_option_descriptor" "$dryad_sprouts_run_dimension_option_dim"; then
            dryad_sprouts_run_dimension_option_value=$dyd_ret0
        else
            dryad_sprouts_run_dimension_option_value=
        fi
        dryad_sprouts_run_dimension_option_exists=0
        [ -n "$dryad_sprouts_run_dimension_option_value" ] &&
            dryad_sprouts_run_dimension_option_exists=1

        if dryad_sprouts_run_options_match_value "$dryad_sprouts_run_dimension_option_options" "$dryad_sprouts_run_dimension_option_value" "$dryad_sprouts_run_dimension_option_exists"; then
            return 0
        fi
    done <<EOF
$dryad_sprouts_run_dimension_option_stems
EOF

    return 1
}

dryad_sprouts_run_trait_matches_options () {
    dryad_sprouts_run_trait_sprout=$1
    dryad_sprouts_run_trait_name=$2
    dryad_sprouts_run_trait_options=$3

    dryad_sprouts_run_trait_path=$dryad_sprouts_run_trait_sprout/dyd/traits/$dryad_sprouts_run_trait_name
    dryad_sprouts_run_trait_exists=0
    dryad_sprouts_run_trait_value=

    if [ -f "$dryad_sprouts_run_trait_path" ]; then
        dryad_sprouts_run_trait_exists=1
        dryad_sprouts_run_trait_value=$(sed 's/^[[:space:]]*//;s/[[:space:]]*$//' "$dryad_sprouts_run_trait_path")
    fi

    dryad_sprouts_run_options_match_value "$dryad_sprouts_run_trait_options" "$dryad_sprouts_run_trait_value" "$dryad_sprouts_run_trait_exists"
}

dryad_sprouts_run_sprout_matches_selector () {
    dryad_sprouts_run_selector_sprout=$1
    dryad_sprouts_run_selector_stems=$2
    dryad_sprouts_run_selector=$3

    [ -n "$dryad_sprouts_run_selector" ] || return 0

    dryad_sprouts_run_selector_old_ifs=$IFS
    IFS=+
    set -- $dryad_sprouts_run_selector
    IFS=$dryad_sprouts_run_selector_old_ifs

    for dryad_sprouts_run_selector_pair do
        dryad_sprouts_run_selector_dim=${dryad_sprouts_run_selector_pair%%=*}
        dryad_sprouts_run_selector_options=${dryad_sprouts_run_selector_pair#*=}

        if dryad_sprouts_run_stems_have_dimension "$dryad_sprouts_run_selector_stems" "$dryad_sprouts_run_selector_dim"; then
            dryad_sprouts_run_stems_match_dimension_options "$dryad_sprouts_run_selector_stems" "$dryad_sprouts_run_selector_dim" "$dryad_sprouts_run_selector_options" || return 1
        else
            dryad_sprouts_run_trait_matches_options "$dryad_sprouts_run_selector_sprout" "$dryad_sprouts_run_selector_dim" "$dryad_sprouts_run_selector_options" || return 1
        fi
    done

    return 0
}

dryad_sprouts_run_filter_matches_one () {
    dryad_sprouts_run_filter_display=$1
    dryad_sprouts_run_filter_sprout=$2
    dryad_sprouts_run_filter_stems=$3
    dryad_sprouts_run_filter=$4

    dryad_sprouts_run_filter_path=$dryad_sprouts_run_filter
    dryad_sprouts_run_filter_selector=
    dryad_sprouts_run_filter_has_selector=0

    case $dryad_sprouts_run_filter in
        "~"* )
            dryad_sprouts_run_filter_path="**"
            dryad_sprouts_run_filter_selector=${dryad_sprouts_run_filter#\~}
            dryad_sprouts_run_filter_has_selector=1
            ;;
        *"~"* )
            dryad_sprouts_run_filter_path=${dryad_sprouts_run_filter%~*}
            dryad_sprouts_run_filter_selector=${dryad_sprouts_run_filter##*~}
            dryad_sprouts_run_filter_has_selector=1
            ;;
    esac

    case $dryad_sprouts_run_filter_display in
        $dryad_sprouts_run_filter_path )
            ;;
        * )
            return 1
            ;;
    esac

    [ "$dryad_sprouts_run_filter_has_selector" = 1 ] ||
        return 0

    dryad_fs_descriptor_normalize_load "$dryad_sprouts_run_filter_selector"
    dryad_sprouts_run_filter_selector=$dyd_ret0
    dryad_sprouts_run_sprout_matches_selector "$dryad_sprouts_run_filter_sprout" "$dryad_sprouts_run_filter_stems" "$dryad_sprouts_run_filter_selector"
}

dryad_sprouts_run_matches_includes () {
    dryad_sprouts_run_filter_display=$1
    dryad_sprouts_run_filter_sprout=$2
    dryad_sprouts_run_filter_stems=$3

    if [ -z "${dryad_roots_include:-}" ]; then
        return 0
    fi

    while IFS= read -r dryad_sprouts_run_filter_include; do
        [ -n "$dryad_sprouts_run_filter_include" ] || continue
        if dryad_sprouts_run_filter_matches_one "$dryad_sprouts_run_filter_display" "$dryad_sprouts_run_filter_sprout" "$dryad_sprouts_run_filter_stems" "$dryad_sprouts_run_filter_include"; then
            return 0
        fi
    done <<EOF
$dryad_roots_include
EOF

    return 1
}

dryad_sprouts_run_matches_excludes () {
    dryad_sprouts_run_filter_display=$1
    dryad_sprouts_run_filter_sprout=$2
    dryad_sprouts_run_filter_stems=$3

    if [ -z "${dryad_roots_exclude:-}" ]; then
        return 1
    fi

    while IFS= read -r dryad_sprouts_run_filter_exclude; do
        [ -n "$dryad_sprouts_run_filter_exclude" ] || continue
        if dryad_sprouts_run_filter_matches_one "$dryad_sprouts_run_filter_display" "$dryad_sprouts_run_filter_sprout" "$dryad_sprouts_run_filter_stems" "$dryad_sprouts_run_filter_exclude"; then
            return 0
        fi
    done <<EOF
$dryad_roots_exclude
EOF

    return 1
}

dryad_sprouts_run_matches_filters () {
    dryad_sprouts_run_filter_display=$1
    dryad_sprouts_run_filter_sprout=$2
    dryad_sprouts_run_filter_stems=$3

    dryad_sprouts_run_matches_includes "$dryad_sprouts_run_filter_display" "$dryad_sprouts_run_filter_sprout" "$dryad_sprouts_run_filter_stems" || return 1
    ! dryad_sprouts_run_matches_excludes "$dryad_sprouts_run_filter_display" "$dryad_sprouts_run_filter_sprout" "$dryad_sprouts_run_filter_stems"
}

dryad_sprouts_run_confirm_prompt () {
    printf "are you sure? type '%s' to continue\n" "$dryad_sprouts_run_confirm"

    IFS= read -r dryad_sprouts_run_confirm_input ||
        dryad_die "error while reading input"

    [ "$dryad_sprouts_run_confirm_input" = "$dryad_sprouts_run_confirm" ] ||
        dryad_die "input does not match confirmation, aborting"
}

dryad_sprouts_run_confirm_print_sprout () {
    dryad_sprouts_run_confirm_garden=$1
    dryad_sprouts_run_confirm_sprout=$2
    dryad_sprouts_run_confirm_ref=${3:-}

    dryad_sprouts_run_confirm_stems=$(dryad_sprouts_run_available_stems "$dryad_sprouts_run_confirm_sprout") ||
        return 1
    dryad_sprouts_run_confirm_display=${dryad_sprouts_run_confirm_sprout#"$dryad_sprouts_run_confirm_garden"/}

    if dryad_sprouts_run_matches_filters "$dryad_sprouts_run_confirm_display" "$dryad_sprouts_run_confirm_sprout" "$dryad_sprouts_run_confirm_stems"; then
        if [ -n "$dryad_sprouts_run_confirm_ref" ]; then
            printf ' - %s\n' "$dryad_sprouts_run_confirm_ref"
        else
            printf ' - %s\n' "$dryad_sprouts_run_confirm_display"
        fi
    fi
}

dryad_sprouts_run_confirm_print_all () {
    dryad_sprouts_run_confirm_garden=$1
    dryad_sprouts_run_confirm_dir=$dryad_sprouts_run_confirm_garden/dyd/sprouts

    printf '%s\n' "dryad sprouts exec will execute these sprouts:"
    dryad_sprouts_find_sprouts "$dryad_sprouts_run_confirm_dir" |
        while IFS= read -r dryad_sprouts_run_confirm_sprout; do
            dryad_sprouts_run_confirm_print_sprout "$dryad_sprouts_run_confirm_garden" "$dryad_sprouts_run_confirm_sprout"
        done
}

dryad_sprouts_run_confirm_print_stdin () {
    dryad_sprouts_run_confirm_garden=$1
    dryad_sprouts_run_confirm_file=$2

    printf '%s\n' "dryad sprouts exec will execute these sprouts:"
    while IFS= read -r dryad_sprouts_run_confirm_ref; do
        dryad_sprouts_run_confirm_ref=$(printf '%s\n' "$dryad_sprouts_run_confirm_ref" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        [ -n "$dryad_sprouts_run_confirm_ref" ] || continue
        dryad_sprouts_run_ref_parse "$dryad_sprouts_run_confirm_ref"
        if [ -n "$dryad_sprouts_run_variant" ] && [ -n "$dryad_sprouts_run_ref_selector" ]; then
            dryad_die "sprouts run selector specified in both stdin sprout_ref and --variant"
        fi
        dryad_sprouts_run_resolve_path_load "$dryad_sprouts_run_confirm_garden" "$dryad_sprouts_run_ref_path"
        dryad_sprouts_run_confirm_sprout=$dyd_ret0
        dryad_sprouts_run_confirm_print_sprout "$dryad_sprouts_run_confirm_garden" "$dryad_sprouts_run_confirm_sprout" "$dryad_sprouts_run_confirm_ref"
    done < "$dryad_sprouts_run_confirm_file"
}

dryad_sprouts_run_stem_path_env () {
    dryad_sprouts_run_stem_path_env_stem=$1
    dryad_sprouts_run_stem_path_env_bin_dir=$2

    dryad_sprouts_run_stem_path_env_value=$dryad_sprouts_run_stem_path_env_stem/dyd/commands:$dryad_sprouts_run_stem_path_env_stem/dyd/path:$dryad_sprouts_run_stem_path_env_bin_dir:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    dryad_host_os_load
    dryad_sprouts_run_stem_path_env_host_os=$dyd_ret0
    case $dryad_sprouts_run_stem_path_env_host_os in
        darwin )
            dryad_sprouts_run_stem_path_env_value=$dryad_sprouts_run_stem_path_env_value:/opt/homebrew/bin:/opt/homebrew/sbin
            ;;
    esac

    printf '%s\n' "$dryad_sprouts_run_stem_path_env_value"
}

dryad_sprouts_run_cli_bin () {
    dryad_sprouts_run_cli_bin_value=$(command -v dryad 2>/dev/null || printf '%s\n' "$0")
    case $dryad_sprouts_run_cli_bin_value in
        /* )
            printf '%s\n' "$dryad_sprouts_run_cli_bin_value"
            ;;
        */* )
            dryad_sprouts_run_cli_bin_dir=$(dirname "$dryad_sprouts_run_cli_bin_value")
            dryad_sprouts_run_cli_bin_name=$(basename "$dryad_sprouts_run_cli_bin_value")
            printf '%s/%s\n' "$(dryad_clean_cd "$dryad_sprouts_run_cli_bin_dir")" "$dryad_sprouts_run_cli_bin_name"
            ;;
        * )
            printf '%s\n' "$dryad_sprouts_run_cli_bin_value"
            ;;
    esac
}

dryad_sprouts_run_env_command () {
    dryad_sprouts_run_env_inherit=$1
    dryad_sprouts_run_env_command_path=$2
    dryad_sprouts_run_env_stem=$3
    dryad_sprouts_run_env_context=$4
    dryad_sprouts_run_env_garden=$5
    dryad_sprouts_run_env_path=$6
    dryad_sprouts_run_env_cli_bin=$7
    shift 7
    dryad_host_os_load
    dryad_sprouts_run_env_host_os=$dyd_ret0
    dryad_host_arch_load
    dryad_sprouts_run_env_host_arch=$dyd_ret0

    if [ "$dryad_sprouts_run_env_inherit" = 1 ]; then
        PATH=$dryad_sprouts_run_env_path \
        HOME=$dryad_sprouts_run_env_context \
        DYD_CONTEXT=$dryad_sprouts_run_env_context \
        DYD_STEM=$dryad_sprouts_run_env_stem \
        DYD_GARDEN=$dryad_sprouts_run_env_garden \
        DYD_CLI_BIN=$dryad_sprouts_run_env_cli_bin \
        DYD_OS=$dryad_sprouts_run_env_host_os \
        DYD_ARCH=$dryad_sprouts_run_env_host_arch \
        DYD_LOG_LEVEL=$dryad_log_level \
        "$dryad_sprouts_run_env_command_path" "$@"
    else
        env -i \
            TERM="${TERM:-}" \
            PATH="$dryad_sprouts_run_env_path" \
            HOME="$dryad_sprouts_run_env_context" \
            DYD_CONTEXT="$dryad_sprouts_run_env_context" \
            DYD_STEM="$dryad_sprouts_run_env_stem" \
            DYD_GARDEN="$dryad_sprouts_run_env_garden" \
            DYD_CLI_BIN="$dryad_sprouts_run_env_cli_bin" \
            DYD_OS="$dryad_sprouts_run_env_host_os" \
            DYD_ARCH="$dryad_sprouts_run_env_host_arch" \
            DYD_LOG_LEVEL="$dryad_log_level" \
            "$dryad_sprouts_run_env_command_path" "$@"
    fi
}

dryad_sprouts_run_resolve_command () {
    dryad_sprouts_run_resolve_command_stem=$1
    dryad_sprouts_run_resolve_command_path_env=$2

    if [ -z "${dryad_sprouts_run_command:-}" ]; then
        printf '%s\n' "$dryad_sprouts_run_resolve_command_stem/dyd/commands/dyd-stem-run"
        return 0
    fi

    case $dryad_sprouts_run_command in
        /* | */* )
            printf '%s\n' "$dryad_sprouts_run_command"
            return 0
            ;;
    esac

    dryad_sprouts_run_resolve_command_old_ifs=$IFS
    IFS=:
    set -- $dryad_sprouts_run_resolve_command_path_env
    IFS=$dryad_sprouts_run_resolve_command_old_ifs

    for dryad_sprouts_run_resolve_command_dir do
        [ -n "$dryad_sprouts_run_resolve_command_dir" ] ||
            dryad_sprouts_run_resolve_command_dir=.
        dryad_sprouts_run_resolve_command_candidate=$dryad_sprouts_run_resolve_command_dir/$dryad_sprouts_run_command
        if [ -f "$dryad_sprouts_run_resolve_command_candidate" ]; then
            printf '%s\n' "$dryad_sprouts_run_resolve_command_candidate"
            return 0
        fi
    done

    printf '%s\n' "dryad-sh: error: missing stem main \"$dryad_sprouts_run_command\"" >&2
    return 1
}

dryad_sprouts_run_exec_with_redirs () {
    dryad_sprouts_run_exec_stdout_mode=$1
    dryad_sprouts_run_exec_stdout_file=$2
    dryad_sprouts_run_exec_stderr_mode=$3
    dryad_sprouts_run_exec_stderr_file=$4
    shift 4

    if [ "$dryad_sprouts_run_exec_stdout_mode" = file ]; then
        if [ "$dryad_sprouts_run_exec_stderr_mode" = file ]; then
            dryad_sprouts_run_env_command "$@" >"$dryad_sprouts_run_exec_stdout_file" 2>"$dryad_sprouts_run_exec_stderr_file"
        elif [ "$dryad_sprouts_run_exec_stderr_mode" = join ]; then
            dryad_sprouts_run_env_command "$@" >"$dryad_sprouts_run_exec_stdout_file"
        else
            dryad_sprouts_run_env_command "$@" >"$dryad_sprouts_run_exec_stdout_file" 2>/dev/null
        fi
    elif [ "$dryad_sprouts_run_exec_stdout_mode" = join ]; then
        if [ "$dryad_sprouts_run_exec_stderr_mode" = file ]; then
            dryad_sprouts_run_env_command "$@" 2>"$dryad_sprouts_run_exec_stderr_file"
        elif [ "$dryad_sprouts_run_exec_stderr_mode" = join ]; then
            dryad_sprouts_run_env_command "$@"
        else
            dryad_sprouts_run_env_command "$@" 2>/dev/null
        fi
    else
        if [ "$dryad_sprouts_run_exec_stderr_mode" = file ]; then
            dryad_sprouts_run_env_command "$@" >/dev/null 2>"$dryad_sprouts_run_exec_stderr_file"
        elif [ "$dryad_sprouts_run_exec_stderr_mode" = join ]; then
            dryad_sprouts_run_env_command "$@" >/dev/null
        else
            dryad_sprouts_run_env_command "$@" >/dev/null 2>/dev/null
        fi
    fi
}

dryad_sprouts_run_one_stem () {
    dryad_sprouts_run_one_stem_garden=$1
    dryad_sprouts_run_one_stem_sprout=$2
    dryad_sprouts_run_one_stem_descriptor=$3
    dryad_sprouts_run_one_stem_path=$4
    dryad_sprouts_run_one_stem_selected_count=$5
    shift 5

    dryad_sprouts_run_one_stem_label=$dryad_sprouts_run_one_stem_descriptor
    [ -n "$dryad_sprouts_run_one_stem_label" ] ||
        dryad_sprouts_run_one_stem_label=default

    dryad_sprouts_run_one_stem_cli_bin=$(dryad_sprouts_run_cli_bin)
    case $dryad_sprouts_run_one_stem_cli_bin in
        */* )
            dryad_sprouts_run_one_stem_cli_dir=$(dirname "$dryad_sprouts_run_one_stem_cli_bin")
            ;;
        * )
            dryad_sprouts_run_one_stem_cli_dir=$(pwd -P)
            ;;
    esac
    dryad_sprouts_run_one_stem_path_env=$(dryad_sprouts_run_stem_path_env "$dryad_sprouts_run_one_stem_path" "$dryad_sprouts_run_one_stem_cli_dir")
    dryad_sprouts_run_one_stem_command=$(dryad_sprouts_run_resolve_command "$dryad_sprouts_run_one_stem_path" "$dryad_sprouts_run_one_stem_path_env") ||
        return 1

    if [ ! -f "$dryad_sprouts_run_one_stem_command" ]; then
        printf '%s\n' "dryad-sh: error: missing stem main \"$dryad_sprouts_run_one_stem_command\"" >&2
        return 1
    fi
    if [ ! -x "$dryad_sprouts_run_one_stem_command" ]; then
        printf '%s\n' "dryad-sh: error: stem main is not executable \"$dryad_sprouts_run_one_stem_command\"" >&2
        return 1
    fi

    dryad_sprouts_run_one_stem_context_name=$dryad_sprouts_run_context
    [ -n "$dryad_sprouts_run_one_stem_context_name" ] ||
        dryad_sprouts_run_one_stem_context_name=default
    dryad_sprouts_run_one_stem_context=$dryad_sprouts_run_one_stem_garden/dyd/heap/contexts/$dryad_sprouts_run_one_stem_context_name
    mkdir -p "$dryad_sprouts_run_one_stem_context"

    dryad_sprouts_run_one_stem_rel=${dryad_sprouts_run_one_stem_sprout#"$dryad_sprouts_run_one_stem_garden"/}
    dryad_sprouts_run_one_stem_stdout_mode=null
    dryad_sprouts_run_one_stem_stdout_file=
    dryad_sprouts_run_one_stem_stderr_mode=null
    dryad_sprouts_run_one_stem_stderr_file=

    if [ -n "$dryad_sprouts_run_log_stdout" ]; then
        dryad_sprouts_run_one_stem_stdout_mode=file
        dryad_sprouts_run_sanitize_segment_load "$dryad_sprouts_run_one_stem_rel"
        dryad_sprouts_run_one_stem_stdout_name=dyd-sprout-run--$dyd_ret0
        if [ "$dryad_sprouts_run_one_stem_selected_count" -gt 1 ] || [ -n "$dryad_sprouts_run_one_stem_descriptor" ]; then
            dryad_sprouts_run_sanitize_segment_load "$dryad_sprouts_run_one_stem_label"
            dryad_sprouts_run_one_stem_stdout_name=$dryad_sprouts_run_one_stem_stdout_name--$dyd_ret0
        fi
        dryad_sprouts_run_one_stem_stdout_file=$dryad_sprouts_run_log_stdout/$dryad_sprouts_run_one_stem_stdout_name.out
    elif [ "$dryad_sprouts_run_join_stdout" = 1 ]; then
        dryad_sprouts_run_one_stem_stdout_mode=join
    fi

    if [ -n "$dryad_sprouts_run_log_stderr" ]; then
        dryad_sprouts_run_one_stem_stderr_mode=file
        dryad_sprouts_run_sanitize_segment_load "$dryad_sprouts_run_one_stem_rel"
        dryad_sprouts_run_one_stem_stderr_name=dyd-sprout-run--$dyd_ret0
        if [ "$dryad_sprouts_run_one_stem_selected_count" -gt 1 ] || [ -n "$dryad_sprouts_run_one_stem_descriptor" ]; then
            dryad_sprouts_run_sanitize_segment_load "$dryad_sprouts_run_one_stem_label"
            dryad_sprouts_run_one_stem_stderr_name=$dryad_sprouts_run_one_stem_stderr_name--$dyd_ret0
        fi
        dryad_sprouts_run_one_stem_stderr_file=$dryad_sprouts_run_log_stderr/$dryad_sprouts_run_one_stem_stderr_name.err
    elif [ "$dryad_sprouts_run_join_stderr" = 1 ]; then
        dryad_sprouts_run_one_stem_stderr_mode=join
    fi

    dryad_sprouts_run_log "sprout run starting sprout=$dryad_sprouts_run_one_stem_sprout variant=$dryad_sprouts_run_one_stem_label"
    dryad_sprouts_run_exec_with_redirs \
        "$dryad_sprouts_run_one_stem_stdout_mode" \
        "$dryad_sprouts_run_one_stem_stdout_file" \
        "$dryad_sprouts_run_one_stem_stderr_mode" \
        "$dryad_sprouts_run_one_stem_stderr_file" \
        "$dryad_sprouts_run_inherit" \
        "$dryad_sprouts_run_one_stem_command" \
        "$dryad_sprouts_run_one_stem_path" \
        "$dryad_sprouts_run_one_stem_context" \
        "$dryad_sprouts_run_one_stem_garden" \
        "$dryad_sprouts_run_one_stem_path_env" \
        "$dryad_sprouts_run_one_stem_cli_bin" \
        "$@" ||
        return $?
    dryad_sprouts_run_log "sprout run finished sprout=$dryad_sprouts_run_one_stem_sprout variant=$dryad_sprouts_run_one_stem_label"
}

dryad_sprouts_run_one_sprout () {
    dryad_sprouts_run_one_garden=$1
    dryad_sprouts_run_one_sprout=$2
    dryad_sprouts_run_one_selector=$3
    dryad_sprouts_run_one_stems=$4
    shift 4

    dryad_sprouts_run_one_selector=$(dryad_sprouts_run_selector_normalize "$dryad_sprouts_run_one_selector") ||
        return 1
    dryad_sprouts_run_selector_validate "$dryad_sprouts_run_one_stems" "$dryad_sprouts_run_one_selector" ||
        return 1

    dryad_sprouts_run_one_selected=
    dryad_sprouts_run_one_selected_count=0
    while IFS='|' read -r dryad_sprouts_run_one_descriptor dryad_sprouts_run_one_stem; do
        [ -n "$dryad_sprouts_run_one_stem" ] || continue
        if dryad_root_variant_selector_matches_descriptor "$dryad_sprouts_run_one_selector" "$dryad_sprouts_run_one_descriptor"; then
            dryad_sprouts_run_one_selected="${dryad_sprouts_run_one_selected}
$dryad_sprouts_run_one_descriptor|$dryad_sprouts_run_one_stem"
            dryad_sprouts_run_one_selected_count=$((dryad_sprouts_run_one_selected_count + 1))
        fi
    done <<EOF
$dryad_sprouts_run_one_stems
EOF

    if [ "$dryad_sprouts_run_one_selected_count" -eq 0 ]; then
        printf '%s\n' "dryad-sh: error: resolved sprout run variants are empty" >&2
        return 1
    fi

    while IFS='|' read -r dryad_sprouts_run_one_descriptor dryad_sprouts_run_one_stem; do
        [ -n "$dryad_sprouts_run_one_stem" ] || continue
        dryad_sprouts_run_one_stem "$dryad_sprouts_run_one_garden" "$dryad_sprouts_run_one_sprout" "$dryad_sprouts_run_one_descriptor" "$dryad_sprouts_run_one_stem" "$dryad_sprouts_run_one_selected_count" "$@" ||
            return $?
    done <<EOF
$dryad_sprouts_run_one_selected
EOF
}

dryad_sprouts_run_target () {
    dryad_sprouts_run_target_garden=$1
    dryad_sprouts_run_target_sprout=$2
    dryad_sprouts_run_target_selector=$3
    shift 3

    dryad_sprouts_run_target_stems=$(dryad_sprouts_run_available_stems "$dryad_sprouts_run_target_sprout") ||
        return 1

    if [ -z "$dryad_sprouts_run_target_stems" ]; then
        printf '%s\n' "dryad-sh: error: sprout has no stem dependencies: $dryad_sprouts_run_target_sprout" >&2
        return 1
    fi

    dryad_sprouts_run_target_display=${dryad_sprouts_run_target_sprout#"$dryad_sprouts_run_target_garden"/}
    if ! dryad_sprouts_run_matches_filters "$dryad_sprouts_run_target_display" "$dryad_sprouts_run_target_sprout" "$dryad_sprouts_run_target_stems"; then
        return 0
    fi

    dryad_sprouts_run_log "sprout run requested sprout=$dryad_sprouts_run_target_sprout variant_selector=${dryad_sprouts_run_target_selector:-default}"
    if dryad_sprouts_run_one_sprout "$dryad_sprouts_run_target_garden" "$dryad_sprouts_run_target_sprout" "$dryad_sprouts_run_target_selector" "$dryad_sprouts_run_target_stems" "$@"; then
        dryad_sprouts_run_log "sprout run completed sprout=$dryad_sprouts_run_target_sprout variant_selector=${dryad_sprouts_run_target_selector:-default}"
        return 0
    else
        dryad_sprouts_run_status=$?
        dryad_sprouts_run_log "sprout threw error during execution sprout=$dryad_sprouts_run_target_sprout variant_selector=${dryad_sprouts_run_target_selector:-default}"
        return "$dryad_sprouts_run_status"
    fi
}

dryad_cmd_sprouts_run () {
    dryad_roots_include=
    dryad_roots_exclude=
    dryad_sprouts_run_from_stdin=0
    dryad_sprouts_run_variant=
    dryad_sprouts_run_context=
    dryad_sprouts_run_inherit=0
    dryad_sprouts_run_command=
    dryad_sprouts_run_ignore_errors=0
    dryad_sprouts_run_join_stdout=0
    dryad_sprouts_run_join_stderr=0
    dryad_sprouts_run_log_stdout=
    dryad_sprouts_run_log_stderr=
    dryad_sprouts_run_confirm=

    while [ "$#" -gt 0 ]; do
        dryad_sprouts_run_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_run_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad sprouts run [--include=<filter>] [--exclude=<filter>] [--variant=<descriptor>] [--from-stdin] -- [args...]
EOF
                return 0
                ;;
            --include=* )
                dryad_roots_include="${dryad_roots_include}
${dryad_sprouts_run_arg#--include=}"
                shift
                ;;
            --include )
                [ "$#" -gt 1 ] || dryad_die "--include requires a value"
                dryad_roots_include="${dryad_roots_include}
$2"
                shift 2
                ;;
            --exclude=* )
                dryad_roots_exclude="${dryad_roots_exclude}
${dryad_sprouts_run_arg#--exclude=}"
                shift
                ;;
            --exclude )
                [ "$#" -gt 1 ] || dryad_die "--exclude requires a value"
                dryad_roots_exclude="${dryad_roots_exclude}
$2"
                shift 2
                ;;
            --from-stdin )
                dryad_sprouts_run_from_stdin=1
                shift
                ;;
            --variant=* )
                dryad_fs_descriptor_normalize_load "${dryad_sprouts_run_arg#--variant=}"
                dryad_sprouts_run_variant=$dyd_ret0
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                dryad_fs_descriptor_normalize_load "$2"
                dryad_sprouts_run_variant=$dyd_ret0
                shift 2
                ;;
            --context=* )
                dryad_sprouts_run_context=${dryad_sprouts_run_arg#--context=}
                shift
                ;;
            --context )
                [ "$#" -gt 1 ] || dryad_die "--context requires a value"
                dryad_sprouts_run_context=$2
                shift 2
                ;;
            --inherit )
                dryad_sprouts_run_inherit=1
                shift
                ;;
            --ignore-errors )
                dryad_sprouts_run_ignore_errors=1
                shift
                ;;
            --join-stdout )
                dryad_sprouts_run_join_stdout=1
                shift
                ;;
            --join-stderr )
                dryad_sprouts_run_join_stderr=1
                shift
                ;;
            --log-stdout=* )
                dryad_sprouts_run_log_stdout=${dryad_sprouts_run_arg#--log-stdout=}
                dryad_sprouts_run_join_stdout=0
                shift
                ;;
            --log-stdout )
                [ "$#" -gt 1 ] || dryad_die "--log-stdout requires a value"
                dryad_sprouts_run_log_stdout=$2
                dryad_sprouts_run_join_stdout=0
                shift 2
                ;;
            --log-stderr=* )
                dryad_sprouts_run_log_stderr=${dryad_sprouts_run_arg#--log-stderr=}
                dryad_sprouts_run_join_stderr=0
                shift
                ;;
            --log-stderr )
                [ "$#" -gt 1 ] || dryad_die "--log-stderr requires a value"
                dryad_sprouts_run_log_stderr=$2
                dryad_sprouts_run_join_stderr=0
                shift 2
                ;;
            --confirm=* )
                dryad_sprouts_run_confirm=${dryad_sprouts_run_arg#--confirm=}
                shift
                ;;
            --confirm )
                [ "$#" -gt 1 ] || dryad_die "--confirm requires a value"
                dryad_sprouts_run_confirm=$2
                shift 2
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
                dryad_die "unsupported sprouts run option: $1"
                ;;
            * )
                break
                ;;
        esac
    done

    dryad_sprouts_run_garden=$(dryad_garden_find)
    dryad_sprouts_run_stdin_file=

    if [ "$dryad_sprouts_run_from_stdin" = 1 ]; then
        dryad_sprouts_run_stdin_file=$(mktemp 2>/dev/null || mktemp -t dryad-sprouts-run)
        while IFS= read -r dryad_sprouts_run_stdin_ref; do
            dryad_sprouts_run_stdin_ref=$(printf '%s\n' "$dryad_sprouts_run_stdin_ref" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
            [ -n "$dryad_sprouts_run_stdin_ref" ] || continue
            printf '%s\n' "$dryad_sprouts_run_stdin_ref" >> "$dryad_sprouts_run_stdin_file"
        done
    fi

    if [ -n "$dryad_sprouts_run_confirm" ]; then
        if [ "$dryad_sprouts_run_from_stdin" = 1 ]; then
            dryad_sprouts_run_confirm_print_stdin "$dryad_sprouts_run_garden" "$dryad_sprouts_run_stdin_file"
        else
            dryad_sprouts_run_confirm_print_all "$dryad_sprouts_run_garden"
        fi
        dryad_sprouts_run_confirm_prompt
    fi

    if [ "$dryad_sprouts_run_from_stdin" = 1 ]; then
        while IFS= read -r dryad_sprouts_run_stdin_ref; do
            dryad_sprouts_run_ref_parse "$dryad_sprouts_run_stdin_ref"
            if [ -n "$dryad_sprouts_run_variant" ] && [ -n "$dryad_sprouts_run_ref_selector" ]; then
                dryad_die "sprouts run selector specified in both stdin sprout_ref and --variant"
            fi
            dryad_sprouts_run_stdin_selector=$dryad_sprouts_run_variant
            [ -z "$dryad_sprouts_run_ref_selector" ] ||
                dryad_sprouts_run_stdin_selector=$dryad_sprouts_run_ref_selector
            dryad_sprouts_run_resolve_path_load "$dryad_sprouts_run_garden" "$dryad_sprouts_run_ref_path"
            dryad_sprouts_run_stdin_sprout=$dyd_ret0
            if dryad_sprouts_run_target "$dryad_sprouts_run_garden" "$dryad_sprouts_run_stdin_sprout" "$dryad_sprouts_run_stdin_selector" "$@"; then
                :
            else
                dryad_sprouts_run_stdin_status=$?
                [ "$dryad_sprouts_run_ignore_errors" = 1 ] ||
                    return "$dryad_sprouts_run_stdin_status"
            fi
        done < "$dryad_sprouts_run_stdin_file"
        rm -f "$dryad_sprouts_run_stdin_file"
        return 0
    fi

    dryad_sprouts_run_dir=$dryad_sprouts_run_garden/dyd/sprouts
    dryad_sprouts_find_sprouts "$dryad_sprouts_run_dir" |
        while IFS= read -r dryad_sprouts_run_sprout; do
            if dryad_sprouts_run_target "$dryad_sprouts_run_garden" "$dryad_sprouts_run_sprout" "$dryad_sprouts_run_variant" "$@"; then
                :
            else
                dryad_sprouts_run_status=$?
                [ "$dryad_sprouts_run_ignore_errors" = 1 ] ||
                    exit "$dryad_sprouts_run_status"
            fi
        done
}

dryad_cmd_sprout_run_bool_option_load () {
    dryad_sprout_run_bool_arg=$1
    dryad_sprout_run_bool_name=$2
    dryad_sprout_run_bool_default=$3

    case $dryad_sprout_run_bool_arg in
        --*=* )
            dryad_bool_value_load "${dryad_sprout_run_bool_arg#--*=}"
            ;;
        * )
            dyd_ret0=$dryad_sprout_run_bool_default
            ;;
    esac
}

dryad_cmd_sprout_run () {
    dryad_roots_include=
    dryad_roots_exclude=
    dryad_sprouts_run_variant=
    dryad_sprouts_run_context=
    dryad_sprouts_run_inherit=0
    dryad_sprouts_run_command=
    dryad_sprouts_run_ignore_errors=0
    dryad_sprouts_run_join_stdout=1
    dryad_sprouts_run_join_stderr=1
    dryad_sprouts_run_log_stdout=
    dryad_sprouts_run_log_stderr=
    dryad_sprouts_run_confirm=
    dryad_sprout_run_ref=

    while [ "$#" -gt 0 ]; do
        dryad_sprout_run_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprout_run_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad sprout run [--variant=<descriptor>] [--context=<name>] [--inherit=<bool>] [--command=<name>] [--confirm=<string>] [--join-stdout=<bool>] [--join-stderr=<bool>] [--log-stdout=<dir>] [--log-stderr=<dir>] [--scope=<scope>] <sprout_ref> -- [args...]
EOF
                return 0
                ;;
            --variant=* )
                dryad_fs_descriptor_normalize_load "${dryad_sprout_run_arg#--variant=}"
                dryad_sprouts_run_variant=$dyd_ret0
                shift
                ;;
            --variant )
                [ "$#" -gt 1 ] || dryad_die "--variant requires a value"
                dryad_fs_descriptor_normalize_load "$2"
                dryad_sprouts_run_variant=$dyd_ret0
                shift 2
                ;;
            --context=* )
                dryad_sprouts_run_context=${dryad_sprout_run_arg#--context=}
                shift
                ;;
            --context )
                [ "$#" -gt 1 ] || dryad_die "--context requires a value"
                dryad_sprouts_run_context=$2
                shift 2
                ;;
            --inherit=* )
                dryad_cmd_sprout_run_bool_option_load "$dryad_sprout_run_arg" inherit 1
                dryad_sprouts_run_inherit=$dyd_ret0
                shift
                ;;
            --inherit )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_bool_value_load "$2"
                            dryad_sprouts_run_inherit=$dyd_ret0
                            shift 2
                            ;;
                        * )
                            dryad_sprouts_run_inherit=1
                            shift
                            ;;
                    esac
                else
                    dryad_sprouts_run_inherit=1
                    shift
                fi
                ;;
            --command=* )
                dryad_sprouts_run_command=${dryad_sprout_run_arg#--command=}
                shift
                ;;
            --command )
                [ "$#" -gt 1 ] || dryad_die "--command requires a value"
                dryad_sprouts_run_command=$2
                shift 2
                ;;
            --confirm=* )
                dryad_sprouts_run_confirm=${dryad_sprout_run_arg#--confirm=}
                shift
                ;;
            --confirm )
                [ "$#" -gt 1 ] || dryad_die "--confirm requires a value"
                dryad_sprouts_run_confirm=$2
                shift 2
                ;;
            --join-stdout=* )
                dryad_cmd_sprout_run_bool_option_load "$dryad_sprout_run_arg" join-stdout 1
                dryad_sprouts_run_join_stdout=$dyd_ret0
                shift
                ;;
            --join-stdout )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_bool_value_load "$2"
                            dryad_sprouts_run_join_stdout=$dyd_ret0
                            shift 2
                            ;;
                        * )
                            dryad_sprouts_run_join_stdout=1
                            shift
                            ;;
                    esac
                else
                    dryad_sprouts_run_join_stdout=1
                    shift
                fi
                ;;
            --join-stderr=* )
                dryad_cmd_sprout_run_bool_option_load "$dryad_sprout_run_arg" join-stderr 1
                dryad_sprouts_run_join_stderr=$dyd_ret0
                shift
                ;;
            --join-stderr )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_bool_value_load "$2"
                            dryad_sprouts_run_join_stderr=$dyd_ret0
                            shift 2
                            ;;
                        * )
                            dryad_sprouts_run_join_stderr=1
                            shift
                            ;;
                    esac
                else
                    dryad_sprouts_run_join_stderr=1
                    shift
                fi
                ;;
            --log-stdout=* )
                dryad_sprouts_run_log_stdout=${dryad_sprout_run_arg#--log-stdout=}
                dryad_sprouts_run_join_stdout=0
                shift
                ;;
            --log-stdout )
                [ "$#" -gt 1 ] || dryad_die "--log-stdout requires a value"
                dryad_sprouts_run_log_stdout=$2
                dryad_sprouts_run_join_stdout=0
                shift 2
                ;;
            --log-stderr=* )
                dryad_sprouts_run_log_stderr=${dryad_sprout_run_arg#--log-stderr=}
                dryad_sprouts_run_join_stderr=0
                shift
                ;;
            --log-stderr )
                [ "$#" -gt 1 ] || dryad_die "--log-stderr requires a value"
                dryad_sprouts_run_log_stderr=$2
                dryad_sprouts_run_join_stderr=0
                shift 2
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
                dryad_die "unsupported sprout run option: $1"
                ;;
            * )
                if [ -n "$dryad_sprout_run_ref" ]; then
                    dryad_die "unsupported sprout run argument: $1"
                fi
                dryad_sprout_run_ref=$1
                shift
                ;;
        esac
    done

    [ -n "$dryad_sprout_run_ref" ] ||
        dryad_die "sprout run requires sprout_ref"

    dryad_sprouts_run_garden=$(dryad_garden_find)
    dryad_sprouts_run_ref_parse "$dryad_sprout_run_ref"
    if [ -n "$dryad_sprouts_run_variant" ] && [ -n "$dryad_sprouts_run_ref_selector" ]; then
        dryad_die "sprout run selector specified in both sprout_ref and --variant"
    fi

    dryad_sprout_run_selector=$dryad_sprouts_run_variant
    [ -z "$dryad_sprouts_run_ref_selector" ] ||
        dryad_sprout_run_selector=$dryad_sprouts_run_ref_selector
    dryad_sprouts_run_resolve_path_load "$dryad_sprouts_run_garden" "$dryad_sprouts_run_ref_path"
    dryad_sprout_run_sprout=$dyd_ret0

    if [ -n "$dryad_sprouts_run_confirm" ]; then
        printf '%s\n' "this package will be executed:"
        printf '%s\n' "$dryad_sprout_run_ref"
        dryad_sprouts_run_confirm_prompt
    fi

    dryad_sprouts_run_target "$dryad_sprouts_run_garden" "$dryad_sprout_run_sprout" "$dryad_sprout_run_selector" "$@"
}

dryad_cmd_sprout () {
    dryad_sprout_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_sprout_action in
        run )
            dryad_cmd_sprout_run "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad sprout run <sprout_ref>
EOF
            ;;
        * )
            dryad_die "unsupported sprout action: $dryad_sprout_action"
            ;;
    esac
}

dryad_path_has_owner_write () {
    dryad_path_has_owner_write_path=$1
    dryad_path_has_owner_write_mode=$(ls -ld "$dryad_path_has_owner_write_path" 2>/dev/null | sed 's/ .*//') ||
        return 1

    case $dryad_path_has_owner_write_mode in
        ??w* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_path_chmod_parent_owner_write () {
    dryad_path_chmod_parent_path=$1
    dryad_path_chmod_parent_parent=$(dirname "$dryad_path_chmod_parent_path")
    if [ -d "$dryad_path_chmod_parent_parent" ] &&
        ! dryad_path_has_owner_write "$dryad_path_chmod_parent_parent"; then
        chmod u+w "$dryad_path_chmod_parent_parent"
    fi
}

dryad_sprouts_ensure_dir_load () {
    dryad_sprouts_ensure_garden=$1
    dryad_sprouts_ensure_dir=$dryad_sprouts_ensure_garden/dyd/sprouts

    if [ -e "$dryad_sprouts_ensure_dir" ] && [ ! -d "$dryad_sprouts_ensure_dir" ]; then
        dryad_die "sprouts path exists and is not a directory: $dryad_sprouts_ensure_dir"
    fi

    if [ ! -d "$dryad_sprouts_ensure_dir" ]; then
        dryad_path_chmod_parent_owner_write "$dryad_sprouts_ensure_dir"
        mkdir -p "$dryad_sprouts_ensure_dir"
        chmod 551 "$dryad_sprouts_ensure_dir"
    fi

    dyd_ret0=$dryad_sprouts_ensure_dir
}

dryad_sprouts_make_tree_removable () {
    dryad_sprouts_make_tree_path=$1

    [ -e "$dryad_sprouts_make_tree_path" ] || [ -L "$dryad_sprouts_make_tree_path" ] || return 0

    dryad_path_chmod_parent_owner_write "$dryad_sprouts_make_tree_path"

    if [ -d "$dryad_sprouts_make_tree_path" ] && [ ! -L "$dryad_sprouts_make_tree_path" ]; then
        find "$dryad_sprouts_make_tree_path" -type d -exec chmod u+w {} \;
    fi
}

dryad_sprouts_remove_children () {
    dryad_sprouts_remove_children_dir=$1

    [ -d "$dryad_sprouts_remove_children_dir" ] || return 0

    dryad_sprouts_remove_restore_owner_write=0
    if ! dryad_path_has_owner_write "$dryad_sprouts_remove_children_dir"; then
        dryad_sprouts_remove_restore_owner_write=1
    fi

    dryad_sprouts_make_tree_removable "$dryad_sprouts_remove_children_dir"

    find "$dryad_sprouts_remove_children_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} \;

    if [ "$dryad_sprouts_remove_restore_owner_write" = 1 ]; then
        chmod u-w "$dryad_sprouts_remove_children_dir"
    fi
}

dryad_sprouts_prune () {
    dryad_sprouts_prune_garden=$1
    dryad_sprouts_ensure_dir_load "$dryad_sprouts_prune_garden"
    dryad_sprouts_prune_dir=$dyd_ret0
    dryad_sprouts_prune_roots=$dryad_sprouts_prune_garden/dyd/roots

    find "$dryad_sprouts_prune_dir" -depth -mindepth 1 -print |
        while IFS= read -r dryad_sprouts_prune_entry; do
            dryad_sprouts_prune_rel=${dryad_sprouts_prune_entry#"$dryad_sprouts_prune_dir"/}
            dryad_sprouts_prune_root_equiv=$dryad_sprouts_prune_roots/$dryad_sprouts_prune_rel

            if [ ! -e "$dryad_sprouts_prune_root_equiv" ] && [ ! -L "$dryad_sprouts_prune_root_equiv" ]; then
                dryad_sprouts_prune_parent=$(dirname "$dryad_sprouts_prune_entry")
                dryad_sprouts_prune_restore_parent=0
                if [ -d "$dryad_sprouts_prune_parent" ] &&
                    ! dryad_path_has_owner_write "$dryad_sprouts_prune_parent"; then
                    dryad_sprouts_prune_restore_parent=1
                fi

                dryad_sprouts_make_tree_removable "$dryad_sprouts_prune_entry"
                dryad_sprouts_prune_remove_status=0
                rm -rf "$dryad_sprouts_prune_entry" ||
                    dryad_sprouts_prune_remove_status=$?

                if [ "$dryad_sprouts_prune_restore_parent" = 1 ] &&
                    [ -d "$dryad_sprouts_prune_parent" ]; then
                    chmod u-w "$dryad_sprouts_prune_parent"
                fi

                [ "$dryad_sprouts_prune_remove_status" = 0 ] ||
                    return "$dryad_sprouts_prune_remove_status"
            fi
        done
}

dryad_cmd_sprouts_no_arg_action () {
    dryad_sprouts_no_arg_action=$1
    shift

    while [ "$#" -gt 0 ]; do
        dryad_sprouts_no_arg_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_no_arg_arg in
            --help | -h )
                cat <<EOF
Usage:
  dryad sprouts $dryad_sprouts_no_arg_action
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
                dryad_die "unsupported sprouts $dryad_sprouts_no_arg_action option: $1"
                ;;
            * )
                dryad_die "unsupported sprouts $dryad_sprouts_no_arg_action argument: $1"
                ;;
        esac
    done

    dryad_sprouts_no_arg_garden=$(dryad_garden_find)
    case $dryad_sprouts_no_arg_action in
        wipe )
            dryad_sprouts_ensure_dir_load "$dryad_sprouts_no_arg_garden"
            dryad_sprouts_no_arg_dir=$dyd_ret0
            dryad_sprouts_remove_children "$dryad_sprouts_no_arg_dir"
            ;;
        prune )
            dryad_sprouts_prune "$dryad_sprouts_no_arg_garden"
            ;;
        * )
            dryad_die "unsupported sprouts action: $dryad_sprouts_no_arg_action"
            ;;
    esac
}

dryad_cmd_sprouts_path () {
    while [ "$#" -gt 0 ]; do
        dryad_sprouts_path_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_path_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad sprouts path
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
                dryad_die "unsupported sprouts path option: $1"
                ;;
            * )
                dryad_die "unsupported sprouts path argument: $1"
                ;;
        esac
    done

    dryad_sprouts_path
}

dryad_cmd_sprouts_list () {
    dryad_roots_include=
    dryad_roots_exclude=
    dryad_sprouts_from_stdin=0
    dryad_sprouts_relative=1

    while [ "$#" -gt 0 ]; do
        dryad_sprouts_arg=$(dryad_strip_option_quotes "$1")
        case $dryad_sprouts_arg in
            --help | -h )
                cat <<'EOF'
Usage:
  dryad sprouts list [--relative=<bool>] [--include=<filter>] [--exclude=<filter>] [--from-stdin]
EOF
                return 0
                ;;
            --include=* )
                dryad_roots_include="${dryad_roots_include}
${dryad_sprouts_arg#--include=}"
                shift
                ;;
            --include )
                [ "$#" -gt 1 ] || dryad_die "--include requires a value"
                dryad_roots_include="${dryad_roots_include}
$2"
                shift 2
                ;;
            --exclude=* )
                dryad_roots_exclude="${dryad_roots_exclude}
${dryad_sprouts_arg#--exclude=}"
                shift
                ;;
            --exclude )
                [ "$#" -gt 1 ] || dryad_die "--exclude requires a value"
                dryad_roots_exclude="${dryad_roots_exclude}
$2"
                shift 2
                ;;
            --from-stdin )
                dryad_sprouts_from_stdin=1
                shift
                ;;
            --relative=* )
                dryad_bool_value_load "${dryad_sprouts_arg#--relative=}"
                dryad_sprouts_relative=$dyd_ret0
                shift
                ;;
            --relative )
                if [ "$#" -gt 1 ]; then
                    case $2 in
                        true | false | 0 | 1 )
                            dryad_bool_value_load "$2"
                            dryad_sprouts_relative=$dyd_ret0
                            shift 2
                            ;;
                        * )
                            dryad_sprouts_relative=1
                            shift
                            ;;
                    esac
                else
                    dryad_sprouts_relative=1
                    shift
                fi
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
                dryad_die "unsupported sprouts list option: $1"
                ;;
            * )
                dryad_die "unsupported sprouts list argument: $1"
                ;;
        esac
    done

    dryad_sprouts_garden=$(dryad_garden_find)

    if [ "$dryad_sprouts_from_stdin" = 1 ]; then
        dryad_sprouts_list_from_stdin "$dryad_sprouts_garden" | sort
        return 0
    fi

    dryad_sprouts_dir=$dryad_sprouts_garden/dyd/sprouts
    dryad_sprouts_find_sprouts "$dryad_sprouts_dir" | while IFS= read -r dryad_sprouts_path_entry; do
        dryad_sprouts_display=${dryad_sprouts_path_entry#"$dryad_sprouts_garden"/}
        dryad_sprouts_ref_descriptor_load "$dryad_sprouts_display"
        dryad_sprouts_descriptor=$dyd_ret0
        if dryad_sprouts_entry_matches_filters "$dryad_sprouts_display" "$dryad_sprouts_path_entry" "$dryad_sprouts_descriptor"; then
            dryad_sprouts_print_ref "$dryad_sprouts_path_entry" "$dryad_sprouts_garden"
        fi
    done | sort
}

dryad_cmd_sprouts () {
    dryad_sprouts_action=${1:-}
    if [ "$#" -gt 0 ]; then
        shift
    fi

    case $dryad_sprouts_action in
        path )
            dryad_cmd_sprouts_path "$@"
            ;;
        list )
            dryad_cmd_sprouts_list "$@"
            ;;
        prune | wipe )
            dryad_cmd_sprouts_no_arg_action "$dryad_sprouts_action" "$@"
            ;;
        run )
            dryad_cmd_sprouts_run "$@"
            ;;
        '' | help | --help | -h )
            cat <<'EOF'
Usage:
  dryad sprouts list
  dryad sprouts path
  dryad sprouts prune
  dryad sprouts run
  dryad sprouts wipe
EOF
            ;;
        * )
            dryad_die "unsupported sprouts action: $dryad_sprouts_action"
            ;;
    esac
}
