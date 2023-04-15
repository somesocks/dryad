---
title: Design Goals
layout: default
nav_order: 2
---

# Design Goals

dryad was built with several design goals in mind:

## Hermetic builds

Hermetic builds are foundational to build performance and reproducible installs.  So, dryad packages should be stable by default. They should only depend on their own assetsand explicitly marked dependendencies.  Any unstable packages must be explicity marked as such.  All builds should happend in a (somewhat) isolated, read-only build environment.  All packages should be addressable by a content-hash-based fingerprint.

## Language agnostic

There should be _no favored language_ for dryad packages, to maximize compatibility with different tools and environments.  dryad packages have no special build language or configuation formats.  All package properties / assets should be organized in a standard filesystem structure, to maximize compatibility and make package operations simple.

## Registry agnostic

The _only_ unique identifier for a dryad package is its fingerprint.  There should be no global package names, no authoritative package registry, no "namespaces", etc.  Package authenticity should be verified by cryptographic signatures independent of package distribution.  Package "upgrades" should be decided heuristically based on package traits, instead of an authoritative versioning scheme.
