---
title: "01 - The Garden"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: true
# bookComments: false
# bookSearchExclude: false
---

# The Garden

A workspace for downloading/installing/developing packages is called a **garden**. All packages belong to a single garden.

A garden has a specific directory structure:
- `dyd/type` - a sentinel file indicating a garden. This file is managed by dryad, and should contain the string `garden`.
- `dyd/heap` - a directory for build artifacts. This directory is managed by dryad.
- `dyd/roots` - a directory for package sources. This directory is unmanaged.
- `dyd/shed` - a configuration directory. This directory is managed by dryad.
- `dyd/sprouts` - an output directory for built packages. This directory is managed by dryad.
