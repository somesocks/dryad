dryad_root_build_log () {
    case $dryad_log_level in
        info | debug | trace )
            printf 'dryad-sh: %s\n' "$*" >&2
            ;;
    esac
}

dryad_root_build_selected_descriptors_uncached () {
    dryad_root_build_selected_root=$1
    dryad_root_build_selected_selector=$2

    dryad_roots_variant_descriptors "$dryad_root_build_selected_root" |
        while IFS= read -r dryad_root_build_selected_descriptor; do
            if dryad_root_variant_selector_matches_descriptor "$dryad_root_build_selected_selector" "$dryad_root_build_selected_descriptor"; then
                printf '%s\n' "$dryad_root_build_selected_descriptor"
            fi
        done
}

dryad_root_build_selected_descriptors () {
    dryad_root_build_selected_root=$1
    dryad_root_build_selected_selector=$2

    if dryad_root_build_selected_result=$(dryad_memo_get root-build-selected-descriptors "$dryad_root_build_selected_root" "$dryad_root_build_selected_selector"); then
        dryad_profile_count memo.hit.root-build-selected-descriptors
        printf '%s\n' "$dryad_root_build_selected_result"
        return 0
    fi

    dryad_profile_count memo.miss.root-build-selected-descriptors
    dryad_root_build_selected_result=$(dryad_root_build_selected_descriptors_uncached "$dryad_root_build_selected_root" "$dryad_root_build_selected_selector") ||
        return $?
    dryad_memo_put_value root-build-selected-descriptors "$dryad_root_build_selected_result" "$dryad_root_build_selected_root" "$dryad_root_build_selected_selector"
    printf '%s\n' "$dryad_root_build_selected_result"
}

dryad_root_build_copy_dir_contents () {
    dryad_root_build_copy_src=$1
    dryad_root_build_copy_dst=$2

    [ -d "$dryad_root_build_copy_src" ] || return 0
    mkdir -p "$dryad_root_build_copy_dst"
    cp -R "$dryad_root_build_copy_src/." "$dryad_root_build_copy_dst/"
}

dryad_root_build_materialize_traits_apply_descriptor () {
    dryad_root_build_traits_descriptor=$1
    dryad_root_build_traits_dest=$2
    dryad_root_build_traits_rel=$3

    [ -n "$dryad_root_build_traits_descriptor" ] || return 0

    dryad_root_build_traits_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_traits_descriptor
    IFS=$dryad_root_build_traits_old_ifs

    for dryad_root_build_traits_pair do
        dryad_root_build_traits_dim=${dryad_root_build_traits_pair%%=*}
        dryad_root_build_traits_value=${dryad_root_build_traits_pair#*=}
        dryad_root_build_traits_path=$dryad_root_build_traits_dest/$dryad_root_build_traits_dim
        if [ -f "$dryad_root_build_traits_path" ]; then
            dryad_root_build_traits_existing=$(cat "$dryad_root_build_traits_path")
            if [ "$dryad_root_build_traits_existing" != "$dryad_root_build_traits_value" ]; then
                printf '%s\n' "dryad-sh: variant option overrides existing trait value path=dyd/roots/$dryad_root_build_traits_rel/dyd/traits/$dryad_root_build_traits_dim" >&2
            fi
        fi
        printf '%s' "$dryad_root_build_traits_value" > "$dryad_root_build_traits_dest/$dryad_root_build_traits_dim"
    done
}

dryad_root_build_materialize_traits () {
    dryad_root_build_traits_root=$1
    dryad_root_build_traits_descriptor=$2
    dryad_root_build_traits_workspace=$3
    dryad_roots_owning_selected_path_load "$dryad_root_build_traits_root" "$dryad_root_build_traits_descriptor" traits
    dryad_root_build_traits_selected=$dyd_ret0
    dryad_root_build_traits_dest=$dryad_root_build_traits_workspace/dyd/traits
    dryad_root_build_traits_garden=$(dryad_garden_find)
    dryad_root_build_traits_rel=${dryad_root_build_traits_root#"$dryad_root_build_traits_garden"/dyd/roots/}

    mkdir -p "$dryad_root_build_traits_dest"
    if [ -n "$dryad_root_build_traits_selected" ]; then
        dryad_root_build_copy_dir_contents "$dryad_root_build_traits_selected" "$dryad_root_build_traits_dest"
    fi

    dryad_root_build_materialize_traits_apply_descriptor "$dryad_root_build_traits_descriptor" "$dryad_root_build_traits_dest" "$dryad_root_build_traits_rel"
}

dryad_root_build_prepare_source () {
    dryad_root_build_prepare_root=$1
    dryad_root_build_prepare_descriptor=$2
    dryad_root_build_prepare_workspace=$3
    dryad_profile_count call.root-build.prepare-source

    mkdir -p "$dryad_root_build_prepare_workspace/dyd/dependencies"
    dryad_profile_time_block root-build.prepare-source.materialize-traits \
        dryad_root_build_materialize_traits "$dryad_root_build_prepare_root" "$dryad_root_build_prepare_descriptor" "$dryad_root_build_prepare_workspace"

    for dryad_root_build_prepare_kind in assets commands secrets docs; do
        dryad_roots_owning_selected_path_load "$dryad_root_build_prepare_root" "$dryad_root_build_prepare_descriptor" "$dryad_root_build_prepare_kind"
        dryad_root_build_prepare_selected=$dyd_ret0
        if [ -n "$dryad_root_build_prepare_selected" ]; then
            ln -s "$dryad_root_build_prepare_selected" "$dryad_root_build_prepare_workspace/dyd/$dryad_root_build_prepare_kind"
        fi
    done

    dryad_roots_owning_selected_path_load "$dryad_root_build_prepare_root" "$dryad_root_build_prepare_descriptor" requirements
    dryad_root_build_prepare_requirements=$dyd_ret0
    if [ -n "$dryad_root_build_prepare_requirements" ]; then
        ln -s "$dryad_root_build_prepare_requirements" "$dryad_root_build_prepare_workspace/dyd/~requirements"
        ln -s "$dryad_root_build_prepare_requirements" "$dryad_root_build_prepare_workspace/dyd/requirements"
    fi
}

dryad_root_build_target_selector_matches_descriptor () {
    dryad_root_build_target_selector=$1
    dryad_root_build_target_descriptor=$2
    dryad_root_build_target_parent_descriptor=$3

    dryad_root_build_target_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_target_selector
    IFS=$dryad_root_build_target_old_ifs

    for dryad_root_build_target_pair do
        dryad_root_build_target_dim=${dryad_root_build_target_pair%%=*}
        dryad_root_build_target_want=${dryad_root_build_target_pair#*=}
        if dryad_descriptor_value_load "$dryad_root_build_target_descriptor" "$dryad_root_build_target_dim"; then
            dryad_root_build_target_got=$dyd_ret0
        else
            dryad_root_build_target_got=
        fi

        case $dryad_root_build_target_want in
            inherit )
                if dryad_descriptor_value_load "$dryad_root_build_target_parent_descriptor" "$dryad_root_build_target_dim"; then
                    dryad_root_build_target_want=$dyd_ret0
                else
                    dryad_root_build_target_want=
                fi
                [ -n "$dryad_root_build_target_want" ] || dryad_root_build_target_want=none
                ;;
            host )
                case $dryad_root_build_target_dim in
                    os ) dryad_host_os_load; dryad_root_build_target_want=$dyd_ret0 ;;
                    arch ) dryad_host_arch_load; dryad_root_build_target_want=$dyd_ret0 ;;
                    * ) ;;
                esac
                ;;
        esac

        if [ "$dryad_root_build_target_want" = any ]; then
            [ -n "$dryad_root_build_target_got" ] || return 1
            continue
        fi

        if [ "$dryad_root_build_target_want" = none ]; then
            [ -z "$dryad_root_build_target_got" ] || return 1
            continue
        fi

        [ -n "$dryad_root_build_target_got" ] || return 1
        dryad_option_list_contains "$dryad_root_build_target_want" "$dryad_root_build_target_got" || return 1
    done

    dryad_root_build_target_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_target_descriptor
    IFS=$dryad_root_build_target_old_ifs

    for dryad_root_build_target_pair do
        dryad_root_build_target_dim=${dryad_root_build_target_pair%%=*}
        dryad_root_build_target_selector_value_load "$dryad_root_build_target_selector" "$dryad_root_build_target_dim" ||
            return 1
    done

    return 0
}

