---
title: "02 - The Heap"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: true
# bookComments: false
# bookSearchExclude: false
---

# The Heap

Every garden has a **heap** - where package assets and other build artifacts are stored in a content-addressed filesystem structure.

The location of the heap within a garden should be `dyd/heap`.

The heap is managed by dryad, however is a disposable directory. If the heap directory is deleted, all packages in the garden should still be able to be rebuilt successfully.
