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
- `dyd/dependencies/stem+<descriptor>` for explicit variants, such as `stem+arch=amd64,os=linux`

Sprouts are build artifacts. Do not edit them directly; rebuild roots to regenerate them.
