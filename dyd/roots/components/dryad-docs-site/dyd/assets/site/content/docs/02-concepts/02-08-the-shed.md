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

## HTTP Remotes

HTTP requirements can use remote configuration from `dyd/shed/remotes/<vhost>/`. The `<vhost>` is the literal host from the requirement URL, including the port if one is present. For example, `https://artifacts.example/data.txt#fingerprint=v2-...` uses `dyd/shed/remotes/artifacts.example/`.

Remote configuration files are:

- `host` - optional network host to fetch from instead of `<vhost>`.
- `auth` - optional authentication directive used for network fetches.

If `host` is missing, Dryad uses `<vhost>` as the network host. If `auth` is missing or empty, Dryad uses `none`.

Example host mapping:

Path: `dyd/shed/remotes/artifacts/host`

Contents:

```txt
artifacts.internal.example
```

With that mapping, Dryad keeps `https://artifacts/data.txt#fingerprint=v2-...` in the requirement file but fetches from `https://artifacts.internal.example/data.txt`.

Supported `auth` forms are:

- `none`
- `bearer env:NAME`
- `bearer inline-token`
- `basic env:USER env:PASSWORD`
- `basic inline-user inline-password`

`env:NAME` credentials are read from the host environment only when Dryad needs to fetch from the network. Cache hits by fingerprint do not require auth credentials. Use `env:` for shared or version-controlled shed config; inline credentials are stored as plaintext in the shed.

Dryad allows auth over `http://`, which is useful for local test servers, but use `https://` for non-local credentials.
