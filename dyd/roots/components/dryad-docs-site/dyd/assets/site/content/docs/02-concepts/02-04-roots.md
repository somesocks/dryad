---
title: "02.04 - Roots"
description: "Source for a dryad package."
type: docs
layout: single
---

# Roots

A **root** is source code and build metadata for producing package artifacts. All roots in a garden are stored under `dyd/roots/`.

Each root includes directories such as:
- `dyd/assets` - source/input files
- `dyd/commands` - command scripts, including `dyd-root-build`
- `dyd/traits` - metadata files
- `dyd/variants` - variant dimension catalogs and include/exclude rules
- `dyd/requirements` - dependency declarations
- `dyd/docs`, `dyd/secrets`

Variant selectors are also supported for root content paths:
- `dyd/assets~<descriptor>`
- `dyd/commands~<descriptor>`
- `dyd/traits~<descriptor>`
- `dyd/secrets~<descriptor>`
- `dyd/docs~<descriptor>`
- `dyd/requirements~<descriptor>`

Example:
- `dyd/assets~arch=amd64,arm64+os=linux`

## Requirements

Each requirement has:
- A filename: `<alias>` or `<alias>~<condition_descriptor>`
  - Alias names may only contain `[A-Za-z0-9._-]+`.
- A file value: target URL, usually `root:<relative-path>` plus an optional variant selector query string, or `env:<name>` for a required host environment variable.

Examples:
- `dyd/requirements/foo` with `root:../../../foo`
- `dyd/requirements/foo~arch=any+os=linux` with `root:../../../foo?arch=amd64&os=linux`
- `dyd/requirements~os=linux/foo` with `root:../../../foo`
- `dyd/requirements/display` with `env:DISPLAY`

In practice:
- Filename descriptor (`~arch=...+os=...`) is a **condition**: when this requirement is active.
- URL query descriptor (`?arch=...&os=...`) is a **target selector**: which dependency variant(s) to link.
- Directory descriptor on `dyd/requirements~...` is a **path selector**: which requirements directory is active for the root variant.

For each variant, dryad selects at most one matching requirements directory. If there are no matching requirements directories, the root is built with no requirements. If there are multiple matching requirements directories, the build fails due to ambiguity.

Environment requirements are execution requirements. When a stem is executed, each `env:<name>` requirement reads a host environment variable and injects it into the process environment. The requirement alias is canonicalized to the injected environment variable name by uppercasing ASCII letters and converting `-` and `.` to `_`; the canonical name must match `[A-Z_][A-Z0-9_]*`. The `env:<name>` target is canonicalized the same way before reading the host environment.

For source roots, an `env:<name>` requirement is a build-time dependency because dryad executes the source stem to build the output stem. Dryad records a fingerprint of the host environment value in the materialized source stem requirement before checking the derivation cache. For built stems, an `env:<name>` requirement is a run-time dependency and the current host value is injected when the stem is run.

## Build flow

Root builds are variant-aware:

1. Resolve concrete variant(s) for the root.
2. For each concrete variant, create a source stem from the root plus resolved dependencies.
3. Execute `dyd/commands/dyd-root-build` in an isolated environment (`DYD_BUILD` points at output path) to produce the built stem.
4. Fingerprint and store built stems in `dyd/heap/stems`.
5. Aggregate built stem variant(s) into a sprout package in `dyd/heap/sprouts`.
6. Link the sprout package at `dyd/sprouts/<root_path>`.

Build caching is variant-scoped: different concrete variants are cached independently.

For full variant semantics and selector behavior, see [Root variants](../02-10-root-variants/).
