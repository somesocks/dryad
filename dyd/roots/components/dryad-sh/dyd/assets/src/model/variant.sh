dryad_root_variant_selector_matches_descriptor () {
    dryad_root_variant_selector=$1
    dryad_root_variant_descriptor=$2

    [ -n "$dryad_root_variant_selector" ] || return 0

    dryad_root_variant_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_variant_selector
    IFS=$dryad_root_variant_old_ifs

    for dryad_root_variant_pair do
        dryad_root_variant_dim=${dryad_root_variant_pair%%=*}
        dryad_root_variant_want=${dryad_root_variant_pair#*=}
        if dryad_descriptor_value_load "$dryad_root_variant_descriptor" "$dryad_root_variant_dim"; then
            dryad_root_variant_got=$dyd_ret0
        else
            dryad_root_variant_got=
        fi

        case $dryad_root_variant_want in
            inherit )
                dryad_die "inherit is invalid in root variant selectors"
                ;;
            host )
                case $dryad_root_variant_dim in
                    os ) dryad_root_variant_want=$(dryad_host_os) ;;
                    arch ) dryad_root_variant_want=$(dryad_host_arch) ;;
                esac
                ;;
        esac

        if [ "$dryad_root_variant_want" = any ]; then
            [ -n "$dryad_root_variant_got" ] || return 1
            continue
        fi

        if [ "$dryad_root_variant_want" = none ]; then
            [ -z "$dryad_root_variant_got" ] || return 1
            continue
        fi

        [ -n "$dryad_root_variant_got" ] || return 1
        dryad_option_list_contains "$dryad_root_variant_want" "$dryad_root_variant_got" || return 1
    done

    return 0
}

