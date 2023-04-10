---
title: Roots
layout: default
parent: Concepts
---

# Roots

A **root** is the package environment for building a stem.  All roots in a garden are stored under `dyd/roots/`.  Roots have a similar filesystem structure to stems, with `dyd/assets/`, `dyd/traits`, `dyd/secrets`, and a `dyd/main`, but no path or fingerprint.

Roots may have a `dyd/stems/` directory to specify dependencies.  In addition, roots may have a `dyd/roots/` directory, where other roots in the workspace may be symlinked as dependencies.

The root build process is roughly:

1. For a root, build all root dependencies, and get the fingerprints of the resulting stems to add as dependencies.
2. Naively package a root as a stem by bundling it, adding the root dependencies, fingerprinting it, and adding it into the heap.
3. Execute the main of that root, providing it with a temporary directory in which to build a stem.
4. Fingerprint the resulting stem, and add it into the heap.

The build process may be skipped if the root has already been built and the resulting stem already added to the heap.
