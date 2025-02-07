---
title: "02.04 - Roots"
description: "Source for a dryad package."
type: docs
layout: single
---

# Roots

A **root** is the package environment for building a stem.  All roots in a garden are stored under `dyd/roots/`.  Roots have a similar filesystem structure to stems, with `dyd/assets/`, `dyd/traits`, `dyd/secrets`, and a `dyd/commands/default`, but no path or fingerprint.

Roots have a `dyd/requirements/` directory to specify dependencies.

The root build process is roughly:

1. For a root, build all root dependencies, and get the fingerprints of the resulting stems to add as dependencies.
2. Naively package a root as a stem by bundling it, adding the root dependencies, fingerprinting it, and adding it into the heap.
3. Execute the main of that root, providing it with a temporary directory in which to build a stem.
4. Fingerprint the resulting stem, and add it into the heap.

The build process may be skipped if the root has already been built and the resulting stem already added to the heap.

## Unstable roots

A root is _unstable_ if the package has dependencies that may change between builds.  A common example is using the current time as a property during the build process.

If a root is unstable, you can make it as unstable by adding the trait `dyd/traits/unstable` (the file can be empty; the contents don't matter).  If this file exists, dryad will ignore the build cache, and always rebuild the root.

Keep in mind that any root that has a direct or indirect dependency on an unstable root may also need to be rebuilt.  It's a good practice to add unstable roots at the top of the dependency tree instead of the base.  The later in the build process unstable roots are used, the more dryad can use the build cache to speed up the build process.

