---
title: Concepts
layout: default
nav_order: 3
---
# Concepts


## Gardens

A workspace for downloading/installing/developing packages is called a _garden_.  All packages belong to a single garden.

## The heap

Every garden has a **heap** - where package assets are stored in a content-addressed filesystem structure.

The location of the heap within a garden should be `dyd/heap`.

## Stems

A dryad package is called a **stem**.  Every stem is stored in a read-only filesystem structure inside the heap.  Stems must always be treated as read-only, except during the build process.

Every stem has a fingerprint, a unique, content-hash-based id for that stem.  The fingerprint is formatted as `<algorithm>-<hash as hexcode>`, f.ex. `blake2b-b838157d9ff40dc52b4c832c1a3e50b8`.  The fingerprint for a stem is stored in a plain-text file located at `dyd/fingerprint`.

Every stem may have any number of code/library/data files, called **assets**.  All assets should be stored in the directory `dyd/assets/`.

Every stem may have any number of extra properties to describe the package, called **traits**.  All traits should be stored in the `dyd/traits/` directory for that stem.  The `dyd/traits/` directory should be treated as a key-value text store where possible, so traits should be stored individually in plain-text files `dyd/traits/` where it makes sence.

Every stem may have a _single_ executable file call the **main entrypoint** or **main**.  This file should be located at `dyd/main`.  If multiple executables are packaged in a stem, this should be the entrypoint for invoking/passing commands to them.

Every stem may be dependent on any number of other stems.  These stems should be symlinked into the `dyd/stems/` directory.  Each stem is linked as a package dependency using a package **alias**.  A dependency with the alias `foo` should be symlinked to `dyd/stems/foo`.  The alias  for a dependency has _no relation_ to the dependency itself.  The same stem can be linked as a dependency twice under two different alias names, for example.

When a stem fingerprint is calculated, the **fingerprints** and **traits** of all dependencies are included as part of the fingerprint calculation.  This "fingerprint of fingerprints" is what creates a fully-reproducible package tree where every dependency is packaged to a specific version.

Likewise, when a stem is packaged into an isolated archive, the **fingerprints** and **traits** of all dependencies are considered part of that stem, and must also be included.

Every stem also has a collection of auto-generated **path stubs** for each dependency present in `dyd/path/`.  These stubs are included as part of the executable path when `dyd/main` is invoked, which provides a (semi) isolated execution environment for each stem.  They are auto-generated during the stem build process, and included in calculation of the stem fingerprint.

If a stem needs access to credentials, keys, or other temporary secrets, they will be mounted in the `dyd/secrets/` directory at install or run-time.  The secrets a stem _must be available_ during the stem build process, and will be included (indirectly) as part of the stem fingerprint.  If a stem should have access to secrets at run-time, a **secrets fingerprint** will be generated and stored in the stem at `dyd/secrets-fingerprint`.

## Roots

A _root_ is the package environment for building a stem.  All roots in a garden are stored under `dyd/roots/`.  Roots have a similar filesystem structure to stems, with `dyd/assets/`, `dyd/traits`, `dyd/secrets`, and a `dyd/main`, but no path or fingerprint.

Roots may have a `dyd/stems/` directory to specify dependencies.  In addition, roots may have a `dyd/roots/` directory, where other roots in the workspace may be symlinked as dependencies.

The root build process is roughly:

1. For a root, build all root dependencies, and get the fingerprints of the resulting stems to add as dependencies.
2. Naively package a root as a stem by bundling it, adding the root dependencies, fingerprinting it, and adding it into the heap.
3. Execute the main of that root, providing it with a temporary directory in which to build a stem.
4. Fingerprint the resulting stem, and add it into the heap.

The build process may be skipped if the root has already been built and the resulting stem already added to the heap.

## Sprouts

To track the built versions of each root, A garden will also have a **sprouts** directory at `dyd/sprouts/`.  The sprouts directory will automatically be created with the same filesystem structure as the roots directory during the build process.  So, a garden with two roots at `dyd/roots/tools/foo`, `dyd/roots/tests/foo-tests`, after a build will have two sprouts `dyd/sprouts/tools/foo` and `dyd/sprouts/tests/foo-tests` that link to the stems that resulted from the build.