dryad_roots_validate_filter_rules () {
    dryad_roots_filter_root=$1
    dryad_roots_filter_dir=$dryad_roots_filter_root/dyd/variants

    for dryad_roots_filter_kind in _include _exclude; do
        dryad_roots_filter_kind_dir=$dryad_roots_filter_dir/$dryad_roots_filter_kind
        [ -d "$dryad_roots_filter_kind_dir" ] || continue
        case $dryad_roots_filter_kind in
            _include ) dryad_roots_filter_label=included ;;
            _exclude ) dryad_roots_filter_label=excluded ;;
        esac

        for dryad_roots_filter_file in "$dryad_roots_filter_kind_dir"/*; do
            [ -f "$dryad_roots_filter_file" ] || continue
            dryad_roots_filter_enabled=$(tr -d '[:space:]' < "$dryad_roots_filter_file")
            [ "$dryad_roots_filter_enabled" = true ] || continue
            dryad_roots_filter_descriptor=$(basename "$dryad_roots_filter_file")

            dryad_roots_filter_old_ifs=$IFS
            IFS=+
            set -- $dryad_roots_filter_descriptor
            IFS=$dryad_roots_filter_old_ifs

            for dryad_roots_filter_pair do
                dryad_roots_filter_options=${dryad_roots_filter_pair#*=}
                dryad_roots_filter_opt_old_ifs=$IFS
                IFS=,
                set -- $dryad_roots_filter_options
                IFS=$dryad_roots_filter_opt_old_ifs
                for dryad_roots_filter_option do
                    case $dryad_roots_filter_option in
                        inherit | host )
                            dryad_die "$dryad_roots_filter_option option is not supported for $dryad_roots_filter_label variant selectors"
                            ;;
                    esac
                done
            done
        done
    done
}

dryad_roots_variant_file_enabled () {
    dryad_roots_variant_enabled_file=$1
    dryad_roots_variant_enabled_raw=
    if ! IFS= read -r dryad_roots_variant_enabled_raw < "$dryad_roots_variant_enabled_file"; then
        [ -n "$dryad_roots_variant_enabled_raw" ] || return 1
    fi
    dryad_roots_variant_enabled_value=$(printf '%s\n' "$dryad_roots_variant_enabled_raw" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    [ "$dryad_roots_variant_enabled_value" = true ]
}

dryad_roots_variant_dim_seen () {
    dryad_roots_variant_dim_seen_target=$1

    while IFS= read -r dryad_roots_variant_dim_seen_dim; do
        [ -n "$dryad_roots_variant_dim_seen_dim" ] || continue
        [ "$dryad_roots_variant_dim_seen_dim" = "$dryad_roots_variant_dim_seen_target" ] && return 0
    done <<EOF
$dryad_roots_variant_dims
EOF

    return 1
}

dryad_roots_variant_append_option () {
    dryad_roots_variant_append_dim=$1
    dryad_roots_variant_append_option=$2

    if ! dryad_roots_variant_dim_seen "$dryad_roots_variant_append_dim"; then
        dryad_roots_variant_dims="${dryad_roots_variant_dims}${dryad_roots_variant_append_dim}
"
    fi
    dryad_roots_variant_options="${dryad_roots_variant_options}${dryad_roots_variant_append_dim}	${dryad_roots_variant_append_option}
"
}

dryad_roots_variant_rule_matches () {
    dryad_roots_variant_rule_descriptor=$1
    dryad_roots_variant_rule=$2

    [ -n "$dryad_roots_variant_rule" ] || return 0

    dryad_roots_variant_rule_old_ifs=$IFS
    IFS=+
    set -- $dryad_roots_variant_rule
    IFS=$dryad_roots_variant_rule_old_ifs

    for dryad_roots_variant_rule_pair do
        dryad_roots_variant_rule_dim=${dryad_roots_variant_rule_pair%%=*}
        dryad_roots_variant_rule_want=${dryad_roots_variant_rule_pair#*=}
        if dryad_descriptor_value_load "$dryad_roots_variant_rule_descriptor" "$dryad_roots_variant_rule_dim"; then
            dryad_roots_variant_rule_got=$dyd_ret0
        else
            dryad_roots_variant_rule_got=
        fi

        if [ "$dryad_roots_variant_rule_want" = none ]; then
            [ -z "$dryad_roots_variant_rule_got" ] || return 1
            continue
        fi
        if [ "$dryad_roots_variant_rule_want" = any ]; then
            [ -n "$dryad_roots_variant_rule_got" ] || return 1
            continue
        fi
        dryad_option_list_contains "$dryad_roots_variant_rule_want" "$dryad_roots_variant_rule_got" || return 1
    done

    return 0
}

dryad_roots_variant_included () {
    dryad_roots_variant_included_descriptor=$1

    [ -n "$dryad_roots_variant_include_rules" ] || return 0

    while IFS= read -r dryad_roots_variant_include_rule; do
        [ -n "$dryad_roots_variant_include_rule" ] || continue
        if dryad_roots_variant_rule_matches "$dryad_roots_variant_included_descriptor" "$dryad_roots_variant_include_rule"; then
            return 0
        fi
    done <<EOF
$dryad_roots_variant_include_rules
EOF

    return 1
}

dryad_roots_variant_excluded () {
    dryad_roots_variant_excluded_descriptor=$1

    while IFS= read -r dryad_roots_variant_exclude_rule; do
        [ -n "$dryad_roots_variant_exclude_rule" ] || continue
        if dryad_roots_variant_rule_matches "$dryad_roots_variant_excluded_descriptor" "$dryad_roots_variant_exclude_rule"; then
            return 0
        fi
    done <<EOF
$dryad_roots_variant_exclude_rules
EOF

    return 1
}

dryad_roots_variant_emitted () {
    dryad_roots_variant_emitted_target=D$1

    while IFS= read -r dryad_roots_variant_emitted_record; do
        [ -n "$dryad_roots_variant_emitted_record" ] || continue
        [ "$dryad_roots_variant_emitted_record" = "$dryad_roots_variant_emitted_target" ] && return 0
    done <<EOF
$dryad_roots_variant_emitted_records
EOF

    return 1
}

dryad_roots_variant_emit () {
    dryad_roots_variant_emit_descriptor=$1

    dryad_roots_variant_included "$dryad_roots_variant_emit_descriptor" || return 0
    dryad_roots_variant_excluded "$dryad_roots_variant_emit_descriptor" && return 0
    dryad_roots_variant_emitted "$dryad_roots_variant_emit_descriptor" && return 0

    dryad_roots_variant_emitted_records="${dryad_roots_variant_emitted_records}D${dryad_roots_variant_emit_descriptor}
"
    printf '%s\n' "$dryad_roots_variant_emit_descriptor"
}

dryad_roots_variant_descriptors_uncached () {
    dryad_roots_variant_root=$1
    dryad_roots_variant_dir=$dryad_roots_variant_root/dyd/variants

    if [ ! -d "$dryad_roots_variant_dir" ]; then
        printf '\n'
        return 0
    fi

    dryad_roots_validate_filter_rules "$dryad_roots_variant_root"

    dryad_roots_variant_dims=
    dryad_roots_variant_options=
    dryad_roots_variant_include_rules=
    dryad_roots_variant_exclude_rules=
    dryad_roots_variant_records=$(find "$dryad_roots_variant_dir" -mindepth 2 -maxdepth 2 -type f |
        sed "s|^$dryad_roots_variant_dir/||" |
        sort)

    while IFS=/ read -r dryad_roots_variant_record_kind dryad_roots_variant_record_name; do
        [ -n "$dryad_roots_variant_record_kind" ] || continue
        [ -n "$dryad_roots_variant_record_name" ] || continue
        dryad_roots_variant_record_rel=$dryad_roots_variant_record_kind/$dryad_roots_variant_record_name
        dryad_roots_variant_file_enabled "$dryad_roots_variant_dir/$dryad_roots_variant_record_rel" || continue

        case $dryad_roots_variant_record_kind in
            _include )
                dryad_roots_variant_include_rules="${dryad_roots_variant_include_rules}${dryad_roots_variant_record_name}
"
                ;;
            _exclude )
                dryad_roots_variant_exclude_rules="${dryad_roots_variant_exclude_rules}${dryad_roots_variant_record_name}
"
                ;;
            * )
                dryad_roots_variant_append_option "$dryad_roots_variant_record_kind" "$dryad_roots_variant_record_name"
                ;;
        esac
    done <<EOF
$dryad_roots_variant_records
EOF

    dryad_roots_variant_current='D
'
    while IFS= read -r dryad_roots_variant_dim; do
        [ -n "$dryad_roots_variant_dim" ] || continue
        dryad_roots_variant_next=
        while IFS= read -r dryad_roots_variant_descriptor_record; do
            [ -n "$dryad_roots_variant_descriptor_record" ] || continue
            dryad_roots_variant_descriptor=${dryad_roots_variant_descriptor_record#D}
            while IFS='	' read -r dryad_roots_variant_option_dim dryad_roots_variant_option; do
                [ "$dryad_roots_variant_option_dim" = "$dryad_roots_variant_dim" ] || continue
                if [ "$dryad_roots_variant_option" = none ]; then
                    dryad_roots_variant_next_descriptor=$dryad_roots_variant_descriptor
                elif [ -n "$dryad_roots_variant_descriptor" ]; then
                    dryad_roots_variant_next_descriptor=$dryad_roots_variant_descriptor+$dryad_roots_variant_dim=$dryad_roots_variant_option
                else
                    dryad_roots_variant_next_descriptor=$dryad_roots_variant_dim=$dryad_roots_variant_option
                fi
                dryad_roots_variant_next="${dryad_roots_variant_next}D${dryad_roots_variant_next_descriptor}
"
            done <<EOF
$dryad_roots_variant_options
EOF
        done <<EOF
$dryad_roots_variant_current
EOF
        dryad_roots_variant_current=$dryad_roots_variant_next
    done <<EOF
$dryad_roots_variant_dims
EOF

    dryad_roots_variant_emitted_records=
    while IFS= read -r dryad_roots_variant_descriptor_record; do
        [ -n "$dryad_roots_variant_descriptor_record" ] || continue
        dryad_roots_variant_emit "${dryad_roots_variant_descriptor_record#D}"
    done <<EOF
$dryad_roots_variant_current
EOF
}

dryad_roots_variant_descriptors () {
    dryad_roots_variant_root=$1

    if dryad_roots_variant_result=$(dryad_memo_get roots-variant-descriptors "$dryad_roots_variant_root"); then
        dryad_profile_count memo.hit.roots-variant-descriptors
        printf '%s\n' "$dryad_roots_variant_result"
        return 0
    fi

    dryad_profile_count memo.miss.roots-variant-descriptors
    dryad_profile_count call.roots-variant-descriptors.uncached
    dryad_roots_variant_result=$(dryad_roots_variant_descriptors_uncached "$dryad_roots_variant_root") || return $?
    dryad_memo_put_value roots-variant-descriptors "$dryad_roots_variant_result" "$dryad_roots_variant_root"
    printf '%s\n' "$dryad_roots_variant_result"
}