dryad_root_build_dependency_die () {
    dryad_root_build_dep_die_garden=$1
    dryad_root_build_dep_die_root=$2
    shift 2
    dryad_root_build_dep_die_rel=${dryad_root_build_dep_die_root#"$dryad_root_build_dep_die_garden"/dyd/roots/}
    dryad_die "error resolving root dependencies for dyd/roots/$dryad_root_build_dep_die_rel: $*"
}

dryad_root_build_variant_option_enabled () {
    dryad_root_build_variant_option_root=$1
    dryad_root_build_variant_option_dim=$2
    dryad_root_build_variant_option_name=$3
    dryad_root_build_variant_option_path=$dryad_root_build_variant_option_root/dyd/variants/$dryad_root_build_variant_option_dim/$dryad_root_build_variant_option_name

    [ -f "$dryad_root_build_variant_option_path" ] || return 1
    dryad_root_build_variant_option_value=$(tr -d '[:space:]' < "$dryad_root_build_variant_option_path")
    [ "$dryad_root_build_variant_option_value" = true ]
}

dryad_root_build_validate_target_option () {
    dryad_root_build_validate_option_garden=$1
    dryad_root_build_validate_option_source_root=$2
    dryad_root_build_validate_option_target_root=$3
    dryad_root_build_validate_option_dim=$4
    dryad_root_build_validate_option_name=$5
    dryad_root_build_validate_option_path=$dryad_root_build_validate_option_target_root/dyd/variants/$dryad_root_build_validate_option_dim/$dryad_root_build_validate_option_name

    if [ ! -f "$dryad_root_build_validate_option_path" ]; then
        dryad_root_build_dependency_die "$dryad_root_build_validate_option_garden" "$dryad_root_build_validate_option_source_root" "wrongly-specified requirement variant option: $dryad_root_build_validate_option_dim=$dryad_root_build_validate_option_name"
    fi

    if ! dryad_root_build_variant_option_enabled "$dryad_root_build_validate_option_target_root" "$dryad_root_build_validate_option_dim" "$dryad_root_build_validate_option_name"; then
        dryad_root_build_dependency_die "$dryad_root_build_validate_option_garden" "$dryad_root_build_validate_option_source_root" "disabled requirement variant option: $dryad_root_build_validate_option_dim=$dryad_root_build_validate_option_name"
    fi
}

dryad_root_build_target_selector_value_load () {
    dryad_root_build_target_selector_value_selector=$1
    dryad_root_build_target_selector_value_dim=$2
    dyd_ret0=

    dryad_root_build_target_selector_value_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_target_selector_value_selector
    IFS=$dryad_root_build_target_selector_value_old_ifs

    for dryad_root_build_target_selector_value_pair do
        case $dryad_root_build_target_selector_value_pair in
            "$dryad_root_build_target_selector_value_dim="* )
                dyd_ret0=${dryad_root_build_target_selector_value_pair#*=}
                return 0
                ;;
        esac
    done

    return 1
}

dryad_root_build_validate_target_selector () {
    dryad_root_build_validate_garden=$1
    dryad_root_build_validate_source_root=$2
    dryad_root_build_validate_target_root=$3
    dryad_root_build_validate_selector=$4
    dryad_root_build_validate_parent_descriptor=$5
    dryad_root_build_validate_variants=$dryad_root_build_validate_target_root/dyd/variants

    if [ ! -d "$dryad_root_build_validate_variants" ]; then
        if [ -n "$dryad_root_build_validate_selector" ]; then
            dryad_root_build_validate_old_ifs=$IFS
            IFS=+
            set -- $dryad_root_build_validate_selector
            IFS=$dryad_root_build_validate_old_ifs
            for dryad_root_build_validate_pair do
                dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "over-specified requirement variant dimension: ${dryad_root_build_validate_pair%%=*}"
            done
        fi
        return 0
    fi

    dryad_root_build_validate_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_validate_selector
    IFS=$dryad_root_build_validate_old_ifs
    for dryad_root_build_validate_pair do
        dryad_root_build_validate_dim=${dryad_root_build_validate_pair%%=*}
        case $dryad_root_build_validate_dim in
            '' ) continue ;;
            _include | _exclude ) ;;
        esac
        if [ ! -d "$dryad_root_build_validate_variants/$dryad_root_build_validate_dim" ]; then
            dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "over-specified requirement variant dimension: $dryad_root_build_validate_dim"
        fi
    done

    find "$dryad_root_build_validate_variants" -mindepth 1 -maxdepth 1 -type d |
        while IFS= read -r dryad_root_build_validate_dim_path; do
            dryad_root_build_validate_dim=$(basename "$dryad_root_build_validate_dim_path")
            case $dryad_root_build_validate_dim in
                _include | _exclude ) continue ;;
            esac

            if dryad_root_build_target_selector_value_load "$dryad_root_build_validate_selector" "$dryad_root_build_validate_dim"; then
                dryad_root_build_validate_requested=$dyd_ret0
            else
                dryad_root_build_validate_requested=
            fi
            if [ -z "$dryad_root_build_validate_requested" ]; then
                if ! dryad_root_build_variant_option_enabled "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" none; then
                    dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "under-specified requirement variant dimension: $dryad_root_build_validate_dim"
                fi
                continue
            fi

            dryad_root_build_validate_req_old_ifs=$IFS
            IFS=,
            set -- $dryad_root_build_validate_requested
            IFS=$dryad_root_build_validate_req_old_ifs

            for dryad_root_build_validate_option do
                case $dryad_root_build_validate_option in
                    inherit )
                        if dryad_descriptor_value_load "$dryad_root_build_validate_parent_descriptor" "$dryad_root_build_validate_dim"; then
                            dryad_root_build_validate_inherited=$dyd_ret0
                        else
                            dryad_root_build_validate_inherited=
                        fi
                        [ -n "$dryad_root_build_validate_inherited" ] || dryad_root_build_validate_inherited=none
                        dryad_root_build_validate_target_option "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_inherited"
                        ;;
                    any )
                        dryad_root_build_validate_any=0
                        for dryad_root_build_validate_option_path in "$dryad_root_build_validate_dim_path"/*; do
                            [ -f "$dryad_root_build_validate_option_path" ] || continue
                            dryad_root_build_validate_option_name=$(basename "$dryad_root_build_validate_option_path")
                            [ "$dryad_root_build_validate_option_name" != none ] || continue
                            if dryad_root_build_variant_option_enabled "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_option_name"; then
                                dryad_root_build_validate_any=1
                                break
                            fi
                        done
                        [ "$dryad_root_build_validate_any" = 1 ] ||
                            dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "no enabled variant options for any resolution: $dryad_root_build_validate_dim"
                        ;;
                    host )
                        case $dryad_root_build_validate_dim in
                            os ) dryad_host_os_load; dryad_root_build_validate_host=$dyd_ret0 ;;
                            arch ) dryad_host_arch_load; dryad_root_build_validate_host=$dyd_ret0 ;;
                            * )
                                dryad_root_build_dependency_die "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "host option is only supported for variant dimensions os/arch: $dryad_root_build_validate_dim"
                                ;;
                        esac
                        dryad_root_build_validate_target_option "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_host"
                        ;;
                    * )
                        dryad_root_build_validate_target_option "$dryad_root_build_validate_garden" "$dryad_root_build_validate_source_root" "$dryad_root_build_validate_target_root" "$dryad_root_build_validate_dim" "$dryad_root_build_validate_option"
                        ;;
                esac
            done
        done
}

dryad_root_build_validate_requirement_condition () {
    dryad_root_build_validate_condition_garden=$1
    dryad_root_build_validate_condition_root=$2
    dryad_root_build_validate_condition_name=$3
    dryad_root_build_validate_condition=$4

    case $dryad_root_build_validate_condition_name in
        *~* )
            [ -n "$dryad_root_build_validate_condition" ] ||
                dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
            ;;
        * )
            return 0
            ;;
    esac

    dryad_root_build_validate_condition_old_ifs=$IFS
    IFS=+
    set -- $dryad_root_build_validate_condition
    IFS=$dryad_root_build_validate_condition_old_ifs

    for dryad_root_build_validate_condition_pair do
        case $dryad_root_build_validate_condition_pair in
            [ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*=* )
                ;;
            * )
                dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
                ;;
        esac

        dryad_root_build_validate_condition_dim=${dryad_root_build_validate_condition_pair%%=*}
        dryad_root_build_validate_condition_options=${dryad_root_build_validate_condition_pair#*=}
        case $dryad_root_build_validate_condition_dim in
            *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]* | '' )
                dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
                ;;
        esac

        dryad_root_build_validate_condition_opt_old_ifs=$IFS
        IFS=,
        set -- $dryad_root_build_validate_condition_options
        IFS=$dryad_root_build_validate_condition_opt_old_ifs

        for dryad_root_build_validate_condition_option do
            case $dryad_root_build_validate_condition_option in
                host )
                    case $dryad_root_build_validate_condition_dim in
                        os | arch ) ;;
                        * )
                            dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "host option is only supported for variant dimensions os/arch: $dryad_root_build_validate_condition_dim"
                            ;;
                    esac
                    ;;
                '' )
                    dryad_root_build_dependency_die "$dryad_root_build_validate_condition_garden" "$dryad_root_build_validate_condition_root" "malformed requirement condition descriptor: $dryad_root_build_validate_condition_name"
                    ;;
            esac
        done
    done
}

dryad_root_build_ref_path () {
    dryad_root_build_ref=$1
    case $dryad_root_build_ref in
        *~* )
            printf '%s\n' "${dryad_root_build_ref%%~*}"
            ;;
        * )
            printf '%s\n' "$dryad_root_build_ref"
            ;;
    esac
}

dryad_root_build_ref_descriptor () {
    dryad_root_build_ref=$1
    case $dryad_root_build_ref in
        *~* )
            printf '%s\n' "${dryad_root_build_ref#*~}"
            ;;
    esac
}

dryad_root_build_descriptor_needs_suffix () {
    dryad_root_build_suffix_selector=$1
    dryad_root_build_suffix_count=$2

    [ "$dryad_root_build_suffix_count" -gt 1 ] && return 0
    case $dryad_root_build_suffix_selector in
        *'=any'* ) return 0 ;;
        *','* ) return 0 ;;
    esac
    return 1
}

dryad_root_build_prepare_dependencies () {
    dryad_root_build_deps_garden=$1
    dryad_root_build_deps_root=$2
    dryad_root_build_deps_descriptor=$3
    dryad_root_build_deps_workspace=$4
    dryad_root_build_deps_sources_out=${5:-}
    dryad_root_build_deps_req_dir=$dryad_root_build_deps_workspace/dyd/~requirements
    dryad_profile_count call.root-build.prepare-dependencies

    [ -d "$dryad_root_build_deps_req_dir" ] || return 0

    for dryad_root_build_deps_file in "$dryad_root_build_deps_req_dir"/* "$dryad_root_build_deps_req_dir"/.[!.]* "$dryad_root_build_deps_req_dir"/..?*; do
        [ -f "$dryad_root_build_deps_file" ] || [ -L "$dryad_root_build_deps_file" ] || continue

        dryad_root_build_deps_name=${dryad_root_build_deps_file##*/}
        dryad_root_build_deps_alias=${dryad_root_build_deps_name%%~*}
        dryad_root_build_deps_condition=
        case $dryad_root_build_deps_name in
            *~* )
                dryad_root_build_deps_condition=${dryad_root_build_deps_name#*~}
                ;;
        esac

        dryad_root_build_validate_requirement_condition "$dryad_root_build_deps_garden" "$dryad_root_build_deps_root" "$dryad_root_build_deps_name" "$dryad_root_build_deps_condition"
        if ! dryad_condition_matches_descriptor "$dryad_root_build_deps_condition" "$dryad_root_build_deps_descriptor"; then
            continue
        fi

        dryad_requirement_target_spec_load "$dryad_root_build_deps_file"
        dryad_root_build_deps_spec=$dyd_ret0
        case $dryad_root_build_deps_spec in
            root:* )
                ;;
            * )
                dryad_die "requirement target must use root: scheme: $dryad_root_build_deps_file"
                ;;
        esac

        dryad_root_build_deps_body=${dryad_root_build_deps_spec#root:}
        dryad_root_build_deps_query=
        case $dryad_root_build_deps_body in
            *\?* )
                dryad_root_build_deps_target_rel=${dryad_root_build_deps_body%%\?*}
                dryad_root_build_deps_query=${dryad_root_build_deps_body#*\?}
                ;;
            * )
                dryad_root_build_deps_target_rel=$dryad_root_build_deps_body
                ;;
        esac

        dryad_url_query_warn_if_noncanonical "$dryad_root_build_deps_query"
        dryad_url_query_to_descriptor_load "$dryad_root_build_deps_query"
        dryad_root_build_deps_selector=$dyd_ret0
        dryad_root_build_deps_file_dir=$(dryad_clean_cd "$(dirname "$dryad_root_build_deps_file")")
        dryad_join_path_load "$dryad_root_build_deps_file_dir" "$dryad_root_build_deps_target_rel"
        dryad_root_build_deps_target_path=$dyd_ret0
        dryad_root_path_find_load "$dryad_root_build_deps_target_path"
        dryad_root_build_deps_target_root=$dyd_ret0
        dryad_root_build_validate_target_selector "$dryad_root_build_deps_garden" "$dryad_root_build_deps_root" "$dryad_root_build_deps_target_root" "$dryad_root_build_deps_selector" "$dryad_root_build_deps_descriptor"
        dryad_root_build_deps_matches_file=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-deps.XXXXXX")
        dryad_root_build_selected_descriptors "$dryad_root_build_deps_target_root" "" |
            while IFS= read -r dryad_root_build_deps_candidate; do
                if dryad_root_build_target_selector_matches_descriptor "$dryad_root_build_deps_selector" "$dryad_root_build_deps_candidate" "$dryad_root_build_deps_descriptor"; then
                    printf '%s\n' "$dryad_root_build_deps_candidate"
                fi
            done > "$dryad_root_build_deps_matches_file"

        dryad_root_build_deps_count=$(wc -l < "$dryad_root_build_deps_matches_file" | tr -d ' ')
        [ "$dryad_root_build_deps_count" -gt 0 ] ||
            dryad_die "root requirement target has no matching variants: $dryad_root_build_deps_file"

        while IFS= read -r dryad_root_build_deps_target_descriptor; do
            dryad_root_build_stem_load "$dryad_root_build_deps_garden" "$dryad_root_build_deps_target_root" "$dryad_root_build_deps_target_descriptor"
            dryad_root_build_deps_fingerprint=$dyd_ret0
            dryad_memo_get_line_load root-build-source-fingerprint "$dryad_root_build_deps_garden" "$dryad_root_build_deps_target_root" "$dryad_root_build_deps_target_descriptor" ||
                dryad_die "missing source fingerprint for root dependency: $dryad_root_build_deps_target_root"
            dryad_root_build_deps_source_fingerprint=$dyd_ret0
            dryad_root_build_deps_dep_name=$dryad_root_build_deps_alias
            if dryad_root_build_descriptor_needs_suffix "$dryad_root_build_deps_selector" "$dryad_root_build_deps_count" &&
                [ -n "$dryad_root_build_deps_target_descriptor" ]; then
                dryad_root_build_deps_dep_name=$dryad_root_build_deps_alias~$dryad_root_build_deps_target_descriptor
            fi
            dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_deps_garden" stems "$dryad_root_build_deps_fingerprint"
            dryad_root_build_deps_heap_path=$dyd_ret0
            ln -s "$dryad_root_build_deps_heap_path" "$dryad_root_build_deps_workspace/dyd/dependencies/$dryad_root_build_deps_dep_name"
            if [ -n "$dryad_root_build_deps_sources_out" ]; then
                printf '%s\n' "$dryad_root_build_deps_source_fingerprint" >> "$dryad_root_build_deps_sources_out"
            fi
        done < "$dryad_root_build_deps_matches_file"
        rm -f "$dryad_root_build_deps_matches_file"
    done
}

