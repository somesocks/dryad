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

## Exclusions

Variant exclusions live under:

- `<root>/dyd/traits/variants/_exclude/<descriptor>`

Each exclusion file also contains `true` or `false`:

- `true` means exclusion is active
- `false` means exclusion is ignored

Exclusion filenames must be canonical filesystem descriptors, for example:

- `arch=amd64+os=darwin`

## Materialization

When dryad builds a concrete variant:

- selected variant options are materialized into concrete traits under `dyd/traits/`
- if a dimension is selected as `none`, that trait key is omitted in the concrete build

For quick inspection of resolved build variants for a root, use:

- `dryad root variants list <root_path>`
