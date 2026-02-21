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
- `dyd/requirements` - dependency declarations
- `dyd/docs`, `dyd/secrets`

## Requirements

Each requirement has:
- A filename: `<alias>` or `<alias>+<condition_descriptor>`
- A file value: target URL, usually `root:<relative-path>` plus an optional variant selector query string.

Examples:
- `dyd/requirements/foo` with `root:../../../foo`
- `dyd/requirements/foo+arch=any,os=linux` with `root:../../../foo?arch=amd64&os=linux`

In practice:
- Filename descriptor (`+arch=...,os=...`) is a **condition**: when this requirement is active.
- URL query descriptor (`?arch=...&os=...`) is a **target selector**: which dependency variant(s) to link.

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

## Unstable roots

A root is _unstable_ if the package has dependencies that may change between builds. A common example is using the current time as a property during the build process.

If a root is unstable, add the trait `dyd/traits/unstable` (file content does not matter). If this file exists, dryad bypasses derivation cache reuse and rebuilds the root.

Keep in mind that any root with a direct or indirect dependency on an unstable root may also need to be rebuilt.