dryad_root_build_prepare_path () {
    dryad_root_build_path_workspace=$1
    dryad_root_build_path_dir=$dryad_root_build_path_workspace/dyd/path

    rm -rf "$dryad_root_build_path_dir"
    mkdir -p "$dryad_root_build_path_dir"
    [ -d "$dryad_root_build_path_workspace/dyd/dependencies" ] || return 0

    for dryad_root_build_path_dep in "$dryad_root_build_path_workspace"/dyd/dependencies/*; do
        [ -e "$dryad_root_build_path_dep" ] || [ -L "$dryad_root_build_path_dep" ] || continue
        dryad_root_build_path_dep_name=${dryad_root_build_path_dep##*/}
        [ -d "$dryad_root_build_path_dep/dyd/commands" ] || continue
        for dryad_root_build_path_command in "$dryad_root_build_path_dep"/dyd/commands/*; do
            [ -f "$dryad_root_build_path_command" ] || continue
            dryad_root_build_path_command_name=${dryad_root_build_path_command##*/}
            case $dryad_root_build_path_command_name in
                dyd-stem-run | default )
                    dryad_root_build_path_stub_name=$dryad_root_build_path_dep_name
                    ;;
                * )
                    dryad_root_build_path_stub_name=$dryad_root_build_path_dep_name--$dryad_root_build_path_command_name
                    ;;
            esac
            {
                printf '%s\n' '#!/bin/sh'
                printf '%s\n' 'set -eu'
                printf '%s\n' 'export DYD_STEM="$(dirname "$0")/../dependencies/'"$dryad_root_build_path_dep_name"'"'
                printf '%s\n' 'export PATH="$DYD_STEM/dyd/path:$PATH"'
                printf '%s\n' 'exec "$DYD_STEM/dyd/commands/'"$dryad_root_build_path_command_name"'" "$@"'
            } > "$dryad_root_build_path_dir/$dryad_root_build_path_stub_name"
            chmod 755 "$dryad_root_build_path_dir/$dryad_root_build_path_stub_name"
        done
    done
}

dryad_root_build_prepare_built_requirements () {
    dryad_root_build_reqs_stem=$1
    dryad_root_build_reqs_dir=$dryad_root_build_reqs_stem/dyd/requirements
    dryad_root_build_reqs_deps=$dryad_root_build_reqs_stem/dyd/dependencies
    dryad_profile_count call.root-build.prepare-built-requirements

    rm -rf "$dryad_root_build_reqs_dir"
    mkdir -p "$dryad_root_build_reqs_dir"
    [ -d "$dryad_root_build_reqs_deps" ] || return 0

    for dryad_root_build_reqs_dep in "$dryad_root_build_reqs_deps"/*; do
        [ -e "$dryad_root_build_reqs_dep" ] || [ -L "$dryad_root_build_reqs_dep" ] || continue
        dryad_root_build_reqs_name=${dryad_root_build_reqs_dep##*/}
        [ -f "$dryad_root_build_reqs_dep/dyd/fingerprint" ] ||
            dryad_die "dependency missing fingerprint: $dryad_root_build_reqs_dep"
        cp "$dryad_root_build_reqs_dep/dyd/fingerprint" "$dryad_root_build_reqs_dir/$dryad_root_build_reqs_name"
    done
}

dryad_root_build_prepare_built_dependencies () {
    dryad_root_build_built_deps_source=$1
    dryad_root_build_built_deps_stem=$2
    dryad_root_build_built_deps_source_dir=$dryad_root_build_built_deps_source/dyd/dependencies
    dryad_root_build_built_deps_dest_dir=$dryad_root_build_built_deps_stem/dyd/dependencies
    dryad_profile_count call.root-build.prepare-built-dependencies

    [ -d "$dryad_root_build_built_deps_source_dir" ] || return 0
    mkdir -p "$dryad_root_build_built_deps_dest_dir"

    for dryad_root_build_built_deps_source_link in "$dryad_root_build_built_deps_source_dir"/*; do
        [ -e "$dryad_root_build_built_deps_source_link" ] || [ -L "$dryad_root_build_built_deps_source_link" ] || continue
        dryad_root_build_built_deps_name=${dryad_root_build_built_deps_source_link##*/}
        dryad_root_build_built_deps_dest_link=$dryad_root_build_built_deps_dest_dir/$dryad_root_build_built_deps_name
        [ ! -e "$dryad_root_build_built_deps_dest_link" ] && [ ! -L "$dryad_root_build_built_deps_dest_link" ] || continue

        dryad_root_build_built_deps_target=$(dryad_clean_cd "$dryad_root_build_built_deps_source_link")
        [ -f "$dryad_root_build_built_deps_target/dyd/fingerprint" ] ||
            dryad_die "dependency missing fingerprint: $dryad_root_build_built_deps_source_link"
        ln -s "$dryad_root_build_built_deps_target" "$dryad_root_build_built_deps_dest_link"
    done
}

dryad_root_build_materialize_source_stem () {
    dryad_root_build_source_stem_src=$1
    dryad_root_build_source_stem_dest=$2

    dryad_root_build_init_stem "$dryad_root_build_source_stem_dest"

    for dryad_root_build_source_stem_kind in path assets secrets commands docs traits requirements; do
        if [ -d "$dryad_root_build_source_stem_src/dyd/$dryad_root_build_source_stem_kind" ] ||
            [ -L "$dryad_root_build_source_stem_src/dyd/$dryad_root_build_source_stem_kind" ]; then
            dryad_root_build_copy_dir_contents \
                "$dryad_root_build_source_stem_src/dyd/$dryad_root_build_source_stem_kind" \
                "$dryad_root_build_source_stem_dest/dyd/$dryad_root_build_source_stem_kind"
        fi
    done

    dryad_root_build_prepare_built_dependencies "$dryad_root_build_source_stem_src" "$dryad_root_build_source_stem_dest"
}

dryad_root_build_init_stem () {
    dryad_root_build_init_stem_path=$1

    mkdir -p "$dryad_root_build_init_stem_path/dyd/assets"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/commands"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/dependencies"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/docs"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/requirements"
    mkdir -p "$dryad_root_build_init_stem_path/dyd/traits"
}


dryad_root_build_file_hash () {
    dryad_root_build_hash_file=$1
    {
        printf 'file\000'
        cat "$dryad_root_build_hash_file"
    } | dryad_blake2b_128_stream_base32
}

dryad_root_build_link_hash_load () {
    dryad_root_build_hash_link=$1
    dryad_root_build_hash_link_target=$(readlink "$dryad_root_build_hash_link")
    dryad_root_build_hash_link_value=$({
        printf 'link\000'
        printf '%s' "$dryad_root_build_hash_link_target"
    } | dryad_blake2b_128_stream_base32)
    dyd_ret0=$dryad_root_build_hash_link_value
}

