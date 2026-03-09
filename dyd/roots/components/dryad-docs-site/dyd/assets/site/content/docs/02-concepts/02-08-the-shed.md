---
title: "02.08 - The Shed"
description: "Configuring and extending dryad commands."
type: docs
layout: single
---


# The Shed

The **shed** is the part of the garden that stores workspace configurations and settings, including all the scoped settings.  You can find the contents of the shed under `dyd/shed`.

Dryad also stores heap layout configuration in the shed under `dyd/shed/heap`.

Each fingerprinted heap namespace has a `depth` file that controls fanout for new entries, for example:

- `dyd/shed/heap/files/depth`
- `dyd/shed/heap/secrets/depth`
- `dyd/shed/heap/stems/depth`
- `dyd/shed/heap/sprouts/depth`
- `dyd/shed/heap/derivations/roots/depth`

Each file contains a single non-negative integer depth:

- `0` means a flat layout like `abcdefghijklmnopqrstuvwx23`
- `1` means one 2-character fanout segment like `ab/cdefghijklmnopqrstuvwx23`
- `2` means two 2-character fanout segments like `ab/cd/efghijklmnopqrstuvwx23`

The default depth is `1`.

Changing one of these files affects where new heap entries are written, but it does not change fingerprints or build outputs. Old entries in previous layouts remain disposable cache entries and may be removed later.
