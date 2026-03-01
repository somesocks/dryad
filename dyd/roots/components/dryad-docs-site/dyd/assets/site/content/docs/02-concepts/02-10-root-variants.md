---
title: "02.10 - Root variants"
description: "How root variants are defined and resolved."
type: docs
layout: single
---

# Root variants

Root variants let one root produce multiple concrete build outputs from variant dimensions such as `os`, `arch`, or project-specific traits.

## Filesystem layout

Variant configuration lives under:

- `<root>/dyd/traits/variants/<dimension>/<option>`

Each option file must contain exactly `true` or `false`:

- `true` means option is enabled
- `false` means option is disabled

Example:

```text
dyd/traits/variants/
  os/
    linux      # true
    darwin     # true
    none       # true
  arch/
    amd64      # true
    arm64      # false
  _include/
    arch=amd64+os=any     # true|false
  _exclude/
    arch=amd64+os=darwin  # true|false
```

In that example, enabled concrete variants include:

- `arch=amd64+os=linux`
- `arch=amd64+os=darwin`
- `arch=amd64` (because `os=none` omits the `os` trait)

## Name rules and reserved options

Variant dimension names and option names use:

- `[A-Za-z0-9._-]+`

Reserved option names in the dimension catalog:

- `inherit`
- `any`
- `host`

The `none` option is allowed and means "omit this trait dimension" when selected.

## Descriptor forms

Dryad uses two descriptor encodings:

- Filesystem form: `arch=amd64+os=linux`
- URL form: `?arch=amd64&os=linux`

Canonical order is ascending by dimension name, for example `arch=...` before `os=...`.

## Where selectors are used

Requirement filename (condition):

- `foo~arch=any+os=linux`

Requirement file value (target selector):

- `root:../../../foo?arch=inherit&os=host`

Condition and target selector are independent:

- condition controls **when** requirement is active
- target selector controls **which variant(s)** of the dependency are linked

Root content path selectors:

- `dyd/assets~<descriptor>`
- `dyd/commands~<descriptor>`
- `dyd/secrets~<descriptor>`
- `dyd/docs~<descriptor>`
- example: `dyd/assets~arch=amd64,arm64+os=linux`

## Selector keywords

`none`
- Selects the omitted option for a dimension.

`any`
- Expands to all enabled options for that dimension.

`inherit`
- For requirement target selectors, takes the option from the parent/root variant for the same dimension.
- If the parent omits that dimension, it resolves as `none`.

`host`
- Resolves to current host values for `os` and `arch`.

In requirement *conditions*:

- `any` and `inherit` act as wildcards.
- `none` matches when parent omits that dimension.
- `host` matches parent variant against host `os`/`arch`.

In root content path selectors (`dyd/assets~...`, `dyd/commands~...`, `dyd/secrets~...`, `dyd/docs~...`):

- supported: concrete options, `none`, `any`, and comma lists
- not supported: `inherit` and `host`
- omitted dimensions are implicit `any`; the plain path (for example `dyd/assets`) is the empty selector and therefore matches all variants
- selector descriptors must be canonical filesystem descriptors
- for each path kind, at most one selector may match:
  - no matches are allowed (the path is omitted from the build input)
  - multiple matches fail the build

## Exclusions

Variant exclusions live under:

- `<root>/dyd/traits/variants/_exclude/<descriptor>`

Each exclusion file also contains `true` or `false`:

- `true` means exclusion is active
- `false` means exclusion is ignored

Exclusion filenames must be canonical filesystem descriptors, for example:

- `arch=amd64+os=darwin`

Exclusion selectors must specify every enabled dimension. Unlike requirement selectors, exclusions are root-local and do not support parent/host context:

- supported in `_exclude`: concrete options, `none`, `any`, and comma lists (for example `os=darwin,linux`)
- not supported in `_exclude`: `inherit` and `host`

## Inclusions

Variant inclusions live under:

- `<root>/dyd/traits/variants/_include/<descriptor>`

Each inclusion file also contains `true` or `false`:

- `true` means inclusion is active
- `false` means inclusion is ignored

Inclusion filenames must be canonical filesystem descriptors and specify every enabled dimension.

Inclusion selectors use the same selector rules as exclusions:

- supported in `_include`: concrete options, `none`, `any`, and comma lists
- not supported in `_include`: `inherit` and `host`

Resolution behavior:

- candidate variants are generated from enabled dimension options
- if `_exclude` rules are enabled, matching variants are removed
- if `_include` rules are enabled, only matching variants are kept
- if the effective inclusion map is empty, all variants are treated as included
- final allow rule is: included and not excluded

## Materialization

When dryad builds a concrete variant:

- selected variant options are materialized into concrete traits under `dyd/traits/`
- if a dimension is selected as `none`, that trait key is omitted in the concrete build

For quick inspection of resolved build variants for a root, use:

- `dryad root variants list <root_path>`