dryad_root_build_fingerprint () {
    dryad_root_build_fingerprint_path=$1
    dryad_root_build_fingerprint_file_hashes_out=${2:-}
    dryad_profile_count call.root-build.fingerprint
    dryad_root_build_fingerprint_payload=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint.XXXXXX")
    dryad_root_build_fingerprint_table=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint-table.XXXXXX")
    dryad_root_build_fingerprint_file_manifest=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint-files.XXXXXX")
    dryad_root_build_fingerprint_file_hashes=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint-file-hashes.XXXXXX")
    dryad_root_build_fingerprint_hashes=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-fingerprint-hashes.XXXXXX")

    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_start=$dyd_ret0
    (
        cd "$dryad_root_build_fingerprint_path" || exit 1
        find . -mindepth 1 -print | sort | while IFS= read -r dryad_root_build_fingerprint_entry; do
            dryad_root_build_fingerprint_rel=${dryad_root_build_fingerprint_entry#./}
            case $dryad_root_build_fingerprint_rel in
                dyd/path/* | dyd/assets/* | dyd/secrets/* | dyd/commands/* | dyd/docs/* | dyd/type | dyd/variants/* | dyd/traits/* | dyd/requirements/* )
                    ;;
                * )
                    continue
                    ;;
            esac
            if [ -L "$dryad_root_build_fingerprint_entry" ]; then
                dryad_root_build_link_hash_load "$dryad_root_build_fingerprint_entry"
                dryad_root_build_fingerprint_link_hash=$dyd_ret0
                printf '%s\t%s\n' "$dryad_root_build_fingerprint_rel" "$dryad_root_build_fingerprint_link_hash" >> "$dryad_root_build_fingerprint_hashes"
            elif [ -f "$dryad_root_build_fingerprint_entry" ]; then
                printf '%s\t%s\n' "$dryad_root_build_fingerprint_rel" "$dryad_root_build_fingerprint_entry" >> "$dryad_root_build_fingerprint_file_manifest"
            fi
        done
    )
    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_end=$dyd_ret0
    dryad_profile_time_record_bounds root-build.fingerprint.scan "$dryad_root_build_fingerprint_t_start" "$dryad_root_build_fingerprint_t_end"

    if [ -s "$dryad_root_build_fingerprint_file_manifest" ]; then
        dryad_profile_time_now_ns_load
        dryad_root_build_fingerprint_t_start=$dyd_ret0
        (
            cd "$dryad_root_build_fingerprint_path" || exit 1
            dryad_blake2b_128_files_table_base32 < "$dryad_root_build_fingerprint_file_manifest"
        ) > "$dryad_root_build_fingerprint_file_hashes"
        dryad_profile_time_now_ns_load
        dryad_root_build_fingerprint_t_end=$dyd_ret0
        dryad_profile_time_record_bounds root-build.fingerprint.batch-file-hash "$dryad_root_build_fingerprint_t_start" "$dryad_root_build_fingerprint_t_end"
        cat "$dryad_root_build_fingerprint_file_hashes" >> "$dryad_root_build_fingerprint_hashes"
    fi

    dryad_root_build_fingerprint_sep=$(printf '\t')
    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_start=$dyd_ret0
    sort "$dryad_root_build_fingerprint_hashes" |
        while IFS=$dryad_root_build_fingerprint_sep read -r dryad_root_build_fingerprint_rel dryad_root_build_fingerprint_hash; do
            [ -n "$dryad_root_build_fingerprint_hash" ] || continue
            printf '%s ./%s\n' "$dryad_root_build_fingerprint_hash" "$dryad_root_build_fingerprint_rel"
        done > "$dryad_root_build_fingerprint_table"
    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_end=$dyd_ret0
    dryad_profile_time_record_bounds root-build.fingerprint.sort-table "$dryad_root_build_fingerprint_t_start" "$dryad_root_build_fingerprint_t_end"

    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_start=$dyd_ret0
    printf 'stem\000' > "$dryad_root_build_fingerprint_payload"
    dryad_root_build_fingerprint_first=1
    while IFS= read -r dryad_root_build_fingerprint_line; do
        if [ "$dryad_root_build_fingerprint_first" = 1 ]; then
            dryad_root_build_fingerprint_first=0
        else
            printf '\000' >> "$dryad_root_build_fingerprint_payload"
        fi
            printf '%s' "$dryad_root_build_fingerprint_line" >> "$dryad_root_build_fingerprint_payload"
    done < "$dryad_root_build_fingerprint_table"
    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_end=$dyd_ret0
    dryad_profile_time_record_bounds root-build.fingerprint.payload "$dryad_root_build_fingerprint_t_start" "$dryad_root_build_fingerprint_t_end"

    if [ -n "$dryad_root_build_fingerprint_file_hashes_out" ]; then
        dryad_profile_time_now_ns_load
        dryad_root_build_fingerprint_t_start=$dyd_ret0
        sort "$dryad_root_build_fingerprint_file_hashes" |
            while IFS=$dryad_root_build_fingerprint_sep read -r dryad_root_build_fingerprint_file_rel dryad_root_build_fingerprint_file_hash; do
                [ -n "$dryad_root_build_fingerprint_file_hash" ] || continue
                printf '%s\tv2-%s\n' "$dryad_root_build_fingerprint_file_rel" "$dryad_root_build_fingerprint_file_hash"
            done > "$dryad_root_build_fingerprint_file_hashes_out"
        dryad_profile_time_now_ns_load
        dryad_root_build_fingerprint_t_end=$dyd_ret0
        dryad_profile_time_record_bounds root-build.fingerprint.export-file-hashes "$dryad_root_build_fingerprint_t_start" "$dryad_root_build_fingerprint_t_end"
    fi

    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_start=$dyd_ret0
    dryad_blake2b_128_file_fingerprint "$dryad_root_build_fingerprint_payload"
    dryad_profile_time_now_ns_load
    dryad_root_build_fingerprint_t_end=$dyd_ret0
    dryad_profile_time_record_bounds root-build.fingerprint.final-hash "$dryad_root_build_fingerprint_t_start" "$dryad_root_build_fingerprint_t_end"
    rm -f "$dryad_root_build_fingerprint_payload" "$dryad_root_build_fingerprint_table" "$dryad_root_build_fingerprint_file_manifest" "$dryad_root_build_fingerprint_file_hashes" "$dryad_root_build_fingerprint_hashes"
}

dryad_root_build_source_fingerprint () {
    dryad_root_build_source_fingerprint_path=$1
    dryad_root_build_source_fingerprint_file_hashes_out=${2:-}
    dryad_profile_count call.root-build.source-fingerprint
    printf '%s' stem > "$dryad_root_build_source_fingerprint_path/dyd/type"
    dryad_root_build_source_fingerprint=$(dryad_root_build_fingerprint "$dryad_root_build_source_fingerprint_path" "$dryad_root_build_source_fingerprint_file_hashes_out")
    printf '%s' "$dryad_root_build_source_fingerprint" > "$dryad_root_build_source_fingerprint_path/dyd/fingerprint"
    printf '%s\n' "$dryad_root_build_source_fingerprint"
}

dryad_root_build_heap_package_path () {
    dryad_root_build_heap_fingerprint_path_load "$1" "$2" "$3"
    printf '%s\n' "$dyd_ret0"
}

# Returns:
#   dyd_ret0 = heap depth
dryad_root_build_heap_depth_load () {
    dryad_root_build_heap_depth_garden=$1
    dryad_root_build_heap_depth_kind=$2

    if [ "${dryad_root_build_heap_depth_cache+x}" = x ]; then
        while IFS='	' read -r dryad_root_build_heap_depth_cache_garden dryad_root_build_heap_depth_cache_kind dryad_root_build_heap_depth_cache_value; do
            [ -n "$dryad_root_build_heap_depth_cache_garden" ] || continue
            if [ "$dryad_root_build_heap_depth_cache_garden" = "$dryad_root_build_heap_depth_garden" ] &&
                [ "$dryad_root_build_heap_depth_cache_kind" = "$dryad_root_build_heap_depth_kind" ]; then
                dyd_ret0=$dryad_root_build_heap_depth_cache_value
                return 0
            fi
        done <<EOF
$dryad_root_build_heap_depth_cache
EOF
    fi

    dryad_root_build_heap_depth_file=$dryad_root_build_heap_depth_garden/dyd/shed/heap/$dryad_root_build_heap_depth_kind/depth

    if dryad_memo_get_line_load heap-depth "$dryad_root_build_heap_depth_garden" "$dryad_root_build_heap_depth_kind"; then
        dryad_root_build_heap_depth_value=$dyd_ret0
    elif [ ! -f "$dryad_root_build_heap_depth_file" ]; then
        dryad_memo_put_value heap-depth 1 "$dryad_root_build_heap_depth_garden" "$dryad_root_build_heap_depth_kind"
        dryad_root_build_heap_depth_value=1
    else
        dryad_root_build_heap_depth_value=$(tr -d '[:space:]' < "$dryad_root_build_heap_depth_file")
        case $dryad_root_build_heap_depth_value in
            '' | *[!0-9]* )
                dryad_die "invalid shed heap depth in $dryad_root_build_heap_depth_file"
                ;;
        esac
        dryad_memo_put_value heap-depth "$dryad_root_build_heap_depth_value" "$dryad_root_build_heap_depth_garden" "$dryad_root_build_heap_depth_kind"
    fi

    dryad_root_build_heap_depth_cache="${dryad_root_build_heap_depth_cache:-}${dryad_root_build_heap_depth_garden}	${dryad_root_build_heap_depth_kind}	${dryad_root_build_heap_depth_value}
"
    dyd_ret0=$dryad_root_build_heap_depth_value
}

dryad_root_build_heap_depth () {
    dryad_root_build_heap_depth_load "$1" "$2"
    printf '%s\n' "$dyd_ret0"
}

# Returns:
#   dyd_ret0 = heap path for fingerprint
dryad_root_build_heap_fingerprint_path_load () {
    dryad_root_build_heap_garden=$1
    dryad_root_build_heap_kind=$2
    dryad_root_build_heap_fingerprint=$3
    dryad_root_build_heap_encoded=${dryad_root_build_heap_fingerprint#v2-}

    case $dryad_root_build_heap_encoded in
        "$dryad_root_build_heap_fingerprint" )
            dryad_die "invalid heap fingerprint: $dryad_root_build_heap_fingerprint"
            ;;
    esac

    dryad_root_build_heap_depth_load "$dryad_root_build_heap_garden" "$dryad_root_build_heap_kind"
    dryad_root_build_heap_depth_value=$dyd_ret0
    dryad_root_build_heap_remaining=$dryad_root_build_heap_encoded
    dryad_root_build_heap_path=$dryad_root_build_heap_garden/dyd/heap/$dryad_root_build_heap_kind/v2
    dryad_root_build_heap_i=0
    while [ "$dryad_root_build_heap_i" -lt "$dryad_root_build_heap_depth_value" ]; do
        case $dryad_root_build_heap_remaining in
            ??* )
                ;;
            * )
                dryad_die "invalid shed heap depth $dryad_root_build_heap_depth_value for fingerprint $dryad_root_build_heap_fingerprint"
                ;;
        esac
        dryad_root_build_heap_segment=${dryad_root_build_heap_remaining%"${dryad_root_build_heap_remaining#??}"}
        dryad_root_build_heap_remaining=${dryad_root_build_heap_remaining#??}
        dryad_root_build_heap_path=$dryad_root_build_heap_path/$dryad_root_build_heap_segment
        dryad_root_build_heap_i=$((dryad_root_build_heap_i + 1))
    done
    dyd_ret0=$dryad_root_build_heap_path/$dryad_root_build_heap_remaining
}

dryad_root_build_heap_fingerprint_path () {
    dryad_root_build_heap_fingerprint_path_load "$1" "$2" "$3"
    printf '%s\n' "$dyd_ret0"
}

dryad_root_build_file_fingerprint () {
    dryad_root_build_file_fp_file=$1
    {
        printf 'file\000'
        cat "$dryad_root_build_file_fp_file"
    } | {
        dryad_root_build_file_fp_base32=$(dryad_blake2b_128_stream_base32)
        printf 'v2-%s\n' "$dryad_root_build_file_fp_base32"
    }
}

dryad_root_build_heap_ensure_file_load () {
    dryad_root_build_add_file_garden=$1
    dryad_root_build_add_file_kind=$2
    dryad_root_build_add_file_src=$3
    dryad_root_build_add_file_fingerprint=$4
    dryad_profile_count call.root-build.heap-add-file
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_add_file_garden" "$dryad_root_build_add_file_kind" "$dryad_root_build_add_file_fingerprint"
    dryad_root_build_add_file_dest=$dyd_ret0

    if [ -f "$dryad_root_build_add_file_dest" ]; then
        dryad_profile_count hit.root-build.heap-add-file
        dyd_ret0=$dryad_root_build_add_file_fingerprint
        dyd_ret1=$dryad_root_build_add_file_dest
        return 0
    fi

    dryad_profile_count miss.root-build.heap-add-file
    dryad_root_build_add_file_dest_dir=${dryad_root_build_add_file_dest%/*}
    dryad_root_build_add_file_dest_base=${dryad_root_build_add_file_dest##*/}
    mkdir -p "$dryad_root_build_add_file_dest_dir"
    dryad_root_build_add_file_tmp=$dryad_root_build_add_file_dest_dir/.tmp-$dryad_root_build_add_file_dest_base.$$
    rm -f "$dryad_root_build_add_file_tmp"
    cp "$dryad_root_build_add_file_src" "$dryad_root_build_add_file_tmp"
    chmod 511 "$dryad_root_build_add_file_tmp"
    if ! ln "$dryad_root_build_add_file_tmp" "$dryad_root_build_add_file_dest" 2>/dev/null; then
        if [ ! -f "$dryad_root_build_add_file_dest" ]; then
            rm -f "$dryad_root_build_add_file_tmp"
            dryad_die "could not publish heap file: $dryad_root_build_add_file_dest"
        fi
    fi
    rm -f "$dryad_root_build_add_file_tmp"
    dyd_ret0=$dryad_root_build_add_file_fingerprint
    dyd_ret1=$dryad_root_build_add_file_dest
}

dryad_root_build_heap_add_file_load () {
    dryad_root_build_add_file_garden=$1
    dryad_root_build_add_file_kind=$2
    dryad_root_build_add_file_src=$3
    dryad_root_build_add_file_fingerprint=$(dryad_root_build_file_fingerprint "$dryad_root_build_add_file_src")
    dryad_root_build_heap_ensure_file_load \
        "$dryad_root_build_add_file_garden" "$dryad_root_build_add_file_kind" "$dryad_root_build_add_file_src" "$dryad_root_build_add_file_fingerprint"
}

dryad_root_build_publish_should_include () {
    case $1 in
        dyd | dyd/path | dyd/path/* | dyd/assets | dyd/assets/* | dyd/secrets | dyd/secrets/* | dyd/commands | dyd/commands/* | dyd/docs | dyd/docs/* | dyd/type | dyd/fingerprint | dyd/requirements | dyd/requirements/* | dyd/variants | dyd/variants/* | dyd/dependencies | dyd/traits | dyd/traits/* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_root_build_link_is_internal () {
    dryad_root_build_link_base=$1
    dryad_root_build_link_rel=$2
    dryad_root_build_link_target=$3
    dryad_root_build_link_base_abs=$(dryad_clean_cd "$dryad_root_build_link_base")

    case $dryad_root_build_link_target in
        /* )
            dryad_root_build_link_target_abs=$dryad_root_build_link_target
            ;;
        * )
            dryad_root_build_link_target_abs=$dryad_root_build_link_base_abs/$(dirname "$dryad_root_build_link_rel")/$dryad_root_build_link_target
            ;;
    esac

    dryad_root_build_link_target_dir=$(dirname "$dryad_root_build_link_target_abs")
    if [ -d "$dryad_root_build_link_target_dir" ]; then
        dryad_file_abs_path_load "$dryad_root_build_link_target_abs"
        dryad_root_build_link_target_abs=$dyd_ret0
    fi

    case $dryad_root_build_link_target_abs in
        "$dryad_root_build_link_base_abs" | "$dryad_root_build_link_base_abs"/* )
            return 0
            ;;
        * )
            return 1
            ;;
    esac
}

dryad_root_build_mkdir_parent () {
    dryad_root_build_mkdir_parent_path=$1
    dryad_root_build_mkdir_parent_dir=${dryad_root_build_mkdir_parent_path%/*}
    [ "$dryad_root_build_mkdir_parent_dir" != "$dryad_root_build_mkdir_parent_path" ] || dryad_root_build_mkdir_parent_dir=.
    mkdir -p "$dryad_root_build_mkdir_parent_dir"
}

dryad_root_build_publish_symlink () {
    dryad_root_build_publish_symlink_target=$1
    dryad_root_build_publish_symlink_dest=$2

    ln -s "$dryad_root_build_publish_symlink_target" "$dryad_root_build_publish_symlink_dest" 2>/dev/null || {
        dryad_root_build_mkdir_parent "$dryad_root_build_publish_symlink_dest"
        ln -s "$dryad_root_build_publish_symlink_target" "$dryad_root_build_publish_symlink_dest"
    }
}

dryad_root_build_publish_file_link () {
    dryad_root_build_publish_file_link_src=$1
    dryad_root_build_publish_file_link_dest=$2

    ln "$dryad_root_build_publish_file_link_src" "$dryad_root_build_publish_file_link_dest" 2>/dev/null || {
        dryad_root_build_mkdir_parent "$dryad_root_build_publish_file_link_dest"
        ln "$dryad_root_build_publish_file_link_src" "$dryad_root_build_publish_file_link_dest"
    }
}

dryad_root_build_publish_dir_mkdir () {
    dryad_root_build_publish_dir_mkdir_dest=$1

    mkdir "$dryad_root_build_publish_dir_mkdir_dest" 2>/dev/null || {
        [ -d "$dryad_root_build_publish_dir_mkdir_dest" ] || mkdir -p "$dryad_root_build_publish_dir_mkdir_dest"
    }
}

dryad_root_build_publish_tree_process_entries () {
    while IFS= read -r dryad_root_build_publish_tree_entry; do
        dryad_root_build_publish_tree_rel=${dryad_root_build_publish_tree_entry#./}
        [ "$dryad_root_build_publish_tree_rel" != . ] || continue
        dryad_root_build_publish_should_include "$dryad_root_build_publish_tree_rel" || continue

        dryad_root_build_publish_tree_dest=$dryad_root_build_publish_tree_tmp/$dryad_root_build_publish_tree_rel
        if [ -L "$dryad_root_build_publish_tree_entry" ]; then
            dryad_profile_count root-build.publish-tree.symlink
            dryad_root_build_publish_tree_target=$(dryad_profile_time_block root-build.publish-tree.symlink.readlink \
                readlink "$dryad_root_build_publish_tree_entry")
            if dryad_profile_time_block root-build.publish-tree.symlink.internal-check \
                dryad_root_build_link_is_internal "$dryad_root_build_publish_tree_src" "$dryad_root_build_publish_tree_rel" "$dryad_root_build_publish_tree_target"; then
                dryad_profile_time_block root-build.publish-tree.symlink.link \
                    dryad_root_build_publish_symlink "$dryad_root_build_publish_tree_target" "$dryad_root_build_publish_tree_dest"
            fi
        elif [ -d "$dryad_root_build_publish_tree_entry" ]; then
            dryad_profile_count root-build.publish-tree.dir
            dryad_profile_time_block root-build.publish-tree.dir.mkdir \
                dryad_root_build_publish_dir_mkdir "$dryad_root_build_publish_tree_dest"
        elif [ -f "$dryad_root_build_publish_tree_entry" ]; then
            dryad_profile_count root-build.publish-tree.file
            dryad_root_build_publish_tree_file_kind=files
            case $dryad_root_build_publish_tree_kind:$dryad_root_build_publish_tree_rel in
                stem:dyd/secrets/* )
                    dryad_root_build_publish_tree_file_kind=secrets
                    ;;
            esac
            dryad_root_build_publish_tree_file_fp=''
            if [ "$dryad_root_build_publish_tree_hashes_enabled" = 1 ]; then
                if [ "$dryad_root_build_publish_tree_hashes_loaded" = 0 ] && [ "$dryad_root_build_publish_tree_hashes_done" = 0 ]; then
                    if IFS="$(printf '\t')" read -r dryad_root_build_publish_tree_hash_rel dryad_root_build_publish_tree_hash_fp <&3; then
                        dryad_root_build_publish_tree_hashes_loaded=1
                    else
                        dryad_root_build_publish_tree_hashes_done=1
                    fi
                fi
                if [ "$dryad_root_build_publish_tree_hashes_loaded" = 1 ] &&
                    [ "$dryad_root_build_publish_tree_hash_rel" = "$dryad_root_build_publish_tree_rel" ]; then
                    dryad_profile_count root-build.publish-tree.prehashed-file
                    dryad_profile_time_block root-build.publish-tree.heap-ensure \
                        dryad_root_build_heap_ensure_file_load \
                        "$dryad_root_build_publish_tree_garden" \
                        "$dryad_root_build_publish_tree_file_kind" \
                        "$dryad_root_build_publish_tree_entry" \
                        "$dryad_root_build_publish_tree_hash_fp"
                    dryad_root_build_publish_tree_file_fp=$dyd_ret0
                    dryad_root_build_publish_tree_file_heap=$dyd_ret1
                    dryad_root_build_publish_tree_hashes_loaded=0
                fi
            fi
            if [ -z "$dryad_root_build_publish_tree_file_fp" ]; then
                dryad_profile_count root-build.publish-tree.hash-file
                dryad_profile_time_block root-build.publish-tree.heap-add \
                    dryad_root_build_heap_add_file_load \
                    "$dryad_root_build_publish_tree_garden" "$dryad_root_build_publish_tree_file_kind" "$dryad_root_build_publish_tree_entry"
                dryad_root_build_publish_tree_file_fp=$dyd_ret0
                dryad_root_build_publish_tree_file_heap=$dyd_ret1
            fi
            dryad_profile_time_block root-build.publish-tree.file.link \
                dryad_root_build_publish_file_link "$dryad_root_build_publish_tree_file_heap" "$dryad_root_build_publish_tree_dest"
        fi
    done
}

dryad_root_build_publish_tree () {
    dryad_root_build_publish_tree_garden=$1
    dryad_root_build_publish_tree_kind=$2
    dryad_root_build_publish_tree_src=$3
    dryad_root_build_publish_tree_tmp=$4
    dryad_root_build_publish_tree_file_hashes=${5:-}

    (
        cd "$dryad_root_build_publish_tree_src" || exit 1
        dryad_root_build_publish_tree_hashes_enabled=0
        dryad_root_build_publish_tree_hashes_loaded=0
        dryad_root_build_publish_tree_hashes_done=0
        if [ -n "$dryad_root_build_publish_tree_file_hashes" ] && [ -f "$dryad_root_build_publish_tree_file_hashes" ]; then
            dryad_root_build_publish_tree_hashes_enabled=1
            exec 3< "$dryad_root_build_publish_tree_file_hashes"
        fi
        if [ -n "${DRYAD_SH_PROFILE_FILE:-}" ]; then
            dryad_root_build_publish_tree_entries=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-publish-tree.XXXXXX")
            dryad_profile_time_now_ns_load
            dryad_root_build_publish_tree_t_start=$dyd_ret0
            find . -print | sort > "$dryad_root_build_publish_tree_entries"
            dryad_profile_time_now_ns_load
            dryad_root_build_publish_tree_t_end=$dyd_ret0
            dryad_profile_time_record_bounds root-build.publish-tree.scan-sort "$dryad_root_build_publish_tree_t_start" "$dryad_root_build_publish_tree_t_end"

            dryad_profile_time_now_ns_load
            dryad_root_build_publish_tree_t_start=$dyd_ret0
            dryad_root_build_publish_tree_process_entries < "$dryad_root_build_publish_tree_entries"
            dryad_profile_time_now_ns_load
            dryad_root_build_publish_tree_t_end=$dyd_ret0
            dryad_profile_time_record_bounds root-build.publish-tree.process "$dryad_root_build_publish_tree_t_start" "$dryad_root_build_publish_tree_t_end"
            rm -f "$dryad_root_build_publish_tree_entries"
        else
            find . -print | sort | dryad_root_build_publish_tree_process_entries
        fi
        if [ "$dryad_root_build_publish_tree_hashes_enabled" = 1 ]; then
            exec 3<&-
        fi
    )
}

dryad_root_build_publish_dependency_links () {
    dryad_root_build_publish_deps_garden=$1
    dryad_root_build_publish_deps_src=$2
    dryad_root_build_publish_deps_tmp=$3
    dryad_root_build_publish_deps_src_dir=$dryad_root_build_publish_deps_src/dyd/dependencies
    dryad_root_build_publish_deps_tmp_dir=$dryad_root_build_publish_deps_tmp/dyd/dependencies

    mkdir -p "$dryad_root_build_publish_deps_tmp_dir"
    [ -d "$dryad_root_build_publish_deps_src_dir" ] || return 0

    for dryad_root_build_publish_deps_link in "$dryad_root_build_publish_deps_src_dir"/*; do
        [ -e "$dryad_root_build_publish_deps_link" ] || [ -L "$dryad_root_build_publish_deps_link" ] || continue
        dryad_root_build_publish_deps_name=${dryad_root_build_publish_deps_link##*/}
        dryad_root_build_publish_deps_fingerprint_file=$dryad_root_build_publish_deps_link/dyd/fingerprint
        [ -f "$dryad_root_build_publish_deps_fingerprint_file" ] ||
            dryad_die "dependency missing fingerprint: $dryad_root_build_publish_deps_link"
        dryad_root_build_publish_deps_fingerprint=
        IFS= read -r dryad_root_build_publish_deps_fingerprint < "$dryad_root_build_publish_deps_fingerprint_file" || [ -n "$dryad_root_build_publish_deps_fingerprint" ] || dryad_root_build_publish_deps_fingerprint=
        dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_publish_deps_garden" stems "$dryad_root_build_publish_deps_fingerprint"
        dryad_root_build_publish_deps_heap_path=$dyd_ret0
        dryad_relative_path_literal_load "$dryad_root_build_publish_deps_tmp_dir" "$dryad_root_build_publish_deps_heap_path"
        dryad_root_build_publish_deps_rel=$dyd_ret0
        ln -s "$dryad_root_build_publish_deps_rel" "$dryad_root_build_publish_deps_tmp_dir/$dryad_root_build_publish_deps_name"
    done
}

