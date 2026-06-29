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
- A file value: target URL such as `root:<relative-path>`, `env:<name>`, `file:<relative-path>`, or `https://example.com/data.txt#fingerprint=v2-...`.

Examples:
- `dyd/requirements/foo` with `root:../../../foo`
- `dyd/requirements/foo~arch=any+os=linux` with `root:../../../foo?arch=amd64&os=linux`
- `dyd/requirements~os=linux/foo` with `root:../../../foo`
- `dyd/requirements/display` with `env:DISPLAY`
- `dyd/requirements/local-data` with `file:../data.txt?as=dyd/assets/data.txt`
- `dyd/requirements/remote-data` with `https://example.com/data.txt#as=dyd/assets/data.txt&fingerprint=v2-...`

In practice:
- Filename descriptor (`~arch=...+os=...`) is a **condition**: when this requirement is active.
- URL query descriptor (`?arch=...&os=...`) is a **target selector**: which dependency variant(s) to link.
- Directory descriptor on `dyd/requirements~...` is a **path selector**: which requirements directory is active for the root variant.

For each variant, dryad selects at most one matching requirements directory. If there are no matching requirements directories, the root is built with no requirements. If there are multiple matching requirements directories, the build fails due to ambiguity.

### Environment requirements

Environment requirements are execution requirements. When a stem is executed, each `env:<name>` requirement reads a host environment variable and injects it into the process environment. The requirement alias is canonicalized to the injected environment variable name by uppercasing ASCII letters and converting `-` and `.` to `_`; the canonical name must match `[A-Z_][A-Z0-9_]*`. The `env:<name>` target is canonicalized the same way before reading the host environment.

For source roots, an `env:<name>` requirement is a build-time dependency because dryad executes the source stem to build the output stem. Dryad records a fingerprint of the host environment value in the materialized source stem requirement before checking the derivation cache. For built stems, an `env:<name>` requirement is a run-time dependency and the current host value is injected when the stem is run.

### File requirements

File requirements import files or directories from the local filesystem into a dependency package. Placement is controlled by query parameters:

- `as=<path>` places the imported source at an exact package path.
- `into=<path>` places the imported source under a package directory.
- `unpack=true` extracts a tar, tar.gz, or zip archive before placement.
- `optional=true` allows a missing local source and creates an empty dependency package.
- `fingerprint=<fingerprint>` pins the expected dependency package fingerprint.

Placement paths must stay under `dyd/assets`, `dyd/docs`, `dyd/traits`, or `dyd/secrets`. If neither `as` nor `into` is specified, file requirements default to `dyd/assets`.

### HTTP requirements

HTTP requirements download external assets into dependency packages. Dryad leaves the URL query string unchanged and reads its own options from the fragment:

```txt
https://example.com/archive.tar.gz?download=1#into=dyd/assets/vendor&unpack=true&fingerprint=v2-...
```

Supported HTTP metadata parameters are:

- `as=<path>` places the downloaded source at an exact package path.
- `into=<path>` places the downloaded source under a package directory.
- `unpack=true` extracts a tar, tar.gz, or zip archive. With `into=...`, the extracted archive is placed under a directory named after the archive.
- `fingerprint=<fingerprint>` pins the expected dependency package fingerprint.

If neither `as` nor `into` is specified, HTTP requirements default to `dyd/assets`.

Stored HTTP requirements must include `fingerprint=`. The fingerprint is not the raw downloaded-file checksum; it is the fingerprint of the resulting dependency package after placement and optional unpacking. `optional=` is not supported for HTTP requirements, and HTTP URLs must not include username/password credentials such as `https://user:pass@example.com/data.txt`.

Use `dryad root requirement add` to lock HTTP requirements. If the HTTP target does not include a fingerprint, the command downloads the asset and writes a requirement target containing the computed fingerprint:

```sh
dryad root requirement add 'https://example.com/data.txt?download=1' data --as=dyd/assets/data.txt
```

The written requirement will look like:

```txt
https://example.com/data.txt?download=1#as=dyd/assets/data.txt&fingerprint=v2-...
```

Builds resolve HTTP requirements cache-first by fingerprint. If the dependency package is already cached, Dryad does not contact the remote server or read auth credentials.

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
