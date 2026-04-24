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
        dryad_root_variant_got=$(dryad_descriptor_value "$dryad_root_variant_descriptor" "$dryad_root_variant_dim" || true)

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

dryad_roots_variant_descriptors_uncached () {
    dryad_roots_variant_root=$1
    dryad_roots_variant_dir=$dryad_roots_variant_root/dyd/variants

    if [ ! -d "$dryad_roots_variant_dir" ]; then
        printf '\n'
        return 0
    fi

    dryad_roots_validate_filter_rules "$dryad_roots_variant_root"

    find "$dryad_roots_variant_dir" -mindepth 2 -maxdepth 2 -type f |
        sed "s|^$dryad_roots_variant_dir/||" |
        sort |
        awk -F/ -v base="$dryad_roots_variant_dir" '
            function trim(value) {
                sub(/^[[:space:]]+/, "", value)
                sub(/[[:space:]]+$/, "", value)
                return value
            }

            function read_enabled(rel,    raw, path) {
                path = base "/" rel
                if ((getline raw < path) < 0) {
                    return 0
                }
                close(path)
                raw = trim(raw)
                return raw == "true"
            }

            function append_option(dim, option) {
                if (!(dim in dim_seen)) {
                    dims[++dim_count] = dim
                    dim_seen[dim] = 1
                }
                options[dim] = options[dim] "\034" option
            }

            function descriptor_value(descriptor, dim,    parts, count, i, pair) {
                count = split(descriptor, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    if (pair ~ "^" dim "=") {
                        sub("^[^=]+=", "", pair)
                        return pair
                    }
                }
                return ""
            }

            function option_list_contains(list, target,    parts, count, i) {
                count = split(list, parts, ",")
                for (i = 1; i <= count; i++) {
                    if (parts[i] == target) {
                        return 1
                    }
                }
                return 0
            }

            function rule_matches(descriptor, rule,    parts, count, i, pair, dim, want, got) {
                if (rule == "") {
                    return 1
                }

                count = split(rule, parts, "+")
                for (i = 1; i <= count; i++) {
                    pair = parts[i]
                    dim = pair
                    sub("=.*$", "", dim)
                    want = pair
                    sub("^[^=]+=", "", want)
                    got = descriptor_value(descriptor, dim)

                    if (want == "none") {
                        if (got != "") {
                            return 0
                        }
                    } else if (want == "any") {
                        if (got == "") {
                            return 0
                        }
                    } else if (!option_list_contains(want, got)) {
                        return 0
                    }
                }
                return 1
            }

            function included(descriptor,    i) {
                if (include_count == 0) {
                    return 1
                }
                for (i = 1; i <= include_count; i++) {
                    if (rule_matches(descriptor, include_rules[i])) {
                        return 1
                    }
                }
                return 0
            }

            function excluded(descriptor,    i) {
                for (i = 1; i <= exclude_count; i++) {
                    if (rule_matches(descriptor, exclude_rules[i])) {
                        return 1
                    }
                }
                return 0
            }

            function emit_variant(descriptor) {
                if (!included(descriptor) || excluded(descriptor)) {
                    return
                }
                if (!(descriptor in emitted)) {
                    emitted[descriptor] = 1
                    output[++output_count] = descriptor
                }
            }

            function walk(dim_index, prefix,    dim, parts, count, i, option, next_prefix) {
                if (dim_index > dim_count) {
                    emit_variant(prefix)
                    return
                }

                dim = dims[dim_index]
                count = split(options[dim], parts, "\034")
                for (i = 2; i <= count; i++) {
                    option = parts[i]
                    if (option == "none") {
                        next_prefix = prefix
                    } else if (prefix == "") {
                        next_prefix = dim "=" option
                    } else {
                        next_prefix = prefix "+" dim "=" option
                    }
                    walk(dim_index + 1, next_prefix)
                }
            }

            {
                rel = $0
                if ($1 == "_include") {
                    if (read_enabled(rel)) {
                        include_rules[++include_count] = $2
                    }
                    next
                }
                if ($1 == "_exclude") {
                    if (read_enabled(rel)) {
                        exclude_rules[++exclude_count] = $2
                    }
                    next
                }
                if (read_enabled(rel)) {
                    append_option($1, $2)
                }
            }

            END {
                if (dim_count == 0) {
                    emit_variant("")
                } else {
                    walk(1, "")
                }

                for (i = 1; i <= output_count; i++) {
                    print output[i]
                }
            }
        '
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