dryad_root_build_publish_dir () {
    dryad_root_build_publish_garden=$1
    dryad_root_build_publish_kind=$2
    dryad_root_build_publish_src=$3
    dryad_root_build_publish_dest=$4
    dryad_root_build_publish_file_hashes=${5:-}
    dryad_profile_count call.root-build.publish-dir

    if [ -e "$dryad_root_build_publish_dest" ] || [ -L "$dryad_root_build_publish_dest" ]; then
        return 0
    fi

    dryad_root_build_publish_parent=${dryad_root_build_publish_dest%/*}
    dryad_root_build_publish_name=${dryad_root_build_publish_dest##*/}
    mkdir -p "$dryad_root_build_publish_parent"
    dryad_root_build_publish_tmp=$dryad_root_build_publish_parent/.tmp-$dryad_root_build_publish_name.$$
    rm -rf "$dryad_root_build_publish_tmp"
    mkdir -p "$dryad_root_build_publish_tmp"
    dryad_profile_time_block root-build.publish-dir.tree \
        dryad_root_build_publish_tree "$dryad_root_build_publish_garden" "$dryad_root_build_publish_kind" "$dryad_root_build_publish_src" "$dryad_root_build_publish_tmp" "$dryad_root_build_publish_file_hashes"
    dryad_profile_time_block root-build.publish-dir.dependency-links \
        dryad_root_build_publish_dependency_links "$dryad_root_build_publish_garden" "$dryad_root_build_publish_src" "$dryad_root_build_publish_tmp"
    if ! dryad_profile_time_block root-build.publish-dir.move \
        mv "$dryad_root_build_publish_tmp" "$dryad_root_build_publish_dest" 2>/dev/null; then
        rm -rf "$dryad_root_build_publish_tmp"
        [ -e "$dryad_root_build_publish_dest" ] || return 1
    fi
    dryad_profile_time_block root-build.publish-dir.final-chmod \
        find "$dryad_root_build_publish_dest" -type d -exec chmod 511 {} +
}

