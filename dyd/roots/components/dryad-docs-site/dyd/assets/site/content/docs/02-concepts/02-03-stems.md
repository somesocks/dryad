---
title: "02.03 - Stems"
description: "A built dryad package."
type: docs
layout: single
---

# Stems

A dryad package is called a **stem**. Every stem is stored in a read-only filesystem structure inside the heap, and will only ever be accessible in a read-only format, except during the build process.

Every stem has a fingerprint, a unique, content-hash-based id for that stem. The fingerprint is formatted as `<algorithm>-<hash as hexcode>`, f.ex. `blake2b-b838157d9ff40dc52b4c832c1a3e50b8`. The fingerprint for a stem is stored in a plain-text file located at `dyd/fingerprint`.

Every stem may have any number of code/library/data files, called **assets**. All assets should be stored in the directory `dyd/assets/`.

Every stem may have any number of extra properties to describe the package, called **traits**. All traits should be stored in the `dyd/traits/` directory for that stem. The `dyd/traits/` directory should be treated as a key-value text store where possible, so traits should be stored individually in plain-text files `dyd/traits/` where it makes sense.

Every stem may have a _single_ executable file call the **main entrypoint** or **main**. This file should be located at `dyd/commands/dyd-stem-run`. If multiple executables are packaged in a stem, this should be the entrypoint for invoking/passing commands to them.

Every stem may be dependent on any number of other stems. These stems should be symlinked into the `dyd/dependencies/` directory. Each stem is linked as a package dependency using a package **alias**. A dependency with the alias `foo` should be symlinked to `dyd/dependencies/foo`. The alias for a dependency has _no relation_ to the dependency itself. The same stem can be linked as a dependency twice under two different alias names, for example.

When a stem fingerprint is calculated, the **fingerprints** and **traits** of all dependencies are included as part of the fingerprint calculation. This "fingerprint of fingerprints" (a merkle tree) is what creates a fully-reproducible package tree where every dependency is pinned to a specific version.

Likewise, when a stem is packaged into an isolated archive, the **fingerprints** and **traits** of all dependencies are considered part of that stem, and must also be included.

Every stem also has a collection of generated **path stubs** for each dependency present in `dyd/path/`. These stubs are included as part of the executable path when `dyd/commands/dyd-stem-run` is invoked, which provides a (semi) isolated execution environment for the main of each stem. They are auto-generated during the stem build process, and included in calculation of the stem fingerprint.