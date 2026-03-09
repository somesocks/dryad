---
title: "02.02 - The Heap"
description: "Where builds go."
type: docs
layout: single
---

# The Heap

Every garden has a **heap** - where package assets and other build artifacts are stored in a content-addressed filesystem structure.

The location of the heap within a garden should be `dyd/heap`.

The heap is managed by dryad, however is a disposable directory. If the heap directory is deleted, all packages in the garden should still be able to be rebuilt successfully.

Most heap namespaces are content-addressed by fingerprint. For example:

- `dyd/heap/files`
- `dyd/heap/secrets`
- `dyd/heap/stems`
- `dyd/heap/sprouts`
- `dyd/heap/derivations/roots`

These namespaces can use directory fanout derived from the encoded fingerprint. Dryad uses fixed 2-character base32 fanout segments, so a fingerprint payload may be stored as:

- `abcdefghijklmnopqrstuvwx23`
- `ab/cdefghijklmnopqrstuvwx23`
- `ab/cd/efghijklmnopqrstuvwx23`

depending on the configured fanout depth for that namespace. A depth of `0` is flat, and larger depths add one 2-character segment per directory level.

Changing the fanout depth changes where new heap entries are written, but does not change fingerprint identity. Old entries in the previous layout are treated as stale cache entries and may be removed by later prune operations.