dryad_root_build_publish_derivation () {
    dryad_root_build_derivation_garden=$1
    dryad_root_build_derivation_source_fingerprint=$2
    dryad_root_build_derivation_result_fingerprint=$3
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_derivation_garden" derivations/roots "$dryad_root_build_derivation_source_fingerprint"
    dryad_root_build_derivation_dest=$dyd_ret0

    dryad_root_build_derivation_dir=${dryad_root_build_derivation_dest%/*}
    dryad_root_build_derivation_name=${dryad_root_build_derivation_dest##*/}
    mkdir -p "$dryad_root_build_derivation_dir"
    dryad_root_build_derivation_tmp=$dryad_root_build_derivation_dir/.tmp-$dryad_root_build_derivation_name.$$
    rm -f "$dryad_root_build_derivation_tmp"
    printf '%s' "$dryad_root_build_derivation_result_fingerprint" > "$dryad_root_build_derivation_tmp"
    if ! mv "$dryad_root_build_derivation_tmp" "$dryad_root_build_derivation_dest" 2>/dev/null; then
        rm -f "$dryad_root_build_derivation_tmp"
        [ -f "$dryad_root_build_derivation_dest" ] ||
            dryad_die "could not publish root derivation: $dryad_root_build_derivation_dest"
    fi
}

dryad_root_build_provenance_node_store () {
    dryad_root_build_provenance_source_fingerprint=$1
    dryad_root_build_provenance_result_fingerprint=$2
    dryad_root_build_provenance_dependencies_file=${3:-}
    dryad_root_build_provenance_payload=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-provenance-node.XXXXXX")

    {
        printf 'result\t%s\n' "$dryad_root_build_provenance_result_fingerprint"
        if [ -n "$dryad_root_build_provenance_dependencies_file" ] && [ -f "$dryad_root_build_provenance_dependencies_file" ]; then
            while IFS= read -r dryad_root_build_provenance_dependency_source; do
                [ -n "$dryad_root_build_provenance_dependency_source" ] || continue
                printf 'dep\t%s\n' "$dryad_root_build_provenance_dependency_source"
            done < "$dryad_root_build_provenance_dependencies_file"
        fi
    } > "$dryad_root_build_provenance_payload"

    dryad_memo_put provenance-node "$dryad_root_build_provenance_source_fingerprint" < "$dryad_root_build_provenance_payload"
    rm -f "$dryad_root_build_provenance_payload"
}

dryad_root_build_provenance_collect_node () {
    dryad_root_build_provenance_collect_source=$1
    dryad_root_build_provenance_collect_sources_dir=$2
    dryad_root_build_provenance_collect_results_dir=$3

    [ -n "$dryad_root_build_provenance_collect_source" ] || return 0
    [ ! -e "$dryad_root_build_provenance_collect_sources_dir/$dryad_root_build_provenance_collect_source" ] || return 0
    : > "$dryad_root_build_provenance_collect_sources_dir/$dryad_root_build_provenance_collect_source"

    dryad_root_build_provenance_collect_node_data=$(dryad_memo_get provenance-node "$dryad_root_build_provenance_collect_source") ||
        dryad_die "missing provenance node for source fingerprint: $dryad_root_build_provenance_collect_source"

    dryad_root_build_provenance_collect_result=
    dryad_root_build_provenance_collect_tab=$(printf '\t')
    while IFS=$dryad_root_build_provenance_collect_tab read -r dryad_root_build_provenance_collect_kind dryad_root_build_provenance_collect_value; do
        case $dryad_root_build_provenance_collect_kind in
            result )
                dryad_root_build_provenance_collect_result=$dryad_root_build_provenance_collect_value
                ;;
            dep )
                dryad_root_build_provenance_collect_node \
                    "$dryad_root_build_provenance_collect_value" \
                    "$dryad_root_build_provenance_collect_sources_dir" \
                    "$dryad_root_build_provenance_collect_results_dir"
                ;;
        esac
    done <<EOF
$dryad_root_build_provenance_collect_node_data
EOF

    [ -n "$dryad_root_build_provenance_collect_result" ] ||
        dryad_die "missing provenance result fingerprint for source fingerprint: $dryad_root_build_provenance_collect_source"
    mkdir -p "$dryad_root_build_provenance_collect_results_dir/$dryad_root_build_provenance_collect_result"
    : > "$dryad_root_build_provenance_collect_results_dir/$dryad_root_build_provenance_collect_result/$dryad_root_build_provenance_collect_source"
}

dryad_root_build_provenance_stem_load () {
    dryad_root_build_provenance_garden=$1
    dryad_root_build_provenance_source_fingerprint=$2
    dryad_root_build_provenance_result_fingerprint=$3

    dryad_root_build_provenance_tmp=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-provenance.XXXXXX")
    dryad_root_build_provenance_sources_dir=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-provenance-sources.XXXXXX")
    dryad_root_build_provenance_results_dir=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-provenance-results.XXXXXX")

    dryad_root_build_init_stem "$dryad_root_build_provenance_tmp"
    printf '%s' provenance > "$dryad_root_build_provenance_tmp/dyd/traits/kind"
    printf '%s' "$dryad_root_build_provenance_result_fingerprint" > "$dryad_root_build_provenance_tmp/dyd/traits/result-fingerprint"

    dryad_root_build_provenance_collect_node \
        "$dryad_root_build_provenance_source_fingerprint" \
        "$dryad_root_build_provenance_sources_dir" \
        "$dryad_root_build_provenance_results_dir"

    for dryad_root_build_provenance_source_path in "$dryad_root_build_provenance_sources_dir"/*; do
        [ -e "$dryad_root_build_provenance_source_path" ] || continue
        dryad_root_build_provenance_source_name=${dryad_root_build_provenance_source_path##*/}
        dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_provenance_garden" stems "$dryad_root_build_provenance_source_name"
        dryad_root_build_provenance_source_heap_path=$dyd_ret0
        ln -s "$dryad_root_build_provenance_source_heap_path" \
            "$dryad_root_build_provenance_tmp/dyd/dependencies/$dryad_root_build_provenance_source_name"
    done

    for dryad_root_build_provenance_result_dir in "$dryad_root_build_provenance_results_dir"/*; do
        [ -d "$dryad_root_build_provenance_result_dir" ] || continue
        dryad_root_build_provenance_result_name=${dryad_root_build_provenance_result_dir##*/}
        mkdir -p "$dryad_root_build_provenance_tmp/dyd/assets/results/$dryad_root_build_provenance_result_name"
        for dryad_root_build_provenance_source_path in "$dryad_root_build_provenance_result_dir"/*; do
            [ -e "$dryad_root_build_provenance_source_path" ] || continue
            dryad_root_build_provenance_source_name=${dryad_root_build_provenance_source_path##*/}
            : > "$dryad_root_build_provenance_tmp/dyd/assets/results/$dryad_root_build_provenance_result_name/$dryad_root_build_provenance_source_name"
        done
    done

    dryad_root_build_prepare_built_requirements "$dryad_root_build_provenance_tmp"
    printf '%s' stem > "$dryad_root_build_provenance_tmp/dyd/type"
    dryad_root_build_provenance_file_hashes=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-provenance-file-hashes.XXXXXX")
    dryad_root_build_provenance_fingerprint=$(dryad_profile_time_block root-build.fingerprint.provenance \
        dryad_root_build_fingerprint "$dryad_root_build_provenance_tmp" "$dryad_root_build_provenance_file_hashes")
    printf '%s' "$dryad_root_build_provenance_fingerprint" > "$dryad_root_build_provenance_tmp/dyd/fingerprint"
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_provenance_garden" stems "$dryad_root_build_provenance_fingerprint"
    dryad_root_build_provenance_heap_path=$dyd_ret0
    dryad_profile_time_block root-build.publish-dir.provenance \
        dryad_root_build_publish_dir \
        "$dryad_root_build_provenance_garden" \
        stem \
        "$dryad_root_build_provenance_tmp" \
        "$dryad_root_build_provenance_heap_path" \
        "$dryad_root_build_provenance_file_hashes"

    rm -f "$dryad_root_build_provenance_file_hashes"
    rm -rf "$dryad_root_build_provenance_tmp" "$dryad_root_build_provenance_sources_dir" "$dryad_root_build_provenance_results_dir"
    dyd_ret0=$dryad_root_build_provenance_fingerprint
}

