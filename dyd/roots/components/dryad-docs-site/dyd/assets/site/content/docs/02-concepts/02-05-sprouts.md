---
title: "02.05 - Sprouts"
description: "Output of a build process."
type: docs
layout: single
---

# Sprouts

A **sprout** is the top-level built package for a root.

Sprouts are content-addressed artifacts in `dyd/heap/sprouts`. The `dyd/sprouts` directory mirrors root paths and contains symlinks to those heap sprouts.

For example, if you have roots:
- `dyd/roots/tools/foo`
- `dyd/roots/tests/foo-tests`

after build you will get:
- `dyd/sprouts/tools/foo`
- `dyd/sprouts/tests/foo-tests`

Each sprout package contains metadata and dependency links to one or more built stem variants:
- `dyd/dependencies/stem` for a single default variant
- `dyd/dependencies/stem~<descriptor>` for explicit variants, such as `stem~arch=amd64+os=linux`

Sprouts are build artifacts. Do not edit them directly; rebuild roots to regenerate them.

## Running sprouts

`sprout run` and `sprouts run` execute selected concrete variants with a shared
`--parallel` limit (default 8). This also applies to targets supplied through
`sprouts run --from-stdin`. Use `--parallel=1` for serial execution.

### Stdin

`sprout run` and `sprouts run` accept an optional boolean `--join-stdin`:

- Omitted (`auto`): inherit stdin for a single selected execution or with `--parallel=1`; otherwise provide EOF.
- `false` (`none`): always provide EOF to children.
- `true` (`inherit`): provide parent stdin to children. Multiple selected executions require `--parallel=1`.

The execution count includes all selected sprouts and their selected variants. With
`sprouts run --from-stdin`, stdin supplies the target list: children always receive
EOF, and explicit `--join-stdin=true` is rejected.

`stem run` also accepts `--join-stdin`; its default is to inherit stdin because it
executes one stem. These flags control input only; `--inherit` controls environment variables.