dryad_root_build_lookup_derivation_load () {
    dryad_root_build_lookup_garden=$1
    dryad_root_build_lookup_source_fingerprint=$2
    dryad_profile_count call.root-build.lookup-derivation
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_lookup_garden" derivations/roots "$dryad_root_build_lookup_source_fingerprint"
    dryad_root_build_lookup_path=$dyd_ret0

    [ -f "$dryad_root_build_lookup_path" ] || {
        dryad_profile_count miss.root-build.lookup-derivation
        return 1
    }

    IFS= read -r dryad_root_build_lookup_result < "$dryad_root_build_lookup_path" || [ -n "$dryad_root_build_lookup_result" ] || {
        dryad_profile_count miss.root-build.lookup-derivation
        return 1
    }
    case $dryad_root_build_lookup_result in
        v2-* )
            ;;
        * )
            dryad_profile_count miss.root-build.lookup-derivation
            return 1
            ;;
    esac

    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_lookup_garden" stems "$dryad_root_build_lookup_result"
    dryad_root_build_lookup_stem=$dyd_ret0
    [ -d "$dryad_root_build_lookup_stem" ] || {
        dryad_profile_count miss.root-build.lookup-derivation
        return 1
    }
    [ -f "$dryad_root_build_lookup_stem/dyd/fingerprint" ] || {
        dryad_profile_count miss.root-build.lookup-derivation
        return 1
    }
    IFS= read -r dryad_root_build_lookup_stem_fingerprint < "$dryad_root_build_lookup_stem/dyd/fingerprint" || [ -n "$dryad_root_build_lookup_stem_fingerprint" ] || {
        dryad_profile_count miss.root-build.lookup-derivation
        return 1
    }
    [ "$dryad_root_build_lookup_stem_fingerprint" = "$dryad_root_build_lookup_result" ] || {
        dryad_profile_count miss.root-build.lookup-derivation
        return 1
    }

    dryad_profile_count hit.root-build.lookup-derivation
    dyd_ret0=$dryad_root_build_lookup_result
}

dryad_root_build_ensure_sprout_parent_load () {
    dryad_root_build_parent_garden=$1
    dryad_root_build_parent_rel=$2
    dryad_root_build_parent_dir=${dryad_root_build_parent_rel%/*}
    [ "$dryad_root_build_parent_dir" != "$dryad_root_build_parent_rel" ] || dryad_root_build_parent_dir=.
    dryad_root_build_parent_current=$(dryad_sprouts_ensure_dir "$dryad_root_build_parent_garden")

    if [ "$dryad_root_build_parent_dir" = . ]; then
        dyd_ret0=$dryad_root_build_parent_current
        return 0
    fi

    dryad_root_build_parent_old_ifs=$IFS
    IFS=/
    set -- $dryad_root_build_parent_dir
    IFS=$dryad_root_build_parent_old_ifs

    for dryad_root_build_parent_segment do
        [ -n "$dryad_root_build_parent_segment" ] || continue
        dryad_root_build_parent_next=$dryad_root_build_parent_current/$dryad_root_build_parent_segment
        if [ -e "$dryad_root_build_parent_next" ] && [ ! -d "$dryad_root_build_parent_next" ]; then
            dryad_die "sprout parent path exists and is not a directory: $dryad_root_build_parent_next"
        fi
        if [ ! -d "$dryad_root_build_parent_next" ]; then
            dryad_root_build_parent_restore=0
            if ! dryad_path_has_owner_write "$dryad_root_build_parent_current"; then
                dryad_root_build_parent_restore=1
                chmod u+w "$dryad_root_build_parent_current"
            fi
            mkdir "$dryad_root_build_parent_next"
            chmod 551 "$dryad_root_build_parent_next"
            if [ "$dryad_root_build_parent_restore" = 1 ]; then
                chmod u-w "$dryad_root_build_parent_current"
            fi
        fi
        dryad_root_build_parent_current=$dryad_root_build_parent_next
    done

    dyd_ret0=$dryad_root_build_parent_current
}

dryad_root_build_run_command () {
    dryad_root_build_run_garden=$1
    dryad_root_build_run_rel=$2
    dryad_root_build_run_source=$3
    dryad_root_build_run_dest=$4
    dryad_root_build_run_join_stdout=$5
    dryad_root_build_run_join_stderr=$6
    dryad_root_build_run_log_stdout=$7
    dryad_root_build_run_log_stderr=$8
    dryad_profile_count call.root-build.run-command

    dryad_root_build_run_command=$dryad_root_build_run_source/dyd/commands/dyd-root-build
    [ -f "$dryad_root_build_run_command" ] ||
        dryad_die "missing root build command: $dryad_root_build_run_command"

    dryad_root_build_run_stdout=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-build.stdout.XXXXXX")
    dryad_root_build_run_stderr=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-build.stderr.XXXXXX")
    dryad_root_build_run_path=$dryad_root_build_run_source/dyd/commands:$dryad_root_build_run_source/dyd/path:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    dryad_host_os_load
    dryad_root_build_run_host_os=$dyd_ret0
    dryad_host_arch_load
    dryad_root_build_run_host_arch=$dyd_ret0
    case $dryad_root_build_run_host_os in
        darwin )
            dryad_root_build_run_path=$dryad_root_build_run_path:/opt/homebrew/bin:/opt/homebrew/sbin
            ;;
    esac

    set +e
    (
        cd "$dryad_root_build_run_source" || exit 1
        PATH=$dryad_root_build_run_path \
        DYD_STEM=$dryad_root_build_run_source \
        DYD_BUILD=$dryad_root_build_run_dest \
        DYD_GARDEN=$dryad_root_build_run_garden \
        DYD_OS=$dryad_root_build_run_host_os \
        DYD_ARCH=$dryad_root_build_run_host_arch \
        DYD_LOG_LEVEL=$dryad_log_level \
        "$dryad_root_build_run_command" "$dryad_root_build_run_dest"
    ) > "$dryad_root_build_run_stdout" 2> "$dryad_root_build_run_stderr"
    dryad_root_build_run_status=$?
    set -e

    if [ "$dryad_root_build_run_join_stdout" = 1 ]; then
        cat "$dryad_root_build_run_stdout" >&2
    fi
    if [ "$dryad_root_build_run_join_stderr" = 1 ]; then
        cat "$dryad_root_build_run_stderr" >&2
    fi
    if [ -n "$dryad_root_build_run_log_stdout" ]; then
        mkdir -p "$dryad_root_build_run_log_stdout"
        cp "$dryad_root_build_run_stdout" "$dryad_root_build_run_log_stdout/dyd-root-build--$(dryad_sprouts_run_sanitize_segment "$dryad_root_build_run_rel").out"
    fi
    if [ -n "$dryad_root_build_run_log_stderr" ]; then
        mkdir -p "$dryad_root_build_run_log_stderr"
        cp "$dryad_root_build_run_stderr" "$dryad_root_build_run_log_stderr/dyd-root-build--$(dryad_sprouts_run_sanitize_segment "$dryad_root_build_run_rel").err"
    fi

    rm -f "$dryad_root_build_run_stdout" "$dryad_root_build_run_stderr"
    return "$dryad_root_build_run_status"
}

dryad_root_build_stem_uncached () {
    dryad_root_build_stem_garden=$1
    dryad_root_build_stem_root=$2
    dryad_root_build_stem_descriptor=$3
    dryad_profile_count call.root-build.stem-uncached

    dryad_root_build_stem_rel=${dryad_root_build_stem_root#"$dryad_root_build_stem_garden"/dyd/roots/}
    dryad_root_build_stem_workspace=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-root.XXXXXX")
    dryad_root_build_stem_source=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-source-stem.XXXXXX")
    dryad_root_build_stem_dest=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-stem.XXXXXX")
    dryad_root_build_stem_dependency_sources=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-root-dep-sources.XXXXXX")

    dryad_root_build_log "root build - verifying root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"

    dryad_profile_time_block root-build.prepare-source \
        dryad_root_build_prepare_source "$dryad_root_build_stem_root" "$dryad_root_build_stem_descriptor" "$dryad_root_build_stem_workspace"
    dryad_profile_time_block root-build.prepare-dependencies \
        dryad_root_build_prepare_dependencies "$dryad_root_build_stem_garden" "$dryad_root_build_stem_root" "$dryad_root_build_stem_descriptor" "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_dependency_sources"
    dryad_profile_time_block root-build.prepare-built-requirements.workspace \
        dryad_root_build_prepare_built_requirements "$dryad_root_build_stem_workspace"
    dryad_profile_time_block root-build.prepare-path.workspace \
        dryad_root_build_prepare_path "$dryad_root_build_stem_workspace"
    dryad_root_build_materialize_source_stem "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_source"
    dryad_root_build_source_file_hashes=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-source-file-hashes.XXXXXX")
    dryad_root_build_stem_source_fingerprint=$(dryad_profile_time_block root-build.source-fingerprint \
        dryad_root_build_source_fingerprint "$dryad_root_build_stem_source" "$dryad_root_build_source_file_hashes")
    dryad_memo_put_value root-build-source-fingerprint "$dryad_root_build_stem_source_fingerprint" "$dryad_root_build_stem_garden" "$dryad_root_build_stem_root" "$dryad_root_build_stem_descriptor"
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_stem_garden" stems "$dryad_root_build_stem_source_fingerprint"
    dryad_root_build_stem_source_heap_path=$dyd_ret0
    dryad_profile_time_block root-build.publish-dir.source-stem \
        dryad_root_build_publish_dir \
        "$dryad_root_build_stem_garden" \
        stem \
        "$dryad_root_build_stem_source" \
        "$dryad_root_build_stem_source_heap_path" \
        "$dryad_root_build_source_file_hashes"

    if dryad_root_build_lookup_derivation_load "$dryad_root_build_stem_garden" "$dryad_root_build_stem_source_fingerprint"; then
        dryad_root_build_stem_cached_fingerprint=$dyd_ret0
        dryad_root_build_provenance_node_store "$dryad_root_build_stem_source_fingerprint" "$dryad_root_build_stem_cached_fingerprint" "$dryad_root_build_stem_dependency_sources"
        rm -f "$dryad_root_build_source_file_hashes"
        rm -f "$dryad_root_build_stem_dependency_sources"
        rm -rf "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_source" "$dryad_root_build_stem_dest"
        dryad_root_build_log "root build - done building root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"
        printf '%s\n' "$dryad_root_build_stem_cached_fingerprint"
        return 0
    fi

    dryad_root_build_init_stem "$dryad_root_build_stem_dest"
    dryad_root_build_log "root build - building root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"
    dryad_profile_time_block root-build.run-command \
        dryad_root_build_run_command \
        "$dryad_root_build_stem_garden" \
        "$dryad_root_build_stem_rel" \
        "$dryad_root_build_stem_workspace" \
        "$dryad_root_build_stem_dest" \
        "${dryad_root_build_join_stdout:-0}" \
        "${dryad_root_build_join_stderr:-0}" \
        "${dryad_root_build_log_stdout:-}" \
        "${dryad_root_build_log_stderr:-}" ||
        dryad_die "error executing root to build stem: dyd/roots/$dryad_root_build_stem_rel"

    dryad_profile_time_block root-build.prepare-built-dependencies \
        dryad_root_build_prepare_built_dependencies "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_dest"
    dryad_profile_time_block root-build.prepare-built-requirements.dest \
        dryad_root_build_prepare_built_requirements "$dryad_root_build_stem_dest"
    dryad_profile_time_block root-build.prepare-path.dest \
        dryad_root_build_prepare_path "$dryad_root_build_stem_dest"
    printf '%s' stem > "$dryad_root_build_stem_dest/dyd/type"
    dryad_root_build_stem_file_hashes=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-stem-file-hashes.XXXXXX")
    dryad_root_build_stem_fingerprint=$(dryad_profile_time_block root-build.fingerprint.stem \
        dryad_root_build_fingerprint "$dryad_root_build_stem_dest" "$dryad_root_build_stem_file_hashes")
    printf '%s' "$dryad_root_build_stem_fingerprint" > "$dryad_root_build_stem_dest/dyd/fingerprint"
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_stem_garden" stems "$dryad_root_build_stem_fingerprint"
    dryad_root_build_stem_heap_path=$dyd_ret0
    dryad_profile_time_block root-build.publish-dir.stem \
        dryad_root_build_publish_dir "$dryad_root_build_stem_garden" stem "$dryad_root_build_stem_dest" "$dryad_root_build_stem_heap_path" "$dryad_root_build_stem_file_hashes"
    dryad_root_build_publish_derivation "$dryad_root_build_stem_garden" "$dryad_root_build_stem_source_fingerprint" "$dryad_root_build_stem_fingerprint"
    dryad_root_build_provenance_node_store "$dryad_root_build_stem_source_fingerprint" "$dryad_root_build_stem_fingerprint" "$dryad_root_build_stem_dependency_sources"

    rm -f "$dryad_root_build_source_file_hashes"
    rm -f "$dryad_root_build_stem_dependency_sources"
    rm -f "$dryad_root_build_stem_file_hashes"
    rm -rf "$dryad_root_build_stem_workspace" "$dryad_root_build_stem_source" "$dryad_root_build_stem_dest"
    dryad_root_build_log "root build - done building root path=dyd/roots/$dryad_root_build_stem_rel variant=${dryad_root_build_stem_descriptor:-default}"
    printf '%s\n' "$dryad_root_build_stem_fingerprint"
}

dryad_root_build_stem_load () {
    dryad_root_build_memo_garden=$1
    dryad_root_build_memo_root=$2
    dryad_root_build_memo_descriptor=$3

    if dryad_memo_get_line_load root-build-stem "$dryad_root_build_memo_garden" "$dryad_root_build_memo_root" "$dryad_root_build_memo_descriptor"; then
        dryad_profile_count memo.hit.root-build-stem
        dryad_root_build_memo_fingerprint=$dyd_ret0
        dryad_root_build_memo_rel=${dryad_root_build_memo_root#"$dryad_root_build_memo_garden"/dyd/roots/}
        dryad_debug "root build memo hit path=dyd/roots/$dryad_root_build_memo_rel variant=${dryad_root_build_memo_descriptor:-default}"
        dyd_ret0=$dryad_root_build_memo_fingerprint
        return 0
    fi

    dryad_profile_count memo.miss.root-build-stem
    dryad_root_build_memo_fingerprint=$(dryad_root_build_stem_uncached "$dryad_root_build_memo_garden" "$dryad_root_build_memo_root" "$dryad_root_build_memo_descriptor") ||
        return $?
    dryad_memo_put_value root-build-stem "$dryad_root_build_memo_fingerprint" "$dryad_root_build_memo_garden" "$dryad_root_build_memo_root" "$dryad_root_build_memo_descriptor"
    dyd_ret0=$dryad_root_build_memo_fingerprint
}

dryad_root_build_materialize_sprout () {
    dryad_root_build_sprout_garden=$1
    dryad_root_build_sprout_root=$2
    dryad_root_build_sprout_descriptors_file=$3
    dryad_profile_count call.root-build.materialize-sprout
    dryad_root_build_sprout_tmp=$(mktemp -d "${TMPDIR:-/tmp}/dryad-sh-sprout.XXXXXX")
    dryad_root_build_sprout_rel=${dryad_root_build_sprout_root#"$dryad_root_build_sprout_garden"/dyd/roots/}

    mkdir -p "$dryad_root_build_sprout_tmp/dyd/dependencies"
    mkdir -p "$dryad_root_build_sprout_tmp/dyd/requirements"
    mkdir -p "$dryad_root_build_sprout_tmp/dyd/traits"

    if [ -d "$dryad_root_build_sprout_root/dyd/traits" ]; then
        dryad_root_build_copy_dir_contents "$dryad_root_build_sprout_root/dyd/traits" "$dryad_root_build_sprout_tmp/dyd/traits"
    fi

    while IFS= read -r dryad_root_build_sprout_descriptor; do
        dryad_root_build_stem_load "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_descriptor"
        dryad_root_build_sprout_stem_fingerprint=$dyd_ret0
        dryad_memo_get_line_load root-build-source-fingerprint "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_descriptor" ||
            dryad_die "missing source fingerprint for root build sprout variant: $dryad_root_build_sprout_root"
        dryad_root_build_sprout_source_fingerprint=$dyd_ret0
        dryad_root_build_provenance_stem_load \
            "$dryad_root_build_sprout_garden" \
            "$dryad_root_build_sprout_source_fingerprint" \
            "$dryad_root_build_sprout_stem_fingerprint"
        dryad_root_build_sprout_provenance_fingerprint=$dyd_ret0
        dryad_root_build_sprout_dep_name=stem
        dryad_root_build_sprout_provenance_name=stem.provenance
        if [ -n "$dryad_root_build_sprout_descriptor" ]; then
            dryad_root_build_sprout_dep_name=stem~$dryad_root_build_sprout_descriptor
            dryad_root_build_sprout_provenance_name=stem.provenance~$dryad_root_build_sprout_descriptor
        fi
        dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_sprout_garden" stems "$dryad_root_build_sprout_stem_fingerprint"
        dryad_root_build_sprout_stem_heap_path=$dyd_ret0
        dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_sprout_garden" stems "$dryad_root_build_sprout_provenance_fingerprint"
        dryad_root_build_sprout_provenance_heap_path=$dyd_ret0
        ln -s "$dryad_root_build_sprout_stem_heap_path" "$dryad_root_build_sprout_tmp/dyd/dependencies/$dryad_root_build_sprout_dep_name"
        ln -s "$dryad_root_build_sprout_provenance_heap_path" "$dryad_root_build_sprout_tmp/dyd/dependencies/$dryad_root_build_sprout_provenance_name"
    done < "$dryad_root_build_sprout_descriptors_file"

    dryad_root_build_prepare_built_requirements "$dryad_root_build_sprout_tmp"

    printf '%s' sprout > "$dryad_root_build_sprout_tmp/dyd/type"
    dryad_root_build_sprout_file_hashes=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-sprout-file-hashes.XXXXXX")
    dryad_root_build_sprout_fingerprint=$(dryad_profile_time_block root-build.fingerprint.sprout \
        dryad_root_build_fingerprint "$dryad_root_build_sprout_tmp" "$dryad_root_build_sprout_file_hashes")
    printf '%s' "$dryad_root_build_sprout_fingerprint" > "$dryad_root_build_sprout_tmp/dyd/fingerprint"
    dryad_root_build_heap_fingerprint_path_load "$dryad_root_build_sprout_garden" sprouts "$dryad_root_build_sprout_fingerprint"
    dryad_root_build_sprout_heap_path=$dyd_ret0
    dryad_profile_time_block root-build.publish-dir.sprout \
        dryad_root_build_publish_dir "$dryad_root_build_sprout_garden" sprout "$dryad_root_build_sprout_tmp" "$dryad_root_build_sprout_heap_path" "$dryad_root_build_sprout_file_hashes"

    dryad_root_build_ensure_sprout_parent_load "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_rel"
    dryad_root_build_sprout_link_parent=$dyd_ret0
    dryad_root_build_sprout_link=$dryad_root_build_sprout_link_parent/${dryad_root_build_sprout_rel##*/}
    dryad_root_build_sprout_restore_parent=0
    if [ -d "$dryad_root_build_sprout_link_parent" ] &&
        ! dryad_path_has_owner_write "$dryad_root_build_sprout_link_parent"; then
        dryad_root_build_sprout_restore_parent=1
        chmod u+w "$dryad_root_build_sprout_link_parent"
    fi
    if [ -e "$dryad_root_build_sprout_link" ] || [ -L "$dryad_root_build_sprout_link" ]; then
        dryad_sprouts_make_tree_removable "$dryad_root_build_sprout_link"
        rm -rf "$dryad_root_build_sprout_link"
    fi
    dryad_root_build_sprout_ln_status=0
    dryad_relative_path_literal_load "$dryad_root_build_sprout_link_parent" "$dryad_root_build_sprout_heap_path"
    dryad_root_build_sprout_target=$dyd_ret0
    ln -s "$dryad_root_build_sprout_target" "$dryad_root_build_sprout_link" ||
        dryad_root_build_sprout_ln_status=$?
    if [ "$dryad_root_build_sprout_restore_parent" = 1 ]; then
        chmod u-w "$dryad_root_build_sprout_link_parent"
    fi
    [ "$dryad_root_build_sprout_ln_status" = 0 ] ||
        return "$dryad_root_build_sprout_ln_status"

    rm -f "$dryad_root_build_sprout_file_hashes"
    rm -rf "$dryad_root_build_sprout_tmp"
    dryad_root_build_log "root build - done verifying root path=dyd/roots/$dryad_root_build_sprout_rel"
    printf '%s\n' "$dryad_root_build_sprout_fingerprint"
}

dryad_root_build_sprout () {
    dryad_root_build_sprout_root=$1
    dryad_root_build_sprout_selector=$2
    dryad_root_build_sprout_garden=$3
    dryad_root_build_sprout_descriptors=$(mktemp "${TMPDIR:-/tmp}/dryad-sh-sprout-descriptors.XXXXXX")
    dryad_root_build_selected_descriptors "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_selector" > "$dryad_root_build_sprout_descriptors"

    if [ ! -s "$dryad_root_build_sprout_descriptors" ] &&
        [ -d "$dryad_root_build_sprout_root/dyd/variants" ]; then
        rm -f "$dryad_root_build_sprout_descriptors"
        dryad_die "resolved root build variants are filtered by variants/_include and variants/_exclude"
    fi

    dryad_root_build_materialize_sprout "$dryad_root_build_sprout_garden" "$dryad_root_build_sprout_root" "$dryad_root_build_sprout_descriptors"
    rm -f "$dryad_root_build_sprout_descriptors"
}
